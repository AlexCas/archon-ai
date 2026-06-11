package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupProjectWithConfig(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	archonDir := filepath.Join(tmpDir, ".archon")
	if err := os.MkdirAll(archonDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configContent := `harness_version: "1.0.0"
agent: opencode
skill_count: 2
created_at: 2026-06-10T00:00:00Z
mutation_testing:
  enabled: false
models:
  default: claude-sonnet-4
  phases:
    apply: gpt-4o
`
	if err := os.WriteFile(filepath.Join(archonDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return tmpDir
}

func TestConfigCmd_Get(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "get", "models.default"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	got := strings.TrimSpace(stdout.String())
	if got != "claude-sonnet-4" {
		t.Errorf("config get models.default = %q, want %q", got, "claude-sonnet-4")
	}
}

func TestConfigCmd_GetPhase(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "get", "models.phases.apply"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	got := strings.TrimSpace(stdout.String())
	if got != "gpt-4o" {
		t.Errorf("config get models.phases.apply = %q, want %q", got, "gpt-4o")
	}
}

func TestConfigCmd_GetMissingKey(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "get", "models.phases.verify"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	got := strings.TrimSpace(stdout.String())
	if got != "" {
		t.Errorf("config get models.phases.verify = %q, want empty", got)
	}
}

func TestConfigCmd_SetRoundtrip(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout1, stderr1 bytes.Buffer
	root1 := newRootCmd(&stdout1, &stderr1)
	root1.SetArgs([]string{"config", "set", "models.phases.verify", "o3"})

	if err := root1.Execute(); err != nil {
		t.Fatalf("set Execute() error = %v, stderr = %s", err, stderr1.String())
	}

	var stdout2, stderr2 bytes.Buffer
	root2 := newRootCmd(&stdout2, &stderr2)
	root2.SetArgs([]string{"config", "get", "models.phases.verify"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("get Execute() error = %v, stderr = %s", err, stderr2.String())
	}

	got := strings.TrimSpace(stdout2.String())
	if got != "o3" {
		t.Errorf("after set, get models.phases.verify = %q, want %q", got, "o3")
	}
}

func TestConfigCmd_SetUnknownModelWarning(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "set", "models.default", "future-model-v2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "warning") {
		t.Errorf("stderr = %q, want contains 'warning'", errOutput)
	}
}

func TestConfigCmd_List(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "models.default = claude-sonnet-4") {
		t.Errorf("list output = %q, want contains 'models.default = claude-sonnet-4'", output)
	}
	if !strings.Contains(output, "models.phases.apply = gpt-4o") {
		t.Errorf("list output = %q, want contains 'models.phases.apply = gpt-4o'", output)
	}
}

func TestConfigCmd_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	archonDir := filepath.Join(tmpDir, ".archon")
	if err := os.MkdirAll(archonDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configContent := `harness_version: "1.0.0"
agent: opencode
`
	if err := os.WriteFile(filepath.Join(archonDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output != "(none configured)" {
		t.Errorf("list output = %q, want %q", output, "(none configured)")
	}
}
