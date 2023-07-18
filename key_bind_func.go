package prompt

// GoLineEnd Go to the End of the line.
func GoLineEnd(buf *Buffer) {
	x := []rune(buf.Document().TextAfterCursor())
	buf.CursorRight(len(x))
}

// GoLineBeginning Go to the beginning of the line.
func GoLineBeginning(buf *Buffer) {
	x := []rune(buf.Document().TextBeforeCursor())
	buf.CursorLeft(len(x))
}

// DeleteChar Delete character under the cursor.
func DeleteChar(buf *Buffer) {
	buf.Delete(1)
}

// DeleteWord Delete word before the cursor.
func DeleteWord(buf *Buffer) {
	buf.DeleteBeforeCursor(len([]rune(buf.Document().TextBeforeCursor())) -
		buf.Document().FindStartOfPreviousWordWithSpace())
}

// DeleteBeforeChar Go to Backspace.
func DeleteBeforeChar(buf *Buffer) {
	buf.DeleteBeforeCursor(1)
}

// GoRightChar forwards to one character to the right.
// In the case of a multi-line command the cursor moves down,
// when the end of the line is reached.
func GoRightChar(buf *Buffer) {
	if buf.Document().GetCursorRightPosition(1) == 0 {
		if !buf.Document().OnLastLine() {
			buf.CursorDown(1)
			GoLineBeginning(buf)
		}
		return
	}
	buf.CursorRight(1)
}

// GoLeftChar backward to one character to the left.
// In the case of a multi-line command the cursor moves up,
// when the begin of the line is reached.
func GoLeftChar(buf *Buffer) {
	if buf.Document().GetCursorLeftPosition(1) == 0 {
		if buf.Document().CursorPositionRow() != 0 {
			buf.CursorUp(1)
			GoLineEnd(buf)
		}
		return
	}
	buf.CursorLeft(1)
}

// GoRightWord moves the cursor to the end of the next word.
func GoRightWord(buf *Buffer) {
	buf.setCursorPosition(buf.Document().FindWordEndForwardCursor(func(r rune) bool {
		return r != '\n' && r != ' '
	}))
}

// GoLeftWord moves the cursor to the beginning of the previous word.
func GoLeftWord(buf *Buffer) {
	buf.setCursorPosition(buf.Document().FindWordStartBackwardCursor(func(r rune) bool {
		return r != '\n' && r != ' '
	}))
}

// GoCmdBeginning moves the cursor to the beginning of the command.
func GoCmdBeginning(buf *Buffer) {
	buf.setCursorPosition(0)
}

// GoCmdEnd moves the cursor to the end of the command.
func GoCmdEnd(buf *Buffer) {
	buf.setCursorPosition(len([]rune(buf.Text())))
}
