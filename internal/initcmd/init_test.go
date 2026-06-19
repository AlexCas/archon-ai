package initcmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/archon-ai/archon/internal/config"
)

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
		"sdd-propose/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-propose\n---\n# Propose"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Agent != "opencode" {
		t.Errorf("Agent = %q, want %q", result.Agent, "opencode")
	}

	if result.ExtractedCount != 2 {
		t.Errorf("ExtractedCount = %d, want %d", result.ExtractedCount, 2)
	}

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file not created: %s", configPath)
	}

	rollbackPath := filepath.Join(projectDir, ".archon", "rollback.json")
	if _, err := os.Stat(rollbackPath); os.IsNotExist(err) {
		t.Errorf("Rollback manifest not created: %s", rollbackPath)
	}

	agentsPath := filepath.Join(projectDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		t.Errorf("AGENTS.md not created: %s", agentsPath)
	}

	openspecDir := filepath.Join(projectDir, "openspec", "changes")
	if _, err := os.Stat(openspecDir); os.IsNotExist(err) {
		t.Errorf("openspec/changes directory not created: %s", openspecDir)
	}
}

func TestRun_ClaudeAgent(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", result.Agent, "claude")
	}

	claudeMD := filepath.Join(projectDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMD); os.IsNotExist(err) {
		t.Errorf("CLAUDE.md not created: %s", claudeMD)
	}
}

func TestRun_WithAgentFlag(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Agent:      "claude",
		EmbeddedFS: embeddedFS,
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", result.Agent, "claude")
	}
}

func TestRun_MissingHomeDir(t *testing.T) {
	opts := Options{
		ProjectDir: "/tmp/project",
		EmbeddedFS: fstest.MapFS{},
	}

	_, err := Run(opts)
	if err == nil {
		t.Error("Run() should fail with missing HomeDir")
	}
}

func TestRun_MissingProjectDir(t *testing.T) {
	opts := Options{
		HomeDir:    "/tmp/home",
		EmbeddedFS: fstest.MapFS{},
	}

	_, err := Run(opts)
	if err == nil {
		t.Error("Run() should fail with missing ProjectDir")
	}
}

func TestRun_MissingEmbeddedFS(t *testing.T) {
	opts := Options{
		HomeDir:    "/tmp/home",
		ProjectDir: "/tmp/project",
	}

	_, err := Run(opts)
	if err == nil {
		t.Error("Run() should fail with missing EmbeddedFS")
	}
}

func TestRun_NoAgentDetected(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}

	_, err := Run(opts)
	if err == nil {
		t.Error("Run() should fail when no agent detected and no flag provided")
	}
}

func TestRun_Idempotency(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}

	_, err := Run(opts)
	if err != nil {
		t.Fatalf("First Run() error = %v", err)
	}

	_, err = Run(opts)
	if err != nil {
		t.Fatalf("Second Run() error = %v", err)
	}
}

func TestRun_WithModelFlags(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:      homeDir,
		ProjectDir:   projectDir,
		EmbeddedFS:   embeddedFS,
		ModelDefault: "claude-sonnet-4",
		ModelPhases: map[string]string{
			"apply": "gpt-4o",
		},
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Agent != "opencode" {
		t.Errorf("Agent = %q, want %q", result.Agent, "opencode")
	}

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "claude-sonnet-4") {
		t.Errorf("config should contain default model, got:\n%s", content)
	}
	if !strings.Contains(content, "gpt-4o") {
		t.Errorf("config should contain apply phase model, got:\n%s", content)
	}

	// Hermeticity: opencode.json must be written under homeDir (the temp dir),
	// NOT under the real $HOME. Verify the file exists at the expected temp path
	// and that the real home's opencode.json was not touched.
	expectedSettings := filepath.Join(homeDir, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(expectedSettings); os.IsNotExist(err) {
		t.Errorf("opencode.json not created at temp homeDir path %s", expectedSettings)
	}

	// Parse the settings file and verify model assignments landed correctly.
	rawSettings, err := os.ReadFile(expectedSettings)
	if err != nil {
		t.Fatalf("ReadFile opencode.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(rawSettings, &settings); err != nil {
		t.Fatalf("Unmarshal opencode.json: %v", err)
	}
	agents, ok := settings["agent"].(map[string]any)
	if !ok {
		t.Fatal("opencode.json: agent key missing or not an object")
	}
	sddApply, ok := agents["sdd-apply"].(map[string]any)
	if !ok {
		t.Fatal("opencode.json: sdd-apply agent not found")
	}
	if sddApply["model"] != "openai/gpt-4o" {
		t.Errorf("sdd-apply.model = %v, want openai/gpt-4o", sddApply["model"])
	}
	sddExplore, ok := agents["sdd-explore"].(map[string]any)
	if !ok {
		t.Fatal("opencode.json: sdd-explore agent not found")
	}
	if sddExplore["model"] != "anthropic/claude-sonnet-4" {
		t.Errorf("sdd-explore.model = %v, want anthropic/claude-sonnet-4 (static map fallback with no cache)", sddExplore["model"])
	}
}

// TestRun_OpenCodeOverlay_Hermetic verifies that Run never writes to or reads
// from the real $HOME when opts.HomeDir points to a temp directory. This is
// the regression test for the non-hermetic Apply defect: previously,
// opencode.Apply was called with SettingsPath() / CachePath() which resolved
// against the real os.UserHomeDir(), silently writing to the developer's live
// ~/.config/opencode/opencode.json during test runs.
func TestRun_OpenCodeOverlay_Hermetic(t *testing.T) {
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine real home dir; skipping hermeticity check")
	}
	realSettings := filepath.Join(realHome, ".config", "opencode", "opencode.json")

	// Record the real settings file state BEFORE the test run.
	realBefore, _ := os.ReadFile(realSettings)
	realBackupsBefore := countOpenCodeBackups(t, realHome)

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	for _, d := range []string{
		homeDir,
		projectDir,
		filepath.Join(projectDir, ".opencode"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	_, runErr := Run(Options{
		HomeDir:      homeDir,
		ProjectDir:   projectDir,
		EmbeddedFS:   embeddedFS,
		ModelDefault: "claude-sonnet-4",
	})
	if runErr != nil {
		t.Fatalf("Run() error = %v", runErr)
	}

	// The overlay MUST appear under the temp homeDir, not the real home.
	tempSettings := filepath.Join(homeDir, ".config", "opencode", "opencode.json")
	if _, err := os.Stat(tempSettings); os.IsNotExist(err) {
		t.Errorf("opencode.json not written to temp homeDir at %s", tempSettings)
	}

	// The real settings file must be byte-for-byte identical to what it was before.
	realAfter, _ := os.ReadFile(realSettings)
	if string(realBefore) != string(realAfter) {
		t.Errorf("real $HOME opencode.json was modified by test run (hermeticity violation!)")
	}

	// No new .backup.* files must appear in the real home.
	realBackupsAfter := countOpenCodeBackups(t, realHome)
	if realBackupsAfter != realBackupsBefore {
		t.Errorf("test run created %d new backup file(s) in real $HOME (hermeticity violation!)",
			realBackupsAfter-realBackupsBefore)
	}
}

// countOpenCodeBackups returns the number of opencode.json.backup.* files
// in the real user's ~/.config/opencode/ directory.
func countOpenCodeBackups(t *testing.T, homeDir string) int {
	t.Helper()
	dir := filepath.Join(homeDir, ".config", "opencode")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "opencode.json.backup.") {
			count++
		}
	}
	return count
}

