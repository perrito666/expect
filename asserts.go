package expect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

const snapShotDir = "__expectations"

type Args struct {
	shouldUpdate  bool
	shouldCleanup bool
}

var currentRunArgs *Args
var registeredName map[string]bool
var registerNameMutex sync.Mutex

func init() {
	var args Args
	for _, arg := range os.Args {
		argList := strings.Split(arg, "=")
		if len(argList) > 0 {
			arg = argList[0]
		}
		switch arg {
		case "-u":
			args.shouldUpdate = true
		case "-cleanup":
			args.shouldCleanup = true
		}
	}
	currentRunArgs = &args
	registeredName = map[string]bool{}
}

const snapshotFilePerm = 0755

type ErrRepeated struct {
	repeatedSnapshot string
}

func (e *ErrRepeated) Error() string {
	return fmt.Sprintf("expectation %q is already set", e.repeatedSnapshot)
}

func (e *ErrRepeated) Is(err error) bool {
	_, isErr := err.(*ErrRepeated)
	return isErr
}

func registerTestName(testName string) error {
	registerNameMutex.Lock()
	defer registerNameMutex.Unlock()
	if registeredName[testName] {
		return &ErrRepeated{repeatedSnapshot: testName}
	}
	registeredName[testName] = true
	return nil
}

type fileHeader struct {
	OS        string `json:"os"`
	LimitToOS bool   `json:"limit_to_os"`
}

func (f *fileHeader) dump() ([]byte, error) {
	return json.Marshal(f)
}

func (f *fileHeader) load(h []byte) error {
	return json.Unmarshal(h, f)
}

func (f *fileHeader) considerForCleanup() bool {
	return (f.LimitToOS && runtime.GOOS == f.OS) || !f.LimitToOS
}

type fileContents struct {
	header *fileHeader
	body   []byte
}

var headerSep = []byte("\n\n")

func (f *fileContents) dump(fileName string) error {
	h, err := f.header.dump()
	if err != nil {
		return fmt.Errorf("dumping header %w", err)
	}
	return os.WriteFile(fileName,
		append(h, append(headerSep, f.body...)...),
		snapshotFilePerm)
}

func (f *fileContents) load(fileName string) error {
	fContent, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("reading file for loading: %w", err)
	}
	sep := bytes.Index(fContent, headerSep)
	if sep == -1 {
		return fmt.Errorf("malformed expectation, cannot find separator")
	}
	f.header = &fileHeader{}
	f.header.load(fContent[:sep])
	f.body = fContent[sep+len(headerSep):]
	return nil
}
func AsSnapshot(t *testing.T, name string, comparable Comparable) {
	asSnapshot(t, name, comparable, false)
}

func AsOSDependentSnapshot(t *testing.T, name string, comparable Comparable) {
	asSnapshot(t, name, comparable, true)
}

func asSnapshot(t *testing.T, name string, comparable Comparable, limitOS bool) {
	pathName := url.PathEscape(name)
	if err := registerTestName(pathName); err == nil {
		t.Error(fmt.Errorf("setting new expectation: %w", err))
	}
	snapshotFilePath := filepath.Join(snapShotDir, pathName)
	if currentRunArgs != nil && currentRunArgs.shouldUpdate {
		fc := fileContents{
			header: &fileHeader{OS: runtime.GOOS, LimitToOS: limitOS},
			body:   comparable.Dump(),
		}
		if err := fc.dump(snapshotFilePath); err != nil {
			panic(err)
		}

		// we just updated the snapshot, makes no sense to compare
		return
	}
	f, err := os.Open(snapshotFilePath)
	if err != nil {
		t.Fatal(fmt.Errorf("opening snapshot file for test %q: %w", name, err))
	}
	defer f.Close()
	snapshotContents, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(fmt.Errorf("reading snapshot for test %q: %w", name, err))
	}
	expectation := comparable.Load(snapshotContents)
	diff, err := expectation.CompareTo(comparable)
	if err != nil {
		t.Fatal(fmt.Errorf("comparing expectation to result: %w", err))
	}
	if diff == "" {
		return
	}
	t.Logf("found a difference between expectation and result on test %q, difference follows:", name)
	t.Log(diff)
	t.FailNow()
}

func Cleanup() error {
	registerNameMutex.Lock()
	defer registerNameMutex.Unlock()
	if currentRunArgs == nil || (currentRunArgs != nil && !currentRunArgs.shouldCleanup) {
		return nil
	}
	dirContents, err := os.ReadDir(snapShotDir)
	if err != nil {
		return fmt.Errorf("reading snapshot directory contents: %w", err)
	}
	var deletable []string
	for _, entry := range dirContents {
		p := filepath.Join(snapShotDir, entry.Name())
		fc := &fileContents{}
		fc.load(p)
		if !fc.header.considerForCleanup() {
			continue
		}
		if registeredName[entry.Name()] {
			continue
		}
		deletable = append(deletable, p)
	}
	for i, d := range deletable {
		if err := os.Remove(d); err != nil {
			return fmt.Errorf("deleting stale snapshot, %d were deleted before failure: %w", i, err)
		}
	}
	return nil
}
