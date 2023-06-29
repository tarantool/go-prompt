package prompt

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func ExampleDocument_CurrentLine() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.CurrentLine())
	// Output:
	// This is a example of Document component.
}

func ExampleDocument_DisplayCursorPosition() {
	d := &Document{
		Text:           `Hello! my name is c-bata.`,
		cursorPosition: len(`Hello`),
	}
	fmt.Println("DisplayCursorPosition", d.DisplayCursorPosition())
	// Output:
	// DisplayCursorPosition 5
}

func ExampleDocument_CursorPositionRow() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println("CursorPositionRow", d.CursorPositionRow())
	// Output:
	// CursorPositionRow 1
}

func ExampleDocument_CursorPositionCol() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println("CursorPositionCol", d.CursorPositionCol())
	// Output:
	// CursorPositionCol 14
}

func ExampleDocument_TextBeforeCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.TextBeforeCursor())
	// Output:
	// Hello! my name is c-bata.
	// This is a exam
}

func ExampleDocument_TextAfterCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.TextAfterCursor())
	// Output:
	// ple of Document component.
	// This component has texts displayed in terminal and cursor position.
}

func ExampleDocument_DisplayCursorPosition_withJapanese() {
	d := &Document{
		Text:           `こんにちは、芝田 将です。`,
		cursorPosition: 3,
	}
	fmt.Println("DisplayCursorPosition", d.DisplayCursorPosition())
	// Output:
	// DisplayCursorPosition 6
}

func ExampleDocument_CurrentLineBeforeCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.CurrentLineBeforeCursor())
	// Output:
	// This is a exam
}

func ExampleDocument_CurrentLineAfterCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
This component has texts displayed in terminal and cursor position.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.CurrentLineAfterCursor())
	// Output:
	// ple of Document component.
}

func ExampleDocument_GetWordBeforeCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.GetWordBeforeCursor())
	// Output:
	// exam
}

func ExampleDocument_GetWordAfterCursor() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a exam`),
	}
	fmt.Println(d.GetWordAfterCursor())
	// Output:
	// ple
}

func ExampleDocument_GetWordBeforeCursorWithSpace() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a example `),
	}
	fmt.Println(d.GetWordBeforeCursorWithSpace())
	// Output:
	// example
}

func ExampleDocument_GetWordAfterCursorWithSpace() {
	d := &Document{
		Text: `Hello! my name is c-bata.
This is a example of Document component.
`,
		cursorPosition: len(`Hello! my name is c-bata.
This is a`),
	}
	fmt.Println(d.GetWordAfterCursorWithSpace())
	// Output:
	//  example
}

func ExampleDocument_GetWordBeforeCursorUntilSeparator() {
	d := &Document{
		Text:           `hello,i am c-bata`,
		cursorPosition: len(`hello,i am c`),
	}
	fmt.Println(d.GetWordBeforeCursorUntilSeparator(","))
	// Output:
	// i am c
}

func ExampleDocument_GetWordAfterCursorUntilSeparator() {
	d := &Document{
		Text:           `hello,i am c-bata,thank you for using go-prompt`,
		cursorPosition: len(`hello,i a`),
	}
	fmt.Println(d.GetWordAfterCursorUntilSeparator(","))
	// Output:
	// m c-bata
}

func ExampleDocument_GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor() {
	d := &Document{
		Text:           `hello,i am c-bata,thank you for using go-prompt`,
		cursorPosition: len(`hello,i am c-bata,`),
	}
	fmt.Println(d.GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor(","))
	// Output:
	// i am c-bata,
}

func ExampleDocument_GetWordAfterCursorUntilSeparatorIgnoreNextToCursor() {
	d := &Document{
		Text:           `hello,i am c-bata,thank you for using go-prompt`,
		cursorPosition: len(`hello`),
	}
	fmt.Println(d.GetWordAfterCursorUntilSeparatorIgnoreNextToCursor(","))
	// Output:
	// ,i am c-bata
}

