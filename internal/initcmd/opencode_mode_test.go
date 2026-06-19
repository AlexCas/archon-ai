package initcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/archon-ai/archon/internal/config"
)

const testLeaderModel = "anthropic/claude-sonnet-4-20250514"

// readOpencodeDoc reads and parses <dir>/opencode.json into a generic map.
func readOpencodeDoc(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal opencode.json: %v", err)
	}
	return doc
}

// leaderAgentFrom extracts agent.archon-leader as a map from a parsed doc.
func leaderAgentFrom(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"agent\"] is not an object: %T", doc["agent"])
	}
	leader, ok := agents[leaderAgentName].(map[string]any)
	if !ok {
		t.Fatalf("agent.archon-leader is not an object: %T", agents[leaderAgentName])
	}
	return leader
}

// phaseAgentFrom extracts agent.archon-<phase> as a map from a parsed doc.
func phaseAgentFrom(t *testing.T, doc map[string]any, phase string) map[string]any {
	t.Helper()
	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"agent\"] is not an object: %T", doc["agent"])
	}
	key := phaseAgentName(phase)
	agent, ok := agents[key].(map[string]any)
	if !ok {
		t.Fatalf("agent.%s is not an object: %T", key, agents[key])
	}
	return agent
}

// S2: merge creates opencode.json with the correct shape and the init run
// registers the written path for rollback.
func TestMergeOpencodeAgent_CreatesAgent(t *testing.T) {
	dir := t.TempDir()

	written, err := mergeOpencodeAgent(dir, config.ModelConfig{Leader: config.ParseModelRef(testLeaderModel)})
	if err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}
	if want := filepath.Join(dir, "opencode.json"); written != want {
		t.Errorf("written path = %q, want %q", written, want)
	}

	doc := readOpencodeDoc(t, dir)
	if _, exists := doc["default_agent"]; exists {
		t.Error("default_agent key must not be written")
	}

	leader := leaderAgentFrom(t, doc)
	if leader["mode"] != "primary" {
		t.Errorf("mode = %v, want %q", leader["mode"], "primary")
	}
	if leader["prompt"] != "{file:./AGENTS.md}" {
		t.Errorf("prompt = %v, want %q", leader["prompt"], "{file:./AGENTS.md}")
	}
	if leader["model"] != testLeaderModel {
		t.Errorf("model = %v, want %q", leader["model"], testLeaderModel)
	}
}

// S2 (rollback registration): an opencode init run with a leader model
// registers the opencode.json path in the rollback manifest.
func TestRun_RegistersOpencodeJSONForRollback(t *testing.T) {
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

	opts := Options{
		HomeDir:     homeDir,
		ProjectDir:  projectDir,
		Agent:       "opencode",
		ModelLeader: testLeaderModel,
		EmbeddedFS:  embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	opencodePath := filepath.Join(projectDir, "opencode.json")
	if _, err := os.Stat(opencodePath); err != nil {
		t.Fatalf("opencode.json not created: %v", err)
	}

	manifest, err := config.LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	found := false
	for _, p := range manifest.CreatedPaths {
		if p == opencodePath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("opencode.json path %q not registered for rollback; manifest paths = %v", opencodePath, manifest.CreatedPaths)
	}
}

// S3: merging into an existing opencode.json adds agent.archon-leader while
// leaving every pre-existing key and agent untouched.
func TestMergeOpencodeAgent_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	seed := map[string]any{
		"$schema": "https://opencode.ai/config.json",
		"theme":   "dark",
		"agent": map[string]any{
			"build": map[string]any{
				"mode":  "subagent",
				"model": "some/other-model",
			},
		},
	}
	seedData, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), append(seedData, '\n'), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	models := config.ModelConfig{
		Leader:  config.ParseModelRef(testLeaderModel),
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	if doc["$schema"] != "https://opencode.ai/config.json" {
		t.Errorf("$schema = %v, want preserved", doc["$schema"])
	}
	if doc["theme"] != "dark" {
		t.Errorf("theme = %v, want %q", doc["theme"], "dark")
	}
	if _, exists := doc["default_agent"]; exists {
		t.Error("default_agent key must not be written")
	}

	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		t.Fatalf("agent is not an object: %T", doc["agent"])
	}
	build, ok := agents["build"].(map[string]any)
	if !ok {
		t.Fatalf("pre-existing agent.build is missing or wrong type: %T", agents["build"])
	}
	if build["mode"] != "subagent" || build["model"] != "some/other-model" {
		t.Errorf("agent.build was modified: %v", build)
	}
	if _, ok := agents[leaderAgentName]; !ok {
		t.Error("agent.archon-leader was not added")
	}
	// Subagents should also be present (default is set).
	for _, phase := range config.PhaseOrder {
		key := phaseAgentName(phase)
		if _, ok := agents[key]; !ok {
			t.Errorf("agent.%s was not added", key)
		}
	}
}

