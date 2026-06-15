package initcmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
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

	// OverwriteTemplate mirrors a confirmed/forced re-init: re-running must
	// remain idempotent once the user has authorized replacing the file.
	opts := Options{
		HomeDir:           homeDir,
		ProjectDir:        projectDir,
		EmbeddedFS:        embeddedFS,
		OverwriteTemplate: true,
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

func TestRun_AbortsOnExistingTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Pre-existing CLAUDE.md with hand-written content.
	claudeMD := filepath.Join(projectDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMD, []byte("# my custom orchestrator\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: sdd-init\n---\n# Init")},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Agent:      "claude",
		EmbeddedFS: embeddedFS,
	}

	_, err := Run(opts)
	if !errors.Is(err, ErrTemplateExists) {
		t.Fatalf("Run() error = %v, want ErrTemplateExists", err)
	}

	// The original file must be untouched and no .archon dir created.
	data, _ := os.ReadFile(claudeMD)
	if string(data) != "# my custom orchestrator\n" {
		t.Errorf("CLAUDE.md was modified: %q", string(data))
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".archon")); err == nil {
		t.Errorf(".archon dir should not be created when init is aborted")
	}

	// With OverwriteTemplate, init proceeds and replaces the file.
	opts.OverwriteTemplate = true
	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() with overwrite error = %v", err)
	}
	data, _ = os.ReadFile(claudeMD)
	if string(data) == "# my custom orchestrator\n" {
		t.Errorf("CLAUDE.md should have been replaced")
	}
}

func TestRun_CreatesAgentDirWhenMissing(t *testing.T) {
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
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: sdd-init\n---\n# Init")},
	}

	// No agent folder exists; selecting claude must create .claude.
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
		t.Errorf("Agent = %q, want claude", result.Agent)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err != nil {
		t.Errorf(".claude dir was not created: %v", err)
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
