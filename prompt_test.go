package prompt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentPrefix(t *testing.T) {
	prompt := &Prompt{
		prefix: "prefix",
		livePrefixCallback: func() (prefix string, useLivePrefix bool) {
			return "", false
		},
		history: NewHistory(),
	}

	t.Run("basic", func(t *testing.T) {
		actualPrefix := prompt.getCurrentPrefix()
		assert.Equal(t, prompt.prefix, actualPrefix)
	})

	t.Run("live-prefix", func(t *testing.T) {
		prompt.livePrefixCallback = func() (prefix string, useLivePrefix bool) {
			return "live_prefix", true
		}
		actualPrefix := prompt.getCurrentPrefix()
		assert.Equal(t, "live_prefix", actualPrefix)
	})

	t.Run("reverse-search enabled", func(t *testing.T) {
		prompt.reverseSearch = NewReverseSearch(prompt.history)
		prompt.reverseSearch.matchedIndex = 1
		prompt.buf = NewBuffer()
		prompt.buf.InsertText("input", false, true)
		actualPrefix := prompt.getCurrentPrefix()
		assert.Equal(t, fmt.Sprintf(matchSearchPrefixFmt, "input"),
			actualPrefix)
	})
}

func TestGetCmdToRender(t *testing.T) {
	prompt := &Prompt{
		prefix: "prefix> ",
		livePrefixCallback: func() (prefix string, useLivePrefix bool) {
			return "", false
		},
		history: NewHistory(),
	}

	input := "Ïа_#строка"
	prompt.buf = NewBuffer()
	prompt.buf.InsertText(input, false, true)

	matchedCmd := fmt.Sprintf("print(`%s`)", input)
	prompt.history.Add(matchedCmd)

	t.Run("basic", func(t *testing.T) {
		cmdBuf, prefixLen := prompt.getCmdToRender()
		assert.Equal(t, "prefix> "+input, cmdBuf.Text())
		assert.Equal(t, 8, prefixLen)
		assert.Equal(t, 8+len([]rune(input)), cmdBuf.cursorPosition)
	})

	t.Run("reverse-search enabled", func(t *testing.T) {
		prompt.reverseSearch = NewReverseSearch(prompt.history)
		prompt.reverseSearch.update(input)

		cmdBuf, prefixLen := prompt.getCmdToRender()
		expectedPrefx := fmt.Sprintf(matchSearchPrefixFmt, input)
		assert.Equal(t, expectedPrefx+matchedCmd, cmdBuf.Text())
		assert.Equal(t, len(expectedPrefx), prefixLen)
		assert.Equal(t, len([]rune(matchedCmd))+len([]rune(expectedPrefx)), cmdBuf.cursorPosition)
	})
}

func TestInReverseSearchMode(t *testing.T) {
	prompt := &Prompt{history: NewHistory()}
	assert.False(t, prompt.inReverseSearchMode())
	prompt.reverseSearch = NewReverseSearch(prompt.history)
	assert.True(t, prompt.inReverseSearchMode())
	prompt.reverseSearch = nil
	assert.False(t, prompt.inReverseSearchMode())
}

func TestEnableReverseSearch(t *testing.T) {
	prompt := &Prompt{
		isReverseSearchEnabled: true,
		history:                NewHistory(),
		buf:                    NewBuffer(),
	}

	t.Run("basic", func(t *testing.T) {
		prompt.buf.InsertText("multi\nline", false, true)
		prompt.enableReverseSearch()
		assert.Equal(t, "", prompt.buf.Text())
		assert.NotNil(t, prompt.reverseSearch)
		prompt.reverseSearch = nil
	})

	t.Run("disabled option", func(t *testing.T) {
		prompt.buf.InsertText("multi\nline", false, true)
		prompt.isReverseSearchEnabled = false
		prompt.enableReverseSearch()
		assert.Nil(t, prompt.reverseSearch)
		assert.Equal(t, "multi\nline", prompt.buf.Text())
	})
}

func TestDisableReverseSearch(t *testing.T) {
	history := NewHistory()
	history.Add("entry 11")
	history.Add("entry 2")

	prompt := &Prompt{
		isReverseSearchEnabled: true,
		history:                history,
		reverseSearch:          NewReverseSearch(history),
		buf:                    NewBuffer(),
	}

	// Have match.
	prompt.reverseSearch.update("entry 1")
	prompt.disableReverseSearch()
	assert.Nil(t, prompt.reverseSearch)
	assert.Equal(t, "entry 11", prompt.buf.Text())
	assert.Equal(t, 0, history.selected)

	// Haven't match.
	prompt.reverseSearch = NewReverseSearch(history)
	prompt.reverseSearch.update("eentry")
	prompt.disableReverseSearch()
	assert.Nil(t, prompt.reverseSearch)
	assert.Equal(t, "", prompt.buf.Text())
	assert.Equal(t, len(history.tmp)-1, history.selected)
}
