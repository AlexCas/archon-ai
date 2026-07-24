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

// initForUpdate runs `init` in a temp project rooted at cwd with HOME pointed at
// an isolated home, so the global skills dir is also isolated. It returns the
// project dir (== cwd) and the home dir.
func initForUpdate(t *testing.T) (projectDir, homeDir string) {
	t.Helper()
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	t.Cleanup(func() { os.Chdir(origDir) })

	homeDir = filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", homeDir)

	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"init", "--agent", "claude"})
	if err := root.Execute(); err != nil {
		t.Fatalf("init Execute() error = %v, stderr = %s", err, stderr.String())
	}

	return tmpDir, homeDir
}

// TestUpdateCommand_CheckNoMutation asserts `update --check` performs no
// filesystem mutation. Backs "Check reports the diff without writing".
func TestUpdateCommand_CheckNoMutation(t *testing.T) {
	projectDir, homeDir := initForUpdate(t)

	globalSkillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
	// Remove one installed skill so it shows up as "added" (a real gap).
	removed := filepath.Join(globalSkillsDir, "sdd-apply")
	if err := os.RemoveAll(removed); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	configPath := filepath.Join(projectDir, ".archon", "config.yaml")
	cfgBefore, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"update", "--check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("update --check Execute() error = %v, stderr = %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "no changes will be made") {
		t.Errorf("update --check output = %q, want dry-run notice", output)
	}
	if !strings.Contains(output, "machine-wide") {
		t.Errorf("update --check output = %q, want machine-wide scope note", output)
	}

	// --check must not re-extract the removed skill.
	if _, statErr := os.Stat(removed); statErr == nil {
		t.Error("update --check re-extracted a skill; it must not write")
	}
	cfgAfter, _ := os.ReadFile(configPath)
	if string(cfgBefore) != string(cfgAfter) {
		t.Error("update --check modified config; it must not write")
	}
}

