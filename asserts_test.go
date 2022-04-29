package expect

import (
	"os"
	"path/filepath"
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
	type args struct {
		config *Config
	}
	tests := []struct {
		name         string
		deletables   []string
		conservables []string
		args         args
		wantErr      bool
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
			for _, dc := range append(tt.deletables, tt.conservables...) {
				fd, err := os.OpenFile(filepath.Join(tt.args.config.SnapShotDir, dc),
					os.O_CREATE|os.O_TRUNC, snapshotFilePerm)
				if err != nil {
					t.Fatal(err)
				}
				fd.Close()
			}
			registeredName = map[string]bool{}
			for _, c := range tt.conservables {
				registeredName[c] = true
			}
			if err := cleanup(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("cleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
			for _, c := range tt.conservables {
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
		})
	}
}
