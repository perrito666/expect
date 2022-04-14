package expect

// Kind is used to represent a kind of comparable, ideally is used to know if two Comparables
// can compare themselves with custom method, or they need string comparison
type Kind string

type Comparable interface {
	CompareTo(Comparable) string
	String() string
	Kind() Kind
}
