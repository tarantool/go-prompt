package prompt

import (
	"runtime"

	"github.com/c-bata/go-prompt/internal/debug"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	// basicRenderEvent renders context and completion.
	basicRenderEvent = iota
	// breakLineRenderEvent renders context with break-line.
	breakLineRenderEvent
	windowResizeRenderEvent
)

// Render to render prompt information from state of Buffer.
type Render struct {
	out ConsoleWriter

	breakLineCallback func(*Document)
	row               uint16
	col               uint16

	// Colors.
	prefixTextColor              Color
	prefixBGColor                Color
	inputTextColor               Color
	inputBGColor                 Color
	previewSuggestionTextColor   Color
	previewSuggestionBGColor     Color
	suggestionTextColor          Color
	suggestionBGColor            Color
	selectedSuggestionTextColor  Color
	selectedSuggestionBGColor    Color
	descriptionTextColor         Color
	descriptionBGColor           Color
	selectedDescriptionTextColor Color
	selectedDescriptionBGColor   Color
	scrollbarThumbColor          Color
	scrollbarBGColor             Color
}

// Setup to initialize console output.
func (r *Render) Setup(title string) {
	if title != "" {
		r.out.SetTitle(title)
		debug.AssertNoError(r.out.Flush())
	}
}

// TearDown to clear title and erasing.
func (r *Render) TearDown() {
	r.out.ClearTitle()
	r.out.EraseDown()
	debug.AssertNoError(r.out.Flush())
}

func (r *Render) prepareArea(lines int) {
	for i := 0; i < lines; i++ {
		r.out.ScrollDown()
	}
	for i := 0; i < lines; i++ {
		r.out.ScrollUp()
	}
}

// UpdateWinSize called when window size is changed.
func (r *Render) UpdateWinSize(ws *WinSize) {
	r.row = ws.Row
	r.col = ws.Col
}

func (r *Render) renderWindowTooSmall() {
	r.out.CursorGoTo(0, 0)
	r.out.EraseScreen()
	r.out.SetColor(DarkRed, White, false)
	r.out.WriteStr("Your console window is too small...")
}

