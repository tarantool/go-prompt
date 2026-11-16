package prompt

import (
	"bytes"
	"unicode/utf8"
)

var (
	// bracketedPasteStart is what a terminal in bracketed paste mode sends
	// before the text of a paste.
	bracketedPasteStart = []byte{0x1b, 0x5b, 0x32, 0x30, 0x30, 0x7e}
	// bracketedPasteEnd is what a terminal in bracketed paste mode sends
	// after the text of a paste.
	bracketedPasteEnd = []byte{0x1b, 0x5b, 0x32, 0x30, 0x31, 0x7e}
)

const (
	escByte = 0x1b
	delByte = 0x7f
)

// inputTokenKind tells a key from text.
type inputTokenKind int

const (
	// keyToken is a key: a control character or an escape sequence, which
	// is looked up among the key sequences and the key bindings.
	keyToken inputTokenKind = iota
	// textToken is text, which is inserted as it is.
	textToken
)

// inputToken is one key, or a run of text, of what is read from the
// terminal.
type inputToken struct {
	kind  inputTokenKind
	bytes []byte
	// orText is true for a key that is text in a read that holds text: a
	// tab, or a line break of a command.
	orText bool
}

// inputLexer cuts what is read from the terminal into keys and runs of text.
//
// One read holds one key when the keys are typed, but a paste, or keys typed
// faster than they are read, come in a single read: "ab\x1b[D" is the text
// "ab" and Left. A read may also end in the middle of a key or of a
// character: such a tail is kept until the next read.
//
// A tab in a read that holds text is text, as is every character between the
// start and the end of a bracketed paste: a pasted tab does not complete a
// word. A line break there is a line break of the command, or Enter if
// lineBreakIsEnter is true: "5 + 5\r" is then the text "5 + 5" and Enter, so
// that every pasted line runs as if it were typed. Keys alone, such as two
// tabs or two Enters typed at once, stay keys.
type inputLexer struct {
	// lineBreakIsEnter is true if a line break in a read with other keys is
	// Enter, not a line break of the command.
	lineBreakIsEnter bool
	// tail is the end of the last read, which is the start of a key or of a
	// character the read cut.
	tail []byte
	// inPaste is true between the start and the end of a bracketed paste.
	inPaste bool
}

// split returns the keys and the runs of text of b, a read from the
// terminal, binds being the custom key bindings by their byte sequences.
func (l *inputLexer) split(b []byte, binds []ASCIICodeBind) []inputToken {
	if len(l.tail) > 0 {
		b = append(l.tail, b...)
		l.tail = nil
	}

	// A read that is a key as a whole is that key, as it has always been.
	if !l.inPaste && isWholeKey(b, binds) {
		return []inputToken{{kind: keyToken, bytes: b}}
	}

	var tokens []inputToken
	for len(b) > 0 {
		var (
			token *inputToken
			n     int
		)
		if l.inPaste {
			token, n = l.nextPasted(b)
		} else {
			token, n = l.nextTyped(b, binds)
		}
		if n == 0 {
			l.tail = append([]byte(nil), b...)
			break
		}
		if token != nil {
			tokens = append(tokens, *token)
		}
		b = b[n:]
	}

	// A tab, or a line break of a command, that comes with text is text.
	hasText := false
	for _, token := range tokens {
		hasText = hasText || token.kind == textToken
	}
	if hasText {
		for i := range tokens {
			if tokens[i].orText {
				tokens[i].kind = textToken
			}
		}
	}

	return tokens
}

// isWholeKey reports whether b is a key sequence or the sequence of a key
// binding.
func isWholeKey(b []byte, binds []ASCIICodeBind) bool {
	if GetKey(b) != NotDefined {
		return true
	}
	for _, bind := range binds {
		if bytes.Equal(bind.ASCIICode, b) {
			return true
		}
	}
	return false
}

// nextTyped returns the first key or run of text of b and the number of
// bytes it takes. A token of nil skips the bytes, and no bytes taken means
// that b is the start of a key or of a character, which the next read
// continues.
func (l *inputLexer) nextTyped(b []byte, binds []ASCIICodeBind) (*inputToken, int) {
	switch c := b[0]; {
	case c == escByte:
		return l.nextEscape(b, binds)
	case c == '\r', c == '\n':
		return l.nextLineBreak(b, false)
	case c == '\t':
		return &inputToken{kind: keyToken, bytes: b[:1], orText: true}, 1
	case isControl(c):
		return &inputToken{kind: keyToken, bytes: b[:1]}, 1
	}

	n := textLen(b, false)
	if n == 0 {
		return nil, 0
	}
	return &inputToken{kind: textToken, bytes: b[:n]}, n
}

