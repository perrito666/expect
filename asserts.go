package expect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"perri.to/expect/snapshots"
)

type Args struct {
	shouldUpdate   bool
	shouldCleanup  bool
	runInArguments bool
}

var currentRunArgs *Args
var registeredName map[string]bool
var registerNameMutex sync.Mutex

// ran should be set to true if we ran at least one test.
var ran bool

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
		case "-run", "-test.run":
			args.runInArguments = true
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
	ran = true // we ran at least once
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
	return json.MarshalIndent(f, "", "  ")
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
	abs, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("getting abs path for dump file %w", err)
	}
	fd := filepath.Dir(abs)
	err = os.MkdirAll(fd, snapshotFilePerm)
	if err != nil {
		return fmt.Errorf("creating snapshot folders %w", err)
	}
	return os.WriteFile(fileName,
		append(h, append(headerSep, f.body...)...),
		snapshotFilePerm)
}

var ErrNotSnapshotted = errors.New("not snapshotted")

func readFileContents(fileName string) (*fileContents, error) {
	fc := &fileContents{}
	if _, err := os.Stat(fileName); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotSnapshotted
		}
		return nil, err
	}
	if err := fc.load(fileName); err != nil {
		return nil, err
	}
	return fc, nil
}

func (f *fileContents) load(fileName string) error {
	fContent, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("reading file for loading: %w", err)
	}
	if len(fContent) == 0 {
		f.header = &fileHeader{
			OS:        runtime.GOOS,
			LimitToOS: false,
		}
		return nil
	}
	sep := bytes.Index(fContent, headerSep)
	if sep == -1 {
		return fmt.Errorf("malformed expectation, cannot find separator")
	}
	f.header = &fileHeader{}
	if err := f.header.load(fContent[:sep]); err != nil {
		return fmt.Errorf("loading header: %w", err)
	}
	f.body = fContent[sep+len(headerSep):]
	return nil
}

// FromSnapshot will fail if the stored information is not equal (in a non-agnostic comparison) to the passed comparabletypes.
func FromSnapshot(t *testing.T, name string, comparable snapshots.Comparable) {
	doCompareAndEvaluateResult(t, name, comparable, false)
}

// FromSnapshotWithConfig will fail if the stored information is not equal (in a non-agnostic comparison) to the passed
// comparabletypes, the config will be overriden/updated with the passed one.
func FromSnapshotWithConfig(t *testing.T, name string, comparable snapshots.Comparable, config *Config) {
	doCompareAndEvaluateResultWithConfig(t, name, comparable, false, config)
}

// FromOSDependentSnapshot will fail if the stored information is not equal (in a non-agnostic comparison) to the passed
// comparabletypes but only if the OS of both matches, this should prevent weird side effect of snapshotting in different
// machines.
func FromOSDependentSnapshot(t *testing.T, name string, comparable snapshots.Comparable) {
	doCompareAndEvaluateResult(t, name, comparable, true)
}

// FromOSDependentSnapshotWithConfig will fail if the stored information is not equal (in a non-agnostic comparison) to the passed
// comparabletypes but only if the OS of both matches, this should prevent weird side effect of snapshotting in different
// machines, the config will be overriden/updated with the passed one
func FromOSDependentSnapshotWithConfig(t *testing.T, name string, comparable snapshots.Comparable, config *Config) {
	doCompareAndEvaluateResultWithConfig(t, name, comparable, true, config)
}

// doCompareAndEvaluateResult does the actual snapshot running and decides if and how to fail according to results.
func doCompareAndEvaluateResult(t *testing.T, name string, comparable snapshots.Comparable, limitOs bool) {
	config, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}
	doCompareAndEvaluateResultWithConfig(t, name, comparable, limitOs, config)
}

// doCompareAndEvaluateResult does the actual snapshot running and decides if and how to fail according to results.
func doCompareAndEvaluateResultWithConfig(t *testing.T, name string, comparable snapshots.Comparable, limitOs bool,
	config *Config) {

	if err := fromSnapshot(name, comparable, limitOs, config); err != nil {
		if errors.Is(err, ErrNotSnapshotted) {
			t.Fatal(fmt.Errorf("expected snapshot for %s to exist: %w", name, err))
			return
		}
		if errors.Is(err, &ErrTestErrored{}) {
			t.Fatal(errors.Unwrap(err))
			return
		}
		if errors.Is(err, &ErrTestFailed{}) {
			t.Logf("found a difference between expectation and result on test %q, difference follows:", name)
			/*for _, line := range strings.Split(err.Error(), "\n"){
				t.Log(line)
			}*/
			t.Log(err)
			t.FailNow()
			return
		}
		panic(err)
	}
}

// ErrTestFailed should be returned when a comparison test fails.
type ErrTestFailed struct {
	failure string
}

// Error implements errors for ErrTestFailed
func (e *ErrTestFailed) Error() string {
	return e.failure
}

// Is implements the (for some unspoken reason) tacit errors.Is interface for ErrTestFailed
func (e *ErrTestFailed) Is(err error) bool {
	_, is := err.(*ErrTestFailed)
	return is
}

// ErrTestErrored should be returned when one of the preconditions or setups of the tests errored
type ErrTestErrored struct {
	err error
}

// Error implements errors for ErrTestErrored
func (e *ErrTestErrored) Error() string {
	return fmt.Sprintf("test errored: %s", e.err)
}

// Is implements the (for some unspoken reason) tacit errors.Is interface for ErrTestErrored
func (e *ErrTestErrored) Is(err error) bool {
	_, is := err.(*ErrTestErrored)
	return is
}

