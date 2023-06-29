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

}
