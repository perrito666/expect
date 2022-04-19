package comparables

import (
	"expect"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var _ expect.Comparable = (*StringComparable)(nil)

type StringComparable string
type PrettyStringComparable struct {
	StringComparable
}

func (s StringComparable) CompareTo(c expect.Comparable) (string, error) {
	return s.compareTo(c, false)
}
func (s PrettyStringComparable) CompareTo(c expect.Comparable) (string, error) {
	return s.compareTo(c, true)
}

func (s StringComparable) compareTo(c expect.Comparable, pretty bool) (string, error) {
	otherStr := c.String()
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(s), otherStr, false)
	if pretty {
		prettyDiff := dmp.DiffPrettyText(diffs)
		if prettyDiff == string(s) {
			return "", nil
		}
		return prettyDiff, nil
	}
	patches := dmp.PatchMake(diffs)
	return dmp.PatchToText(patches), nil
}

func (s StringComparable) String() string {
	return string(s)
}

const KindString expect.Kind = "string"

func (s StringComparable) Kind() expect.Kind {
	return KindString
}

func (s StringComparable) Dump() []byte {
	return []byte(s)
}

func (s StringComparable) Load(rawS []byte) expect.Comparable {
	return StringComparable(rawS)
}
