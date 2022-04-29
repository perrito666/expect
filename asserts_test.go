package expect

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"expect/snapshots"
	"expect/snapshots/comparabletypes"
)

func newStringComparableFromLiteral(s string) *comparabletypes.StringComparable {
	c := comparabletypes.StringComparable(s)
	return &c
}

func Test_fromSnapshot(t *testing.T) {
	type args struct {
		name       string
		comparable snapshots.Comparable
		limitOS    bool
		config     *Config
	}
	// currentRunArgs = &Args{shouldUpdate: true} // handy for new tests
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "regular_compare",
			args: args{
				name:       "test_from_snapshot_01",
				comparable: newStringComparableFromLiteral("Hello World"),
				limitOS:    false,
				config: &Config{
					Grouping:    groupByTestFile,
					SnapShotDir: "test_snapshot_sample",
					Replacers:   nil,
				},
			},
			wantErr: false,
		},
		{
			name: "regular_compare_repeated_fails",
			args: args{
				name:       "test_from_snapshot_01",
				comparable: newStringComparableFromLiteral("Hello World"),
				limitOS:    false,
				config: &Config{
					Grouping:    groupByTestFile,
					SnapShotDir: "test_snapshot_sample",
					Replacers:   nil,
				},
			},
			wantErr: true,
		},
		{
			name: "regular_compare_group_by_test_file",
			args: args{
				name:       "test_from_snapshot_02",
				comparable: newStringComparableFromLiteral("Hello World"),
				limitOS:    false,
				config: &Config{
					Grouping:  groupByTestFile,
					Replacers: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "regular_compare_group_by_package",
			args: args{
				name:       "test_from_snapshot_03",
				comparable: newStringComparableFromLiteral("Hello World"),
				limitOS:    false,
				config: &Config{
					Grouping:  groupByPackage,
					Replacers: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fromSnapshot(tt.args.name, tt.args.comparable, tt.args.limitOS, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("fromSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cleanup(t *testing.T) {
	var deletableOS string
	// this accounts for the most common and ones I can try, if you have another and want to run test
	// add a case. Here is the full list https://github.com/golang/go/blob/master/src/go/build/syslist.go
	switch runtime.GOOS {
	case "windows", "darwin":
		deletableOS = "linux"
	default:
		deletableOS = "windows"
	}
	type args struct {
		config *Config
	}
	tests := []struct {
		name           string
		deletables     []string
		conservables   []string
		os_spare_ables []string
		args           args
		wantErr        bool
	}{
		{
			name:         "cleanup_snapshot_folder",
			conservables: []string{"one", "two", "tree"},
			deletables:   []string{"four", "five", "six"},
			args: args{config: &Config{
				Grouping:    "",
				SnapShotDir: t.TempDir(),
				Replacers:   nil,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentRunArgs = &Args{shouldUpdate: true, shouldCleanup: true}
			for _, dc := range append(tt.deletables, tt.conservables...) {
				fd, err := os.OpenFile(filepath.Join(tt.args.config.SnapShotDir, dc),
					os.O_CREATE|os.O_TRUNC, snapshotFilePerm)
				if err != nil {
					t.Fatal(err)
				}
				fd.WriteString(`{
  "os": "windows",
  "limit_to_os": false
}

Hello World`)
				fd.Close()
			}
			for _, dc := range tt.os_spare_ables {
				fd, err := os.OpenFile(filepath.Join(tt.args.config.SnapShotDir, dc),
					os.O_CREATE|os.O_TRUNC, snapshotFilePerm)
				if err != nil {
					t.Fatal(err)
				}
				fd.WriteString(fmt.Sprintf(`{
  "os": "%s",
  "limit_to_os": true
}

Hello World`, deletableOS))
				fd.Close()
			}
			registeredName = map[string]bool{}
			for _, c := range tt.conservables {
				registeredName[c] = true
			}
			if err := cleanup(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("cleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, c := range append(tt.conservables, tt.os_spare_ables...) {
				fPath := filepath.Join(tt.args.config.SnapShotDir, c)
				_, err := os.Stat(fPath)
				if err != nil {
					t.Logf("expected file %q to exist but it does not", fPath)
					t.FailNow()
				}
			}
			for _, d := range tt.deletables {
				fPath := filepath.Join(tt.args.config.SnapShotDir, d)
				_, err := os.Stat(fPath)
				if err == nil {
					t.Logf("expected file %q to no longer exist but it does", fPath)
					t.FailNow()
				}
			}
			currentRunArgs = &Args{shouldUpdate: false, shouldCleanup: false}
		})
	}
}