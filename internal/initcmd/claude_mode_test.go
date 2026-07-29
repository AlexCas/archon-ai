package initcmd

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/archon-ai/archon/internal/config"
)

// parseFrontmatter parses a minimal YAML frontmatter block from the start of
// a .md file. It expects the file to start with "---\n", reads key: value
// lines until the closing "---\n", and returns the key/value map.
func parseFrontmatter(t *testing.T, data []byte) map[string]string {
	t.Helper()
	fm := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(data))

	if !sc.Scan() || sc.Text() != "---" {
		t.Fatal("frontmatter: missing opening ---")
	}
	for sc.Scan() {
		line := sc.Text()
		if line == "---" {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			fm[parts[0]] = parts[1]
		}
	}
	return fm
}

// readAgentFile reads and returns the contents of
// <dir>/.claude/agents/archon-<phase>.md.
func readAgentFile(t *testing.T, dir, phase string) []byte {
	t.Helper()
	path := filepath.Join(dir, ".claude", "agents", "archon-"+phase+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read agent file for phase %q: %v", phase, err)
	}
	return data
}

// 5.1: Init writes one file per resolvable phase; archon-judge.md must be written when judge resolves.
func TestWriteClaudeAgents_WritesOneFilePerResolvablePhase(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		Phases:  map[string]config.ModelRef{"spec": config.ParseModelRef("anthropic/claude-opus-4-8")},
	}

	written, err := writeClaudeAgents(dir, models, io.Discard)
	if err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	resolved := config.ResolvePhaseModels(models)
	if len(written) != len(resolved) {
		t.Errorf("written %d paths, want %d", len(written), len(resolved))
	}

	// Every phase ResolvePhaseModels returned must have a file.
	for _, pm := range resolved {
		path := filepath.Join(dir, ".claude", "agents", "archon-"+pm.Phase+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected agent file for phase %q to exist: %v", pm.Phase, err)
		}
	}

	// archon-judge.md must be written (judge is now in PhaseOrder).
	judgePath := filepath.Join(dir, ".claude", "agents", "archon-judge.md")
	if _, err := os.Stat(judgePath); err != nil {
		t.Errorf("archon-judge.md must be written; stat err = %v", err)
	}
}

// 5.1 edge: A phase with no resolvable model must not produce a file.
func TestWriteClaudeAgents_OmitsPhaseWithNoModel(t *testing.T) {
	dir := t.TempDir()

	// Only spec has a model; all other phases have no default, so they are omitted.
	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{"spec": config.ParseModelRef("anthropic/claude-opus-4-8")},
	}

	written, err := writeClaudeAgents(dir, models, io.Discard)
	if err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	if len(written) != 1 {
		t.Errorf("written %d paths, want exactly 1 (spec only)", len(written))
	}

	// archon-spec.md must be written.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "archon-spec.md")); err != nil {
		t.Errorf("archon-spec.md must be written: %v", err)
	}

	// Phases other than spec must not be written.
	for _, phase := range config.PhaseOrder {
		if phase == "spec" {
			continue
		}
		path := filepath.Join(dir, ".claude", "agents", "archon-"+phase+".md")
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("archon-%s.md must not be written when phase has no model; stat err = %v", phase, err)
		}
	}
}

// 5.2: Frontmatter model is the resolved model id WITHOUT the provider prefix.
// Claude Code rejects "<provider>/<model>" in a subagent's model field, so the
// claude writer must drop the opencode-style provider prefix.
func TestWriteClaudeAgents_FrontmatterModelStripsProvider(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{"spec": config.ParseModelRef("anthropic/claude-opus-4-8")},
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	data := readAgentFile(t, dir, "spec")
	fm := parseFrontmatter(t, data)

	const want = "claude-opus-4-8"
	if fm["model"] != want {
		t.Errorf("archon-spec.md frontmatter model = %q, want %q (bare id, no provider)", fm["model"], want)
	}
	if strings.Contains(fm["model"], "/") {
		t.Errorf("archon-spec.md frontmatter model %q still contains a provider prefix", fm["model"])
	}
}

// 5.3: A phase falls back to the default model when no per-phase model is set,
// also emitted as the bare model id.
func TestWriteClaudeAgents_PhaseFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		// No phases.tasks entry — should fall back to Default.
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	data := readAgentFile(t, dir, "tasks")
	fm := parseFrontmatter(t, data)

	const want = "claude-sonnet-4-6"
	if fm["model"] != want {
		t.Errorf("archon-tasks.md frontmatter model = %q, want %q (bare id, no provider)", fm["model"], want)
	}
}