func TestDocument_DisplayCursorPosition(t *testing.T) {
	patterns := []struct {
		document *Document
		expected int
	}{
		{
			document: &Document{
				Text:           "hello",
				cursorPosition: 2,
			},
			expected: 2,
		},
		{
			document: &Document{
				Text:           "こんにちは",
				cursorPosition: 2,
			},
			expected: 4,
		},
		{
			// If you're facing test failure on this test case and your terminal is iTerm2,
			// please check 'Profile -> Text' configuration. 'Use Unicode version 9 widths'
			// must be checked.
			// https://github.com/c-bata/go-prompt/pull/99
			document: &Document{
				Text:           "Добрый день",
				cursorPosition: 3,
			},
			expected: 3,
		},
	}

	for _, p := range patterns {
		ac := p.document.DisplayCursorPosition()
		assert.Equal(t, p.expected, ac)
	}
}

func TestDocument_GetCharRelativeToCursor(t *testing.T) {
	patterns := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len([]rune("line 1\n" + "lin")),
			},
			expected: "e",
		},
		{
			document: &Document{
				Text:           "あいうえお\nかきくけこ\nさしすせそ\nたちつてと\n",
				cursorPosition: 8,
			},
			expected: "く",
		},
		{
			document: &Document{
				Text:           "Добрый\nдень\nДобрый день",
				cursorPosition: 9,
			},
			expected: "н",
		},
	}

	for _, p := range patterns {
		ac := p.document.GetCharRelativeToCursor(1)
		ex, _ := utf8.DecodeRuneInString(p.expected)
		assert.Equal(t, string(ex), string(ac))
	}
}

func TestDocument_TextBeforeCursor(t *testing.T) {
	patterns := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "lin"),
			},
			expected: "line 1\nlin",
		},
		{
			document: &Document{
				Text:           "あいうえお\nかきくけこ\nさしすせそ\nたちつてと\n",
				cursorPosition: 8,
			},
			expected: "あいうえお\nかき",
		},
		{
			document: &Document{
				Text:           "Добрый\nдень\nДобрый день",
				cursorPosition: 9,
			},
			expected: "Добрый\nде",
		},
	}
	for _, p := range patterns {
		ac := p.document.TextBeforeCursor()
		assert.Equal(t, p.expected, ac)
	}
}

func TestDocument_TextAfterCursor(t *testing.T) {
	pattern := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "lin"),
			},
			expected: "e 2\nline 3\nline 4\n",
		},
		{
			document: &Document{
				Text:           "",
				cursorPosition: 0,
			},
			expected: "",
		},
		{
			document: &Document{
				Text:           "あいうえお\nかきくけこ\nさしすせそ\nたちつてと\n",
				cursorPosition: 8,
			},
			expected: "くけこ\nさしすせそ\nたちつてと\n",
		},
		{
			document: &Document{
				Text:           "Добрый\nдень\nДобрый день",
				cursorPosition: 9,
			},
			expected: "нь\nДобрый день",
		},
	}

	for _, p := range pattern {
		ac := p.document.TextAfterCursor()
		assert.Equal(t, p.expected, ac)
	}
}