// S4: re-running the merge with the same inputs yields byte-identical output,
// covering both leader and all phase subagents.
func TestMergeOpencodeAgent_Idempotent(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Leader:  config.ParseModelRef(testLeaderModel),
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("second mergeOpencodeAgent() error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read second: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("merge is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// W3-3: nothing configured (empty ModelConfig) writes nothing.
func TestMergeOpencodeAgent_NothingConfiguredWritesNothing(t *testing.T) {
	dir := t.TempDir()

	written, err := mergeOpencodeAgent(dir, config.ModelConfig{})
	if err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}
	if written != "" {
		t.Errorf("written = %q, want empty", written)
	}
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Errorf("opencode.json should not exist, stat err = %v", err)
	}
}

// S5: a non-opencode agent init run writes no opencode.json.
func TestRun_NonOpencodeWritesNoOpencodeJSON(t *testing.T) {
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
		HomeDir:     homeDir,
		ProjectDir:  projectDir,
		Agent:       "claude",
		ModelLeader: testLeaderModel,
		EmbeddedFS:  embeddedFS,
	}

	if _, err := Run(opts); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectDir, "opencode.json")); !os.IsNotExist(err) {
		t.Errorf("opencode.json must not be created for a non-opencode agent, stat err = %v", err)
	}
}

// W3-4: _WritesSubagentPerResolvablePhase — default set and a phase override →
// archon-<phase> for every phase ResolvePhaseModels returns; archon-judge must
// not be written (judge is not in PhaseOrder).
func TestMergeOpencodeAgent_WritesSubagentPerResolvablePhase(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		Phases:  map[string]config.ModelRef{"spec": config.ParseModelRef("opencode/deepseek-v4-pro")},
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"agent\"] is not an object: %T", doc["agent"])
	}

	// Every phase in PhaseOrder must be present.
	resolved := config.ResolvePhaseModels(models)
	for _, pm := range resolved {
		key := phaseAgentName(pm.Phase)
		if _, ok := agents[key]; !ok {
			t.Errorf("agent.%s was not written", key)
		}
	}

	// archon-judge must never be written.
	if _, ok := agents["archon-judge"]; ok {
		t.Error("agent.archon-judge must not be written")
	}

	// archon-leader must not be written (leader FullID is empty).
	if _, ok := agents[leaderAgentName]; ok {
		t.Error("agent.archon-leader must not be written when leader is empty")
	}
}

