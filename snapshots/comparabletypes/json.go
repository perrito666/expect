package comparabletypes

import (
	"encoding/json"
	"fmt"

	"github.com/nsf/jsondiff"
	"github.com/tidwall/sjson"

	"perri.to/expect/snapshots"
)

var _ snapshots.Comparable = (*JSON)(nil)

type JSON struct {
	rawJSON json.RawMessage
}

func NewJSONFromString(s string) snapshots.Comparable {
	j := JSON{rawJSON: []byte(s)}
	return &j
}

func NewJSONFromByes(b []byte) snapshots.Comparable {
	j := JSON{rawJSON: b}
	return &j
}

func (j *JSON) Subtypes() bool {
	return false
}

func (j *JSON) ReplaceSubtypes(_ map[snapshots.Kind]map[string]string) {
	return
}

func (j *JSON) CompareTo(c snapshots.Comparable) (string, error) {
	opts := jsondiff.DefaultConsoleOptions()

	// TODO: Take string here too, we should be able to handle it most of the time
	newJSON, isJSON := c.(*JSON)
	if !isJSON {
		return "", snapshots.CantCompare(fmt.Sprintf("%T", j), fmt.Sprintf("%T", c))
	}

	jsonDifference, explanation := jsondiff.Compare(j.rawJSON, newJSON.rawJSON, &opts)
	if jsonDifference == jsondiff.FullMatch {
		return "", nil
	}

	switch jsonDifference {
	case jsondiff.SupersetMatch, jsondiff.NoMatch:
		return explanation, nil
	case jsondiff.FirstArgIsInvalidJson:
		return "", snapshots.InvalidSource(fmt.Sprintf("%T", j), j.Kind())
	case jsondiff.SecondArgIsInvalidJson:
		return "", snapshots.InvalidTarget(fmt.Sprintf("%T", c), c.Kind())
	case jsondiff.BothArgsAreInvalidJson:
		if len(j.rawJSON) == len(newJSON.rawJSON) && len(j.rawJSON) == 0 {
			// empty therefore equal
			return "", nil
		}
		return "", snapshots.BothPartsInvalid(fmt.Sprintf("%T", j), fmt.Sprintf("%T", c), j.Kind())

	}
	return explanation, nil
}

func (j *JSON) String() string {
	return string(j.rawJSON)
}

const KindJSON snapshots.Kind = "json"

func (j *JSON) Kind() snapshots.Kind {
	return KindJSON
}

func (j *JSON) Dump() []byte {
	return j.rawJSON
}

func (j *JSON) Load(rawJSON []byte) snapshots.Comparable {
	return &JSON{rawJSON: rawJSON}
}

func (j *JSON) Replace(rs map[string]string) {
	var err error
	for k, v := range rs {
		j.rawJSON, err = sjson.SetBytes(j.rawJSON, k, v)
		if err != nil {
			panic(err)
		}

	}
}

func (j *JSON) Extension() string {
	return "json"
}