// 5.2b: A bare alias (no provider) is passed through unchanged.
func TestWriteClaudeAgents_BareAliasModelUnchanged(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{"spec": config.ParseModelRef("opus")},
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	fm := parseFrontmatter(t, readAgentFile(t, dir, "spec"))
	if fm["model"] != "opus" {
		t.Errorf("archon-spec.md frontmatter model = %q, want %q", fm["model"], "opus")
	}
}

// 5.4: Body references "skills/sdd-<phase>/SKILL.md" and is non-empty after frontmatter.
// Judge is exempt from the generic sdd-skill-ref rule.
func TestWriteClaudeAgents_BodyPointsAtPhaseSkill(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{"design": config.ParseModelRef("anthropic/claude-opus-4-8")},
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	data := readAgentFile(t, dir, "design")
	content := string(data)

	// Body is the content after the closing "---" of the frontmatter.
	closingIdx := strings.Index(content, "---\n\n")
	if closingIdx == -1 {
		t.Fatal("archon-design.md: cannot find end of frontmatter (expected ---\\n\\n)")
	}
	body := content[closingIdx+5:]

	if body == "" {
		t.Error("archon-design.md body is empty after frontmatter")
	}

	wantRef := "skills/sdd-design/SKILL.md"
	if !strings.Contains(body, wantRef) {
		t.Errorf("archon-design.md body missing %q; body = %q", wantRef, body)
	}
}

// 5.7: All non-judge phases have the generic sdd-<phase>/SKILL.md reference;
// judge is exempt (its body must NOT contain that pattern).
func TestWriteClaudeAgents_JudgeExemptFromSkillRef(t *testing.T) {
	dir := t.TempDir()

	// Use a Default model so all phases resolve.
	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-opus-4-8"),
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	for _, phase := range config.PhaseOrder {
		data := readAgentFile(t, dir, phase)
		content := string(data)
		genericRef := "skills/sdd-" + phase + "/SKILL.md"

		if phase == "judge" {
			// Judge must NOT contain the generic reference.
			if strings.Contains(content, genericRef) {
				t.Errorf("archon-judge.md must not contain %q (judge uses wrapper body)", genericRef)
			}
		} else {
			// Every other phase must contain the generic reference.
			if !strings.Contains(content, genericRef) {
				t.Errorf("archon-%s.md missing %q in body", phase, genericRef)
			}
		}
	}
}

// 5.6: Judge body is a wrapper — references judgment-day and harness-judge, not sdd-judge.
func TestWriteClaudeAgents_JudgeBodyIsWrapper(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{"judge": config.ParseModelRef("anthropic/claude-opus-4-8")},
	}

	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	data := readAgentFile(t, dir, "judge")
	content := string(data)

	// Frontmatter model must be bare (no provider prefix).
	fm := parseFrontmatter(t, data)
	if fm["model"] != "claude-opus-4-8" {
		t.Errorf("archon-judge.md frontmatter model = %q, want %q", fm["model"], "claude-opus-4-8")
	}

	// Body must reference judgment-day and harness-judge.
	if !strings.Contains(content, "judgment-day") {
		t.Errorf("archon-judge.md body missing %q", "judgment-day")
	}
	if !strings.Contains(content, "harness-judge") {
		t.Errorf("archon-judge.md body missing %q", "harness-judge")
	}

	// Body must NOT reference skills/sdd-judge/SKILL.md.
	if strings.Contains(content, "skills/sdd-judge/SKILL.md") {
		t.Errorf("archon-judge.md body must not contain %q", "skills/sdd-judge/SKILL.md")
	}

	// Body must NOT contain the generic "Do NOT delegate" line.
	if strings.Contains(content, "Do NOT delegate") {
		t.Errorf("archon-judge.md body must not contain %q", "Do NOT delegate")
	}
}

// 5.5a: Nothing resolvable writes nothing — no .claude/agents directory created.
func TestWriteClaudeAgents_NothingResolvableWritesNothing(t *testing.T) {
	dir := t.TempDir()

	written, err := writeClaudeAgents(dir, config.ModelConfig{}, io.Discard)
	if err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}
	if len(written) != 0 {
		t.Errorf("written = %v, want empty slice", written)
	}

	agentsDir := filepath.Join(dir, ".claude", "agents")
	if _, err := os.Stat(agentsDir); !os.IsNotExist(err) {
		t.Errorf(".claude/agents must not be created when nothing is resolvable; stat err = %v", err)
	}
}

