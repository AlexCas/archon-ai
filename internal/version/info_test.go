package version

import (
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		want    []string
	}{
		{
			name:    "default values",
			version: "dev",
			commit:  "none",
			date:    "unknown",
			want:    []string{"dev", "none", "unknown"},
		},
		{
			name:    "release values",
			version: "1.0.0",
			commit:  "abc123",
			date:    "2026-06-10",
			want:    []string{"1.0.0", "abc123", "2026-06-10"},
		},
		{
			name:    "semantic version",
			version: "2.3.4",
			commit:  "def456",
			date:    "2026-01-15T10:30:00Z",
			want:    []string{"2.3.4", "def456", "2026-01-15T10:30:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldVersion := Version
			oldCommit := Commit
			oldDate := Date
			defer func() {
				Version = oldVersion
				Commit = oldCommit
				Date = oldDate
			}()

			Version = tt.version
			Commit = tt.commit
			Date = tt.date

			got := Print()

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("Print() = %q, should contain %q", got, want)
				}
			}

			if !strings.Contains(got, "archon version") {
				t.Errorf("Print() = %q, should contain 'archon version'", got)
			}
			if !strings.Contains(got, "commit:") {
				t.Errorf("Print() = %q, should contain 'commit:'", got)
			}
			if !strings.Contains(got, "built:") {
				t.Errorf("Print() = %q, should contain 'built:'", got)
			}
		})
	}
}

func TestPrint_Format(t *testing.T) {
	oldVersion := Version
	oldCommit := Commit
	oldDate := Date
	defer func() {
		Version = oldVersion
		Commit = oldCommit
		Date = oldDate
	}()

	Version = "1.0.0"
	Commit = "abc123"
	Date = "2026-06-10"

	got := Print()
	want := "archon version 1.0.0 (commit: abc123, built: 2026-06-10)"

	if got != want {
		t.Errorf("Print() = %q, want %q", got, want)
	}
}
