package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/status"
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

// TestConfigCmd_UnknownKeyListsImpeccableKeys asserts the unknown-key error
// contains all five impeccable.* keys (design §3.3).
func TestConfigCmd_UnknownKeyListsImpeccableKeys(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "get", "bogus.key"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown key, got none")
	}

	wantKeys := []string{
		"impeccable.enabled",
		"impeccable.auto_install",
		"impeccable.severity",
		"impeccable.product_path",
		"impeccable.design_path",
	}
	for _, k := range wantKeys {
		if !strings.Contains(err.Error(), k) {
			t.Errorf("error = %q, want contains %q", err.Error(), k)
		}
	}
}

// TestConfigCmd_ImpeccableSetGet asserts set/get roundtrip for all five
// impeccable.* keys, and that an invalid severity is rejected.
func TestConfigCmd_ImpeccableSetGet(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cases := []struct {
		key   string
		value string
	}{
		{"impeccable.enabled", "true"},
		{"impeccable.auto_install", "true"},
		{"impeccable.severity", "block-all"},
		{"impeccable.product_path", "docs/PRODUCT.md"},
		{"impeccable.design_path", "docs/DESIGN.md"},
	}

	for _, c := range cases {
		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", c.key, c.value})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set %s Execute() error = %v, stderr = %s", c.key, err, setErr.String())
		}

		var getOut, getErr bytes.Buffer
		getCmd := newRootCmd(&getOut, &getErr)
		getCmd.SetArgs([]string{"config", "get", c.key})
		if err := getCmd.Execute(); err != nil {
			t.Fatalf("get %s Execute() error = %v, stderr = %s", c.key, err, getErr.String())
		}

		if got := strings.TrimSpace(getOut.String()); got != c.value {
			t.Errorf("%s = %q, want %q", c.key, got, c.value)
		}
	}
}

// TestConfigCmd_ImpeccableSeverityInvalid asserts `config set
// impeccable.severity invalid` exits non-zero and names the value plus the
// three valid options.
func TestConfigCmd_ImpeccableSeverityInvalid(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "set", "impeccable.severity", "invalid"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid impeccable.severity, got none")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error = %q, want contains %q", err.Error(), "invalid")
	}
	for _, v := range []string{"block-deterministic", "block-all", "advisory"} {
		if !strings.Contains(err.Error(), v) {
			t.Errorf("error = %q, want contains valid option %q", err.Error(), v)
		}
	}
}

// TestConfigCmd_BaseURLSetGet covers REQ-2: set/get base_url for a phase and
// for models.default, and asserts the sibling provider/model fields are
// untouched by a base_url set.
func TestConfigCmd_BaseURLSetGet(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Scenario: Set and get base_url for a phase.
	var setOut, setErr bytes.Buffer
	setCmd := newRootCmd(&setOut, &setErr)
	setCmd.SetArgs([]string{"config", "set", "models.phases.apply.base_url", "http://localhost:11434/v1"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("set models.phases.apply.base_url Execute() error = %v, stderr = %s", err, setErr.String())
	}

	var getOut, getErr bytes.Buffer
	getCmd := newRootCmd(&getOut, &getErr)
	getCmd.SetArgs([]string{"config", "get", "models.phases.apply.base_url"})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("get models.phases.apply.base_url Execute() error = %v, stderr = %s", err, getErr.String())
	}
	if got := strings.TrimSpace(getOut.String()); got != "http://localhost:11434/v1" {
		t.Errorf("models.phases.apply.base_url = %q, want %q", got, "http://localhost:11434/v1")
	}

	// The provider/model fields of the apply ref must be unchanged (fixture: gpt-4o).
	var modelOut, modelErr bytes.Buffer
	modelCmd := newRootCmd(&modelOut, &modelErr)
	modelCmd.SetArgs([]string{"config", "get", "models.phases.apply"})
	if err := modelCmd.Execute(); err != nil {
		t.Fatalf("get models.phases.apply Execute() error = %v, stderr = %s", err, modelErr.String())
	}
	if got := strings.TrimSpace(modelOut.String()); got != "gpt-4o" {
		t.Errorf("models.phases.apply = %q, want unchanged %q", got, "gpt-4o")
	}

	// Set and get models.default.base_url.
	var setDefOut, setDefErr bytes.Buffer
	setDefCmd := newRootCmd(&setDefOut, &setDefErr)
	setDefCmd.SetArgs([]string{"config", "set", "models.default.base_url", "http://localhost:8080/v1"})
	if err := setDefCmd.Execute(); err != nil {
		t.Fatalf("set models.default.base_url Execute() error = %v, stderr = %s", err, setDefErr.String())
	}

	var getDefOut, getDefErr bytes.Buffer
	getDefCmd := newRootCmd(&getDefOut, &getDefErr)
	getDefCmd.SetArgs([]string{"config", "get", "models.default.base_url"})
	if err := getDefCmd.Execute(); err != nil {
		t.Fatalf("get models.default.base_url Execute() error = %v, stderr = %s", err, getDefErr.String())
	}
	if got := strings.TrimSpace(getDefOut.String()); got != "http://localhost:8080/v1" {
		t.Errorf("models.default.base_url = %q, want %q", got, "http://localhost:8080/v1")
	}

	// The models.default model field must be unchanged (fixture: claude-sonnet-4).
	var defModelOut, defModelErr bytes.Buffer
	defModelCmd := newRootCmd(&defModelOut, &defModelErr)
	defModelCmd.SetArgs([]string{"config", "get", "models.default"})
	if err := defModelCmd.Execute(); err != nil {
		t.Fatalf("get models.default Execute() error = %v, stderr = %s", err, defModelErr.String())
	}
	if got := strings.TrimSpace(defModelOut.String()); got != "claude-sonnet-4" {
		t.Errorf("models.default = %q, want unchanged %q", got, "claude-sonnet-4")
	}
}