// 5.5b: Non-claude agent writes no claude agent files (via Run integration gate).
func TestRun_NonClaudeWritesNoClaudeAgentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	for _, d := range []string{homeDir, projectDir, filepath.Join(projectDir, ".opencode")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", d, err)
		}
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: sdd-init\n---\n# Init")},
	}

	// ModelDefault is set so phases actually resolve — this isolates the
	// agent gate (not the no-op guard) as the reason no claude files appear,
	// matching the spec precondition "opencode project with resolvable phase models".
	opts := Options{
		HomeDir:      homeDir,
		ProjectDir:   projectDir,
		Agent:        "opencode",
		ModelLeader:  testLeaderModel,
		ModelDefault: "anthropic/claude-sonnet-4-6",
		EmbeddedFS:   embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	agentsDir := filepath.Join(projectDir, ".claude", "agents")
	if _, err := os.Stat(agentsDir); !os.IsNotExist(err) {
		t.Errorf(".claude/agents must not be created for a non-claude agent; stat err = %v", err)
	}
}

// 5.6: Re-run is byte-identical and preserves unrelated user files.
func TestWriteClaudeAgents_ReRunByteIdenticalAndPreservesUserFiles(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
	}

	// First run.
	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("first writeClaudeAgents() error = %v", err)
	}

	// Seed a user file that must not be touched.
	agentsDir := filepath.Join(dir, ".claude", "agents")
	userFile := filepath.Join(agentsDir, "my-custom-agent.md")
	if err := os.WriteFile(userFile, []byte("# My custom agent\n"), 0o644); err != nil {
		t.Fatalf("write user file: %v", err)
	}

	// Capture first-run contents of all archon files.
	firstContents := make(map[string][]byte)
	for _, phase := range config.PhaseOrder {
		data, err := os.ReadFile(filepath.Join(agentsDir, "archon-"+phase+".md"))
		if err != nil {
			t.Fatalf("read archon-%s.md after first run: %v", phase, err)
		}
		firstContents[phase] = data
	}

	// Second run.
	if _, err := writeClaudeAgents(dir, models, io.Discard); err != nil {
		t.Fatalf("second writeClaudeAgents() error = %v", err)
	}

	// Each archon file must be byte-identical.
	for _, phase := range config.PhaseOrder {
		second, err := os.ReadFile(filepath.Join(agentsDir, "archon-"+phase+".md"))
		if err != nil {
			t.Fatalf("read archon-%s.md after second run: %v", phase, err)
		}
		if !bytes.Equal(firstContents[phase], second) {
			t.Errorf("archon-%s.md is not byte-identical across runs", phase)
		}
	}

	// User file must be unchanged.
	userContent, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("read user file after second run: %v", err)
	}
	if string(userContent) != "# My custom agent\n" {
		t.Errorf("user file was modified: %q", string(userContent))
	}
}

// 5.7: Integration test — Run(agent=claude) registers all agent paths for rollback.
func TestRun_ClaudeRegistersAgentPathsForRollback(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	for _, d := range []string{homeDir, projectDir, filepath.Join(projectDir, ".claude")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", d, err)
		}
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: sdd-init\n---\n# Init")},
	}

	const defaultModel = "anthropic/claude-sonnet-4-6"
	opts := Options{
		HomeDir:      homeDir,
		ProjectDir:   projectDir,
		Agent:        "claude",
		ModelDefault: defaultModel,
		EmbeddedFS:   embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	manifest, err := config.LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}

	registeredSet := make(map[string]bool, len(manifest.CreatedPaths))
	for _, p := range manifest.CreatedPaths {
		registeredSet[p] = true
	}

	// Every archon-<phase>.md that was written must be in the manifest.
	for _, phase := range config.PhaseOrder {
		path := filepath.Join(projectDir, ".claude", "agents", "archon-"+phase+".md")
		if _, err := os.Stat(path); err != nil {
			// File not written — that's fine (could be omitted), no need to check.
			continue
		}
		if !registeredSet[path] {
			t.Errorf("archon-%s.md path %q not registered for rollback; manifest paths = %v",
				phase, path, manifest.CreatedPaths)
		}
	}
}

