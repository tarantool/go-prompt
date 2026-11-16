//go:build !windows
// +build !windows

package prompt

import (
	"bytes"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testParser is a ConsoleParser that reads what a test sends.
type testParser struct {
	reads chan []byte
}

func (*testParser) Setup() error    { return nil }
func (*testParser) TearDown() error { return nil }

func (*testParser) GetWinSize() *WinSize {
	return &WinSize{Row: 24, Col: 80}
}

func (p *testParser) Read() ([]byte, error) {
	select {
	case b := <-p.reads:
		return b, nil
	default:
		return []byte{}, syscall.EAGAIN
	}
}

// testWriter is a ConsoleWriter that keeps what is written.
type testWriter struct {
	VT100Writer
	mu  sync.Mutex
	out bytes.Buffer
}

func (w *testWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.out.Write(w.buffer)
	w.buffer = []byte{}
	return nil
}

func (w *testWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.out.String()
}

// runPrompt runs a prompt that reads reads, one by one, and returns the
// commands it executed and what it wrote.
func runPrompt(t *testing.T, reads []string, opts ...Option) ([]string, string) {
	t.Helper()

	parser := &testParser{reads: make(chan []byte, len(reads)+1)}
	for _, read := range reads {
		parser.reads <- []byte(read)
	}
	// Ctrl-D on an empty line stops the prompt.
	parser.reads <- []byte{0x4}

	savedParser := getInputParser
	getInputParser = func() *PosixParser { return nil }
	defer func() { getInputParser = savedParser }()

	writer := &testWriter{}
	var executed []string
	opts = append([]Option{OptionParser(parser), OptionWriter(writer)}, opts...)
	p := New(func(in string) { executed = append(executed, in) },
		func(Document) []Suggest { return nil }, opts...)
	p.Run()

	return executed, writer.String()
}

func TestPromptRunExecutesEveryLineOfARead(t *testing.T) {
	executed, _ := runPrompt(t, []string{"return 1\nreturn 2\n", "return 3\r"},
		OptionExecutePastedLines())
	assert.Equal(t, []string{"return 1", "return 2", "return 3"}, executed)
}

func TestPromptRunMultiLineCommandOfARead(t *testing.T) {
	executed, _ := runPrompt(t, []string{"return 1\nreturn\t2", "\r"})
	assert.Equal(t, []string{"return 1\nreturn    2"}, executed)
}

func TestPromptRunCompletesWithTabsAlone(t *testing.T) {
	parser := &testParser{reads: make(chan []byte, 4)}
	for _, read := range []string{"a", "\t\t", "\r", "\x04"} {
		parser.reads <- []byte(read)
	}

	savedParser := getInputParser
	getInputParser = func() *PosixParser { return nil }
	defer func() { getInputParser = savedParser }()

	var executed []string
	p := New(func(in string) { executed = append(executed, in) },
		func(d Document) []Suggest {
			if d.Text == "" {
				return nil
			}
			return []Suggest{{Text: "abc"}, {Text: "abd"}}
		},
		OptionParser(parser), OptionWriter(&testWriter{}), OptionExecutePastedLines())
	p.Run()

	// Two tabs at once select the second suggestion, which Enter takes.
	assert.Equal(t, []string{"abd"}, executed)
}

func TestPromptRunExecutesKeysAfterText(t *testing.T) {
	// Left moves the cursor before "c", and Backspace deletes "b".
	executed, _ := runPrompt(t, []string{"abc\x1b[D\x7fd\r"}, OptionExecutePastedLines())
	assert.Equal(t, []string{"adc"}, executed)
}

func TestPromptRunBracketedPaste(t *testing.T) {
	executed, out := runPrompt(t, []string{
		"\x1b[200~if x then\r\n\treturn 1\r\nend\x1b[201~", "\r",
	}, OptionBracketedPaste(), OptionExecutePastedLines())
	assert.Equal(t, []string{"if x then", "    return 1", "end"}, executed)

	// On while the prompt reads, off while a command runs and at the end.
	on, off := "\x1b[?2004h", "\x1b[?2004l"
	assert.Equal(t, 4, strings.Count(out, on))
	assert.Equal(t, 4, strings.Count(out, off))
	assert.Less(t, strings.Index(out, on), strings.Index(out, off))
	assert.Greater(t, strings.LastIndex(out, off), strings.LastIndex(out, on))
}

func TestPromptRunNoBracketedPasteByDefault(t *testing.T) {
	_, out := runPrompt(t, []string{"a", "\r"})
	assert.NotContains(t, out, "\x1b[?2004")
}