func TestDocument_GetWordBeforeCursor(t *testing.T) {
	pattern := []struct {
		document *Document
		expected string
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: "bana",
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./file/foo.json"),
			},
			expected: "foo.json",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple banana orange",
				cursorPosition: len("apple ba"),
			},
			expected: "ba",
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./fi"),
			},
			expected: "fi",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple ",
				cursorPosition: len("apple "),
			},
			expected: "",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ さしすせそ",
				cursorPosition: 8,
			},
			expected: "かき",
		},
		{
			document: &Document{
				Text:           "Добрый день Добрый день",
				cursorPosition: 9,
			},
			expected: "де",
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.GetWordBeforeCursor()
			assert.Equal(t, p.expected, ac)
			ac = p.document.GetWordBeforeCursorUntilSeparator("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.GetWordBeforeCursorUntilSeparator(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_GetWordBeforeCursorWithSpace(t *testing.T) {
	pattern := []struct {
		document *Document
		expected string
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana ",
				cursorPosition: len("apple bana "),
			},
			expected: "bana ",
		},
		{
			document: &Document{
				Text:           "apply -f /path/to/file/",
				cursorPosition: len("apply -f /path/to/file/"),
			},
			expected: "file/",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple ",
				cursorPosition: len("apple "),
			},
			expected: "apple ",
		},
		{
			document: &Document{
				Text:           "path/",
				cursorPosition: len("path/"),
			},
			expected: "path/",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ ",
				cursorPosition: 12,
			},
			expected: "かきくけこ ",
		},
		{
			document: &Document{
				Text:           "Добрый день ",
				cursorPosition: 12,
			},
			expected: "день ",
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.GetWordBeforeCursorWithSpace()
			assert.Equal(t, p.expected, ac)
			ac = p.document.GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_FindStartOfPreviousWord(t *testing.T) {
	pattern := []struct {
		document *Document
		expected int
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: len("apple "),
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./file/foo.json"),
			},
			expected: len("apply -f ./file/"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple ",
				cursorPosition: len("apple "),
			},
			expected: len("apple "),
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./"),
			},
			expected: len("apply -f ./"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ さしすせそ",
				cursorPosition: 8, // between 'き' and 'く'
			},
			expected: len("あいうえお "), // this function returns index byte in string
		},
		{
			document: &Document{
				Text:           "Добрый день Добрый день",
				cursorPosition: 9,
			},
			expected: len("Добрый "), // this function returns index byte in string
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.FindStartOfPreviousWord()
			assert.Equal(t, p.expected, ac)
			ac = p.document.FindStartOfPreviousWordUntilSeparator("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.FindStartOfPreviousWordUntilSeparator(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_FindStartOfPreviousWordWithSpace(t *testing.T) {
	pattern := []struct {
		document *Document
		expected int
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana ",
				cursorPosition: len("apple bana "),
			},
			expected: len("apple "),
		},
		{
			document: &Document{
				Text:           "apply -f /file/foo/",
				cursorPosition: len("apply -f /file/foo/"),
			},
			expected: len("apply -f /file/"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple ",
				cursorPosition: len("apple "),
			},
			expected: len(""),
		},
		{
			document: &Document{
				Text:           "file/",
				cursorPosition: len("file/"),
			},
			expected: len(""),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ ",
				cursorPosition: 12, // cursor points to last
			},
			expected: len("あいうえお "), // this function returns index byte in string
		},
		{
			document: &Document{
				Text:           "Добрый день ",
				cursorPosition: 12,
			},
			expected: len("Добрый "), // this function returns index byte in string
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.FindStartOfPreviousWordWithSpace()
			assert.Equal(t, p.expected, ac)
			ac = p.document.FindStartOfPreviousWordUntilSeparatorIgnoreNextToCursor("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.FindStartOfPreviousWordUntilSeparatorIgnoreNextToCursor(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_GetWordAfterCursor(t *testing.T) {
	pattern := []struct {
		document *Document
		expected string
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: "",
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./fi"),
			},
			expected: "le",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple "),
			},
			expected: "bana",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple"),
			},
			expected: "",
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ."),
			},
			expected: "",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("ap"),
			},
			expected: "ple",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ さしすせそ",
				cursorPosition: 8,
			},
			expected: "くけこ",
		},
		{
			document: &Document{
				Text:           "Добрый день Добрый день",
				cursorPosition: 9,
			},
			expected: "нь",
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.GetWordAfterCursor()
			assert.Equal(t, p.expected, ac)
			ac = p.document.GetWordAfterCursorUntilSeparator("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.GetWordAfterCursorUntilSeparator(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_GetWordAfterCursorWithSpace(t *testing.T) {
	pattern := []struct {
		document *Document
		expected string
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: "",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple "),
			},
			expected: "bana",
		},
		{
			document: &Document{
				Text:           "/path/to",
				cursorPosition: len("/path/"),
			},
			expected: "to",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "/path/to/file",
				cursorPosition: len("/path/"),
			},
			expected: "to",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple"),
			},
			expected: " bana",
		},
		{
			document: &Document{
				Text:           "path/to",
				cursorPosition: len("path"),
			},
			expected: "/to",
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("ap"),
			},
			expected: "ple",
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ さしすせそ",
				cursorPosition: 5,
			},
			expected: " かきくけこ",
		},
		{
			document: &Document{
				Text:           "Добрый день Добрый день",
				cursorPosition: 6,
			},
			expected: " день",
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.GetWordAfterCursorWithSpace()
			assert.Equal(t, p.expected, ac)
			ac = p.document.GetWordAfterCursorUntilSeparatorIgnoreNextToCursor("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.GetWordAfterCursorUntilSeparatorIgnoreNextToCursor(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_FindEndOfCurrentWord(t *testing.T) {
	pattern := []struct {
		document *Document
		expected int
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: len(""),
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple "),
			},
			expected: len("bana"),
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ./"),
			},
			expected: len("file"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple"),
			},
			expected: len(""),
		},
		{
			document: &Document{
				Text:           "apply -f ./file/foo.json",
				cursorPosition: len("apply -f ."),
			},
			expected: len(""),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("ap"),
			},
			expected: len("ple"),
		},
		{
			// りん(cursor)ご ばなな
			document: &Document{
				Text:           "りんご ばなな",
				cursorPosition: 2,
			},
			expected: len("ご"),
		},
		{
			document: &Document{
				Text:           "りんご ばなな",
				cursorPosition: 3,
			},
			expected: 0,
		},
		{
			// Доб(cursor)рый день
			document: &Document{
				Text:           "Добрый день",
				cursorPosition: 3,
			},
			expected: len("рый"),
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.FindEndOfCurrentWord()
			assert.Equal(t, p.expected, ac)
			ac = p.document.FindEndOfCurrentWordUntilSeparator("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.FindEndOfCurrentWordUntilSeparator(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_FindEndOfCurrentWordWithSpace(t *testing.T) {
	pattern := []struct {
		document *Document
		expected int
		sep      string
	}{
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple bana"),
			},
			expected: len(""),
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple "),
			},
			expected: len("bana"),
		},
		{
			document: &Document{
				Text:           "apply -f /file/foo.json",
				cursorPosition: len("apply -f /"),
			},
			expected: len("file"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("apple"),
			},
			expected: len(" bana"),
		},
		{
			document: &Document{
				Text:           "apply -f /path/to",
				cursorPosition: len("apply -f /path"),
			},
			expected: len("/to"),
			sep:      " /",
		},
		{
			document: &Document{
				Text:           "apple bana",
				cursorPosition: len("ap"),
			},
			expected: len("ple"),
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ",
				cursorPosition: 6,
			},
			expected: len("かきくけこ"),
		},
		{
			document: &Document{
				Text:           "あいうえお かきくけこ",
				cursorPosition: 5,
			},
			expected: len(" かきくけこ"),
		},
		{
			document: &Document{
				Text:           "Добрый день",
				cursorPosition: 6,
			},
			expected: len(" день"),
		},
	}

	for _, p := range pattern {
		if p.sep == "" {
			ac := p.document.FindEndOfCurrentWordWithSpace()
			assert.Equal(t, p.expected, ac)
			ac = p.document.FindEndOfCurrentWordUntilSeparatorIgnoreNextToCursor("")
			assert.Equal(t, p.expected, ac)
		} else {
			ac := p.document.FindEndOfCurrentWordUntilSeparatorIgnoreNextToCursor(p.sep)
			assert.Equal(t, p.expected, ac)
		}
	}
}

