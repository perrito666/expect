package snapshots

import "fmt"

// Kind is used to represent a kind of comparable types, ideally is used to know if two Comparables
// can compare themselves with custom method, or they need string comparison
type Kind string

// Comparable represents a type that can be compared with another that fulfills the same interface
// within reason. Mist types should implement a fallback to string compare.
type Comparable interface {
	CompareTo(Comparable) (string, error)
	String() string
	Kind() Kind
	Dump() []byte
	Load([]byte) Comparable
	Replace(map[string]string)
	Extension() string
	Subtypes() bool
	ReplaceSubtypes(map[Kind]map[string]string)
}

// CantCompare constructs a valid ErrCannotCompare
func CantCompare(source, target string) error {
	return &ErrCannotCompare{
		Source: source,
		Target: target,
	}
}

// ErrCannotCompare should be raised when Source is passed Target to compare itself to.
type ErrCannotCompare struct {
	Source string
	Target string
}

// Error implements error for ErrCannotCompare.
func (err *ErrCannotCompare) Error() string {
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

// ErrTargetInvalid should be returned when one of the comparables is not valid (meaning we really
// do not know how to compare it)
type ErrTargetInvalid struct {
	Type string
	Kind Kind
}

// Error implements error for ErrTargetInvalid.
func (err *ErrTargetInvalid) Error() string {
	return fmt.Sprintf("target of type %s is not a valid %s", err.Type, err.Kind)
}

// ErrBothInvalid should be returned when both comparables are invalid.
type ErrBothInvalid struct {
	Source string
	Target string
	Kind   Kind
}

// BothPartsInvalid returns an ErrBothInvalid instance.
func BothPartsInvalid(source, target string, kind Kind) error {
	return &ErrBothInvalid{
		Source: source,
		Target: target,
		Kind:   kind,
	}
}

// Error implements error for ErrBothInvalid.
func (err *ErrBothInvalid) Error() string {
	return fmt.Sprintf("neither source, of type %s nor target of type %s are valid %s",
		err.Source, err.Target, err.Kind)
}
