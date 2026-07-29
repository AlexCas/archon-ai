package initcmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// providerBlockFrom extracts provider.<id> as a map from a parsed doc.
func providerBlockFrom(t *testing.T, doc map[string]any, id string) map[string]any {
	t.Helper()
	providers, ok := doc["provider"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"provider\"] is not an object: %T", doc["provider"])
	}
	block, ok := providers[id].(map[string]any)
	if !ok {
		t.Fatalf("provider.%s is not an object: %T", id, providers[id])
	}
	return block
}

// S2: merge creates opencode.json with the correct shape and the init run
// registers the written path for rollback.
func TestMergeOpencodeAgent_CreatesAgent(t *testing.T) {
	dir := t.TempDir()

	written, err := mergeOpencodeAgent(dir, config.ModelConfig{Leader: config.ParseModelRef(testLeaderModel)}, io.Discard)
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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

	written, err := mergeOpencodeAgent(dir, config.ModelConfig{}, io.Discard)
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
// archon-<phase> for every phase ResolvePhaseModels returns, including archon-judge
// (judge is now in PhaseOrder).
func TestMergeOpencodeAgent_WritesSubagentPerResolvablePhase(t *testing.T) {
	dir := t.TempDir()

	models := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		Phases:  map[string]config.ModelRef{"spec": config.ParseModelRef("opencode/deepseek-v4-pro")},
	}
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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

	// archon-judge must be written (judge is in PhaseOrder and Default model resolves it).
	if _, ok := agents["archon-judge"]; !ok {
		t.Error("agent.archon-judge must be written when judge resolves via Default model")
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, models, io.Discard); err != nil {
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
	written, err := mergeOpencodeAgent(dir, models, io.Discard)
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

// REQ-4: Ollama happy path — single phase produces the full provider.ollama
// shape and the agent still references the FullID.
func TestMergeOpencodeAgent_ProviderBlock_OllamaHappyPath(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	block := providerBlockFrom(t, doc, "ollama")
	if block["npm"] != "@ai-sdk/openai-compatible" {
		t.Errorf("provider.ollama.npm = %v, want %q", block["npm"], "@ai-sdk/openai-compatible")
	}
	options, ok := block["options"].(map[string]any)
	if !ok {
		t.Fatalf("provider.ollama.options is not an object: %T", block["options"])
	}
	if options["baseURL"] != "http://localhost:11434/v1" {
		t.Errorf("provider.ollama.options.baseURL = %v, want %q", options["baseURL"], "http://localhost:11434/v1")
	}
	if _, exists := options["apiKey"]; exists {
		t.Error("provider.ollama.options must not contain an apiKey key")
	}
	modelsMap, ok := block["models"].(map[string]any)
	if !ok {
		t.Fatalf("provider.ollama.models is not an object: %T", block["models"])
	}
	llama3, ok := modelsMap["llama3"].(map[string]any)
	if !ok {
		t.Fatalf("provider.ollama.models[\"llama3\"] is not an object: %T", modelsMap["llama3"])
	}
	if llama3["name"] != "llama3" {
		t.Errorf("provider.ollama.models.llama3.name = %v, want %q", llama3["name"], "llama3")
	}

	applyAgent := phaseAgentFrom(t, doc, "apply")
	if applyAgent["model"] != "ollama/llama3" {
		t.Errorf("agent.archon-apply.model = %v, want %q", applyAgent["model"], "ollama/llama3")
	}
}

// REQ-4: LocalAI happy path — single phase.
func TestMergeOpencodeAgent_ProviderBlock_LocalAIHappyPath(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"spec": {Provider: "localai", Model: "gpt-4-vision", BaseURL: "http://localhost:8080/v1"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	block := providerBlockFrom(t, doc, "localai")
	if block["npm"] != "@ai-sdk/openai-compatible" {
		t.Errorf("provider.localai.npm = %v, want %q", block["npm"], "@ai-sdk/openai-compatible")
	}
	options, ok := block["options"].(map[string]any)
	if !ok {
		t.Fatalf("provider.localai.options is not an object: %T", block["options"])
	}
	if options["baseURL"] != "http://localhost:8080/v1" {
		t.Errorf("provider.localai.options.baseURL = %v, want %q", options["baseURL"], "http://localhost:8080/v1")
	}
	modelsMap, ok := block["models"].(map[string]any)
	if !ok {
		t.Fatalf("provider.localai.models is not an object: %T", block["models"])
	}
	if _, ok := modelsMap["gpt-4-vision"]; !ok {
		t.Errorf("provider.localai.models missing key %q", "gpt-4-vision")
	}

	specAgent := phaseAgentFrom(t, doc, "spec")
	if specAgent["model"] != "localai/gpt-4-vision" {
		t.Errorf("agent.archon-spec.model = %v, want %q", specAgent["model"], "localai/gpt-4-vision")
	}
}

// REQ-4: multiple phases sharing a provider id are coalesced into ONE
// provider block whose models map is the union of both phases' models.
func TestMergeOpencodeAgent_ProviderBlock_CoalescesMultiplePhases(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply":  {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			"verify": {Provider: "ollama", Model: "mistral", BaseURL: "http://localhost:11434/v1"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	providers, ok := doc["provider"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"provider\"] is not an object: %T", doc["provider"])
	}
	if len(providers) != 1 {
		t.Fatalf("provider object has %d keys, want exactly 1 (\"ollama\")", len(providers))
	}

	modelsMap := providerBlockFrom(t, doc, "ollama")["models"].(map[string]any)
	for _, want := range []string{"llama3", "mistral"} {
		if _, ok := modelsMap[want]; !ok {
			t.Errorf("provider.ollama.models missing key %q", want)
		}
	}
}

// REQ-4: a local phase and a remote phase coexist — only the local phase's
// provider id gets a block, the remote phase's agent entry is unaffected.
func TestMergeOpencodeAgent_ProviderBlock_MixedLocalAndRemote(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			"spec":  {Provider: "anthropic", Model: "claude-sonnet-4-6"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	providers, ok := doc["provider"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"provider\"] is not an object: %T", doc["provider"])
	}
	if _, ok := providers["ollama"]; !ok {
		t.Error("provider.ollama missing")
	}
	if _, ok := providers["anthropic"]; ok {
		t.Error("provider.anthropic must not be present — the anthropic ref has no BaseURL")
	}

	specAgent := phaseAgentFrom(t, doc, "spec")
	if specAgent["model"] != "anthropic/claude-sonnet-4-6" {
		t.Errorf("agent.archon-spec.model = %v, want %q", specAgent["model"], "anthropic/claude-sonnet-4-6")
	}
}

// REQ-4: two refs sharing a provider id with DIFFERENT BaseURLs emit a
// conflict warning and keep the first-encountered (PhaseOrder traversal)
// BaseURL. NOTE: the spec.md example for this scenario names phases "apply"
// and "spec" with apply's URL expected to win, but PhaseOrder places "spec"
// (index 2) before "apply" (index 5) — the literal example is inconsistent
// with the locked "PhaseOrder traversal, first-encountered wins" rule (see
// REQ-4 body and design.md's Coalescing order decision). This test uses
// "apply" and "verify" instead (apply index 5 < verify index 6), which keeps
// PhaseOrder-consistent with the given/then narrative order. Flagged as a
// deviation in the apply return summary.
func TestMergeOpencodeAgent_ProviderBlock_ConflictingBaseURLsWarn(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply":  {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			"verify": {Provider: "ollama", Model: "llama3", BaseURL: "http://remote-ollama:11434/v1"},
		},
	}
	var stderr bytes.Buffer
	if _, err := mergeOpencodeAgent(dir, cfg, &stderr); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	wantWarn := `warning: provider "ollama" declared with conflicting baseURLs — using first occurrence "http://localhost:11434/v1"`
	if !strings.Contains(stderr.String(), wantWarn) {
		t.Errorf("stderr = %q, want contains %q", stderr.String(), wantWarn)
	}

	doc := readOpencodeDoc(t, dir)
	options := providerBlockFrom(t, doc, "ollama")["options"].(map[string]any)
	if options["baseURL"] != "http://localhost:11434/v1" {
		t.Errorf("provider.ollama.options.baseURL = %v, want first-occurrence %q", options["baseURL"], "http://localhost:11434/v1")
	}
}