// W3-5: _PhaseModelMatchesResolvedFullID — phases.spec = opencode/deepseek-v4-pro →
// archon-spec.model == "opencode/deepseek-v4-pro".
func TestMergeOpencodeAgent_PhaseModelMatchesResolvedFullID(t *testing.T) {
	dir := t.TempDir()

	const specModel = "opencode/deepseek-v4-pro"
	models := config.ModelConfig{
		Leader: config.ParseModelRef(testLeaderModel),
		Phases: map[string]config.ModelRef{"spec": config.ParseModelRef(specModel)},
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	specAgent := phaseAgentFrom(t, doc, "spec")
	if specAgent["model"] != specModel {
		t.Errorf("agent.archon-spec.model = %v, want %q", specAgent["model"], specModel)
	}
}

// W3-6: _PhaseFallsBackToDefault — phases.tasks empty, default set →
// archon-tasks.model == default FullID.
func TestMergeOpencodeAgent_PhaseFallsBackToDefault(t *testing.T) {
	dir := t.TempDir()

	const defaultModel = "anthropic/claude-sonnet-4-6"
	models := config.ModelConfig{
		Default: config.ParseModelRef(defaultModel),
		// No phases.tasks entry — should fall back to Default.
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	tasksAgent := phaseAgentFrom(t, doc, "tasks")
	if tasksAgent["model"] != defaultModel {
		t.Errorf("agent.archon-tasks.model = %v, want %q", tasksAgent["model"], defaultModel)
	}
}

// W3-7: _SubagentFixedFields — every phase subagent carries the fixed shape:
// mode "subagent", hidden true, description "Archon SDD <phase> phase",
// prompt "{file:./AGENTS.md}".
func TestMergeOpencodeAgent_SubagentFixedFields(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	for _, phase := range config.PhaseOrder {
		a := phaseAgentFrom(t, doc, phase)
		if a["mode"] != "subagent" {
			t.Errorf("agent.%s.mode = %v, want \"subagent\"", phaseAgentName(phase), a["mode"])
		}
		if a["hidden"] != true {
			t.Errorf("agent.%s.hidden = %v, want true", phaseAgentName(phase), a["hidden"])
		}
		wantDesc := "Archon SDD " + phase + " phase"
		if a["description"] != wantDesc {
			t.Errorf("agent.%s.description = %v, want %q", phaseAgentName(phase), a["description"], wantDesc)
		}
		if a["prompt"] != "{file:./AGENTS.md}" {
			t.Errorf("agent.%s.prompt = %v, want %q", phaseAgentName(phase), a["prompt"], "{file:./AGENTS.md}")
		}
	}
}

// E2-3a: variant key present when the leader Effort is non-empty.
func TestMergeOpencodeAgent_LeaderVariantPresent(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Leader: config.ModelRef{Provider: "anthropic", Model: "claude-opus-4-8", Effort: "high"},
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	leader := leaderAgentFrom(t, doc)
	if leader["variant"] != "high" {
		t.Errorf("agent.archon-leader.variant = %v, want %q", leader["variant"], "high")
	}
}

// E2-3b: variant key absent (omitted) when the leader Effort is empty.
func TestMergeOpencodeAgent_LeaderVariantAbsentWhenEmpty(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Leader: config.ParseModelRef(testLeaderModel), // no Effort
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	leader := leaderAgentFrom(t, doc)
	if _, exists := leader["variant"]; exists {
		t.Errorf("agent.archon-leader.variant must be absent when Effort is empty, got %v", leader["variant"])
	}
}

// E2-3c: variant key present on a phase subagent when Effort is set.
func TestMergeOpencodeAgent_PhaseVariantPresent(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"spec": {Provider: "opencode", Model: "deepseek-v4-pro", Effort: "medium"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	specAgent := phaseAgentFrom(t, doc, "spec")
	if specAgent["variant"] != "medium" {
		t.Errorf("agent.archon-spec.variant = %v, want %q", specAgent["variant"], "medium")
	}
}

// E2-3d: variant key absent on a phase subagent when Effort is empty.
func TestMergeOpencodeAgent_PhaseVariantAbsentWhenEmpty(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"), // no Effort
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	for _, phase := range config.PhaseOrder {
		a := phaseAgentFrom(t, doc, phase)
		if _, exists := a["variant"]; exists {
			t.Errorf("agent.archon-%s.variant must be absent when Effort is empty, got %v", phase, a["variant"])
		}
	}
}

// E2-3e: re-run with mixed effort/empty is byte-identical (idempotency).
func TestMergeOpencodeAgent_IdempotentWithMixedEffort(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Leader:  config.ModelRef{Provider: "anthropic", Model: "claude-opus-4-8", Effort: "high"},
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"), // empty Effort
		Phases: map[string]config.ModelRef{
			"spec": {Provider: "opencode", Model: "deepseek-v4-pro", Effort: "low"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, models); err != nil {
		t.Fatalf("second mergeOpencodeAgent() error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read second: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("merge is not idempotent with mixed effort:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// W3-8: _PhasesSetEmptyLeaderWritesSubagentsNoLeader — leader empty, default
// set → subagents present, archon-leader absent.
func TestMergeOpencodeAgent_PhasesSetEmptyLeaderWritesSubagentsNoLeader(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		// Leader intentionally empty.
	}
	written, err := mergeOpencodeAgent(dir, models)
	if err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}
	if written == "" {
		t.Fatal("expected opencode.json to be written but got empty path")
	}

	doc := readOpencodeDoc(t, dir)
	agents, ok := doc["agent"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"agent\"] is not an object: %T", doc["agent"])
	}

	// archon-leader must be absent.
	if _, ok := agents[leaderAgentName]; ok {
		t.Error("agent.archon-leader must not be written when leader FullID is empty")
	}

	// Phase subagents must be present.
	for _, phase := range config.PhaseOrder {
		key := phaseAgentName(phase)
		if _, ok := agents[key]; !ok {
			t.Errorf("agent.%s was not written", key)
		}
	}
}