// Unwrap  implements the (for some unspoken reason) tacit errors.Unwrap() interface for ErrTestErrored
func (e *ErrTestErrored) Unwrap() error {
	return e.err
}

// fromSnapshot loads and compares the snapshot,  it is separated form the logic that handles testing.T to ease
// unit testing.
func fromSnapshot(name string, comparable snapshots.Comparable, limitOS bool, config *Config) error {
	pathName := url.PathEscape(name)
	if err := registerTestName(pathName); err != nil {
		return &ErrTestErrored{
			err: fmt.Errorf("setting new expectation: %w", err),
		}
	}

	// get the test file name just in case snapshot dir needs it
	_, fileName, _, _ := runtime.Caller(0)
	packageSnapshotDir := config.SnapshotDir(fileName)
	snapshotFilePath := filepath.Join(packageSnapshotDir, pathName)
	if ext := comparable.Extension(); ext != "" {
		snapshotFilePath = fmt.Sprintf("%s.%s", snapshotFilePath, ext)
	}

	updatingSnapshot := currentRunArgs != nil && currentRunArgs.shouldUpdate

	fc, err := readFileContents(snapshotFilePath)
	if err != nil {
		return &ErrTestErrored{
			err: fmt.Errorf("loading expectations file: %w", err),
		}
	}

	expectation := comparable.Load(fc.body)
	// time to replace, comparable will know how to.
	if replaceable, ok := config.Replacers[expectation.Kind()]; ok {
		expectation.Replace(replaceable)
		comparable.Replace(replaceable)
	}
	// if this is a composed type lets pass the replacers to the subtypes
	if expectation.Subtypes() {
		expectation.ReplaceSubtypes(config.Replacers)
	}
	if comparable.Subtypes() {
		comparable.ReplaceSubtypes(config.Replacers)
	}

	diff, err := expectation.CompareTo(comparable)
	if err != nil {
		// we are updating, don't care
		if updatingSnapshot {
			fcNew := fileContents{
				header: &fileHeader{OS: runtime.GOOS, LimitToOS: limitOS},
				body:   comparable.Dump(),
			}
			if err := fcNew.dump(snapshotFilePath); err != nil {
				panic(err)
			}
			return nil
		}
		return &ErrTestErrored{
			err: fmt.Errorf("comparing expectation to result: %w", err),
		}
	}
	if diff != "" {
		// we are updating, we only do so if there are differences
		if updatingSnapshot {
			fcNew := fileContents{
				header: &fileHeader{OS: runtime.GOOS, LimitToOS: limitOS},
				body:   comparable.Dump(),
			}
			if err := fcNew.dump(snapshotFilePath); err != nil {
				panic(err)
			}
			return nil
		}
		return &ErrTestFailed{failure: diff}
	}
	return nil
}

// Cleanup should be called in TestMain AFTER m.Run() to remove stale snapshots
func Cleanup() error {
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("cleaning up stale snapshots: %w", err)
	}
	return cleanup(config, false)
}

// MustCleanup will do exactly as Cleanup but also fail if a cleanup was due and no flag was passed
func MustCleanup() error {
	config, err := ReadConfig()
	if err != nil {
		return fmt.Errorf("cleaning up stale snapshots: %w", err)
	}
	return cleanup(config, true)
}

func cleanup(config *Config, must bool) error {
	registerNameMutex.Lock()
	defer registerNameMutex.Unlock()

	if currentRunArgs != nil && currentRunArgs.runInArguments {
		fmt.Println("skipping cleanup because -run was used")
		return fmt.Errorf("skipping cleanup because -run was used")
	}

	if !ran {
		return fmt.Errorf("must run a test before cleaning up")
	}

	shouldCleanup := currentRunArgs != nil && currentRunArgs.shouldCleanup
	if !must && (currentRunArgs == nil || !shouldCleanup) {
		return nil
	}

	_, fileName, _, _ := runtime.Caller(0)
	packageSnapshotDir := config.SnapshotDir(fileName)
	dirContents, err := os.ReadDir(packageSnapshotDir)
	if err != nil {
		return fmt.Errorf("reading snapshot directory contents: %w", err)
	}
	var deletable []string
	for _, entry := range dirContents {
		p := filepath.Join(packageSnapshotDir, entry.Name())
		fc, err := readFileContents(p)
		if err != nil {
			return fmt.Errorf("loading file contents: %w", err)
		}
		if !fc.header.considerForCleanup() {
			continue
		}
		fName := entry.Name()
		ext := path.Ext(entry.Name())
		if ext != fName && len(ext) > 0 {
			fName = strings.TrimSuffix(fName, ext)
		}
		if registeredName[fName] {
			continue
		}
		deletable = append(deletable, p)
		if must && !shouldCleanup {
			cleanName, err := url.PathUnescape(entry.Name())
			if err != nil {
				cleanName = entry.Name()
			}
			fmt.Printf("CLEANUP: There is a snapshot for expectation %q but the expectation no longer exist\n", cleanName)
		}
	}
	if !shouldCleanup {
		if must && len(deletable) > 0 {
			return fmt.Errorf("we found %d expectation snapshots that need cleanup", len(deletable))
		}
		return nil
	}
	for i, d := range deletable {
		if err := os.Remove(d); err != nil {
			return fmt.Errorf("deleting stale snapshot, %d were deleted before failure: %w", i, err)
		}
	}
	return nil
}
