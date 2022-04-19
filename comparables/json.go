package comparables

import (
	"encoding/json"
	"fmt"

	"github.com/nsf/jsondiff"

	"expect"
)

var _ expect.Comparable = (*JSON)(nil)

type JSON struct {
	rawJSON json.RawMessage
}

func (j JSON) CompareTo(c expect.Comparable) (string, error) {
	opts := jsondiff.DefaultConsoleOptions()

	// TODO: Take string here too, we should be able to handle it most of the time
	newJSON, isJSON := c.(JSON)
	if !isJSON {
		return "", expect.CantCompare(fmt.Sprintf("%T", j), fmt.Sprintf("%T", c))
	}

	jsonDifference, explanation := jsondiff.Compare(j.rawJSON, newJSON.rawJSON, &opts)
	if jsonDifference == jsondiff.FullMatch {
		return "", nil
	}

	switch jsonDifference {
	case jsondiff.SupersetMatch, jsondiff.NoMatch:
		return explanation, nil
	case jsondiff.FirstArgIsInvalidJson:
		return "", expect.InvalidSource(fmt.Sprintf("%T", j), j.Kind())
	case jsondiff.SecondArgIsInvalidJson:
		return "", expect.InvalidTarget(fmt.Sprintf("%T", c), c.Kind())
	case jsondiff.BothArgsAreInvalidJson:
		if len(j.rawJSON) == len(newJSON.rawJSON) && len(j.rawJSON) == 0 {
			// empty therefore equal
			return "", nil
		}
		return "", expect.BothPartsInvalid(fmt.Sprintf("%T", j), fmt.Sprintf("%T", c), j.Kind())

	}
	return explanation, nil
}

func (j JSON) String() string {
	return string(j.rawJSON)
}

const KindJSON expect.Kind = "json"

func (j JSON) Kind() expect.Kind {
	return KindJSON
}

func (j JSON) Dump() []byte {
	return j.rawJSON
}

func (j JSON) Load(rawJSON []byte) expect.Comparable {
	return JSON{rawJSON: rawJSON}
}
