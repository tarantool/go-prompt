package prompt

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoryClear(t *testing.T) {
	h := NewHistory()
	h.Add("foo")
	h.Clear()
	expected := &History{
		histories: []string{"foo"},
		tmp:       []string{"foo", ""},
		selected:  1,
	}
	if !reflect.DeepEqual(expected, h) {
		t.Errorf("Should be %#v, but got %#v", expected, h)
	}
}

func TestHistoryAdd(t *testing.T) {
	h := NewHistory()
	h.Add("echo 1")
	expected := &History{
		histories: []string{"echo 1"},
		tmp:       []string{"echo 1", ""},
		selected:  1,
	}
	if !reflect.DeepEqual(h, expected) {
		t.Errorf("Should be %v, but got %v", expected, h)
	}
}

func TestHistoryOlder(t *testing.T) {
	h := NewHistory()
	h.Add("echo 1")

	// Prepare buffer
	buf := NewBuffer()
	buf.InsertText("echo 2", false, true)

	// [1 time] Call Older function
	buf1, changed := h.Older(buf)
	if !changed {
		t.Error("Should be changed history but not changed.")
	}
	if buf1.Text() != "echo 1" {
		t.Errorf("Should be %#v, but got %#v", "echo 1", buf1.Text())
	}

	// [2 times] Call Older function
	buf = NewBuffer()
	buf.InsertText("echo 1", false, true)
	buf2, changed := h.Older(buf)
	if changed {
		t.Error("Should be not changed history but changed.")
	}
	if !reflect.DeepEqual("echo 1", buf2.Text()) {
		t.Errorf("Should be %#v, but got %#v", "echo 1", buf2.Text())
	}
}

func TestHistorySetCurrent(t *testing.T) {
	history := NewHistory()
	history.Add("entry0")
	history.Add("entry1")
	history.Add("entry2")

	newCmd := "alternative"
	history.SetCurrentCmd(newCmd)
	assert.Equal(t, newCmd, history.tmp[history.selected])
}

func TestHistoryFindMatch(t *testing.T) {
	history := NewHistory()
	history.Add("line 0")
	history.Add("cmd1")
	history.Add("cmd2")
	history.Add("")
	history.Add("x")
	history.SetCurrentCmd("echo")

	assert.Equal(t, 5, history.FindMatch("", 5))
	assert.Equal(t, 5, history.FindMatch("c", 5))
	assert.Equal(t, 2, history.FindMatch("cmd", 2))
	assert.Equal(t, 1, history.FindMatch("cmd1", 2))
	assert.Equal(t, 0, history.FindMatch("line", 5))
	assert.Equal(t, -1, history.FindMatch("line 10", 5))
}
