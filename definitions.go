package expect

import "fmt"

// Kind is used to represent a kind of comparable, ideally is used to know if two Comparables
// can compare themselves with custom method, or they need string comparison
type Kind string

type Comparable interface {
	CompareTo(Comparable) (string, error)
	String() string
	Kind() Kind
	Dump() []byte
	Load([]byte) Comparable
}

// CantCompare constructs a valid ErrCantCompare
func CantCompare(source, target string) error {
	return &ErrCantCompare{
		Source: source,
		Target: target,
	}
}

// ErrCantCompare should be raised when Source is passed Target to compare itself to.
type ErrCantCompare struct {
	Source string
	Target string
}

func (err *ErrCantCompare) Error() string {
	return fmt.Sprintf("%q does not know how to compare itself to %q", err.Source, err.Target)
}

// InvalidSource constructs a valid ErrSourceInvalid
func InvalidSource(srcType string, srcKind Kind) error {
	return &ErrSourceInvalid{
		Type: srcType,
		Kind: srcKind,
	}
}

// ErrSourceInvalid should be returned when the source of the comparison is not of the valid type.
type ErrSourceInvalid struct {
	Type string
	Kind Kind
}

func (err *ErrSourceInvalid) Error() string {
	return fmt.Sprintf("source of type %s is not a valid %s", err.Type, err.Kind)
}

// InvalidTarget constructs a valid ErrTargetInvalid
func InvalidTarget(tgtType string, tgtKind Kind) error {
	return &ErrTargetInvalid{
		Type: tgtType,
		Kind: tgtKind,
	}
}

type ErrTargetInvalid struct {
	Type string
	Kind Kind
}

func (err *ErrTargetInvalid) Error() string {
	return fmt.Sprintf("target of type %s is not a valid %s", err.Type, err.Kind)
}

type ErrBothInvalid struct {
	Source string
	Target string
	Kind   Kind
}

func BothPartsInvalid(source, target string, kind Kind) error {
	return &ErrBothInvalid{
		Source: source,
		Target: target,
		Kind:   kind,
	}
}

func (err *ErrBothInvalid) Error() string {
	return fmt.Sprintf("neither source, of type %s nor targetm of type %s are valid %s",
		err.Source, err.Target, err.Kind)
}
