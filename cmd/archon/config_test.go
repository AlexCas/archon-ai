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

// TestConfigCmd_SetProviderQualified confirms the ParseModelRef + FullID seam
// end-to-end: setting a provider-qualified value and getting it back returns the
// same string. (S1f-1, optional)
func TestConfigCmd_SetProviderQualified(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout1, stderr1 bytes.Buffer
	root1 := newRootCmd(&stdout1, &stderr1)
	root1.SetArgs([]string{"config", "set", "models.default", "opencode/deepseek-v4-pro"})

	if err := root1.Execute(); err != nil {
		t.Fatalf("set Execute() error = %v, stderr = %s", err, stderr1.String())
	}

	var stdout2, stderr2 bytes.Buffer
	root2 := newRootCmd(&stdout2, &stderr2)
	root2.SetArgs([]string{"config", "get", "models.default"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("get Execute() error = %v, stderr = %s", err, stderr2.String())
	}

	got := strings.TrimSpace(stdout2.String())
	if got != "opencode/deepseek-v4-pro" {
		t.Errorf("config get models.default = %q, want %q", got, "opencode/deepseek-v4-pro")
	}
}

// TestConfigCmd_SecuritySetGet asserts set then get roundtrip for security.enabled
// and security.profile (S1-7).
func TestConfigCmd_SecuritySetGet(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set security.enabled = true
	var stdout1, stderr1 bytes.Buffer
	root1 := newRootCmd(&stdout1, &stderr1)
	root1.SetArgs([]string{"config", "set", "security.enabled", "true"})
	if err := root1.Execute(); err != nil {
		t.Fatalf("set security.enabled Execute() error = %v, stderr = %s", err, stderr1.String())
	}

	// Get security.enabled
	var stdout2, stderr2 bytes.Buffer
	root2 := newRootCmd(&stdout2, &stderr2)
	root2.SetArgs([]string{"config", "get", "security.enabled"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("get security.enabled Execute() error = %v, stderr = %s", err, stderr2.String())
	}
	if got := strings.TrimSpace(stdout2.String()); got != "true" {
		t.Errorf("security.enabled = %q, want %q", got, "true")
	}

	// Set security.profile = cli
	var stdout3, stderr3 bytes.Buffer
	root3 := newRootCmd(&stdout3, &stderr3)
	root3.SetArgs([]string{"config", "set", "security.profile", "cli"})
	if err := root3.Execute(); err != nil {
		t.Fatalf("set security.profile cli Execute() error = %v, stderr = %s", err, stderr3.String())
	}

	// Get security.profile
	var stdout4, stderr4 bytes.Buffer
	root4 := newRootCmd(&stdout4, &stderr4)
	root4.SetArgs([]string{"config", "get", "security.profile"})
	if err := root4.Execute(); err != nil {
		t.Fatalf("get security.profile Execute() error = %v, stderr = %s", err, stderr4.String())
	}
	if got := strings.TrimSpace(stdout4.String()); got != "cli" {
		t.Errorf("security.profile = %q, want %q", got, "cli")
	}

	// Set security.profile = web (valid)
	var stdout5, stderr5 bytes.Buffer
	root5 := newRootCmd(&stdout5, &stderr5)
	root5.SetArgs([]string{"config", "set", "security.profile", "web"})
	if err := root5.Execute(); err != nil {
		t.Fatalf("set security.profile web Execute() error = %v, stderr = %s", err, stderr5.String())
	}
	var stdout6, stderr6 bytes.Buffer
	root6 := newRootCmd(&stdout6, &stderr6)
	root6.SetArgs([]string{"config", "get", "security.profile"})
	if err := root6.Execute(); err != nil {
		t.Fatalf("get security.profile Execute() error = %v, stderr = %s", err, stderr6.String())
	}
	if got := strings.TrimSpace(stdout6.String()); got != "web" {
		t.Errorf("security.profile = %q, want %q", got, "web")
	}
}

// TestConfigCmd_SecurityProfileInvalidValues asserts that llm, agentic, and a
// garbage value are rejected with the exact profile error (S1-7).
func TestConfigCmd_SecurityProfileInvalidValues(t *testing.T) {
	invalid := []string{"llm", "agentic", "garbage"}

	for _, val := range invalid {
		t.Run(val, func(t *testing.T) {
			tmpDir := setupProjectWithConfig(t)
			origDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			var stdout, stderr bytes.Buffer
			root := newRootCmd(&stdout, &stderr)
			root.SetArgs([]string{"config", "set", "security.profile", val})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error for security.profile = %q, got none", val)
			}
			wantSubstr := "invalid profile"
			if !strings.Contains(err.Error(), wantSubstr) {
				t.Errorf("error = %q, want contains %q", err.Error(), wantSubstr)
			}
			wantSupported := "(supported: cli, web)"
			if !strings.Contains(err.Error(), wantSupported) {
				t.Errorf("error = %q, want contains %q", err.Error(), wantSupported)
			}
		})
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