func TestDocument_CurrentLineBeforeCursor(t *testing.T) {
	cases := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "lin"),
			},
			expected: "lin",
		},
		{
			document: &Document{
				Text:           "желание # ржавый",
				cursorPosition: 10,
			},
			expected: "желание # ",
		},
		{
			document: &Document{
				Text:           "семнадцать\nрассвет\nпечь",
				cursorPosition: 23,
			},
			expected: "печь",
		},
		{
			document: &Document{
				Text:           "",
				cursorPosition: 0,
			},
			expected: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.document.Text, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.document.CurrentLineBeforeCursor())
		})
	}
}

func TestDocument_CurrentLineAfterCursor(t *testing.T) {
	cases := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "lin"),
			},
			expected: "e 2",
		},
		{
			document: &Document{
				Text:           "желание # ржавый",
				cursorPosition: 10,
			},
			expected: "ржавый",
		},
		{
			document: &Document{
				Text:           "семнадцать\nрассвет\nпечь",
				cursorPosition: 12,
			},
			expected: "ассвет",
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 2,
			},
			expected: "леный",
		},
		{
			document: &Document{
				Text:           "",
				cursorPosition: 0,
			},
			expected: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.document.Text, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.document.CurrentLineAfterCursor())
		})
	}
}

func TestDocument_CurrentLine(t *testing.T) {
	cases := []struct {
		document *Document
		expected string
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "lin"),
			},
			expected: "line 2",
		},
		{
			document: &Document{
				Text:           "желание # ржавый",
				cursorPosition: 10,
			},
			expected: "желание # ржавый",
		},
		{
			document: &Document{
				Text:           "семнадцать\nрассвет\nпечь",
				cursorPosition: 12,
			},
			expected: "рассвет",
		},
		{
			document: &Document{
				Text:           "",
				cursorPosition: 0,
			},
			expected: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.document.Text, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.document.CurrentLine())
		})
	}
}
func TestDocument_CursorPositionRowAndCol(t *testing.T) {
	var cursorPositionTests = []struct {
		document    *Document
		expectedRow int
		expectedCol int
	}{
		{
			document:    &Document{Text: "line 1\nline 2\nline 3\n", cursorPosition: len("line 1\n" + "lin")},
			expectedRow: 1,
			expectedCol: 3,
		},
		{
			document:    &Document{Text: "", cursorPosition: 0},
			expectedRow: 0,
			expectedCol: 0,
		},
		{
			document: &Document{
				Text:           "Однострочник",
				cursorPosition: 12,
			},
			expectedRow: 0,
			expectedCol: 12,
		},
		{
			document: &Document{
				Text:           "Документ\nСтрока",
				cursorPosition: 10, // `т`
			},
			expectedRow: 1,
			expectedCol: 1,
		},
	}
	for _, test := range cursorPositionTests {
		ac := test.document.CursorPositionRow()
		assert.Equal(t, test.expectedRow, ac)
		ac = test.document.CursorPositionCol()
		assert.Equal(t, test.expectedCol, ac)
	}
}