// TestConfigCmd_BaseURLPreservedOnModelSet covers the BaseURL-preservation
// invariant: setting a model-id AFTER a base_url has been set must not wipe
// the previously-set BaseURL. This is the primary regression guard for the
// ref.BaseURL = existing.BaseURL copy in setConfigValue.
func TestConfigCmd_BaseURLPreservedOnModelSet(t *testing.T) {
	t.Run("default: set base_url then set model preserves base_url", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		for _, args := range [][]string{
			{"config", "set", "models.default.base_url", "http://localhost:11434/v1"},
			{"config", "set", "models.default", "ollama/llama3"},
		} {
			var out, err bytes.Buffer
			cmd := newRootCmd(&out, &err)
			cmd.SetArgs(args)
			if e := cmd.Execute(); e != nil {
				t.Fatalf("set %v Execute() error = %v, stderr = %s", args, e, err.String())
			}
		}

		var out, err bytes.Buffer
		cmd := newRootCmd(&out, &err)
		cmd.SetArgs([]string{"config", "get", "models.default.base_url"})
		if e := cmd.Execute(); e != nil {
			t.Fatalf("get models.default.base_url Execute() error = %v, stderr = %s", e, err.String())
		}
		if got := strings.TrimSpace(out.String()); got != "http://localhost:11434/v1" {
			t.Errorf("models.default.base_url after model set = %q, want %q (base_url must be preserved)", got, "http://localhost:11434/v1")
		}
	})

	t.Run("leader: set base_url then set model preserves base_url", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		for _, args := range [][]string{
			{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"},
			{"config", "set", "models.leader", "ollama/llama3"},
		} {
			var out, err bytes.Buffer
			cmd := newRootCmd(&out, &err)
			cmd.SetArgs(args)
			if e := cmd.Execute(); e != nil {
				t.Fatalf("set %v Execute() error = %v, stderr = %s", args, e, err.String())
			}
		}

		var out, err bytes.Buffer
		cmd := newRootCmd(&out, &err)
		cmd.SetArgs([]string{"config", "get", "models.leader.base_url"})
		if e := cmd.Execute(); e != nil {
			t.Fatalf("get models.leader.base_url Execute() error = %v, stderr = %s", e, err.String())
		}
		if got := strings.TrimSpace(out.String()); got != "http://localhost:11434/v1" {
			t.Errorf("models.leader.base_url after model set = %q, want %q (base_url must be preserved)", got, "http://localhost:11434/v1")
		}
	})

	t.Run("phase: set base_url then set model preserves base_url", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		for _, args := range [][]string{
			{"config", "set", "models.phases.apply.base_url", "http://localhost:11434/v1"},
			{"config", "set", "models.phases.apply", "ollama/llama3"},
		} {
			var out, err bytes.Buffer
			cmd := newRootCmd(&out, &err)
			cmd.SetArgs(args)
			if e := cmd.Execute(); e != nil {
				t.Fatalf("set %v Execute() error = %v, stderr = %s", args, e, err.String())
			}
		}

		var out, err bytes.Buffer
		cmd := newRootCmd(&out, &err)
		cmd.SetArgs([]string{"config", "get", "models.phases.apply.base_url"})
		if e := cmd.Execute(); e != nil {
			t.Fatalf("get models.phases.apply.base_url Execute() error = %v, stderr = %s", e, err.String())
		}
		if got := strings.TrimSpace(out.String()); got != "http://localhost:11434/v1" {
			t.Errorf("models.phases.apply.base_url after model set = %q, want %q (base_url must be preserved)", got, "http://localhost:11434/v1")
		}
	})
}

