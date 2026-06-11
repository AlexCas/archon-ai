package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_InitStatusRollback(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", homeDir)

	opencodeDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	t.Run("init", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"init", "--agent", "opencode"})

		if err := root.Execute(); err != nil {
			t.Fatalf("init Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "initialized successfully") {
			t.Errorf("init output = %q, want 'initialized successfully'", output)
		}

		configPath := filepath.Join(tmpDir, ".archon", "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("config.yaml not created at %s", configPath)
		}

		rollbackPath := filepath.Join(tmpDir, ".archon", "rollback.json")
		if _, err := os.Stat(rollbackPath); os.IsNotExist(err) {
			t.Errorf("rollback.json not created at %s", rollbackPath)
		}

		agentsPath := filepath.Join(tmpDir, "AGENTS.md")
		if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
			t.Errorf("AGENTS.md not created at %s", agentsPath)
		}
	})

	t.Run("status", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"status"})

		if err := root.Execute(); err != nil {
			t.Fatalf("status Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "Archon Harness Status") {
			t.Errorf("status output = %q, want 'Archon Harness Status'", output)
		}
		if !strings.Contains(output, "opencode") {
			t.Errorf("status output = %q, want contains 'opencode'", output)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		root := newRootCmd(&stdout, &stderr)
		root.SetArgs([]string{"rollback"})

		if err := root.Execute(); err != nil {
			t.Fatalf("rollback Execute() error = %v, stderr = %s", err, stderr.String())
		}

		output := stdout.String()
		if !strings.Contains(output, "Rollback complete") {
			t.Errorf("rollback output = %q, want 'Rollback complete'", output)
		}

		configPath := filepath.Join(tmpDir, ".archon", "config.yaml")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Errorf("config.yaml still exists after rollback at %s", configPath)
		}

		rollbackPath := filepath.Join(tmpDir, ".archon", "rollback.json")
		if _, err := os.Stat(rollbackPath); !os.IsNotExist(err) {
			t.Errorf("rollback.json still exists after rollback at %s", rollbackPath)
		}
	})
}

func TestE2E_RollbackWithoutInit(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"rollback"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Nothing to rollback") {
		t.Errorf("rollback output = %q, want 'Nothing to rollback'", output)
	}
}

func TestE2E_StatusWithoutInit(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"status"})

	err := root.Execute()
	if err == nil {
		t.Error("Execute() should fail when not initialized")
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "No archon configuration found") {
		t.Errorf("stderr = %q, want 'No archon configuration found'", errOutput)
	}
}

func TestE2E_VersionOutput(t *testing.T) {
	var stdout bytes.Buffer
	root := newRootCmd(&stdout, &bytes.Buffer{})
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "archon version") {
		t.Errorf("version output = %q, want 'archon version'", output)
	}
}

func TestE2E_InitNoAgent(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", homeDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"init"})

	err := root.Execute()
	if err == nil {
		t.Error("Execute() should fail when no agent detected")
	}
}

func TestE2E_RollbackDryRun(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", homeDir)

	opencodeDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	var stdout1, stderr1 bytes.Buffer
	root1 := newRootCmd(&stdout1, &stderr1)
	root1.SetArgs([]string{"init", "--agent", "opencode"})
	if err := root1.Execute(); err != nil {
		t.Fatalf("init Execute() error = %v", err)
	}

	var stdout2, stderr2 bytes.Buffer
	root2 := newRootCmd(&stdout2, &stderr2)
	root2.SetArgs([]string{"rollback", "--dry-run"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("rollback --dry-run Execute() error = %v", err)
	}

	output := stdout2.String()
	if !strings.Contains(output, "Dry run") {
		t.Errorf("rollback --dry-run output = %q, want 'Dry run'", output)
	}

	configPath := filepath.Join(tmpDir, ".archon", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.yaml should still exist after dry-run rollback at %s", configPath)
	}
}
