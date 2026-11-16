package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func key(s string) inputToken {
	return inputToken{kind: keyToken, bytes: []byte(s)}
}

func text(s string) inputToken {
	return inputToken{kind: textToken, bytes: []byte(s)}
}

func TestInputLexerSplit(t *testing.T) {
	binds := []ASCIICodeBind{{ASCIICode: []byte{0x1b, 0x62}}}

	tests := []struct {
		name  string
		enter bool
		reads []string
		want  []inputToken
	}{
		{"one key", false, []string{"\r"}, []inputToken{key("\r")}},
		{"one line feed", false, []string{"\n"}, []inputToken{key("\n")}},
		{"one tab", false, []string{"\t"}, []inputToken{key("\t")}},
		{"one escape sequence", false, []string{"\x1b[A"}, []inputToken{key("\x1b[A")}},
		{"one binding", false, []string{"\x1bb"}, []inputToken{key("\x1bb")}},
		{"text", false, []string{"abc"}, []inputToken{text("abc")}},
		{
			"tabs", false, []string{"a\tb\t"},
			[]inputToken{text("a"), text("\t"), text("b"), text("\t")},
		},
		{"tabs alone", false, []string{"\t\t"}, []inputToken{key("\t"), key("\t")}},
		{"enters alone", false, []string{"\r\r"}, []inputToken{key("\n"), key("\n")}},
		{"enters alone executed", true, []string{"\r\n\r"}, []inputToken{key("\r"), key("\r")}},
		{
			"line breaks of a command", false, []string{"a\nb\r\nc\r"},
			[]inputToken{
				text("a"), text("\n"), text("b"), text("\n"), text("c"), text("\n"),
			},
		},
		{"line", true, []string{"5 + 5\r"}, []inputToken{text("5 + 5"), key("\r")}},
		{
			"lines", true, []string{"return 1\nreturn 2\n"},
			[]inputToken{text("return 1"), key("\r"), text("return 2"), key("\r")},
		},
		{
			"line breaks of a paste", true, []string{"a\r\nb\rc"},
			[]inputToken{text("a"), key("\r"), text("b"), key("\r"), text("c")},
		},
		{
			"keys in text", false, []string{"ab\x1b[Dc\x7f\t"},
			[]inputToken{text("ab"), key("\x1b[D"), text("c"), key("\x7f"), text("\t")},
		},
		{
			"longest sequence", false, []string{"\x1b[1;5Cx"},
			[]inputToken{key("\x1b[1;5C"), text("x")},
		},
		{
			"binding in text", false, []string{"a\x1bbc"},
			[]inputToken{text("a"), key("\x1bb"), text("c")},
		},
		{
			"unknown control sequence", false, []string{"a\x1b[9;9Xb"},
			[]inputToken{text("a"), key("\x1b[9;9X"), text("b")},
		},
		{
			"meta", false, []string{"a\x1bxy"},
			[]inputToken{text("a"), key("\x1bx"), text("y")},
		},
		{
			"escape at the end", false, []string{"a\x1b"},
			[]inputToken{text("a"), key("\x1b")},
		},
		{
			"control sequence cut", false, []string{"a\x1b[1;", "5Cb"},
			[]inputToken{text("a"), key("\x1b[1;5C"), text("b")},
		},
		{
			"character cut", true, []string{"да\xd0", "\xb0\r"},
			[]inputToken{text("да"), text("а"), key("\r")},
		},
		{
			"bracketed paste", true, []string{"\x1b[200~a\tb\r\nc\x03\rd\x1b[201~\r"},
			[]inputToken{text("a\tb"), key("\r"), text("c"), key("\r"), text("d"), key("\r")},
		},
		{
			"bracketed paste of a command", false, []string{"\x1b[200~a\r\nb\x1b[201~"},
			[]inputToken{text("a"), text("\n"), text("b")},
		},
		{
			"bracketed paste cut", true, []string{"\x1b[20", "0~a\r\x1b[2", "01~\x1b[A"},
			[]inputToken{text("a"), key("\r"), key("\x1b[A")},
		},
		{
			"keys pasted", false, []string{"\x1b[200~\x1b[A\x7f\x1b[201~"},
			[]inputToken{text("[A")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []inputToken
			lexer := inputLexer{lineBreakIsEnter: tt.enter}
			for _, read := range tt.reads {
				got = append(got, lexer.split([]byte(read), binds)...)
			}
			for i := range got {
				got[i].orText = false
			}
			assert.Equal(t, tt.want, got)
			assert.Empty(t, lexer.tail)
			assert.False(t, lexer.inPaste)
		})
	}
}

func TestInputLexerSplitKeepsTheTail(t *testing.T) {
	var lexer inputLexer

	assert.Equal(t, []inputToken{text("a")}, lexer.split([]byte("a\x1b[20"), nil))
	assert.Equal(t, []byte("\x1b[20"), lexer.tail)

	assert.Equal(t, []inputToken{text("b")}, lexer.split([]byte("0~b\x1b[201"), nil))
	assert.True(t, lexer.inPaste)
	assert.Equal(t, []byte("\x1b[201"), lexer.tail)
}
