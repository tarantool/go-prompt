package prompt

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/c-bata/go-prompt/internal/debug"
)

// Executor is called when user input something text.
type Executor func(string)

// ExitChecker is called after user input to check if prompt must stop and exit go-prompt Run loop.
// User input means: selecting/typing an entry, then, if said entry content matches the ExitChecker function criteria:
// - immediate exit (if breakline is false) without executor called
// - exit after typing <return> (meaning breakline is true), and the executor is called first, before exit.
// Exit means exit go-prompt (not the overall Go program)
type ExitChecker func(in string, breakline bool) bool

// Completer should return the suggest item from Document.
type Completer func(Document) []Suggest

// location indicates the relative location of the cursor on the screen.
type location struct {
	row int
	col int
}

// renderCtx describes render context.
type renderCtx struct {
	cmd              *Buffer
	cursor           location
	endCursor        location
	completion       *CompletionManager
	prefixColor      Color
	prefix           string
	renderCompletion bool
	renderEvent      int
}

// fillCtx fills render context.
func (prompt *Prompt) fillCtx(renderEvent int) renderCtx {
	cmd, _ := prompt.getCmdToRender()
	prefix := prompt.getCurrentPrefix()

	ctx := renderCtx{
		cmd:         cmd,
		cursor:      prompt.cursor,
		endCursor:   prompt.endCursor,
		completion:  prompt.completion,
		prefixColor: prompt.renderer.prefixTextColor,
		prefix:      prefix,
		renderCompletion: !prompt.inReverseSearchMode() &&
			!(prompt.buf.NewLineCount() > 0),
		renderEvent: renderEvent,
	}
	return ctx
}

// Prompt is core struct of go-prompt.
type Prompt struct {
	in                ConsoleParser
	buf               *Buffer
	cursor            location
	endCursor         location
	renderer          *Render
	executor          Executor
	history           *History
	completion        *CompletionManager
	keyBindings       []KeyBind
	ASCIICodeBindings []ASCIICodeBind
	keyBindMode       KeyBindMode
	completionOnDown  bool
	exitChecker       ExitChecker
	skipTearDown      bool

	prefix             string
	livePrefixCallback func() (prefix string, useLivePrefix bool)
	title              string

	// reverseSearch is a pointer to the current reverse-search state.
	// not nil pointer means that reverse-search mode is active.
	reverseSearch *reverseSearchState

	// isReverseSearchEnabled is true if such option was provided.
	isReverseSearchEnabled bool

	// isAutoHistoryEnabled is true if automatic writing to the history is enabled.
	isAutoHistoryEnabled bool
}

// Exec is the struct contains user input context.
type Exec struct {
	input string
}

// ClearScreen clears the screen.
func (p *Prompt) ClearScreen() {
	p.renderer.ClearScreen()
}

