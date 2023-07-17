// Example application to run integration tests.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
)

// Console describes the console.
type Console struct {
	title             string
	input             string
	prefix            string
	livePrefix        string
	livePrefixFunc    func() (string, bool)
	livePrefixEnabled bool
	prompt            *prompt.Prompt
}

// NewConsole creates a new Console instance.
func NewConsole() *Console {
	console := Console{}
	console.title = "prompt_app"
	console.prefix = "prompt_app> "
	console.livePrefix = "> "
	console.livePrefixFunc = func() (string, bool) {
		return console.livePrefix, console.livePrefixEnabled
	}
	console.prompt = prompt.New(getExecutor(&console),
		completer, getPromptOptions(&console)...)
	return &console
}

// addStmt merges two commands.
// For multi-lines commands use `#` at the beginning and at the end of the command.
func addStmt(have string, input string) (string, bool) {
	isInputFinish := len(input) > 0 && input[len(input)-1] == '#'
	if len(have) == 0 {
		if len(input) > 0 {
			return input, input[0] != '#'
		}
		return input, true
	}
	isMultiline := have[0] == '#'
	isCompleted := isInputFinish
	return have + "\n" + input, !isMultiline || isCompleted
}

// getExecutor creates an executor.
func getExecutor(console *Console) prompt.Executor {
	return func(in string) {
		if in == "exit" {
			fmt.Println()
			os.Exit(0)
		}
		var completed bool
		console.input, completed = addStmt(console.input, in)
		if !completed {
			console.livePrefixEnabled = true
			return
		}
		fmt.Printf("cmd: %s\n", console.input)
		console.prompt.PushToHistory(console.input)
		console.input = ""
		console.livePrefixEnabled = false
	}
}

// getPromptOptions returns prompt options.
func getPromptOptions(console *Console) []prompt.Option {
	options := []prompt.Option{
		prompt.OptionTitle(console.title),
		prompt.OptionPrefix(console.prefix),
		prompt.OptionLivePrefix(console.livePrefixFunc),

		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionPreviewSuggestionTextColor(prompt.DefaultColor),

		prompt.OptionDisableAutoHistory(),
		prompt.OptionReverseSearch(),
	}
	args := os.Args
	if len(args) > 1 {
		histories := strings.Split(args[1], ";")
		options = append(options, prompt.OptionHistory(histories))
	}
	if len(args) > 2 {
		socketUri := args[2]
		options = append(options, prompt.OptionEnableRenderSubscribeMode(socketUri))
	}
	return options
}

// completer is function used for completion in the prompt.
func completer(d prompt.Document) []prompt.Suggest {
	if len(d.Text) == 0 {
		return []prompt.Suggest{}
	}
	words := []string{
		"abc",
		"aad",
		"aba",
		"aart",
		"apple",
		"git",
		"gggit",
		"gist",
	}
	result := make([]prompt.Suggest, 0)
	for _, word := range words {
		if strings.HasPrefix(word, d.Text) {
			result = append(result, prompt.Suggest{Text: word})
		}
	}
	return result
}

func main() {
	console := NewConsole()
	console.prompt.Run()
}