// TestConfigCmd_BaseURLGetUnsetReturnsEmpty covers REQ-2: getting base_url on
// a ref with no BaseURL exits 0 and prints nothing.
func TestConfigCmd_BaseURLGetUnsetReturnsEmpty(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "get", "models.default.base_url"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	if got := strings.TrimSpace(stdout.String()); got != "" {
		t.Errorf("models.default.base_url = %q, want empty", got)
	}
}

// TestConfigCmd_ListShowsBaseURLLines covers REQ-2: `config list` includes a
// base_url line grouped with its sibling models.phases.<phase> line.
func TestConfigCmd_ListShowsBaseURLLines(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var setOut, setErr bytes.Buffer
	setCmd := newRootCmd(&setOut, &setErr)
	setCmd.SetArgs([]string{"config", "set", "models.phases.apply.base_url", "http://localhost:11434/v1"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "models.phases.apply.base_url = http://localhost:11434/v1") {
		t.Errorf("list output = %q, want contains %q", output, "models.phases.apply.base_url = http://localhost:11434/v1")
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

// setupEmptyProject returns a project with no models.* keys set at all, for
// tests that build up a base_url-only or leader-only fixture via `config set`.
func setupEmptyProject(t *testing.T) string {
	t.Helper()
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
	return tmpDir
}

// TestConfigCmd_ListBaseURLOnlyIsConfigured covers REQ-8: a ref with only a
// BaseURL (no provider/model) counts as configured on both surfaces.
func TestConfigCmd_ListBaseURLOnlyIsConfigured(t *testing.T) {
	t.Run("default base_url only", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.default.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if strings.Contains(output, "(none configured)") {
			t.Errorf("list output = %q, want NOT contains %q", output, "(none configured)")
		}
		if !strings.Contains(output, "models.default.base_url = http://localhost:11434/v1") {
			t.Errorf("list output = %q, want contains %q", output, "models.default.base_url = http://localhost:11434/v1")
		}
	})

	t.Run("phase base_url only", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.phases.apply.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if strings.Contains(output, "(none configured)") {
			t.Errorf("list output = %q, want NOT contains %q", output, "(none configured)")
		}
		if !strings.Contains(output, "models.phases.apply.base_url = http://localhost:11434/v1") {
			t.Errorf("list output = %q, want contains %q", output, "models.phases.apply.base_url = http://localhost:11434/v1")
		}
	})

	t.Run("leader base_url only", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if strings.Contains(output, "(none configured)") {
			t.Errorf("list output = %q, want NOT contains %q", output, "(none configured)")
		}
		if !strings.Contains(output, "models.leader.base_url = http://localhost:11434/v1") {
			t.Errorf("list output = %q, want contains %q", output, "models.leader.base_url = http://localhost:11434/v1")
		}
	})
}

// TestConfigCmd_ListSuppressesEmptyPrimary covers REQ-9: the primary
// `models.X = ` line is omitted when FullID() is empty, while the base_url
// line is always printed when set.
func TestConfigCmd_ListSuppressesEmptyPrimary(t *testing.T) {
	t.Run("base_url-only default suppresses primary line", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.default.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "models.default.base_url = ") {
			t.Errorf("list output = %q, want contains %q", output, "models.default.base_url = ")
		}
		if strings.Contains(output, "models.default = ") {
			t.Errorf("list output = %q, want NOT contains %q", output, "models.default = ")
		}
	})

	t.Run("phase with both id and base_url emits both lines", func(t *testing.T) {
		tmpDir := setupProjectWithConfig(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.phases.apply.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "models.phases.apply = gpt-4o") {
			t.Errorf("list output = %q, want contains %q", output, "models.phases.apply = gpt-4o")
		}
		if !strings.Contains(output, "models.phases.apply.base_url = http://localhost:11434/v1") {
			t.Errorf("list output = %q, want contains %q", output, "models.phases.apply.base_url = http://localhost:11434/v1")
		}
	})

	t.Run("leader base_url-only suppresses primary line", func(t *testing.T) {
		tmpDir := setupEmptyProject(t)
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs([]string{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"})
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set Execute() error = %v, stderr = %s", err, setErr.String())
		}

		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"config", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "models.leader.base_url = ") {
			t.Errorf("list output = %q, want contains %q", output, "models.leader.base_url = ")
		}
		if strings.Contains(output, "models.leader = ") {
			t.Errorf("list output = %q, want NOT contains %q", output, "models.leader = ")
		}
	})
}

