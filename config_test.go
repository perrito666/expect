package expect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"perri.to/expect/snapshots"
	"perri.to/expect/snapshots/comparabletypes"
)

func TestReadConfigNoFile(t *testing.T) {
	d := t.TempDir()
	c, err := readConfig(func() (string, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}
	eq := reflect.DeepEqual(c, &Config{
		Grouping:    "",
		SnapShotDir: "",
		Replacers:   map[snapshots.Kind]map[string]string{},
	})
	if !eq {
		t.Logf("returned configuration is not empty: %#v", c)
		t.FailNow()
	}
}

func TestReadConfig(t *testing.T) {
	d := t.TempDir()
	expectedConfig := &Config{
		Grouping:    groupByTestFile,
		SnapShotDir: "some_cool_name",
		Replacers: map[snapshots.Kind]map[string]string{
			comparabletypes.KindJSON:   {"Date": "replaced-date"},
			comparabletypes.KindString: {"Time": "replaced-time"},
		},
	}
	m, err := json.Marshal(expectedConfig)
	if err != nil {
		t.Fatal(fmt.Errorf("marshaling sample config: %w", err))
	}
	err = os.WriteFile(filepath.Join(d, configFileName), m, snapshotFilePerm)
	if err != nil {
		t.Fatal(fmt.Errorf("writing sample config: %w", err))
	}
	c, err := readConfig(func() (string, error) { return d, nil })
	if err != nil {
		t.Fatal(err)
	}

	eq := reflect.DeepEqual(c, expectedConfig)
	if !eq {
		t.Logf("returned configuration is not the one we stored: %#v", c)
		t.FailNow()
	}
}

func TestConfig_SnapshotDir(t *testing.T) {
	c := &Config{
		Grouping:    groupByTestFile,
		SnapShotDir: "some_cool_name",
		Replacers: map[snapshots.Kind]map[string]string{
			comparabletypes.KindJSON:   {"Date": "replaced-date"},
			comparabletypes.KindString: {"Time": "replaced-time"},
		},
	}
	f := c.SnapshotDir("config_test.go")
	exp := "some_cool_name"
	if f != exp {
		t.Logf("expected %q but got %q", exp, f)
	}

	c.SnapShotDir = ""
	f = c.SnapshotDir("config_test.go")
	exp = "config_test.expectations"
	if f != exp {
		t.Logf("expected %q but got %q", exp, f)
	}

	c.Grouping = groupByPackage
	f = c.SnapshotDir("config_test.go")
	exp = "TestExpectationsSnapshots"
	if f != exp {
		t.Logf("expected %q but got %q", exp, f)
	}
}
