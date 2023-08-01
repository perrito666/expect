package comparabletypes

import (
	"bytes"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"perri.to/expect/snapshots"
)

const DefaultContextSize = 3

var _ snapshots.Comparable = (*StringComparable)(nil)

type StringComparable struct {
	string
	// how many lines of context we want to print before and after a difference, -1 to turn off
	contextSize int
}
type PrettyStringComparable struct {
	StringComparable
}

// NewStringComparable constructs a StringComparable from a string.
func NewStringComparable(s string) snapshots.Comparable {
	sc := StringComparable{s, DefaultContextSize}
	return &sc
}

// NewPrettyStringComparable constructs a PrettyStringComparable from a string.
func NewPrettyStringComparable(s string) snapshots.Comparable {
	sc := StringComparable{s, DefaultContextSize}
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
	diffs := dmp.DiffMain(s.string, otherStr, false)

	var buffer bytes.Buffer
	for j, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			if s.contextSize == -1 || !strings.ContainsRune(diff.Text, '\n') {
				buffer.WriteString(diff.Text)
			} else {
				hasLastNewLine := diff.Text[len(diff.Text)-1] == '\n'
				lines := strings.Split(diff.Text, "\n")
				l := len(lines)
				lower, upper := s.contextSize, l - s.contextSize
				// if there's no newline at the end, we need to take one more line, since the last one will
				// immediately be followed by an edit, so it doesn't really count as context
				if !hasLastNewLine {
					upper -= 1
				}
				if lower >= upper {
					// print everything
					lower, upper = -1, 0
				}
				// no context before this, we skip printing upper context
				if j != 0 {
					for _, line := range lines[:lower+1] {
						buffer.WriteString(line)
						buffer.WriteRune('\n')
					}
				}
				if lower != -1 {
					if pretty {
						buffer.WriteString("\x1b[33m[...]\x1b[0m\n")
					} else {
						buffer.WriteString("{=...=}\n")
					}
				}
				// no context after this, we skip printing lower context
				if j != len(diffs)-1 {
					last := l - upper
					for i, line := range lines[upper:] {
						buffer.WriteString(line)
						if i != last-1 {
							buffer.WriteRune('\n')
						}
					}
					if hasLastNewLine {
						buffer.WriteRune('\n')
					}
				}
			}
		case diffmatchpatch.DiffDelete:
			if pretty {
				buffer.WriteString("\x1b[31m")
			} else {
				buffer.WriteString("{-")
			}
			buffer.WriteString(diff.Text)
			if pretty {
				buffer.WriteString("\x1b[0m")
			} else {
				buffer.WriteString("-}")
			}
		case diffmatchpatch.DiffInsert:
			if pretty {
				buffer.WriteString("\x1b[32m")
			} else {
				buffer.WriteString("{+")
			}
			buffer.WriteString(diff.Text)
			if pretty {
				buffer.WriteString("\x1b[0m")
			} else {
				buffer.WriteString("+}")
			}
		}
	}
	return buffer.String(), nil
}

func (s *StringComparable) String() string {
	return s.string
}

const KindString snapshots.Kind = "string"

func (s *StringComparable) Kind() snapshots.Kind {
	return KindString
}

func (s *StringComparable) Dump() []byte {
	return []byte(s.string)
}

func (s *StringComparable) Load(rawS []byte) snapshots.Comparable {
	sc := StringComparable{string(rawS), s.contextSize}
	return &sc
}

func (s *PrettyStringComparable) Load(rawS []byte) snapshots.Comparable {
	sc := StringComparable{string(rawS), s.contextSize}
	psc := PrettyStringComparable{sc}
	return &psc
}

func (s *StringComparable) Replace(r map[string]string) {
	// TODO: Should this support regexps? how?
	repl := make([]string, 0, len(r)*2)
	for k, v := range r {
		repl = append(repl, k, v)
	}
	rs := strings.NewReplacer(repl...).Replace(s.string)
	*s = StringComparable{rs, s.contextSize}
}

func (s *StringComparable) Extension() string {
	return "txt"
}
