package comparabletypes

import (
	"testing"

	"expect/snapshots"
)

func newStringComparableFromLiteral(s string) *StringComparable {
	c := StringComparable(s)
	return &c
}

func TestStringComparable_CompareTo(t *testing.T) {
	type args struct {
		c snapshots.Comparable
	}
	var tests = []struct {
		name string
		s    *StringComparable
		args args
		want string
	}{{
		name: "all equal",
		s:    newStringComparableFromLiteral("Lorem ipsum dolor."),
		args: args{c: newStringComparableFromLiteral("Lorem ipsum dolor.")},
		want: "",
	},
		{
			name: "half different",
			s:    newStringComparableFromLiteral("Lorem ipsum dolor."),
			args: args{c: newStringComparableFromLiteral("Lorem dolor sit amet.")},
			want: "@@ -3,16 +3,19 @@\n rem \n-ipsum \n dolor\n+ sit amet\n .\n",
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.CompareTo(tt.args.c)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("CompareTo() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrettyStringComparable_CompareTo(t *testing.T) {
	type args struct {
		c snapshots.Comparable
	}
	var tests = []struct {
		name string
		s    *PrettyStringComparable
		args args
		want string
	}{{
		name: "all equal",
		s:    &PrettyStringComparable{"Lorem ipsum dolor."},
		args: args{c: newStringComparableFromLiteral("Lorem ipsum dolor.")},
		want: "",
	},
		{
			name: "half different",
			s:    &PrettyStringComparable{"Lorem ipsum dolor."},
			args: args{c: &PrettyStringComparable{StringComparable("Lorem dolor sit amet.")}},
			want: "Lorem \x1b[31mipsum \x1b[0mdolor\x1b[32m sit amet\x1b[0m.",
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.CompareTo(tt.args.c)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("CompareTo() = %q, want %q", got, tt.want)
			}
		})
	}
}
