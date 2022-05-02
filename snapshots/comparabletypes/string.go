package comparabletypes

import (
	"strings"

	"expect/snapshots"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var _ snapshots.Comparable = (*StringComparable)(nil)

type StringComparable string
type PrettyStringComparable struct {
	StringComparable
}

// NewStringComparable constructs a StringComparable from a string.
func NewStringComparable(s string) snapshots.Comparable {
	sc := StringComparable(s)
	return &sc
}

// NewPrettyStringComparable constructs a PrettyStringComparable from a string.
func NewPrettyStringComparable(s string) snapshots.Comparable {
	sc := StringComparable(s)
	return &PrettyStringComparable{sc}
}

func (s *StringComparable) CompareTo(c snapshots.Comparable) (string, error) {
	return s.compareTo(c, false)
}
func (s *PrettyStringComparable) CompareTo(c snapshots.Comparable) (string, error) {
	return s.compareTo(c, true)
}

func (s *StringComparable) compareTo(c snapshots.Comparable, pretty bool) (string, error) {
	otherStr := c.String()
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(*s), otherStr, false)
	if pretty {
		prettyDiff := dmp.DiffPrettyText(diffs)
		if prettyDiff == string(*s) {
			return "", nil
		}
		return prettyDiff, nil
	}
	patches := dmp.PatchMake(diffs)
	return dmp.PatchToText(patches), nil
}

func (s *StringComparable) String() string {
	return string(*s)
}

const KindString snapshots.Kind = "string"

func (s *StringComparable) Kind() snapshots.Kind {
	return KindString
}

func (s *StringComparable) Dump() []byte {
	return []byte(*s)
}

func (s *StringComparable) Load(rawS []byte) snapshots.Comparable {
	sc := StringComparable(rawS)
	return &sc
}

func (s *PrettyStringComparable) Load(rawS []byte) snapshots.Comparable {
	sc := StringComparable(rawS)
	psc := PrettyStringComparable{sc}
	return &psc
}

func (s *StringComparable) Replace(r map[string]string) {
	repl := make([]string, 0, len(r)*2)
	for k, v := range r {
		repl = append(repl, k, v)
	}
	rs := strings.NewReplacer(repl...).Replace(string(*s))
	*s = StringComparable(rs)
}

func (s *StringComparable) Extension() string {
	return "txt"
}
