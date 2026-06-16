package scaffold

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
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
		name     string
		setup    func(t *testing.T) (embedded map[string]string, installed map[string]string)
		wantGaps int
		wantErr  bool
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
			// FIX 3 / decision D4: same version "1.0" but DIFFERENT content is a
			// real content update and MUST be reported as a gap. Version-only
			// diffing would have silently missed this.
			name: "gap when version is identical 1.0 but content differs",
			setup: func(t *testing.T) (map[string]string, map[string]string) {
				embedded := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---
# New body`,
				}
				installed := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "1.0"
---
# Old body`,
				}
				return embedded, installed
			},
			wantGaps: 1,
			wantErr:  false,
		},
		{
			// Same version AND identical content → no gap.
			name: "no gap when version and content are identical",
			setup: func(t *testing.T) (map[string]string, map[string]string) {
				embedded := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "2.0"
---
# Body`,
				}
				installed := map[string]string{
					"sdd-init/SKILL.md": `---
metadata:
  version: "2.0"
---
# Body`,
				}
				return embedded, installed
			},
			wantGaps: 0,
			wantErr:  false,
		},
		{
			name: "gap when version and content differ",
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
  version: "1.5"
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

func TestSkillVersion(t *testing.T) {
	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: sdd-init
metadata:
  version: "2.0"
---
# Init`)},
		"sdd-noversion/SKILL.md": &fstest.MapFile{Data: []byte(`---
name: sdd-noversion
---
# No version`)},
	}

	tests := []struct {
		name  string
		skill string
		want  string
	}{
		{name: "version present", skill: "sdd-init", want: "2.0"},
		{name: "version missing in frontmatter", skill: "sdd-noversion", want: ""},
		{name: "skill absent", skill: "does-not-exist", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SkillVersion(embeddedFS, tt.skill); got != tt.want {
				t.Errorf("SkillVersion(%q) = %q, want %q", tt.skill, got, tt.want)
			}
		})
	}
}

func TestClassifyGaps(t *testing.T) {
	tests := []struct {
		name         string
		embedded     map[string]string
		installed    map[string]string
		wantAdded    []string
		wantChanged  []string
		wantOrphaned []string
	}{
		{
			name: "added when embedded skill is not installed",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"2.0\"\n---",
			},
			installed:    map[string]string{},
			wantAdded:    []string{"sdd-init"},
			wantChanged:  nil,
			wantOrphaned: nil,
		},
		{
			name: "changed when version and content differ",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"2.0\"\n---",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.5\"\n---",
			},
			wantAdded:    nil,
			wantChanged:  []string{"sdd-init"},
			wantOrphaned: nil,
		},
		{
			// FIX 3 / decision D4: content is the source of truth. Same version
			// string "1.0" but DIFFERENT body content must be reported Changed,
			// because version-only diffing would silently miss content updates
			// for the many skills that genuinely ship version "1.0".
			// Covers harness-update scenario:
			//   "Content change with unchanged version is detected".
			name: "changed when version is identical 1.0 but content differs",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---\n# New body",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---\n# Old body",
			},
			wantAdded:    nil,
			wantChanged:  []string{"sdd-init"},
			wantOrphaned: nil,
		},
		{
			// Same version AND identical content → not changed.
			name: "no change when version and content are identical",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---\n# Same body",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---\n# Same body",
			},
			wantAdded:    nil,
			wantChanged:  nil,
			wantOrphaned: nil,
		},
		{
			// Different version + different content → Changed.
			name: "changed when both version and content differ",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"2.0\"\n---\n# New body",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---\n# Old body",
			},
			wantAdded:    nil,
			wantChanged:  []string{"sdd-init"},
			wantOrphaned: nil,
		},
		{
			name: "orphaned when installed skill is no longer embedded",
			embedded: map[string]string{
				"sdd-init/SKILL.md": "---\nmetadata:\n  version: \"2.0\"\n---",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md":    "---\nmetadata:\n  version: \"2.0\"\n---",
				"sdd-retired/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---",
			},
			wantAdded:    nil,
			wantChanged:  nil,
			wantOrphaned: []string{"sdd-retired"},
		},
		{
			name: "mixed added changed and orphaned",
			embedded: map[string]string{
				"sdd-init/SKILL.md":    "---\nmetadata:\n  version: \"2.0\"\n---",
				"sdd-propose/SKILL.md": "---\nmetadata:\n  version: \"3.0\"\n---",
			},
			installed: map[string]string{
				"sdd-init/SKILL.md":    "---\nmetadata:\n  version: \"1.5\"\n---",
				"sdd-retired/SKILL.md": "---\nmetadata:\n  version: \"1.0\"\n---",
			},
			wantAdded:    []string{"sdd-propose"},
			wantChanged:  []string{"sdd-init"},
			wantOrphaned: []string{"sdd-retired"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			embeddedFS := make(fstest.MapFS)
			for path, content := range tt.embedded {
				embeddedFS[path] = &fstest.MapFile{Data: []byte(content)}
			}

			installedDir := filepath.Join(tmpDir, "installed")
			if err := os.MkdirAll(installedDir, 0o755); err != nil {
				t.Fatalf("MkdirAll() error = %v", err)
			}
			for path, content := range tt.installed {
				fullPath := filepath.Join(installedDir, path)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
					t.Fatalf("MkdirAll() error = %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			}

			report, err := ClassifyGaps(embeddedFS, installedDir)
			if err != nil {
				t.Fatalf("ClassifyGaps() error = %v", err)
			}

			assertSkillNames(t, "Added", report.Added, tt.wantAdded)
			assertSkillNames(t, "Changed", report.Changed, tt.wantChanged)
			assertSkillNames(t, "Orphaned", report.Orphaned, tt.wantOrphaned)
		})
	}
}

func assertSkillNames(t *testing.T, bucket string, got []SkillChange, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %d entries (%v), want %d (%v)", bucket, len(got), names(got), len(want), want)
	}
	gotSet := make(map[string]struct{}, len(got))
	for _, c := range got {
		gotSet[c.Name] = struct{}{}
	}
	for _, name := range want {
		if _, ok := gotSet[name]; !ok {
			t.Errorf("%s missing expected skill %q (got %v)", bucket, name, names(got))
		}
	}
}

func names(changes []SkillChange) []string {
	out := make([]string, len(changes))
	for i, c := range changes {
		out[i] = c.Name
	}
	return out
}
