package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkOrCopy(t *testing.T) {
	tmpDir := t.TempDir()

	globalDir := filepath.Join(tmpDir, "global")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	skillDir := filepath.Join(globalDir, "sdd-init")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Test Skill"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := SymlinkOrCopy(globalDir, projectDir, "sdd-init"); err != nil {
		t.Fatalf("SymlinkOrCopy() error = %v", err)
	}

	linkPath := filepath.Join(projectDir, "sdd-init")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("Lstat() error = %v", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(linkPath)
		if err != nil {
			t.Fatalf("Readlink() error = %v", err)
		}
		if target != skillDir {
			t.Errorf("Symlink target = %q, want %q", target, skillDir)
		}
	} else {
		data, err := os.ReadFile(linkPath + "/SKILL.md")
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if string(data) != "# Test Skill" {
			t.Errorf("Copied content = %q, want %q", string(data), "# Test Skill")
		}
	}
}

func TestSymlinkOrCopy_CopyFallback(t *testing.T) {
	tmpDir := t.TempDir()

	globalDir := filepath.Join(tmpDir, "global")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	skillDir := filepath.Join(globalDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	subDir := filepath.Join(skillDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	files := map[string]string{
		"SKILL.md":          "# Main skill",
		"subdir/README.md":  "# Subdir readme",
	}

	for path, content := range files {
		fullPath := filepath.Join(skillDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	if err := os.Symlink(globalDir, filepath.Join(tmpDir, "dummy")); err != nil {
		t.Skip("Symlinks not supported, skipping fallback test")
	}

	if err := SymlinkOrCopy(globalDir, projectDir, "test-skill"); err != nil {
		t.Fatalf("SymlinkOrCopy() error = %v", err)
	}

	for path, expectedContent := range files {
		fullPath := filepath.Join(projectDir, "test-skill", path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("ReadFile(%s) error = %v", path, err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("File %s content = %q, want %q", path, string(data), expectedContent)
		}
	}
}

func TestSymlinkOrCopy_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()

	globalDir := filepath.Join(tmpDir, "global")
	projectDir := filepath.Join(tmpDir, "deep", "nested", "project")

	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	skillDir := filepath.Join(globalDir, "sdd-init")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# Test"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := SymlinkOrCopy(globalDir, projectDir, "sdd-init"); err != nil {
		t.Fatalf("SymlinkOrCopy() error = %v", err)
	}

	linkPath := filepath.Join(projectDir, "sdd-init")
	if _, err := os.Lstat(linkPath); err != nil {
		t.Errorf("Symlink not created: %v", err)
	}
}

func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	files := map[string]string{
		"file1.txt":        "content1",
		"sub/file2.txt":    "content2",
	}

	for path, content := range files {
		fullPath := filepath.Join(srcDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir() error = %v", err)
	}

	for path, expectedContent := range files {
		fullPath := filepath.Join(dstDir, path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("ReadFile(%s) error = %v", path, err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("File %s content = %q, want %q", path, string(data), expectedContent)
		}
	}
}