// TestUpdateCommand_CopyModeWarning asserts that a copy-mode project (real
// directory instead of symlink) gets a warning and is not re-linked.
// Backs "Copy-mode install warns without re-linking".
func TestUpdateCommand_CopyModeWarning(t *testing.T) {
	projectDir, homeDir := initForUpdate(t)

	globalSkillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
	// Create a gap so update does not short-circuit on "up to date".
	if err := os.RemoveAll(filepath.Join(globalSkillsDir, "sdd-init")); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	// Force copy-mode: replace a project symlink with a real directory.
	projectSkillsDir := filepath.Join(projectDir, ".claude", "skills")
	link := filepath.Join(projectSkillsDir, "sdd-apply")
	if err := os.RemoveAll(link); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if err := os.MkdirAll(link, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(link, "SKILL.md"), []byte("---\nname: sdd-apply\n---\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"update"})
	if err := root.Execute(); err != nil {
		t.Fatalf("update Execute() error = %v, stderr = %s", err, stderr.String())
	}

	if !strings.Contains(stderr.String(), "copy-mode") {
		t.Errorf("stderr = %q, want copy-mode warning", stderr.String())
	}

	// In copy-mode the CLI must NOT claim this project was refreshed; it should
	// be honest that the project keeps its own copy and was not updated.
	out := stdout.String()
	if strings.Contains(out, "Skills refreshed from the embedded set.") {
		t.Errorf("stdout = %q, must not claim the project was refreshed in copy-mode", out)
	}
	if !strings.Contains(out, "keeps its own copy") {
		t.Errorf("stdout = %q, want honest copy-mode outcome", out)
	}

	// The real directory must remain a real directory (not re-linked).
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("Lstat() error = %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("copy-mode project skill was re-linked; it must not be")
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

// TestInitCommand_JudgeDefaultsToVerifyModel exercises the flag-assembly branch
// in newInitCmd (main.go) where --model-judge is absent: the judge phase must
// inherit the --model-verify value. It is an integration test that runs a real
// init into a temp dir and reads the generated .claude/agents/archon-judge.md.
//
// Non-tautology proof: if the line
//
//	modelFlags["judge"] = modelVerifyFlag
//
// is removed from newInitCmd, modelFlags will have no "judge" entry, buildConfig
// will not populate Models.Phases["judge"], ResolvePhaseModels will find no judge
// model (no default is set either), writeClaudeAgents will omit archon-judge.md,
// and os.ReadFile below will return an error — making this test fail.
func TestInitCommand_JudgeDefaultsToVerifyModel(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	homeDir := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	t.Setenv("HOME", homeDir)

	// Create the .claude dir so the claude agent is accepted.
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const verifyModel = "claude-opus-4-8"

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	// --model-verify is set; --model-judge is intentionally absent.
	root.SetArgs([]string{
		"init", "--agent", "claude",
		"--model-verify", verifyModel,
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	// Read the generated archon-judge.md. If the defaulting wiring is absent,
	// the file won't exist and ReadFile returns an error here.
	judgeAgentPath := filepath.Join(tmpDir, ".claude", "agents", "archon-judge.md")
	data, err := os.ReadFile(judgeAgentPath)
	if err != nil {
		t.Fatalf("archon-judge.md not generated (missing --model-judge defaulting wiring?): %v", err)
	}

	// Parse the YAML frontmatter to extract the model field.
	content := string(data)
	frontmatterModel := ""
	inFrontmatter := false
	for _, line := range strings.Split(content, "\n") {
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break // closing ---
		}
		if inFrontmatter && strings.HasPrefix(line, "model: ") {
			frontmatterModel = strings.TrimPrefix(line, "model: ")
		}
	}

	if frontmatterModel == "" {
		t.Fatalf("archon-judge.md frontmatter has no 'model:' field; content:\n%s", content)
	}

	// claudeFrontmatterModel strips the provider prefix, so "anthropic/claude-opus-4-8"
	// becomes "claude-opus-4-8". The bare id we passed is already stripped.
	if frontmatterModel != verifyModel {
		t.Errorf("archon-judge.md model = %q, want %q (should default to --model-verify when --model-judge is absent)",
			frontmatterModel, verifyModel)
	}
}

// TestMapCommand_GeneratesManagedRegion covers the Generate path: running
// `archon map` seeds map.md with a managed region containing the scanned
// capability.
func TestMapCommand_GeneratesManagedRegion(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	openspecDir := filepath.Join(tmpDir, "openspec")
	if err := os.MkdirAll(filepath.Join(openspecDir, "specs", "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "specs", "alpha", "spec.md"),
		[]byte("## Purpose\n\nAlpha capability.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"map"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, stderr = %s", err, stderr.String())
	}

	mapContent, err := os.ReadFile(filepath.Join(openspecDir, "map.md"))
	if err != nil {
		t.Fatalf("read map.md: %v", err)
	}
	if !strings.Contains(string(mapContent), "<!-- MAP:START -->") ||
		!strings.Contains(string(mapContent), "[[alpha]]") {
		t.Errorf("map.md missing managed region or capability entry; got:\n%s", mapContent)
	}
}

// TestMapCommand_Check covers the real --check path: it exits 0 on a freshly
// generated vault and non-zero after the managed region is hand-mutated.
func TestMapCommand_Check(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	openspecDir := filepath.Join(tmpDir, "openspec")
	if err := os.MkdirAll(filepath.Join(openspecDir, "specs", "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "specs", "alpha", "spec.md"),
		[]byte("## Purpose\n\nAlpha capability.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var genOut, genErr bytes.Buffer
	genRoot := newRootCmd(&genOut, &genErr)
	genRoot.SetArgs([]string{"map"})
	if err := genRoot.Execute(); err != nil {
		t.Fatalf("Execute() map error = %v, stderr = %s", err, genErr.String())
	}

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"map", "--check"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() map --check on fresh vault error = %v, stderr = %s", err, stderr.String())
	}

	mapPath := filepath.Join(openspecDir, "map.md")
	data, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatalf("read map.md: %v", err)
	}
	mutated := strings.Replace(string(data), "<!-- MAP:END -->", "hand-edited\n<!-- MAP:END -->", 1)
	if err := os.WriteFile(mapPath, []byte(mutated), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout2, stderr2 bytes.Buffer
	root2 := newRootCmd(&stdout2, &stderr2)
	root2.SetArgs([]string{"map", "--check"})
	if err := root2.Execute(); err == nil {
		t.Error("Execute() map --check after mutating managed region should return an error")
	}
	if !strings.Contains(stderr2.String(), "stale") {
		t.Errorf("stderr = %q, want contains %q", stderr2.String(), "stale")
	}
}

// TestMapCommand_Backfill covers the real --backfill path: it rewrites a
// stale boundary-crossing link, reports progress, and leaves --check green.
func TestMapCommand_Backfill(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	openspecDir := filepath.Join(tmpDir, "openspec")
	if err := os.MkdirAll(filepath.Join(openspecDir, "specs", "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "specs", "alpha", "spec.md"),
		[]byte("## Purpose\n\nAlpha capability.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	archiveDir := filepath.Join(openspecDir, "changes", "archive", "2026-01-01-my-feature")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveDir, "proposal.md"),
		[]byte("Implements [[alpha]]. See [spec](../../specs/alpha/spec.md).\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr bytes.Buffer
	root := newRootCmd(&stdout, &stderr)
	root.SetArgs([]string{"map", "--backfill"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() map --backfill error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "my-feature") {
		t.Errorf("stdout = %q, want per-change progress mentioning %q", stdout.String(), "my-feature")
	}

	proposal, err := os.ReadFile(filepath.Join(archiveDir, "proposal.md"))
	if err != nil {
		t.Fatalf("read proposal.md: %v", err)
	}
	if !strings.Contains(string(proposal), "../../../specs/alpha/spec.md") {
		t.Errorf("proposal.md = %q, want rewritten link with an extra ../ level", proposal)
	}

	var checkOut, checkErr bytes.Buffer
	checkRoot := newRootCmd(&checkOut, &checkErr)
	checkRoot.SetArgs([]string{"map", "--check"})
	if err := checkRoot.Execute(); err != nil {
		t.Fatalf("Execute() map --check after backfill error = %v, stderr = %s", err, checkErr.String())
	}
}
