package initcmd

import (
	"errors"
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

// TestRun_StillWritesTemplateConfigRollback guards against the refreshSkills
// refactor regressing init's contract: Run must still write the orchestrator
// template, config, and rollback manifest after the skill-side work moved into
// refreshSkills.
func TestRun_StillWritesTemplateConfigRollback(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"2.0\"\n---\n# Init"),
		},
		"sdd-propose/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-propose\nmetadata:\n  version: \"3.0\"\n---\n# Propose"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Agent:      "claude",
		EmbeddedFS: embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Template, config, and rollback manifest must all exist.
	for _, p := range []string{
		filepath.Join(projectDir, "CLAUDE.md"),
		filepath.Join(projectDir, ".archon", "config.yaml"),
		filepath.Join(projectDir, ".archon", "rollback.json"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist after Run: %v", p, err)
		}
	}
}

// TestBuildConfig_SecurityFlag covers spec scenario "Init flag enables the gate"
// and "Init without flag leaves security off". Same for the pre-existing
// Playwright gap: both flags must be faithfully forwarded to the config.
func TestBuildConfig_SecurityFlag(t *testing.T) {
	for _, tt := range []struct {
		name       string
		security   bool
		playwright bool
	}{
		{"security on, playwright on", true, true},
		{"security off, playwright off", false, false},
		{"security on, playwright off", true, false},
		{"security off, playwright on", false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := buildConfig("opencode", nil, nil, "", "", nil, tt.playwright, tt.security)
			if cfg.Security.Enabled != tt.security {
				t.Errorf("Security.Enabled = %v, want %v", cfg.Security.Enabled, tt.security)
			}
			if cfg.Playwright.Enabled != tt.playwright {
				t.Errorf("Playwright.Enabled = %v, want %v", cfg.Playwright.Enabled, tt.playwright)
			}
		})
	}
}

// TestRun_RecordsRealFrontmatterVersions backs the harness-init spec
// "Init records real frontmatter versions": the inventory must carry each
// skill's real metadata.version, never the legacy hardcoded "1.0".
func TestRun_RecordsRealFrontmatterVersions(t *testing.T) {
	tmpDir := t.TempDir()

	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"2.0\"\n---\n# Init"),
		},
		"sdd-propose/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-propose\nmetadata:\n  version: \"3.0\"\n---\n# Propose"),
		},
		// A skill with no metadata.version must still be recorded, with an
		// empty version, without aborting init (spec @edge).
		"sdd-noversion/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-noversion\n---\n# No version"),
		},
	}

	opts := Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Agent:      "claude",
		EmbeddedFS: embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var cfg config.Config
	if err := cfg.Load(os.DirFS(projectDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.SkillInventory) != 3 {
		t.Fatalf("SkillInventory has %d entries, want 3 (incl. the versionless skill)", len(cfg.SkillInventory))
	}

	versions := make(map[string]string, len(cfg.SkillInventory))
	for _, inv := range cfg.SkillInventory {
		if inv.Version == "1.0" {
			t.Errorf("inventory entry %q uses the hardcoded legacy version %q", inv.Name, inv.Version)
		}
		if inv.Source != "embedded" {
			t.Errorf("inventory entry %q source = %q, want embedded", inv.Name, inv.Source)
		}
		versions[inv.Name] = inv.Version
	}

	if versions["sdd-init"] != "2.0" {
		t.Errorf("sdd-init version = %q, want 2.0", versions["sdd-init"])
	}
	if versions["sdd-propose"] != "3.0" {
		t.Errorf("sdd-propose version = %q, want 3.0", versions["sdd-propose"])
	}
	if versions["sdd-noversion"] != "" {
		t.Errorf("sdd-noversion version = %q, want empty string", versions["sdd-noversion"])
	}
}
