package scaffold

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name        string
		fs          fstest.MapFS
		wantSkills  []string
		wantErr     bool
	}{
		{
			name: "extract single skill",
			fs: fstest.MapFS{
				"sdd-init/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: sdd-init\n---\n# Skill content"),
				},
			},
			wantSkills: []string{"sdd-init"},
			wantErr:    false,
		},
		{
			name: "extract multiple skills",
			fs: fstest.MapFS{
				"sdd-init/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: sdd-init\n---\n# Init"),
				},
				"sdd-propose/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: sdd-propose\n---\n# Propose"),
				},
				"judgment-day/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: judgment-day\n---\n# Judge"),
				},
			},
			wantSkills: []string{"sdd-init", "sdd-propose", "judgment-day"},
			wantErr:    false,
		},
		{
			name:       "empty filesystem",
			fs:         fstest.MapFS{},
			wantSkills: nil,
			wantErr:    false,
		},
		{
			name: "skip non-directory entries",
			fs: fstest.MapFS{
				"README.md": &fstest.MapFile{
					Data: []byte("# Skills"),
				},
				"sdd-init/SKILL.md": &fstest.MapFile{
					Data: []byte("---\nname: sdd-init\n---\n# Init"),
				},
			},
			wantSkills: []string{"sdd-init"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			got, err := Extract(tt.fs, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.wantSkills) {
				t.Errorf("Extract() returned %d skills, want %d", len(got), len(tt.wantSkills))
				return
			}

			gotSet := make(map[string]bool)
			for _, skill := range got {
				gotSet[skill] = true
			}

			for _, wantSkill := range tt.wantSkills {
				if !gotSet[wantSkill] {
					t.Errorf("Extract() missing skill %q", wantSkill)
				}

				skillPath := filepath.Join(tmpDir, wantSkill, "SKILL.md")
				if _, err := os.Stat(skillPath); os.IsNotExist(err) {
					t.Errorf("Skill file not created: %s", skillPath)
				}
			}
		})
	}
}

func TestExtract_Idempotency(t *testing.T) {
	fs := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\nversion: 1.0\n---\n# Init"),
		},
	}

	tmpDir := t.TempDir()

	first, err := Extract(fs, tmpDir)
	if err != nil {
		t.Fatalf("First Extract() error = %v", err)
	}

	second, err := Extract(fs, tmpDir)
	if err != nil {
		t.Fatalf("Second Extract() error = %v", err)
	}

	if len(first) != len(second) {
		t.Errorf("Idempotency failed: first=%d, second=%d", len(first), len(second))
	}

	skillPath := filepath.Join(tmpDir, "sdd-init", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	expected := "---\nname: sdd-init\nversion: 1.0\n---\n# Init"
	if string(data) != expected {
		t.Errorf("File content changed after re-extract")
	}
}

func TestExtract_CreatesDirectoryStructure(t *testing.T) {
	fs := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("# Init"),
		},
	}

	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, "skills")

	_, err := Extract(fs, skillsDir)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	expectedDir := filepath.Join(skillsDir, "sdd-init")
	info, err := os.Stat(expectedDir)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if !info.IsDir() {
		t.Errorf("Expected directory at %s", expectedDir)
	}
}