// Run starts prompt.
func (p *Prompt) Run() {
	p.skipTearDown = false
	defer debug.Teardown()
	debug.Log("start prompt")
	p.setUp()
	defer p.tearDown()

	if p.completion.showAtStart {
		p.completion.Update(*p.buf.Document())
	}

	p.render(basicRenderEvent)

	bufCh := make(chan []byte, 128)
	stopReadBufCh := make(chan struct{})
	go p.readBuffer(bufCh, stopReadBufCh)

	exitCh := make(chan int)
	winSizeCh := make(chan *WinSize)
	stopHandleSignalCh := make(chan struct{})
	go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)

	for {
		select {
		case b := <-bufCh:
			shouldExit, e := p.feed(b)

			// Run onUpdate hook.
			p.onInputUpdate()

			if shouldExit {
				p.render(breakLineRenderEvent)
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}
				return
			} else if e != nil {
				// Stop goroutine to run readBuffer function
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}

				// Unset raw mode
				// Reset to Blocking mode because returned EAGAIN when still set non-blocking mode.
				debug.AssertNoError(p.in.TearDown())

				p.executor(e.input)
				p.render(basicRenderEvent)

				if p.exitChecker != nil && p.exitChecker(e.input, true) {
					p.skipTearDown = true
					return
				}
				// Set raw mode
				debug.AssertNoError(p.in.Setup())
				go p.readBuffer(bufCh, stopReadBufCh)
				go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)
			} else {
				p.render(basicRenderEvent)
			}
		case w := <-winSizeCh:
			p.onInputUpdate()
			p.renderer.UpdateWinSize(w)
			p.render(windowResizeRenderEvent)
		case code := <-exitCh:
			p.onInputUpdate()
			p.render(breakLineRenderEvent)
			p.tearDown()
			os.Exit(code)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (p *Prompt) feed(b []byte) (shouldExit bool, exec *Exec) {
	key := GetKey(b)

	p.buf.lastKeyStroke = key
	// completion
	completing := p.completion.Completing()
	p.handleCompletionKeyBinding(key, completing)

	switch key {
	case Enter, ControlJ, ControlM:
		execCmd := p.buf.Text()
		if p.inReverseSearchMode() {
			// Execute last matched command in case of enabled reverse search.
			execCmd = p.reverseSearch.matchedCmd
			p.disableReverseSearch()

			// Render executed command before breakline.
			p.render(basicRenderEvent)
		}
		p.render(breakLineRenderEvent)
		p.buf = NewBuffer()
		exec = &Exec{input: execCmd}
		if exec.input != "" && p.isAutoHistoryEnabled {
			p.history.Add(exec.input)
		}
	case ControlC:
		if p.inReverseSearchMode() {
			p.disableReverseSearch()
		}
		p.render(breakLineRenderEvent)
		p.buf = NewBuffer()
		p.history.Clear()
	case Up, ControlP:
		if p.inReverseSearchMode() {
			p.disableReverseSearch()
		} else if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if newBuf, changed := p.history.Older(p.buf); changed {
				p.buf = newBuf
			}
		}
	case Down, ControlN:
		if p.inReverseSearchMode() {
			p.disableReverseSearch()
		} else if !completing { // Don't use p.completion.Completing() because it takes double operation when switch to selected=-1.
			if newBuf, changed := p.history.Newer(p.buf); changed {
				p.buf = newBuf
			}
		}
	case Left, Right:
		if p.inReverseSearchMode() {
			p.disableReverseSearch()
		}
	case ControlD:
		if p.buf.Text() == "" {
			shouldExit = true
			return
		}
	case ControlR:
		if p.inReverseSearchMode() {
			p.reverseSearch.reducePrefix()
		} else {
			p.enableReverseSearch()
		}
	case NotDefined:
		if p.handleASCIICodeBinding(b) {
			return
		}
		p.buf.InsertText(string(b), false, true)
	}

	shouldExit = p.handleKeyBinding(key)
	return
}

func (p *Prompt) handleCompletionKeyBinding(key Key, completing bool) {
	switch key {
	case Down:
		if completing || p.completionOnDown {
			p.completion.Next()
		}
	case Tab, ControlI:
		p.completion.Next()
	case Up:
		if completing {
			p.completion.Previous()
		}
	case BackTab:
		p.completion.Previous()
	default:
		if s, ok := p.completion.GetSelectedSuggestion(); ok {
			w := p.buf.Document().GetWordBeforeCursorUntilSeparator(p.completion.wordSeparator)
			if w != "" {
				p.buf.DeleteBeforeCursor(len([]rune(w)))
			}
			p.buf.InsertText(s.Text, false, true)
		}
		p.completion.Reset()
	}
}

func (p *Prompt) handleKeyBinding(key Key) bool {
	shouldExit := false
	for i := range commonKeyBindings {
		kb := commonKeyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}

	if p.keyBindMode == EmacsKeyBind {
		for i := range emacsKeyBindings {
			kb := emacsKeyBindings[i]
			if kb.Key == key {
				kb.Fn(p.buf)
			}
		}
	}

	// Custom key bindings
	for i := range p.keyBindings {
		kb := p.keyBindings[i]
		if kb.Key == key {
			kb.Fn(p.buf)
		}
	}
	if p.exitChecker != nil && p.exitChecker(p.buf.Text(), false) {
		shouldExit = true
	}
	return shouldExit
}

func (p *Prompt) handleASCIICodeBinding(b []byte) bool {
	checked := false
	for _, kb := range p.ASCIICodeBindings {
		if bytes.Equal(kb.ASCIICode, b) {
			kb.Fn(p.buf)
			checked = true
		}
	}
	return checked
}

