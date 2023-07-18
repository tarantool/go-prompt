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

func TestGoCmdBeginning(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		input := "зеленый\nред\nсиний"
		buf := NewBuffer()
		buf.InsertText(input, false, true)

		GoCmdBeginning(buf)
		assert.Equal(t, 0, buf.cursorPosition)

		buf.cursorPosition = 2
		GoCmdBeginning(buf)
		assert.Equal(t, 0, buf.cursorPosition)
	})

	t.Run("empty buffer", func(t *testing.T) {
		buf := NewBuffer()
		GoCmdBeginning(buf)
		assert.Equal(t, 0, buf.cursorPosition)
	})
}

func TestGoCmdEnd(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		input := "зеленый\nред\nсиний"
		buf := NewBuffer()
		buf.InsertText(input, false, true)

		GoCmdEnd(buf)
		assert.Equal(t, 17, buf.cursorPosition)

		buf.cursorPosition = 0
		GoCmdEnd(buf)
		assert.Equal(t, 17, buf.cursorPosition)
	})

	t.Run("empty buffer", func(t *testing.T) {
		buf := NewBuffer()
		GoCmdBeginning(buf)
		assert.Equal(t, 0, buf.cursorPosition)
	})
}

func TestGoLeftWord(t *testing.T) {
	cases := []struct {
		description    string
		input          string
		cursor         int
		expectedCursor int
	}{
		{
			"wo[r]d  -> [w]ord",
			"word",
			2,
			0,
		},
		{
			" word[] -> [w]ord ",
			" word  ",
			5,
			1,
		},
		{
			"строка 1\n[с]трока2\nстрока3 -> строка [1]\nстрока2\nстрока3",
			"строка 1\nстрока2\nстрока3 ",
			9,
			7,
		},
		{
			"слово\nслово2\n[с]лово3 -> слово\n[с]лово2\nслово3",
			"слово\nслово2\nслово3",
			13,
			6,
		},
		{
			"some long []  line -> some [l]ong    line",
			"some long    line",
			10,
			5,
		},
		{
			"[] -> []",
			"",
			0,
			0,
		},
		{
			"a\n\n\n\n\n\n[b]a -> [a]\n\n\n\n\n\nba",
			"a\n\n\n\n\n\nba",
			7,
			0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			buf := NewBuffer()
			buf.InsertText(tc.input, false, true)
			buf.setCursorPosition(tc.cursor)
			GoLeftWord(buf)
			assert.Equal(t, tc.expectedCursor, buf.cursorPosition)
		})
	}
}

func TestGoRightWord(t *testing.T) {
	cases := []struct {
		description    string
		input          string
		cursor         int
		expectedCursor int
	}{
		{
			"wo[r]d -> word[]",
			"word",
			2,
			4,
		},
		{
			" []  word ->    word[]",
			"    word",
			1,
			8,
		},
		{
			"строка1\n[]строка2\nстрока3 -> строка1\n строка2[\n]строка3",
			"строка1\n строка2\nстрока3",
			8,
			16,
		},
		{
			"[] -> []",
			"",
			0,
			0,
		},
		{
			"[\n]\n\n\nслово->\n\n\n\nслово[]",
			"\n\n\n\nслово",
			0,
			9,
		},
	}
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			buf := NewBuffer()
			buf.InsertText(tc.input, false, true)
			buf.setCursorPosition(tc.cursor)
			GoRightWord(buf)
			assert.Equal(t, tc.expectedCursor, buf.cursorPosition)
		})
	}
}
