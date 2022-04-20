package expect

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
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

func registerTestName(testName string) {
	registerNameMutex.Lock()
	defer registerNameMutex.Unlock()
	registeredName[testName] = true
}

func AsSnapshot(t *testing.T, name string, comparable Comparable) {
	pathName := url.PathEscape(name)
	registerTestName(pathName)
	snapshotFilePath := filepath.Join(snapShotDir, pathName)
	if currentRunArgs != nil && currentRunArgs.shouldUpdate {
		if err := os.WriteFile(snapshotFilePath, comparable.Dump(), snapshotFilePerm); err != nil {
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