// Input just returns user input text.
func (p *Prompt) Input() string {
	defer debug.Teardown()
	debug.Log("start prompt")
	p.setUp()
	defer p.tearDown()

	if p.completion.showAtStart {
		p.completion.Update(*p.buf.Document())
	}

	p.render(basicRenderEvent)
	bufCh := make(chan []byte, 128)
	stopReadBufCh := make(chan struct{})
	go p.readBuffer(bufCh, stopReadBufCh)

	for {
		select {
		case b := <-bufCh:
			if shouldExit, e := p.feed(b); shouldExit {
				p.render(breakLineRenderEvent)
				stopReadBufCh <- struct{}{}
				return ""
			} else if e != nil {
				// Stop goroutine to run readBuffer function
				stopReadBufCh <- struct{}{}
				return e.input
			} else {
				p.onInputUpdate()
				p.render(basicRenderEvent)
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (p *Prompt) readBuffer(bufCh chan []byte, stopCh chan struct{}) {
	debug.Log("start reading buffer")
	for {
		select {
		case <-stopCh:
			debug.Log("stop reading buffer")
			return
		default:
			if b, err := p.in.Read(); err == nil && !(len(b) == 1 && b[0] == 0) {
				bufCh <- b
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (p *Prompt) setUp() {
	debug.AssertNoError(p.in.Setup())
	p.renderer.Setup(p.title)
	p.renderer.UpdateWinSize(p.in.GetWinSize())
}

func (p *Prompt) tearDown() {
	if !p.skipTearDown {
		debug.AssertNoError(p.in.TearDown())
	}
	p.renderer.TearDown()
}

// getCmdToRender builds command to render.
// Returns (buffer with command to render, prefix length in bytes).
func (prompt *Prompt) getCmdToRender() (cmd *Buffer, prefixLen int) {
	input := prompt.buf.Text()
	prefix := prompt.getCurrentPrefix()
	cmdBuf := NewBuffer()
	cmdBuf.InsertText(prefix, false, true)
	if prompt.inReverseSearchMode() {
		cmdBuf.InsertText(prompt.reverseSearch.matchedCmd, false, true)
	} else {
		cmdBuf.InsertText(input, false, true)
		cmdBuf.setCursorPosition(len(prefix) + prompt.buf.cursorPosition)
	}
	return cmdBuf, len(prefix)
}

// getCurrentPrefix returns current prefix.
// If reverse search is enabled, its prefix extracted.
// If live-prefix is enabled, return live-prefix.
func (prompt *Prompt) getCurrentPrefix() string {
	if prompt.inReverseSearchMode() {
		rsPrefixFmt := matchSearchPrefixFmt
		if prompt.reverseSearch.matchedIndex == -1 {
			rsPrefixFmt = failSearchPrefixFmt
		}
		return fmt.Sprintf(rsPrefixFmt, prompt.buf.Text())
	}
	if prefix, ok := prompt.livePrefixCallback(); ok {
		return prefix
	}
	return prompt.prefix
}

// onInputUpdate does necessary actions at the input update moment.
func (prompt *Prompt) onInputUpdate() {
	prompt.buf = prompt.buf.ReplaceTabs(defaultTabWidth)
	if prompt.inReverseSearchMode() {
		prompt.reverseSearch.update(prompt.buf.Text())
		return
	}
	prompt.history.SetCurrentCmd(prompt.buf.Text())
	prompt.completion.Update(*prompt.buf.Document())
}

// render renders current prompt state to the attached renderer,
// updates current cursor position.
func (prompt *Prompt) render(event int) {
	ctx := prompt.fillCtx(event)
	prompt.cursor, prompt.endCursor = prompt.renderer.Render(ctx)
}

// inReverseSearchMode returns true if the prompt is in reverse-search mode.
func (p *Prompt) inReverseSearchMode() bool {
	return p.reverseSearch != nil
}

// enableReverseSearch enables reverse-search mode.
func (p *Prompt) enableReverseSearch() {
	if !p.isReverseSearchEnabled || p.inReverseSearchMode() {
		return
	}

	p.buf = NewBuffer()
	p.reverseSearch = NewReverseSearch(p.history)
}

// disableReverseSearch disables reverse-search mode,
// sets history pointer to the last matched command.
func (p *Prompt) disableReverseSearch() {
	if !p.isReverseSearchEnabled || !p.inReverseSearchMode() {
		return
	}

	matchedIndex := p.reverseSearch.matchedIndex
	matchedCmd := p.reverseSearch.matchedCmd
	p.history.Clear()
	p.buf = NewBuffer()

	if matchedIndex != -1 {
		p.history.selected = matchedIndex
		p.history.SetCurrentCmd(matchedCmd)
		p.buf.InsertText(matchedCmd, false, true)
	}

	p.reverseSearch = nil
}

// pushToHistory takes command, replaces tabs with spaces,
// pushes it to the history.
func (p *Prompt) pushToHistory(cmd string) {
	cmdBuf := NewBuffer()
	cmdBuf.InsertText(cmd, false, true)
	cmdBuf = cmdBuf.ReplaceTabs(defaultTabWidth)
	p.history.Add(cmdBuf.Text())
}

// PushToHistory pushes to the history, if auto history is disabled.
func (p *Prompt) PushToHistory(cmd string) error {
	if p.isAutoHistoryEnabled {
		return fmt.Errorf("external pushes to the history are forbidden, " +
			"use `OptionDisableAutoHistory`")
	}
	p.pushToHistory(cmd)
	return nil
}
