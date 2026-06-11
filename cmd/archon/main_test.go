package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func fakeEmbeddedFS() fstest.MapFS {
	return fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
		"sdd-propose/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-propose\n---\n# Propose"),
		},
	}
}

func setupProjectDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	opencodeDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	return tmpDir
}

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "archon version") {
		t.Errorf("version output = %q, want contains 'archon version'", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("version output = %q, want contains 'commit:'", output)
	}
}

func TestInitCommand_DryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"init", "--dry-run"})

	origDir, _ := os.Getwd()
	tmpDir := setupProjectDir(t)
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Dry run") {
		t.Errorf("init --dry-run output = %q, want contains 'Dry run'", output)
	}
	if !strings.Contains(output, "Project dir:") {
		t.Errorf("init --dry-run output = %q, want contains 'Project dir:'", output)
	}
}

func TestInitCommand_NoAgentDetected(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"init"})

	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := root.Execute()
	if err == nil {
		t.Error("Execute() should fail when no agent detected")
	}
}

func TestStatusCommand_NotInitialized(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"status"})

	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := root.Execute()
	if err == nil {
		t.Error("Execute() should fail when not initialized")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "No archon configuration found") {
		t.Errorf("stderr = %q, want contains 'No archon configuration found'", errOutput)
	}
}

func TestRollbackCommand_NothingToRollback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"rollback"})

	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Nothing to rollback") {
		t.Errorf("rollback output = %q, want contains 'Nothing to rollback'", output)
	}
}

func TestRollbackCommand_DryRun(t *testing.T) {
	var stdout, stderr bytes.Buffer

	origDir, _ := os.Getwd()
	tmpDir := setupProjectDir(t)
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	archonDir := filepath.Join(tmpDir, ".archon")
	if err := os.MkdirAll(archonDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	manifestContent := `{
  "version": "dev",
  "paths": ["` + filepath.Join(tmpDir, ".archon", "config.yaml") + `"],
  "original_agents_md_backup": ""
}`
	if err := os.WriteFile(filepath.Join(archonDir, "rollback.json"), []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"rollback", "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Dry run") {
		t.Errorf("rollback --dry-run output = %q, want contains 'Dry run'", output)
	}
	if !strings.Contains(output, "config.yaml") {
		t.Errorf("rollback --dry-run output = %q, want contains 'config.yaml'", output)
	}
}

func TestInitCommand_WithAgentFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	t.Setenv("HOME", homeDir)

	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"init", "--agent", "opencode"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "initialized successfully") {
		t.Errorf("init output = %q, want contains 'initialized successfully'", output)
	}
	if !strings.Contains(output, "opencode") {
		t.Errorf("init output = %q, want contains 'opencode'", output)
	}
}