func TestDocument_GetCursorLeftPosition(t *testing.T) {
	cases := []struct {
		document *Document
		shift    []int
		expected []int
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "line 2\n" + "lin"),
			},
			shift:    []int{2, 10},
			expected: []int{-2, -3},
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 18, // `н`
			},
			shift:    []int{1, 7},
			expected: []int{-1, -2},
		},
		{
			document: &Document{
				Text:           "",
				cursorPosition: 0,
			},
			shift:    []int{5},
			expected: []int{0},
		},
	}
	for _, tc := range cases {
		for i, sh := range tc.shift {
			ac := tc.document.GetCursorLeftPosition(sh)
			assert.Equal(t, tc.expected[i], ac)
		}
	}
}

func TestDocument_GetCursorUpPosition(t *testing.T) {
	cases := []struct {
		document *Document
		shift    []int
		column   []int
		expected []int
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "line 2\n" + "lin"),
			},
			shift:    []int{2, 100, 2},
			column:   []int{-1, -1, 0},
			expected: []int{-14, -14, -17},
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 18, // `н`
			},
			shift:    []int{1, 1, 2, 3},
			column:   []int{-1, 0, 2, 0},
			expected: []int{-8, -10, -16, -18},
		},
	}

	for _, tc := range cases {
		for i, sh := range tc.shift {
			ac := tc.document.GetCursorUpPosition(sh, tc.column[i])
			assert.Equal(t, tc.expected[i], ac)
		}
	}
}
func TestDocument_GetCursorDownPosition(t *testing.T) {
	cases := []struct {
		document *Document
		shift    []int
		column   []int
		expected []int
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("lin"),
			},
			shift:    []int{2, 100, 3, 2},
			column:   []int{-1, -1, 2, 4},
			expected: []int{14, 25, 20, 15},
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 2, // `л`
			},
			shift:    []int{1, 1, 2, 3, 4},
			column:   []int{-1, 0, 2, 0, -1},
			expected: []int{8, 6, 16, 19, 19},
		},
	}

	for _, tc := range cases {
		for i, sh := range tc.shift {
			ac := tc.document.GetCursorDownPosition(sh, tc.column[i])
			assert.Equal(t, tc.expected[i], ac)
		}
	}

}

