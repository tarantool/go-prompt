package prompt

import (
	"strings"
)

const (
	matchSearchPrefixFmt = "(reverse-i-search)`%s':"
	failSearchPrefixFmt  = "(failed reverse-i-search)`%s':"
)

// reverseSearchState contains info about reverseSearch state.
type reverseSearchState struct {
	// history, with which reverse-search works.
	history *History
	// searchFromIndex is the last history index, included to the search scope.
	searchFromIndex int
	// matchedIndex is the index of the last matched.
	matchedIndex int
	// matchedCmd is the last matched command.
	matchedCmd string
}

// NewReverseSearch returns new reverseSearchState instance.
func NewReverseSearch(history *History) *reverseSearchState {
	return &reverseSearchState{
		history:         history,
		searchFromIndex: history.selected,
		matchedIndex:    -1,
		matchedCmd:      "",
	}
}

// reducePrefix reduces the search scope.
func (rs *reverseSearchState) reducePrefix() {
	if rs.matchedIndex != -1 {
		rs.searchFromIndex = rs.matchedIndex - 1
	}
	if rs.searchFromIndex < 0 {
		rs.searchFromIndex = 0
	}
}

// update updates current reverse-search state.
func (rs *reverseSearchState) update(input string) {
	rs.matchedIndex = rs.history.FindMatch(
		strings.TrimSpace(input), rs.searchFromIndex,
	)
	if rs.matchedIndex != -1 {
		rs.matchedCmd = rs.history.tmp[rs.matchedIndex]
	} else {
		rs.matchedCmd = ""
	}
}
