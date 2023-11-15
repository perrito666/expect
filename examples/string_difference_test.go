package examples

import (
	"testing"

	"perri.to/expect"
	"perri.to/expect/snapshots/comparabletypes"
)

func TestStringSimpleMoreTextFails(t *testing.T) {
	moreTestThan := `Hello, World
Hello Universe`
	c := comparabletypes.NewStringComparable(moreTestThan)
	expect.FromSnapshot(t, "comparable has more text than snapshot", c)
}

func TestStringSimpleLessTextFails(t *testing.T) {
	moreTestThan := `Hello World`
	c := comparabletypes.NewStringComparable(moreTestThan)
	expect.FromSnapshot(t, "comparable has less text than snapshot", c)
}

func TestPrettyStringSimpleMoreTextFails(t *testing.T) {
	moreTestThan := `Hello, World
Hello Universe`
	c := comparabletypes.NewPrettyStringComparable(moreTestThan)
	expect.FromSnapshot(t, "pretty comparable has more text than snapshot", c)
}

func TestPrettyStringSimpleLessTextFails(t *testing.T) {
	moreTestThan := `Hello World`
	c := comparabletypes.NewPrettyStringComparable(moreTestThan)
	expect.FromSnapshot(t, "pretty comparable has less text than snapshot", c)
}