func TestDocument_GetCursorRightPosition(t *testing.T) {
	cases := []struct {
		document *Document
		shift    []int
		expected []int
	}{
		{
			document: &Document{
				Text:           "line 1\nline 2\nline 3\nline 4\n",
				cursorPosition: len("line 1\n" + "line 2\n" + "lin"),
			},
			shift:    []int{2, 10, 3, 2},
			expected: []int{2, 3, 3, 2},
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 2, // `л`
			},
			shift:    []int{1, 5, 9},
			expected: []int{1, 5, 5},
		},
		{
			document: &Document{
				Text:           "зеленый\nкрасный\nсиний",
				cursorPosition: 8, // `к`
			},
			shift:    []int{-1, 3, 8},
			expected: []int{0, 3, 7},
		},
	}

	for _, tc := range cases {
		for i, sh := range tc.shift {
			ac := tc.document.GetCursorRightPosition(sh)
			assert.Equal(t, tc.expected[i], ac)
		}
	}
}

func TestDocument_Lines(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3\nline 4\n",
		cursorPosition: len("line 1\n" + "lin"),
	}
	ac := d.Lines()
	ex := []string{"line 1", "line 2", "line 3", "line 4", ""}
	assert.Equal(t, ex, ac)
	d = &Document{
		Text:           "зеленый\nкрасный\nсиний\nжелтый",
		cursorPosition: 13,
	}
	ac = d.Lines()
	ex = []string{"зеленый", "красный", "синий", "желтый"}
	assert.Equal(t, ex, ac)
}

func TestDocument_LineCount(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3\nline 4\n",
		cursorPosition: len("line 1\n" + "lin"),
	}
	ac := d.LineCount()
	ex := 5
	assert.Equal(t, ex, ac)
	d = &Document{
		Text:           "зеленый\nкрасный\nсиний\nжелтый",
		cursorPosition: 13,
	}
	ac = d.LineCount()
	ex = 4
	assert.Equal(t, ex, ac)
}

func TestDocument_TranslateIndexToPosition(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3\nline 4\n",
		cursorPosition: len("line 1\n" + "lin"),
	}
	row, col := d.TranslateIndexToPosition(len("line 1\nline 2\nlin"))
	assert.Equal(t, 2, row)
	assert.Equal(t, 3, col)
	row, col = d.TranslateIndexToPosition(0)
	assert.Equal(t, 0, row)
	assert.Equal(t, 0, col)
}