// renderCompletion renders completion.
func (r *Render) renderCompletion(ctx renderCtx) {
	suggestions := ctx.completion.GetSuggestions()
	if len(suggestions) == 0 {
		return
	}
	prefix := ctx.prefix
	formatted, width := formatSuggestions(
		suggestions,
		int(r.col)-runewidth.StringWidth(prefix)-1, // -1 means a width of scrollbar
	)
	// +1 means a width of scrollbar.
	width++

	windowHeight := len(formatted)
	if windowHeight > int(ctx.completion.max) {
		windowHeight = int(ctx.completion.max)
	}
	formatted = formatted[ctx.completion.verticalScroll : ctx.completion.verticalScroll+
		windowHeight]
	r.prepareArea(windowHeight)

	cursor := runewidth.StringWidth(ctx.cmd.Document().TextBeforeCursor())
	x, _ := r.toPos(cursor)
	if x+width >= int(r.col) {
		cursor = r.backward(cursor, x+width-int(r.col))
	}

	contentHeight := len(ctx.completion.tmp)

	fractionVisible := float64(windowHeight) / float64(contentHeight)
	fractionAbove := float64(ctx.completion.verticalScroll) / float64(contentHeight)

	scrollbarHeight := int(clamp(float64(windowHeight), 1, float64(windowHeight)*fractionVisible))
	scrollbarTop := int(float64(windowHeight) * fractionAbove)

	isScrollThumb := func(row int) bool {
		return scrollbarTop <= row && row <= scrollbarTop+scrollbarHeight
	}

	selected := ctx.completion.selected - ctx.completion.verticalScroll
	r.out.SetColor(White, Cyan, false)
	for i := 0; i < windowHeight; i++ {
		r.out.CursorDown(1)
		if i == selected {
			r.out.SetColor(r.selectedSuggestionTextColor, r.selectedSuggestionBGColor, true)
		} else {
			r.out.SetColor(r.suggestionTextColor, r.suggestionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Text)

		if i == selected {
			r.out.SetColor(r.selectedDescriptionTextColor, r.selectedDescriptionBGColor, false)
		} else {
			r.out.SetColor(r.descriptionTextColor, r.descriptionBGColor, false)
		}
		r.out.WriteStr(formatted[i].Description)

		if isScrollThumb(i) {
			r.out.SetColor(DefaultColor, r.scrollbarThumbColor, false)
		} else {
			r.out.SetColor(DefaultColor, r.scrollbarBGColor, false)
		}
		r.out.WriteStr(" ")
		r.out.SetColor(DefaultColor, DefaultColor, false)

		r.lineWrap(cursor + width)
		r.backward(cursor+width, width)
	}

	if x+width >= int(r.col) {
		r.out.CursorForward(x + width - int(r.col))
	}

	r.out.CursorUp(windowHeight)
	r.out.SetColor(DefaultColor, DefaultColor, false)
}

// ClearScreen clears the screen and moves the cursor to home.
func (r *Render) ClearScreen() {
	r.out.EraseScreen()
	r.out.CursorGoTo(0, 0)
}

// writeCmdWithPrefix writes cmd with prefix to the out.
func writeCmdWithPrefix(
	out ConsoleWriter,
	cmd string,
	prefixLen int,
	prefixColor Color,
	bgColor Color,
	defaultColor Color,
) {
	out.SetColor(prefixColor, bgColor, false)
	out.WriteStr(cmd[:prefixLen])
	out.SetColor(DefaultColor, DefaultColor, false)
	out.WriteStr(cmd[prefixLen:])
}

// preprocessCtx preprocesses the context before rendering.
func (r *Render) preprocessCtx(ctx renderCtx) renderCtx {
	ctx.cmd = ctx.cmd.SplitWideLines(int(r.col))
	return ctx
}

// renderCtx renders context to the out, returns
// (new location of cursor, new location of the end of the rendered command).
func (r *Render) renderCtx(ctx renderCtx) (newCursor location, newEndCursor location) {
	// Preprocess the context.
	ctx = r.preprocessCtx(ctx)

	// Calculate current cursor position.
	cursorRow, cursorCol := ctx.cmd.Document().GetCursorPosition()
	// Calculate cursor position of the line end.
	endRow, endCol := ctx.cmd.Document().GetCustomCursorPosition(
		len([]rune(ctx.cmd.Text())),
	)

	ctxCursorPos := ctx.cursor.row*int(r.col) + ctx.cursor.col
	ctxEndCursorPos := ctx.endCursor.row*int(r.col) + ctx.endCursor.col

	if ctx.renderEvent == windowResizeRenderEvent {
		r.clear(ctxCursorPos, true)
	} else {
		// Move to the end of rendered recently.
		r.move(ctxCursorPos, ctxEndCursorPos)
		// Erase rendered recently.
		r.clear(ctxEndCursorPos, false)
	}

	// Render.
	writeCmdWithPrefix(r.out, ctx.cmd.Text(), len(ctx.prefix),
		ctx.prefixColor, r.prefixBGColor, DefaultColor)
	r.lineWrap(endCol)

	// Move cursor back to the position inside cmd.
	r.move(endRow*int(r.col)+endCol, cursorRow*int(r.col)+cursorCol)

	return location{cursorRow, cursorCol}, location{endRow, endCol}
}

// Render is main render function.
// It calls suitable sub-render function (as `renderBreakLine`) in dependence of render event.
// Returns new cursor position.
func (r *Render) Render(ctx renderCtx) (location, location) {
	// In situations where a pseudo tty is allocated (e.g. within a docker container),
	// window size via TIOCGWINSZ is not immediately available and will result in 0,0 dimensions.
	if ctx.renderEvent == breakLineRenderEvent {
		return r.renderBreakLine(ctx)
	}
	if r.col == 0 {
		return location{}, location{}
	}
	defer func() { debug.AssertNoError(r.out.Flush()) }()

	// Hide cursor to prevent blinking.
	r.out.HideCursor()
	defer func() {
		r.out.ShowCursor()
	}()

	// Render current state.
	curCursor, endCursor := r.renderCtx(ctx)

	if ctx.renderCompletion {
		r.renderCompletion(ctx)
		if suggest, ok := ctx.completion.GetSelectedSuggestion(); ok {
			curCursor.col = r.backward(curCursor.col, runewidth.StringWidth(
				ctx.cmd.Document().GetWordBeforeCursorUntilSeparator(
					ctx.completion.wordSeparator,
				),
			))

			r.out.SetColor(r.previewSuggestionTextColor, r.previewSuggestionBGColor, false)
			r.out.WriteStr(suggest.Text)
			r.out.SetColor(DefaultColor, DefaultColor, false)
			curCursor.col += runewidth.StringWidth(suggest.Text)

			rest := ctx.cmd.Document().TextAfterCursor()
			r.out.WriteStr(rest)
			curCursor.col += runewidth.StringWidth(rest)
			r.lineWrap(curCursor.col)

			curCursor.col = r.backward(curCursor.col, runewidth.StringWidth(rest))
		}
	}

	return curCursor, endCursor
}

// renderBreakline renders state with linebreak and calls breakline callback.
func (r *Render) renderBreakLine(ctx renderCtx) (location, location) {
	defer func() { debug.AssertNoError(r.out.Flush()) }()

	// Hide cursor to prevent blinking.
	r.out.HideCursor()
	defer func() {
		r.out.ShowCursor()
	}()

	cmdBuf := NewBuffer()
	cmdBuf.InsertText(ctx.cmd.Text()+"\n", false, true)

	cmdDocument := ctx.cmd.Document()
	ctx.cmd = cmdBuf

	// Render state.
	r.renderCtx(ctx)

	if r.breakLineCallback != nil {
		r.breakLineCallback(cmdDocument)
	}

	return location{}, location{}
}

// clear erases the screen from a beginning of input
//
// function supports two work modes.
// `hard=true` when `EraseDown` is used for clearing the rendered text.
//
// `hard=false`, when softly clear line by line is used.
//
// `EraseDown` may cause flickering in some terminals (e.g. `urxvt`).
func (r *Render) clear(cursor int, hard bool) {
	r.out.EraseDown()
	if !hard {
		fromX, fromY := r.toPos(cursor)
		r.out.CursorBackward(fromX)
		for i := 0; i < fromY; i++ {
			r.out.EraseEndOfLine()
			r.out.CursorUp(1)
		}
		r.out.EraseEndOfLine()
	} else {
		r.move(cursor, 0)
		r.out.EraseDown()
	}
}

// backward moves cursor to backward from a current cursor position
// regardless there is a line break.
func (r *Render) backward(from, n int) int {
	return r.move(from, from-n)
}

// move moves cursor to specified position from the beginning of input
// even if there is a line break.
func (r *Render) move(from, to int) int {
	fromX, fromY := r.toPos(from)
	toX, toY := r.toPos(to)

	r.out.CursorUp(fromY - toY)
	r.out.CursorBackward(fromX - toX)
	return to
}

// toPos returns the relative position from the beginning of the string.
func (r *Render) toPos(cursor int) (x, y int) {
	col := int(r.col)
	return cursor % col, cursor / col
}

func (r *Render) lineWrap(cursor int) {
	if runtime.GOOS != "windows" && cursor > 0 && cursor%int(r.col) == 0 {
		r.out.WriteRaw([]byte{'\n'})
	}
}

func clamp(high, low, x float64) float64 {
	switch {
	case high < x:
		return high
	case x < low:
		return low
	default:
		return x
	}
}
