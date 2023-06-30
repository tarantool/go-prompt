package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoRightChar(t *testing.T) {
	input := "зеленый\nred\nсиний"

	buf := NewBuffer()
	buf.InsertText(input, false, true)

	GoRightChar(buf)
	assert.Equal(t, len([]rune(input)), buf.cursorPosition)

	buf.setCursorPosition(0)
	GoRightChar(buf)
	assert.Equal(t, 1, buf.cursorPosition)

	buf.setCursorPosition(6)
	GoRightChar(buf)
	assert.Equal(t, 7, buf.cursorPosition)
}

func TestGoLeftChar(t *testing.T) {
	input := "зеленый\nred\nсиний"

	buf := NewBuffer()
	buf.InsertText(input, false, true)

	GoLeftChar(buf)
	assert.Equal(t, len([]rune(input))-1, buf.cursorPosition)

	buf.setCursorPosition(0)
	GoLeftChar(buf)
	assert.Equal(t, 0, buf.cursorPosition)

	buf.setCursorPosition(8)
	GoLeftChar(buf)
	assert.Equal(t, 7, buf.cursorPosition)
}
