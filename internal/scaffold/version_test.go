package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "version with quotes",
			content: `---
name: sdd-init
metadata:
  version: "3.0"
---
# Content`,
			want: "3.0",
		},
		{
			name: "version without quotes",
			content: `---
name: sdd-init
metadata:
  version: 2.5
---
# Content`,
			want: "2.5",
		},
		{
			name: "version with single quotes",
			content: `---
name: sdd-init
metadata:
  version: '1.5'
---
# Content`,
			want: "1.5",
		},
		{
			name: "no metadata section",
			content: `---
name: sdd-init
---
# Content`,
			want: "",
		},
		{
			name: "no version in metadata",
			content: `---
name: sdd-init
metadata:
  author: test
---
# Content`,
			want: "",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name: "version with extra spaces",
			content: `---
metadata:
  version:   "4.2"  
---`,
			want: "4.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractVersion(tt.content)
			if got != tt.want {
				t.Errorf("extractVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectVersionGaps(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (embedded map[string]string, installed map[string]string)
		wantGaps  int
		wantErr   bool
	}{
		{
			name: "no gaps when versions match",
			setup: func(t *testing.T) (map[string]string, map[string]string) {
				embedded := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---`,
				}
				installed := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---`,
				}
				return embedded, installed
			},
			wantGaps: 0,
			wantErr:  false,
		},
		{
			name: "gap when version mismatch",
			setup: func(t *testing.T) (map[string]string, map[string]string) {
				embedded := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "2.0"
---`,
				}
				installed := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---`,
				}
				return embedded, installed
			},
			wantGaps: 1,
			wantErr:  false,
		},
		{
			name: "gap when skill not installed",
			setup: func(t *testing.T) (map[string]string, map[string]string) {
				embedded := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---`,
				}
				installed := map[string]string{}
				return embedded, installed
			},
			wantGaps: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedded, installed := tt.setup(t)

			tmpDir := t.TempDir()

			for path, content := range embedded {
				fullPath := filepath.Join(tmpDir, "embedded", path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("MkdirAll() error = %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}

			installedDir := filepath.Join(tmpDir, "installed")
			for path, content := range installed {
				fullPath := filepath.Join(installedDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("MkdirAll() error = %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}

			embeddedFS := os.DirFS(filepath.Join(tmpDir, "embedded"))
			gaps, err := DetectVersionGaps(embeddedFS, installedDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectVersionGaps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(gaps) != tt.wantGaps {
				t.Errorf("DetectVersionGaps() returned %d gaps, want %d", len(gaps), tt.wantGaps)
			}
		})
	}
}