func TestDocument_TranslateRowColToIndex(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3\nline 4\n",
		cursorPosition: len("line 1\n" + "lin"),
	}
	ac := d.TranslateRowColToIndex(2, 3)
	ex := len("line 1\nline 2\nlin")
	assert.Equal(t, ex, ac)
	ac = d.TranslateRowColToIndex(0, 0)
	ex = 0
	assert.Equal(t, ex, ac)
}

func TestDocument_TranslateRowColToCursor(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3\nline 4\n",
		cursorPosition: len("line 1\n" + "lin"),
	}
	ac := d.TranslateRowColToCursor(2, 3)
	ex := len("line 1\nline 2\nlin")
	assert.Equal(t, ex, ac)
	ac = d.TranslateRowColToCursor(0, 0)
	ex = 0
	assert.Equal(t, ex, ac)

	d = &Document{
		Text:           "строка 1\nстрока 2\nстрока 3",
		cursorPosition: 5,
	}
	assert.Equal(t, 21, d.TranslateRowColToCursor(2, 3))
	assert.Equal(t, 9, d.TranslateRowColToCursor(1, 0))
	assert.Equal(t, 4, d.TranslateRowColToCursor(0, 4))
}

func TestDocument_OnLastLine(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3",
		cursorPosition: len("line 1\nline"),
	}
	ac := d.OnLastLine()
	assert.Equal(t, false, ac)
	d.cursorPosition = len("line 1\nline 2\nline")
	ac = d.OnLastLine()
	assert.Equal(t, true, ac)
}

func TestDocument_GetEndOfLinePosition(t *testing.T) {
	d := &Document{
		Text:           "line 1\nline 2\nline 3",
		cursorPosition: len("line 1\nli"),
	}
	ac := d.GetEndOfLinePosition()
	ex := len("ne 2")
	assert.Equal(t, ex, ac)
}

func Test_getCursorIndex(t *testing.T) {
	cases := []struct {
		input    string
		expected int
	}{
		{"abcdef", 6},
		{"line1\nline2", 11},
		{"строка 1\nстрока 2", 29},
		{"line1\nline2\nlonglongline", 24},
		{"", 0},
		{"аба", 6},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			d := &Document{
				Text:           tc.input,
				cursorPosition: len([]rune(tc.input)),
			}
			assert.Equal(t, tc.expected, d.getCursorIndex())
		})
	}
}

func Test_GetCursorPosition(t *testing.T) {
	cases := []struct {
		input       string
		expectedRow int
		expectedCol int
	}{
		{"abcdef", 0, 6},
		{"line1\nline2", 1, 5},
		{"строка 1\nстрока 2", 1, 8},
		{"line1\nline2\nlonglongline", 2, 12},
		{"", 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			d := &Document{
				Text:           tc.input,
				cursorPosition: len([]rune(tc.input)),
			}
			row, col := d.GetCursorPosition()
			assert.Equal(t, tc.expectedRow, row)
			assert.Equal(t, tc.expectedCol, col)
		})
	}
}

func Test_GetCustomCursorPosition(t *testing.T) {
	cases := []struct {
		input       string
		position    int
		expectedRow int
		expectedCol int
	}{
		{"abcdef", 0, 0, 0},
		{"abcdef", 1, 0, 1},
		{"line1\nline2", 2, 0, 2},
		{"line1\nline2\nlonglongline", 12, 2, 0},
		{"line1\nline2\nlonglongline", 23, 2, 11},
		{"", 0, 0, 0},
		{"строка1\nстрока2\nстрока3", 9, 1, 1},
		{"строка1\nстрока2\nстрока3", 16, 2, 0},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			d := &Document{
				Text:           tc.input,
				cursorPosition: len([]rune(tc.input)),
			}
			row, col := d.GetCustomCursorPosition(tc.position)
			assert.Equal(t, tc.expectedRow, row)
			assert.Equal(t, tc.expectedCol, col)
		})
	}
}