// 5.7: Integration test — scenario "Undo removes the generated agent files".
// Run(agent=claude), then Cleanup() via the loaded manifest, and assert every
// generated archon-<phase>.md is gone end-to-end.
func TestRun_ClaudeUndoRemovesGeneratedAgentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	projectDir := filepath.Join(tmpDir, "project")
	for _, d := range []string{homeDir, projectDir, filepath.Join(projectDir, ".claude")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s) error = %v", d, err)
		}
	}

	embeddedFS := fstest.MapFS{
		"sdd-init/SKILL.md": &fstest.MapFile{Data: []byte("---\nname: sdd-init\n---\n# Init")},
	}

	opts := Options{
		HomeDir:      homeDir,
		ProjectDir:   projectDir,
		Agent:        "claude",
		ModelDefault: "anthropic/claude-sonnet-4-6",
		EmbeddedFS:   embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Collect the archon-<phase>.md files that were actually written.
	var written []string
	for _, phase := range config.PhaseOrder {
		path := filepath.Join(projectDir, ".claude", "agents", "archon-"+phase+".md")
		if _, err := os.Stat(path); err == nil {
			written = append(written, path)
		}
	}
	if len(written) == 0 {
		t.Fatal("no archon-<phase>.md files were written; nothing to undo")
	}

	manifest, err := config.LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if err := manifest.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// Every generated agent file must be gone after undo.
	for _, path := range written {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("archon agent file %q still exists after Cleanup(); stat err = %v", path, err)
		}
	}
}

// B1.2 [REQ-6]: "Local ref on Claude path triggers warn-and-skip" — a phase ref
// with BaseURL set emits the exact warning to the injected writer AND the
// agent file is still written with the bare model id (warn-and-skip, never
// silent, never a hard failure).
func TestWriteClaudeAgents_LocalRefWarnsAndStillWritesBareModel(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
		},
	}

	var buf bytes.Buffer
	written, err := writeClaudeAgents(dir, models, &buf)
	if err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("written %d paths, want 1", len(written))
	}

	const wantWarning = `warning: phase "apply" has base_url set but agent is "claude" — local endpoint ignored; claude agents do not support custom baseURLs`
	if !strings.Contains(buf.String(), wantWarning) {
		t.Errorf("warning output = %q, want it to contain %q", buf.String(), wantWarning)
	}

	data := readAgentFile(t, dir, "apply")
	fm := parseFrontmatter(t, data)
	if fm["model"] != "llama3" {
		t.Errorf("archon-apply.md frontmatter model = %q, want %q (bare model, BaseURL dropped)", fm["model"], "llama3")
	}
}

// B1.2 [REQ-6]: "Remote ref on Claude path has no warning" — a ref with no
// BaseURL produces no warning output at all.
func TestWriteClaudeAgents_RemoteRefNoWarning(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		},
	}

	var buf bytes.Buffer
	if _, err := writeClaudeAgents(dir, models, &buf); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no warning output for a remote ref, got %q", buf.String())
	}

	data := readAgentFile(t, dir, "apply")
	fm := parseFrontmatter(t, data)
	if fm["model"] != "claude-sonnet-4-6" {
		t.Errorf("archon-apply.md frontmatter model = %q, want %q", fm["model"], "claude-sonnet-4-6")
	}
}

// B1.2 [REQ-6]: "Multiple local phases each emit a warning" — two phases each
// with a BaseURL produce one warning per phase, and both agent files are
// written with bare model ids.
func TestWriteClaudeAgents_MultipleLocalPhasesEachWarn(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply":  {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			"verify": {Provider: "ollama", Model: "mistral", BaseURL: "http://localhost:11434/v1"},
		},
	}

	var buf bytes.Buffer
	if _, err := writeClaudeAgents(dir, models, &buf); err != nil {
		t.Fatalf("writeClaudeAgents() error = %v", err)
	}

	out := buf.String()
	wantApply := `warning: phase "apply" has base_url set but agent is "claude"`
	wantVerify := `warning: phase "verify" has base_url set but agent is "claude"`
	if strings.Count(out, wantApply) != 1 {
		t.Errorf("expected exactly one warning for phase apply, got output %q", out)
	}
	if strings.Count(out, wantVerify) != 1 {
		t.Errorf("expected exactly one warning for phase verify, got output %q", out)
	}

	applyFM := parseFrontmatter(t, readAgentFile(t, dir, "apply"))
	if applyFM["model"] != "llama3" {
		t.Errorf("archon-apply.md frontmatter model = %q, want %q", applyFM["model"], "llama3")
	}
	verifyFM := parseFrontmatter(t, readAgentFile(t, dir, "verify"))
	if verifyFM["model"] != "mistral" {
		t.Errorf("archon-verify.md frontmatter model = %q, want %q", verifyFM["model"], "mistral")
	}
}