// nextEscape returns the key b starts with, b starting with an escape. It is
// the longest key sequence or key binding b starts with, an escape sequence
// no key has, an escape with the character after it (Meta and the
// character), or the escape key itself.
func (l *inputLexer) nextEscape(b []byte, binds []ASCIICodeBind) (*inputToken, int) {
	if bytes.HasPrefix(b, bracketedPasteStart) {
		l.inPaste = true
		return nil, len(bracketedPasteStart)
	}
	if bytes.HasPrefix(b, bracketedPasteEnd) {
		return nil, len(bracketedPasteEnd)
	}

	n := 0
	for _, code := range ASCIISequences {
		if len(code.ASCIICode) > n && bytes.HasPrefix(b, code.ASCIICode) {
			n = len(code.ASCIICode)
		}
	}
	for _, bind := range binds {
		if len(bind.ASCIICode) > n && bytes.HasPrefix(b, bind.ASCIICode) {
			n = len(bind.ASCIICode)
		}
	}

	csi, complete := csiLen(b)
	if csi > 0 && !complete {
		// A control sequence the read cut.
		return nil, 0
	}
	if csi > n {
		n = csi
	}

	if n <= 1 && len(b) > 1 && b[1] != escByte {
		if !utf8.FullRune(b[1:]) {
			return nil, 0
		}
		_, size := utf8.DecodeRune(b[1:])
		n = 1 + size
	}
	if n == 0 {
		n = 1
	}

	return &inputToken{kind: keyToken, bytes: b[:n]}, n
}

// csiLen returns the length of the control sequence b starts with and
// whether it is complete: an escape, "[" or "O", the parameter and the
// intermediate bytes, and the final byte. It returns 0 if b starts with no
// control sequence.
func csiLen(b []byte) (int, bool) {
	if len(b) < 2 || b[0] != escByte {
		return 0, false
	}
	switch b[1] {
	case 'O':
		if len(b) < 3 {
			return 2, false
		}
		return 3, true
	case '[':
	default:
		return 0, false
	}
	for i := 2; i < len(b); i++ {
		switch c := b[i]; {
		case c >= 0x20 && c <= 0x3f:
			// A parameter or an intermediate byte.
		case c >= 0x40 && c <= 0x7e:
			return i + 1, true
		default:
			// Not a control sequence: the escape and "[" are keys.
			return 0, false
		}
	}
	return len(b), false
}

// nextPasted returns the first line break or run of text of b, b being
// pasted. The end of the paste, and every control character but a tab, take
// bytes and give no token.
func (l *inputLexer) nextPasted(b []byte) (*inputToken, int) {
	if bytes.HasPrefix(b, bracketedPasteEnd) {
		l.inPaste = false
		return nil, len(bracketedPasteEnd)
	}
	if len(b) < len(bracketedPasteEnd) && bytes.HasPrefix(bracketedPasteEnd, b) {
		// The end of the paste the read cut.
		return nil, 0
	}

	switch c := b[0]; {
	case c == '\r', c == '\n':
		return l.nextLineBreak(b, true)
	case c != '\t' && isControl(c):
		return nil, 1
	}

	n := textLen(b, true)
	if n == 0 {
		return nil, 0
	}
	return &inputToken{kind: textToken, bytes: b[:n]}, n
}

// nextLineBreak returns the line break b starts with: "\r\n", "\r" or "\n".
// It is Enter if lineBreakIsEnter is true. Otherwise it is a line break of
// the command if it is pasted, or if it comes with text.
func (l *inputLexer) nextLineBreak(b []byte, pasted bool) (*inputToken, int) {
	n := 1
	if b[0] == '\r' && len(b) > 1 && b[1] == '\n' {
		n = 2
	}
	switch {
	case l.lineBreakIsEnter:
		return &inputToken{kind: keyToken, bytes: []byte{'\r'}}, n
	case pasted:
		return &inputToken{kind: textToken, bytes: []byte{'\n'}}, n
	}
	return &inputToken{kind: keyToken, bytes: []byte{'\n'}, orText: true}, n
}

// textLen returns the length of the run of text b starts with: up to a
// control character, a tab being text if tabs is true. A character the end
// of b cuts is left out, so a run of nothing but such a character has no
// length.
func textLen(b []byte, tabs bool) int {
	n := 0
	for n < len(b) && (tabs && b[n] == '\t' || !isControl(b[n])) {
		n++
	}
	if n < len(b) {
		return n
	}

	// The run goes to the end of b: the read may have cut its last
	// character.
	start := n - 1
	for start > 0 && n-start < utf8.UTFMax && !utf8.RuneStart(b[start]) {
		start--
	}
	if !utf8.FullRune(b[start:n]) {
		return start
	}
	return n
}

// isControl reports whether c is a control character of ASCII.
func isControl(c byte) bool {
	return c < 0x20 || c == delByte
}
