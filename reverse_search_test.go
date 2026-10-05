package prompt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchPrefixFmt(t *testing.T) {
	// The space after ':' is required to match the GNU readline output.
	assert.Equal(t, "(reverse-i-search)`abc': ", fmt.Sprintf(matchSearchPrefixFmt, "abc"))
	assert.Equal(t, "(failed reverse-i-search)`abc': ", fmt.Sprintf(failSearchPrefixFmt, "abc"))
}

func TestRevSearch(t *testing.T) {
	history := NewHistory()
	history.Add("aa")
	history.Add("ab")
	history.Add("ac")
	history.Add("ad")
	history.Add("ae")

	revSearchState := NewReverseSearch(history)

	assert.Equal(t, len(history.tmp)-1, revSearchState.searchFromIndex)
	assert.Equal(t, "", revSearchState.matchedCmd)
	assert.Equal(t, -1, revSearchState.matchedIndex)

	// Update 1.
	revSearchState.update("a")
	assert.Equal(t, "ae", revSearchState.matchedCmd)
	assert.Equal(t, 4, revSearchState.matchedIndex)
	assert.Equal(t, 5, revSearchState.searchFromIndex)

	// Activate again.
	revSearchState.reducePrefix()
	assert.Equal(t, 3, revSearchState.searchFromIndex)

	// Update 2.
	revSearchState.update("a")
	assert.Equal(t, "ad", revSearchState.matchedCmd)
	assert.Equal(t, 3, revSearchState.matchedIndex)

	// Accurate update.
	revSearchState.update("aa")
	assert.Equal(t, "aa", revSearchState.matchedCmd)
	assert.Equal(t, 0, revSearchState.matchedIndex)

	// Activate too many times.
	for i := 0; i < 10; i++ {
		revSearchState.reducePrefix()
	}
	assert.Equal(t, 0, revSearchState.searchFromIndex)
}
