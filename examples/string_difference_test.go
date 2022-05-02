package examples

import (
	"testing"

	"expect"
	"expect/snapshots/comparabletypes"
)

func TestStringSimpleMoreTextFails(t *testing.T) {
	moreTestThan := `Hello, World
Hello Universe`
	c := comparabletypes.StringComparable(moreTestThan)
	expect.FromSnapshot(t, "comparable has more text than snapshot", &c)
}

func TestStringSimpleLessTextFails(t *testing.T) {
	moreTestThan := `Hello World`
	c := comparabletypes.StringComparable(moreTestThan)
	expect.FromSnapshot(t, "comparable has less text than snapshot", &c)
}

func TestPrettyStringSimpleMoreTextFails(t *testing.T) {
	moreTestThan := `Hello, World
Hello Universe`
	c := comparabletypes.PrettyStringComparable{StringComparable: comparabletypes.StringComparable(moreTestThan)}
	expect.FromSnapshot(t, "pretty comparable has more text than snapshot", &c)
}

func TestPrettyStringSimpleLessTextFails(t *testing.T) {
	moreTestThan := `Hello World`
	c := comparabletypes.PrettyStringComparable{StringComparable: comparabletypes.StringComparable(moreTestThan)}
	expect.FromSnapshot(t, "pretty comparable has less text than snapshot", &c)
}
