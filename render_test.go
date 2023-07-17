//go:build !windows
// +build !windows

package prompt

import (
	"bytes"
	"io"
	"reflect"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCompletion(t *testing.T) {
	scenarioTable := []struct {
		scenario      string
		completions   []Suggest
		prefix        string
		suffix        string
		expected      []Suggest
		maxWidth      int
		expectedWidth int
	}{
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select"},
				{Text: "from"},
				{Text: "insert"},
				{Text: "where"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select "},
				{Text: " from   "},
				{Text: " insert "},
				{Text: " where  "},
			},
			maxWidth:      20,
			expectedWidth: 8,
		},
		{
			scenario: "",
			completions: []Suggest{
				{Text: "select", Description: "select description"},
				{Text: "from", Description: "from description"},
				{Text: "insert", Description: "insert description"},
				{Text: "where", Description: "where description"},
			},
			prefix: " ",
			suffix: " ",
			expected: []Suggest{
				{Text: " select ", Description: " select description "},
				{Text: " from   ", Description: " from description   "},
				{Text: " insert ", Description: " insert description "},
				{Text: " where  ", Description: " where description  "},
			},
			maxWidth:      40,
			expectedWidth: 28,
		},
	}

	for _, s := range scenarioTable {
		ac, width := formatSuggestions(s.completions, s.maxWidth)
		if !reflect.DeepEqual(ac, s.expected) {
			t.Errorf("Should be %#v, but got %#v", s.expected, ac)
		}
		if width != s.expectedWidth {
			t.Errorf("Should be %#v, but got %#v", s.expectedWidth, width)
		}
	}
}

func TestBreakLineCallback(t *testing.T) {
	var i int
	r := &Render{
		out: &PosixWriter{
			fd: syscall.Stdin, // "write" to stdin just so we don't mess with the output of the
			// tests
		},
		prefixTextColor:              Blue,
		prefixBGColor:                DefaultColor,
		inputTextColor:               DefaultColor,
		inputBGColor:                 DefaultColor,
		previewSuggestionTextColor:   Green,
		previewSuggestionBGColor:     DefaultColor,
		suggestionTextColor:          White,
		suggestionBGColor:            Cyan,
		selectedSuggestionTextColor:  Black,
		selectedSuggestionBGColor:    Turquoise,
		descriptionTextColor:         Black,
		descriptionBGColor:           Turquoise,
		selectedDescriptionTextColor: White,
		selectedDescriptionBGColor:   Cyan,
		scrollbarThumbColor:          DarkGray,
		scrollbarBGColor:             Cyan,
		col:                          1,
	}
	b := NewBuffer()
	r.renderBreakLine(renderCtx{cmd: b})

	if i != 0 {
		t.Errorf("i should initially be 0, before applying a break line callback")
	}

	r.breakLineCallback = func(doc *Document) {
		i++
	}
	r.renderBreakLine(renderCtx{cmd: b})
	r.renderBreakLine(renderCtx{cmd: b})
	r.renderBreakLine(renderCtx{cmd: b})

	if i != 3 {
		t.Errorf("BreakLine callback not called, i should be 3")
	}
}

type mockConsoleWriter struct {
	VT100Writer
	w       io.Writer
	flushed bool
}

func (m *mockConsoleWriter) Flush() error {
	m.flushed = true
	m.w.Write(m.buffer)
	m.buffer = m.buffer[:0]
	return nil
}

var _ ConsoleWriter = &mockConsoleWriter{}

func TestWriteCmd(t *testing.T) {
	buffer := bytes.Buffer{}
	consoleWriter := &mockConsoleWriter{w: &buffer}
	prefixColor := DarkBlue

	cases := []struct {
		cmd       string
		prefixLen int
		expected  string
	}{
		{
			cmd:       "command1\ncommand2\n",
			prefixLen: 0,
			expected:  "\x1b[0;34;49m\x1b[0;39;49mcommand1\ncommand2\n",
		},
		{
			cmd:       "prefix> command1\ncommand2\n¥¥¼",
			prefixLen: 7,
			expected:  "\x1b[0;34;49mprefix>\x1b[0;39;49m command1\ncommand2\n¥¥¼",
		},
	}

	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			writeCmdWithPrefix(consoleWriter, tc.cmd, tc.prefixLen,
				prefixColor, DefaultColor, DefaultColor)
			consoleWriter.Flush()
			assert.Equal(t, tc.expected, string(buffer.Bytes()))
			buffer.Reset()
		})
	}
}
