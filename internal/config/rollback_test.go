package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollbackManifest_WriteManifest(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &RollbackManifest{
		Version: "1.0.0",
		CreatedPaths: []string{
			filepath.Join(tmpDir, ".archon", "config.yaml"),
			filepath.Join(tmpDir, "skills", "sdd-init"),
		},
		HomeDir: tmpDir,
	}

	if err := manifest.WriteManifest(); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	manifestPath := manifest.manifestPath()
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("Manifest file not created at %s", manifestPath)
	}
}

func TestRollbackManifest_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")
	dir1 := filepath.Join(tmpDir, "testdir")

	os.WriteFile(file1, []byte("test1"), 0o644)
	os.WriteFile(file2, []byte("test2"), 0o644)
	os.MkdirAll(dir1, 0o755)

	manifest := &RollbackManifest{
		Version: "1.0.0",
		CreatedPaths: []string{
			file1,
			file2,
			dir1,
		},
		HomeDir: tmpDir,
	}

	if err := manifest.WriteManifest(); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	for _, path := range manifest.CreatedPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Path %s still exists after cleanup", path)
		}
	}

	manifestPath := manifest.manifestPath()
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Errorf("Manifest file still exists after cleanup")
	}
}

func TestRollbackManifest_BackupAgentsMD(t *testing.T) {
	tmpDir := t.TempDir()

	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	agentsContent := []byte("# Original AGENTS.md\n")
	if err := os.WriteFile(agentsPath, agentsContent, 0o644); err != nil {
		t.Fatalf("Failed to create AGENTS.md: %v", err)
	}

	manifest := &RollbackManifest{
		Version:      "1.0.0",
		CreatedPaths: []string{},
		HomeDir:      tmpDir,
	}

	if err := manifest.BackupAgentsMD(); err != nil {
		t.Fatalf("BackupAgentsMD() error = %v", err)
	}

	if manifest.BackupPath == "" {
		t.Error("BackupPath not set after backup")
	}

	if _, err := os.Stat(agentsPath); !os.IsNotExist(err) {
		t.Error("Original AGENTS.md still exists after backup")
	}

	if _, err := os.Stat(manifest.BackupPath); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}

	backupContent, err := os.ReadFile(manifest.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(backupContent) != string(agentsContent) {
		t.Errorf("Backup content = %q, want %q", string(backupContent), string(agentsContent))
	}
}

func TestRollbackManifest_CleanupWithRestore(t *testing.T) {
	tmpDir := t.TempDir()

	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	originalContent := []byte("# Original AGENTS.md\n")
	if err := os.WriteFile(agentsPath, originalContent, 0o644); err != nil {
		t.Fatalf("Failed to create AGENTS.md: %v", err)
	}

	manifest := &RollbackManifest{
		Version:      "1.0.0",
		CreatedPaths: []string{},
		HomeDir:      tmpDir,
	}

	if err := manifest.BackupAgentsMD(); err != nil {
		t.Fatalf("BackupAgentsMD() error = %v", err)
	}

	newContent := []byte("# New AGENTS.md\n")
	if err := os.WriteFile(agentsPath, newContent, 0o644); err != nil {
		t.Fatalf("Failed to create new AGENTS.md: %v", err)
	}

	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	restoredContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("Failed to read restored AGENTS.md: %v", err)
	}
	if string(restoredContent) != string(originalContent) {
		t.Errorf("Restored content = %q, want %q", string(restoredContent), string(originalContent))
	}
}

func TestRollbackManifest_BackupAgentsMD_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &RollbackManifest{
		Version:      "1.0.0",
		CreatedPaths: []string{},
		HomeDir:      tmpDir,
	}

	if err := manifest.BackupAgentsMD(); err != nil {
		t.Fatalf("BackupAgentsMD() error = %v", err)
	}

	if manifest.BackupPath != "" {
		t.Errorf("BackupPath should be empty when AGENTS.md doesn't exist, got %s", manifest.BackupPath)
	}
}

func TestRollbackManifest_Cleanup_NonexistentPaths(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &RollbackManifest{
		Version: "1.0.0",
		CreatedPaths: []string{
			filepath.Join(tmpDir, "nonexistent1.txt"),
			filepath.Join(tmpDir, "nonexistent2.txt"),
		},
		HomeDir: tmpDir,
	}

	if err := manifest.WriteManifest(); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup() should not error on nonexistent paths: %v", err)
	}
}