func TestRun_WithoutModelFlags(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}

	_, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if strings.Contains(content, "models:") {
		t.Errorf("config should not contain models section when no flags set, got:\n%s", content)
	}
}

// TestRun_OpenCode_RollbackManifest_FreshFile verifies Fix 3: when opencode.json
// did not exist before init, the rollback manifest must contain a FileBackup with
// Backup=="" for that file. Rollback must then remove the file (no prior content).
func TestRun_OpenCode_RollbackManifest_FreshFile(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	for _, d := range []string{
		homeDir,
		projectDir,
		filepath.Join(projectDir, ".opencode"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	_, err := Run(Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	settingsPath := filepath.Join(homeDir, ".config", "opencode", "opencode.json")

	// opencode.json must have been created.
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatalf("opencode.json not created at %s", settingsPath)
	}

	// Rollback manifest must contain a FileBackup for opencode.json with Backup=="".
	manifest, err := config.LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	found := false
	for _, fb := range manifest.FileBackups {
		if fb.Target == settingsPath {
			found = true
			if fb.Backup != "" {
				t.Errorf("FileBackup.Backup should be empty for a freshly-created opencode.json, got %q", fb.Backup)
			}
			break
		}
	}
	if !found {
		t.Errorf("rollback manifest has no FileBackup for opencode.json target %s", settingsPath)
	}

	// Calling Cleanup must remove the freshly-created opencode.json.
	manifest.HomeDir = projectDir
	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Error("opencode.json should be removed by rollback when no prior file existed")
	}
}

// TestRun_OpenCode_RollbackManifest_PriorFile verifies Fix 3 for the case where
// opencode.json existed before init: after init, the rollback manifest has a
// FileBackup with a non-empty Backup path; rollback restores the prior bytes.
func TestRun_OpenCode_RollbackManifest_PriorFile(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	for _, d := range []string{
		homeDir,
		projectDir,
		filepath.Join(projectDir, ".opencode"),
		filepath.Join(homeDir, ".config", "opencode"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}

	settingsPath := filepath.Join(homeDir, ".config", "opencode", "opencode.json")
	priorContent := []byte(`{"prior": true}`)
	if err := os.WriteFile(settingsPath, priorContent, 0o644); err != nil {
		t.Fatalf("WriteFile prior opencode.json: %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\n---\n# Init"),
		},
	}

	_, err := Run(Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Rollback manifest must contain a FileBackup with a non-empty Backup path.
	manifest, err := config.LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	var fb *config.FileBackup
	for i := range manifest.FileBackups {
		if manifest.FileBackups[i].Target == settingsPath {
			fb = &manifest.FileBackups[i]
			break
		}
	}
	if fb == nil {
		t.Fatalf("rollback manifest has no FileBackup for opencode.json target %s", settingsPath)
	}
	if fb.Backup == "" {
		t.Error("FileBackup.Backup must be non-empty when a prior opencode.json existed")
	}

	// Cleanup must restore the prior bytes.
	manifest.HomeDir = projectDir
	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	restored, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile after Cleanup: %v", err)
	}
	if string(restored) != string(priorContent) {
		t.Errorf("Cleanup restored %q, want %q", string(restored), string(priorContent))
	}
}
