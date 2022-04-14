package expect

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

var _ Comparable = (*StringComparable)(nil)

type StringComparable string
type PrettyStringComparable struct {
	StringComparable
}

func (s StringComparable) CompareTo(c Comparable) string {
	return s.compareTo(c, false)
}
func (s PrettyStringComparable) CompareTo(c Comparable) string {
	return s.compareTo(c, true)
}

func (s StringComparable) compareTo(c Comparable, pretty bool) string {
	otherStr := c.String()
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(s), otherStr, false)
	if pretty {
		prettyDiff := dmp.DiffPrettyText(diffs)
		if prettyDiff == string(s) {
			return ""
		}
		return prettyDiff
	}
	patches := dmp.PatchMake(diffs)
	return dmp.PatchToText(patches)
}

func (s StringComparable) String() string {
	return string(s)
}

const KindString Kind = "string"

func (s StringComparable) Kind() Kind {
	return KindString
}