// REQ-4: existing user-defined provider entries survive the merge untouched,
// and the archon-built provider id is added alongside them.
func TestMergeOpencodeAgent_ProviderBlock_PreservesUserDefinedProviders(t *testing.T) {
	dir := t.TempDir()
	seed := map[string]any{
		"provider": map[string]any{
			"myprovider": map[string]any{
				"npm":    "custom-npm-package",
				"custom": "value",
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

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	providers, ok := doc["provider"].(map[string]any)
	if !ok {
		t.Fatalf("doc[\"provider\"] is not an object: %T", doc["provider"])
	}
	my, ok := providers["myprovider"].(map[string]any)
	if !ok {
		t.Fatalf("provider.myprovider is missing or wrong type: %T", providers["myprovider"])
	}
	if my["npm"] != "custom-npm-package" || my["custom"] != "value" {
		t.Errorf("provider.myprovider was modified: %v", my)
	}
	if _, ok := providers["ollama"]; !ok {
		t.Error("provider.ollama was not added")
	}
}

// REQ-4: no ref carries a BaseURL — no top-level "provider" key is emitted.
func TestMergeOpencodeAgent_ProviderBlock_NoBaseURLNoProviderKey(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	if _, exists := doc["provider"]; exists {
		t.Error("doc[\"provider\"] must not be present when no ref carries a BaseURL")
	}
}

// REQ-4 guard: a ref with empty Provider and non-empty BaseURL must NOT produce
// a provider block keyed "" — the ref is silently skipped (ValidateBaseURL
// already warned at config-set time; the warn-never-fail contract means we do
// not hard-fail here, we just drop the invalid ref).
func TestMergeOpencodeAgent_ProviderBlock_EmptyProviderSkipped(t *testing.T) {
	dir := t.TempDir()

	// A bare-model ref (Provider == "") with a BaseURL — reaches buildProviderBlock
	// via the phase list, should produce no provider block at all.
	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply": {Provider: "", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
		},
	}
	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("mergeOpencodeAgent() error = %v", err)
	}

	doc := readOpencodeDoc(t, dir)
	if _, exists := doc["provider"]; exists {
		t.Error("doc[\"provider\"] must not be present when the only ref has an empty Provider id")
	}
}

// REQ-5: the provider block is idempotent across repeated merges, and
// provider ids appear in lexicographic order regardless of input map order.
func TestMergeOpencodeAgent_ProviderBlock_IdempotentAndSorted(t *testing.T) {
	dir := t.TempDir()

	cfg := config.ModelConfig{
		Phases: map[string]config.ModelRef{
			"apply":  {Provider: "zzz", Model: "model-z", BaseURL: "http://zzz.local/v1"},
			"verify": {Provider: "aaa", Model: "model-a", BaseURL: "http://aaa.local/v1"},
		},
	}

	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, cfg, io.Discard); err != nil {
		t.Fatalf("second mergeOpencodeAgent() error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read second: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("provider block merge is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	content := string(first)
	aaaIdx := strings.Index(content, `"aaa"`)
	zzzIdx := strings.Index(content, `"zzz"`)
	if aaaIdx == -1 || zzzIdx == -1 {
		t.Fatalf("expected both provider ids in output, got:\n%s", content)
	}
	if aaaIdx >= zzzIdx {
		t.Errorf("provider keys not in lexicographic order: \"aaa\" at %d, \"zzz\" at %d", aaaIdx, zzzIdx)
	}
}