// TestConfigCmd_LeaderSetGet covers REQ-11: set/get round-trip for
// models.leader and models.leader.base_url, with cross-field isolation.
func TestConfigCmd_LeaderSetGet(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var setOut, setErr bytes.Buffer
	setCmd := newRootCmd(&setOut, &setErr)
	setCmd.SetArgs([]string{"config", "set", "models.leader", "ollama/llama3"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("set models.leader Execute() error = %v, stderr = %s", err, setErr.String())
	}

	var getOut, getErr bytes.Buffer
	getCmd := newRootCmd(&getOut, &getErr)
	getCmd.SetArgs([]string{"config", "get", "models.leader"})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("get models.leader Execute() error = %v, stderr = %s", err, getErr.String())
	}
	if got := strings.TrimSpace(getOut.String()); got != "ollama/llama3" {
		t.Errorf("models.leader = %q, want %q", got, "ollama/llama3")
	}

	var setURLOut, setURLErr bytes.Buffer
	setURLCmd := newRootCmd(&setURLOut, &setURLErr)
	setURLCmd.SetArgs([]string{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"})
	if err := setURLCmd.Execute(); err != nil {
		t.Fatalf("set models.leader.base_url Execute() error = %v, stderr = %s", err, setURLErr.String())
	}

	var getURLOut, getURLErr bytes.Buffer
	getURLCmd := newRootCmd(&getURLOut, &getURLErr)
	getURLCmd.SetArgs([]string{"config", "get", "models.leader.base_url"})
	if err := getURLCmd.Execute(); err != nil {
		t.Fatalf("get models.leader.base_url Execute() error = %v, stderr = %s", err, getURLErr.String())
	}
	if got := strings.TrimSpace(getURLOut.String()); got != "http://localhost:11434/v1" {
		t.Errorf("models.leader.base_url = %q, want %q", got, "http://localhost:11434/v1")
	}

	// Setting the base_url must not have altered the provider/model fields.
	var getIDOut, getIDErr bytes.Buffer
	getIDCmd := newRootCmd(&getIDOut, &getIDErr)
	getIDCmd.SetArgs([]string{"config", "get", "models.leader"})
	if err := getIDCmd.Execute(); err != nil {
		t.Fatalf("get models.leader Execute() error = %v, stderr = %s", err, getIDErr.String())
	}
	if got := strings.TrimSpace(getIDOut.String()); got != "ollama/llama3" {
		t.Errorf("models.leader after base_url set = %q, want unchanged %q", got, "ollama/llama3")
	}
}

// TestConfigCmd_ListShowsLeaderBlock covers REQ-11: `config list` renders the
// leader id and base_url lines, ordered before any models.phases. line.
func TestConfigCmd_ListShowsLeaderBlock(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	for _, args := range [][]string{
		{"config", "set", "models.leader", "ollama/llama3"},
		{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"},
	} {
		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs(args)
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set %v Execute() error = %v, stderr = %s", args, err, setErr.String())
		}
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"config", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("list Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "models.leader = ollama/llama3") {
		t.Errorf("list output = %q, want contains %q", output, "models.leader = ollama/llama3")
	}
	if !strings.Contains(output, "models.leader.base_url = http://localhost:11434/v1") {
		t.Errorf("list output = %q, want contains %q", output, "models.leader.base_url = http://localhost:11434/v1")
	}

	defaultIdx := strings.Index(output, "models.default = ")
	leaderIdx := strings.Index(output, "models.leader = ")
	phaseIdx := strings.Index(output, "models.phases.")
	if defaultIdx == -1 || leaderIdx == -1 || phaseIdx == -1 || !(defaultIdx < leaderIdx && leaderIdx < phaseIdx) {
		t.Errorf("list output = %q, want order: models.default < models.leader < models.phases.", output)
	}
}

// TestConfigCmd_LeaderBaseURLAdvisory covers REQ-11: leader base_url
// validation is advisory-only — an invalid scheme warns but never blocks.
func TestConfigCmd_LeaderBaseURLAdvisory(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var setOut, setErr bytes.Buffer
	setCmd := newRootCmd(&setOut, &setErr)
	setCmd.SetArgs([]string{"config", "set", "models.leader.base_url", "ftp://bad-url"})
	if err := setCmd.Execute(); err != nil {
		t.Fatalf("set models.leader.base_url Execute() error = %v (want exit 0), stderr = %s", err, setErr.String())
	}
	if !strings.Contains(setErr.String(), "warning") {
		t.Errorf("stderr = %q, want contains 'warning'", setErr.String())
	}

	var getOut, getErr bytes.Buffer
	getCmd := newRootCmd(&getOut, &getErr)
	getCmd.SetArgs([]string{"config", "get", "models.leader.base_url"})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("get models.leader.base_url Execute() error = %v, stderr = %s", err, getErr.String())
	}
	if got := strings.TrimSpace(getOut.String()); got != "ftp://bad-url" {
		t.Errorf("models.leader.base_url = %q, want %q (stored despite advisory warning)", got, "ftp://bad-url")
	}
}

// TestConfigCmd_LeaderUnknownKey covers REQ-11: an unrecognized
// models.leader.* key is rejected on both set and get, with the new leader
// keys listed in the supported-keys error message.
func TestConfigCmd_LeaderUnknownKey(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var setOut, setErr bytes.Buffer
	setCmd := newRootCmd(&setOut, &setErr)
	setCmd.SetArgs([]string{"config", "set", "models.leader.typo", "somevalue"})
	setErrResult := setCmd.Execute()
	if setErrResult == nil {
		t.Fatal("set models.leader.typo: expected error, got none")
	}
	if !strings.Contains(setErrResult.Error(), "models.leader, models.leader.base_url") {
		t.Errorf("set error = %q, want contains %q", setErrResult.Error(), "models.leader, models.leader.base_url")
	}

	var getOut, getErr bytes.Buffer
	getCmd := newRootCmd(&getOut, &getErr)
	getCmd.SetArgs([]string{"config", "get", "models.leader.typo"})
	getErrResult := getCmd.Execute()
	if getErrResult == nil {
		t.Fatal("get models.leader.typo: expected error, got none")
	}
	if !strings.Contains(getErrResult.Error(), "models.leader, models.leader.base_url") {
		t.Errorf("get error = %q, want contains %q", getErrResult.Error(), "models.leader, models.leader.base_url")
	}
}

// TestConfigCmd_LeaderSymmetry covers REQ-12's cross-surface invariant:
// `config list` and `status.Format` must agree on the leader id and base_url
// values (Invariant 4).
func TestConfigCmd_LeaderSymmetry(t *testing.T) {
	tmpDir := setupProjectWithConfig(t)
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	for _, args := range [][]string{
		{"config", "set", "models.leader", "ollama/llama3"},
		{"config", "set", "models.leader.base_url", "http://localhost:11434/v1"},
	} {
		var setOut, setErr bytes.Buffer
		setCmd := newRootCmd(&setOut, &setErr)
		setCmd.SetArgs(args)
		if err := setCmd.Execute(); err != nil {
			t.Fatalf("set %v Execute() error = %v, stderr = %s", args, err, setErr.String())
		}
	}

	var listOut, listErr bytes.Buffer
	listCmd := newRootCmd(&listOut, &listErr)
	listCmd.SetArgs([]string{"config", "list"})
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("list Execute() error = %v, stderr = %s", err, listErr.String())
	}
	listOutput := listOut.String()

	cfg := &config.Config{HomeDir: tmpDir}
	if err := cfg.Load(os.DirFS(tmpDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	statusOutput := status.Format(cfg)

	if !strings.Contains(listOutput, "models.leader = ollama/llama3") {
		t.Errorf("list output = %q, want contains %q", listOutput, "models.leader = ollama/llama3")
	}
	if !strings.Contains(listOutput, "models.leader.base_url = http://localhost:11434/v1") {
		t.Errorf("list output = %q, want contains %q", listOutput, "models.leader.base_url = http://localhost:11434/v1")
	}
	if !strings.Contains(statusOutput, "Leader:") {
		t.Errorf("status output = %q, want contains %q", statusOutput, "Leader:")
	}
	if !strings.Contains(statusOutput, "ollama/llama3") {
		t.Errorf("status output = %q, want contains %q", statusOutput, "ollama/llama3")
	}
	if !strings.Contains(statusOutput, "http://localhost:11434/v1") {
		t.Errorf("status output = %q, want contains %q", statusOutput, "http://localhost:11434/v1")
	}
}
