package prompt

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/tarantool/go-prompt/internal/debug"
)

// Executor is called when user input something text.
type Executor func(string)

// ExitChecker is called after user input to check if prompt must stop and exit go-prompt Run loop.
// User input means: selecting/typing an entry, then, if said entry content matches the ExitChecker
// function criteria:
// - immediate exit (if breakline is false) without executor called
// - exit after typing <return> (meaning breakline is true), and the executor is called first,
// before exit.
// Exit means exit go-prompt (not the overall Go program).
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
func (p *Prompt) fillCtx(renderEvent int) renderCtx {
	cmd, _ := p.getCmdToRender()
	prefix := p.getCurrentPrefix()

	ctx := renderCtx{
		cmd:         cmd,
		cursor:      p.cursor,
		endCursor:   p.endCursor,
		completion:  p.completion,
		prefixColor: p.renderer.prefixTextColor,
		prefix:      prefix,
		renderCompletion: !p.inReverseSearchMode() &&
			!(p.buf.NewLineCount() > 0),
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

	// notifyConn is a connection used for rendering notifications.
	notifyConn net.Conn

	// isBracketedPasteEnabled is true if the terminal is asked to mark
	// pastes.
	isBracketedPasteEnabled bool

	// lexer cuts what is read from the terminal into keys.
	lexer inputLexer

	// pending holds the keys read but not fed yet: those after an Enter
	// wait until the command before them is executed.
	pending []inputToken
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

	// feedPending feeds the keys read, executing a command at every Enter,
	// and reports whether the prompt must stop.
	feedPending := func() bool {
		updated := true
		for len(p.pending) > 0 {
			shouldExit, e := p.feedNext(updated)
			updated = false

			if shouldExit {
				// Run onUpdate hook.
				p.onInputUpdate()
				p.render(breakLineRenderEvent)
				p.notifyRender()
				stopReadBufCh <- struct{}{}
				stopHandleSignalCh <- struct{}{}
				return true
			} else if e == nil {
				continue
			}

			// Run onUpdate hook.
			p.onInputUpdate()
			updated = true

			// Stop goroutine to run readBuffer function
			stopReadBufCh <- struct{}{}
			stopHandleSignalCh <- struct{}{}

			// Unset raw mode
			// Reset to Blocking mode because returned EAGAIN when still set non-blocking mode.
			p.setBracketedPaste(false)
			debug.AssertNoError(p.in.TearDown())

			p.executor(e.input)
			p.render(basicRenderEvent)
			p.notifyRender()

			if p.exitChecker != nil && p.exitChecker(e.input, true) {
				p.skipTearDown = true
				return true
			}
			// Set raw mode
			debug.AssertNoError(p.in.Setup())
			p.setBracketedPaste(true)
			go p.readBuffer(bufCh, stopReadBufCh)
			go p.handleSignals(exitCh, winSizeCh, stopHandleSignalCh)
		}

		if !updated {
			// Run onUpdate hook.
			p.onInputUpdate()
			p.render(basicRenderEvent)
			p.notifyRender()
		}
		return false
	}

	// The keys an Input call left.
	if feedPending() {
		return
	}

	for {
		select {
		case b := <-bufCh:
			p.pending = append(p.pending, p.lexer.split(b, p.ASCIICodeBindings)...)
			if feedPending() {
				return
			}
		case w := <-winSizeCh:
			p.onInputUpdate()
			p.renderer.UpdateWinSize(w)
			p.render(windowResizeRenderEvent)
			p.notifyRender()
		case code := <-exitCh:
			p.onInputUpdate()
			p.render(breakLineRenderEvent)
			p.notifyRender()
			p.tearDown()
			os.Exit(code)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// feedNext feeds the first of the pending keys. The completion and the
// history follow the input once a read is fed (onInputUpdate), not after
// every key of it, the completer being possibly slow; so a key other than
// text and Enter that follows other keys gets them updated first, unless
// updated tells that they are.
func (p *Prompt) feedNext(updated bool) (shouldExit bool, exec *Exec) {
	token := p.pending[0]
	p.pending = p.pending[1:]

	if !updated && token.kind == keyToken {
		switch GetKey(token.bytes) {
		case Enter, ControlJ, ControlM:
		default:
			p.onInputUpdate()
		}
	}
	return p.feedToken(token)
}

// feed feeds b as one key.
func (p *Prompt) feed(b []byte) (shouldExit bool, exec *Exec) {
	return p.feedToken(inputToken{kind: keyToken, bytes: b})
}

// feedToken feeds a key, or a run of text, to the prompt.
func (p *Prompt) feedToken(token inputToken) (shouldExit bool, exec *Exec) {
	b := token.bytes
	key := NotDefined
	if token.kind == keyToken {
		key = GetKey(b)
	}

	p.buf.lastKeyStroke = key
	// completion
	completing := p.completion.Completing()
	p.handleCompletionKeyBinding(key, completing)

	switch key {
	case Enter, ControlJ, ControlM:
		// Tabs that came with the Enter, in a paste, are replaced as those
		// typed before it are (onInputUpdate).
		p.buf = p.buf.ReplaceTabs(defaultTabWidth)
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
		} else if !completing { // Don't use p.completion.Completing() because it takes double
			// operation when switch to selected=-1.
			if newBuf, changed := p.history.Older(p.buf); changed {
				p.buf = newBuf
			}
		}
	case Down, ControlN:
		if p.inReverseSearchMode() {
			p.disableReverseSearch()
		} else if !completing { // Don't use p.completion.Completing() because it takes double
			// operation when switch to selected=-1.
			if newBuf, changed := p.history.Newer(p.buf); changed {
				p.buf = newBuf
			}
		}
	case Left, Right, ControlB, ControlF:
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
		if token.kind == keyToken && p.handleASCIICodeBinding(b) {
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

	// feedPending feeds the keys read up to an Enter, which the keys after
	// it wait for the next call, and reports whether the input is done.
	feedPending := func() (string, bool) {
		updated := true
		for len(p.pending) > 0 {
			shouldExit, e := p.feedNext(updated)
			updated = false

			if shouldExit {
				p.render(breakLineRenderEvent)
				stopReadBufCh <- struct{}{}
				return "", true
			} else if e != nil {
				// Stop goroutine to run readBuffer function
				stopReadBufCh <- struct{}{}
				return e.input, true
			}
		}

		if !updated {
			p.onInputUpdate()
			p.render(basicRenderEvent)
		}
		return "", false
	}

	// The keys the last call left.
	if input, done := feedPending(); done {
		return input
	}

	for {
		select {
		case b := <-bufCh:
			p.pending = append(p.pending, p.lexer.split(b, p.ASCIICodeBindings)...)
			if input, done := feedPending(); done {
				return input
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
	p.setBracketedPaste(true)
}

// tearDown restores the terminal state when prompt is turned off.
func (p *Prompt) tearDown() {
	if !p.skipTearDown {
		p.setBracketedPaste(false)
		debug.AssertNoError(p.in.TearDown())
	}
	p.renderer.TearDown()
}

// setBracketedPaste turns the bracketed paste mode of the terminal on or
// off, if the mode is enabled (OptionBracketedPaste). It is on while the
// prompt reads, and off while a command runs, so that neither the command
// nor what runs after the prompt gets pastes marked.
func (p *Prompt) setBracketedPaste(on bool) {
	if !p.isBracketedPasteEnabled {
		return
	}
	if on {
		p.renderer.out.WriteRawStr("\x1b[?2004h")
	} else {
		p.renderer.out.WriteRawStr("\x1b[?2004l")
	}
	debug.AssertNoError(p.renderer.out.Flush())
}

// getCmdToRender builds command to render.
// Returns (buffer with command to render, prefix length in bytes).
func (p *Prompt) getCmdToRender() (cmd *Buffer, prefixLen int) {
	input := p.buf.Text()
	prefix := p.getCurrentPrefix()
	cmdBuf := NewBuffer()
	cmdBuf.InsertText(prefix, false, true)
	if p.inReverseSearchMode() {
		cmdBuf.InsertText(p.reverseSearch.matchedCmd, false, true)
	} else {
		cmdBuf.InsertText(input, false, true)
		cmdBuf.setCursorPosition(len(prefix) + p.buf.cursorPosition)
	}
	return cmdBuf, len(prefix)
}

// getCurrentPrefix returns current prefix.
// If reverse search is enabled, its prefix extracted.
// If live-prefix is enabled, return live-prefix.
func (p *Prompt) getCurrentPrefix() string {
	if p.inReverseSearchMode() {
		rsPrefixFmt := matchSearchPrefixFmt
		if p.reverseSearch.matchedIndex == -1 {
			rsPrefixFmt = failSearchPrefixFmt
		}
		return fmt.Sprintf(rsPrefixFmt, p.buf.Text())
	}
	if prefix, ok := p.livePrefixCallback(); ok {
		return prefix
	}
	return p.prefix
}

// onInputUpdate does necessary actions at the input update moment.
func (p *Prompt) onInputUpdate() {
	p.buf = p.buf.ReplaceTabs(defaultTabWidth)
	if p.inReverseSearchMode() {
		p.reverseSearch.update(p.buf.Text())
		return
	}
	p.history.SetCurrentCmd(p.buf.Text())
	p.completion.Update(*p.buf.Document())
}

// render renders current prompt state to the attached renderer,
// updates current cursor position.
func (p *Prompt) render(event int) {
	ctx := p.fillCtx(event)
	p.cursor, p.endCursor = p.renderer.Render(ctx)
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

// notifyAboutRender notifying about rendering if such option was provided.
func (p *Prompt) notifyRender() {
	if p.notifyConn != nil {
		p.notifyConn.Write([]byte{67})
	}
}
