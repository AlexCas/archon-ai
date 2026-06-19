package initcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/archon-ai/archon/internal/config"
)

// updateEmbeddedFS returns a small embedded skill set for update tests.
func updateEmbeddedFS() fstest.MapFS {
	return fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"2.0\"\n---\n# Init"),
		},
		"sdd-propose/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: sdd-propose\nmetadata:\n  version: \"3.0\"\n---\n# Propose"),
		},
	}
}

// setupInitializedProject runs init so update has a real config + installed
// skills to operate on. It returns homeDir and projectDir.
func setupInitializedProject(t *testing.T, embeddedFS fstest.MapFS) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if _, err := Run(Options{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Agent:      "claude",
		EmbeddedFS: embeddedFS,
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	return homeDir, projectDir
}

// S8: "Update leaves the opencode agent untouched". An opencode project with an
// existing opencode.json (agent.archon-leader) must have that file left
// unwritten and unchanged after archon update, since update is skill-only and
// never calls the merge.
func TestUpdate_LeavesOpencodeJSONUntouched(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	for _, d := range []string{homeDir, projectDir, filepath.Join(projectDir, ".opencode")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", d, err)
		}
	}

	// Initialize an opencode project that writes opencode.json via the merge.
	if _, err := Run(Options{
		HomeDir:     homeDir,
		ProjectDir:  projectDir,
		Agent:       "opencode",
		ModelLeader: "anthropic/claude-sonnet-4-20250514",
		EmbeddedFS:  embeddedFS,
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	opencodePath := filepath.Join(projectDir, "opencode.json")
	before, err := os.ReadFile(opencodePath)
	if err != nil {
		t.Fatalf("opencode.json not created by init: %v", err)
	}
	infoBefore, err := os.Stat(opencodePath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Introduce a real skill gap so update actually does work and writes skills.
	embeddedFS["sdd-init/SKILL.md"] = &fstest.MapFile{
		Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"9.0\"\n---\n# Init"),
	}

	if _, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	after, err := os.ReadFile(opencodePath)
	if err != nil {
		t.Fatalf("opencode.json missing after update: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("opencode.json was rewritten by update:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	infoAfter, err := os.Stat(opencodePath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !infoAfter.ModTime().Equal(infoBefore.ModTime()) {
		t.Errorf("opencode.json mtime changed: before %v, after %v", infoBefore.ModTime(), infoAfter.ModTime())
	}
}

// Scenario: "Update before init reports an actionable error".
func TestUpdate_BeforeInitReportsActionableError(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	_, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: updateEmbeddedFS(),
	})
	if err == nil {
		t.Fatal("Update() should fail when no config exists")
	}
	if !strings.Contains(err.Error(), "archon init") {
		t.Errorf("error = %q, want actionable hint to run 'archon init'", err.Error())
	}

	// Nothing must be written.
	if _, statErr := os.Stat(filepath.Join(projectDir, ".archon")); statErr == nil {
		t.Error(".archon dir should not be created when update fails before init")
	}
}

// Scenario: "No gaps reports already up to date".
func TestUpdate_NoGapsReportsAlreadyUpToDate(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	result, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !result.UpToDate {
		t.Errorf("UpToDate = false, want true when embedded matches installed")
	}
	if result.Wrote {
		t.Errorf("Wrote = true, want false when already up to date")
	}

	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(before) != string(after) {
		t.Error("config was modified when already up to date")
	}
}

// Scenario: "Check reports the diff without writing".
func TestUpdate_CheckReportsDiffWithoutWriting(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	// Introduce a gap: bump an embedded version so sdd-init is "changed".
	embeddedFS["sdd-init/SKILL.md"] = &fstest.MapFile{
		Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"9.0\"\n---\n# Init"),
	}

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	globalSkill := filepath.Join(homeDir, ".config", "opencode", "skills", "sdd-init", "SKILL.md")
	globalBefore, err := os.ReadFile(globalSkill)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	result, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Check:      true,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(result.GapReport.Changed) != 1 {
		t.Errorf("Changed = %d, want 1", len(result.GapReport.Changed))
	}
	if result.Wrote {
		t.Error("Wrote = true, want false for --check")
	}

	after, _ := os.ReadFile(configPath)
	if string(before) != string(after) {
		t.Error("config was modified under --check")
	}
	globalAfter, _ := os.ReadFile(globalSkill)
	if string(globalBefore) != string(globalAfter) {
		t.Error("global skill was re-extracted under --check")
	}
}

// Scenario: "Prune removes orphaned skills".
func TestUpdate_PruneRemovesOrphanedSkills(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	// Plant an orphan in the global dir (installed but no longer embedded).
	orphanDir := filepath.Join(homeDir, ".config", "opencode", "skills", "old-skill")
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("---\nname: old-skill\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Prune:      true,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(result.Pruned) != 1 || result.Pruned[0] != "old-skill" {
		t.Errorf("Pruned = %v, want [old-skill]", result.Pruned)
	}
	if _, statErr := os.Stat(orphanDir); statErr == nil {
		t.Error("orphan skill should have been removed under --prune")
	}
}

// Scenario: "Prune removes orphaned skills" under copy-mode. The project owns a
// real (non-symlink) copy of its skills, so --prune MUST remove the global
// orphan but MUST NOT delete the project's real orphan directory.
func TestUpdate_PruneCopyModeKeepsProjectRealDir(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	// Make the project copy-mode: replace at least one symlinked skill with a
	// real directory containing a SKILL.md so detectCopyMode trips.
	projectSkillsDir := filepath.Join(projectDir, ".claude", "skills")
	realSkill := filepath.Join(projectSkillsDir, "sdd-init")
	if err := os.RemoveAll(realSkill); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if err := os.MkdirAll(realSkill, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(realSkill, "SKILL.md"), []byte("---\nname: sdd-init\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Plant the orphan in BOTH the global dir and the project dir as a real copy.
	orphanGlobal := filepath.Join(homeDir, ".config", "opencode", "skills", "old-skill")
	if err := os.MkdirAll(orphanGlobal, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanGlobal, "SKILL.md"), []byte("---\nname: old-skill\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	orphanProject := filepath.Join(projectSkillsDir, "old-skill")
	if err := os.MkdirAll(orphanProject, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanProject, "SKILL.md"), []byte("---\nname: old-skill\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		Prune:      true,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !result.CopyMode {
		t.Fatal("CopyMode = false, want true for a project with real skill directories")
	}
	if len(result.Pruned) != 1 || result.Pruned[0] != "old-skill" {
		t.Errorf("Pruned = %v, want [old-skill]", result.Pruned)
	}

	// The global orphan MUST be removed.
	if _, statErr := os.Stat(orphanGlobal); statErr == nil {
		t.Error("global orphan should have been removed under --prune")
	}
	// The project's real orphan copy MUST remain intact (copy-mode promised not
	// to touch the project).
	if _, statErr := os.Stat(orphanProject); statErr != nil {
		t.Errorf("project's real orphan dir should be intact in copy-mode, stat error = %v", statErr)
	}
}

// Scenario: "Orphans are kept without prune".
func TestUpdate_OrphansKeptWithoutPrune(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	orphanDir := filepath.Join(homeDir, ".config", "opencode", "skills", "old-skill")
	if err := os.MkdirAll(orphanDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("---\nname: old-skill\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(result.GapReport.Orphaned) != 1 {
		t.Errorf("Orphaned = %d, want 1 (reported)", len(result.GapReport.Orphaned))
	}
	if len(result.Pruned) != 0 {
		t.Errorf("Pruned = %v, want empty without --prune", result.Pruned)
	}
	if _, statErr := os.Stat(orphanDir); statErr != nil {
		t.Error("orphan skill should be kept without --prune")
	}
}

// Scenario: "Update refreshes skills without touching template or user config".
// The Clone round-trip must preserve models, playwright, mutation_testing,
// judge, created_at, and agent; only harness_version/skill_count/inventory move.
func TestUpdate_PreservesUserConfigAndTemplate(t *testing.T) {
	embeddedFS := updateEmbeddedFS()
	homeDir, projectDir := setupInitializedProject(t, embeddedFS)

	// Customize the config and the template so we can assert they survive.
	cfg := &config.Config{HomeDir: projectDir}
	if err := cfg.Load(os.DirFS(projectDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	createdAt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg.CreatedAt = createdAt
	cfg.MutationTesting = config.MutationTesting{Enabled: true, Tool: "gremlins", Threshold: 0.9}
	cfg.Playwright = config.Playwright{Enabled: true, TestDir: "e2e", BaseURL: "http://localhost"}
	cfg.Models = config.ModelConfig{Default: config.ModelRef{Model: "claude-x"}, Phases: map[string]config.ModelRef{"apply": {Model: "gpt-4o"}}}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	claudePath := filepath.Join(projectDir, "CLAUDE.md")
	customTemplate := []byte("# my customized orchestrator\n")
	if err := os.WriteFile(claudePath, customTemplate, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Introduce a real gap so update actually writes.
	embeddedFS["sdd-init/SKILL.md"] = &fstest.MapFile{
		Data: []byte("---\nname: sdd-init\nmetadata:\n  version: \"9.0\"\n---\n# Init"),
	}

	if _, err := Update(UpdateOptions{
		HomeDir:    homeDir,
		ProjectDir: projectDir,
		EmbeddedFS: embeddedFS,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Template must be untouched.
	got, _ := os.ReadFile(claudePath)
	if string(got) != string(customTemplate) {
		t.Errorf("CLAUDE.md was modified by update:\n%s", string(got))
	}

	// Preserved user config values must survive.
	var after config.Config
	if err := after.Load(os.DirFS(projectDir)); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !after.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want preserved %v", after.CreatedAt, createdAt)
	}
	if after.Agent != "claude" {
		t.Errorf("Agent = %q, want preserved claude", after.Agent)
	}
	if !after.MutationTesting.Enabled || after.MutationTesting.Tool != "gremlins" {
		t.Errorf("MutationTesting not preserved: %+v", after.MutationTesting)
	}
	if !after.Playwright.Enabled || after.Playwright.TestDir != "e2e" {
		t.Errorf("Playwright not preserved: %+v", after.Playwright)
	}
	if after.Models.Default.FullID() != "claude-x" || after.Models.Phases["apply"].FullID() != "gpt-4o" {
		t.Errorf("Models not preserved: %+v", after.Models)
	}
}
