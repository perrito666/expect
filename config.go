package expect

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"perri.to/expect/snapshots"
)

// Grouping represents a possible criteria to group the snapshot files
type Grouping string

const (
	groupByTestFile Grouping = "by_test_file"
	groupByPackage  Grouping = "by_package"

	snapShotDir = "TestExpectationsSnapshots"
)

// Config holds the config we allow for a test suite.
type Config struct {
	Grouping    Grouping `json:"grouping,omitempty"`
	SnapShotDir string   `json:"snapshot_dir,omitempty"`
	// Replacers holds possible kw replacement as map[comparabletypes.Kind]map[from]to
	Replacers map[snapshots.Kind]map[string]string `json:"replacers,omitempty"`
}

const configFileName = "expectations.json"

// ReadConfig will try to read a config file from cwd and return that or a sane default.
func ReadConfig() (*Config, error) {
	for i := 2; i < 5; i++ {
		_, fileCalling, _, ok := runtime.Caller(1)
		if ok && strings.HasSuffix(fileCalling, "_test.go") {
			return readConfig(func() (string, error) { return path.Dir(fileCalling), nil })
		}
		// if we could not find a file calling OR the file is oddly empty OR the file is just the file name.
		if !ok || fileCalling == "" || !strings.ContainsRune(fileCalling, os.PathSeparator) {
			return readConfig(os.Getwd)
		}
	}
	return readConfig(os.Getwd)
}

func readConfig(getcwd func() (string, error)) (*Config, error) {
	wd, err := getcwd()
	if err != nil {
		// I have no clue why Getwd() would fail in this context
		return nil, fmt.Errorf("determining test working directory: %w", err)
	}
	cFilePath := filepath.Join(wd, configFileName)
	if _, err := os.Stat(cFilePath); err != nil {
		return &Config{
			Grouping:    "",
			SnapShotDir: "",
			Replacers:   map[snapshots.Kind]map[string]string{},
		}, nil
	}
	fd, err := os.Open(cFilePath)
	if err != nil {
		return nil, fmt.Errorf("opening expectations configuration file: %w", err)
	}
	defer fd.Close()
	var config Config
	if err = json.NewDecoder(fd).Decode(&config); err != nil {
		return nil, fmt.Errorf("unmarshaling expectations configuration file: %w", err)
	}
	return &config, nil
}

// GroupBy returns the configured (or default) grouping
func (c *Config) GroupBy() Grouping {
	if c.Grouping != "" {
		return c.Grouping
	}
	return groupByPackage
}

// SnapshotDir will return either:
// * The user configured snapshot directory
// * The default directory name we use for snapshots
// * In case "per file" snapshots are chosen: File_test.expectations
func (c *Config) SnapshotDir(fileName string) string {
	if c.SnapShotDir != "" {
		return c.SnapShotDir
	}

	if c.Grouping == groupByTestFile {
		// Remove . in case there is no . ... cthulu knows why
		snapName := strings.TrimRight(fileName, "."+filepath.Ext(fileName))
		return fmt.Sprintf("%s.expectations", snapName)
	}
	return snapShotDir
}
