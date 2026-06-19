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

// S2: merge creates opencode.json with the correct shape and the init run
// registers the written path for rollback.
func TestMergeOpencodeAgent_CreatesAgent(t *testing.T) {
	dir := t.TempDir()

	written, err := mergeOpencodeAgent(dir, testLeaderModel)
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

	if _, err := mergeOpencodeAgent(dir, testLeaderModel); err != nil {
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
}

// S4: re-running the merge with the same inputs yields byte-identical output.
func TestMergeOpencodeAgent_Idempotent(t *testing.T) {
	dir := t.TempDir()

	if _, err := mergeOpencodeAgent(dir, testLeaderModel); err != nil {
		t.Fatalf("first mergeOpencodeAgent() error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	if _, err := mergeOpencodeAgent(dir, testLeaderModel); err != nil {
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

// S6: an empty leader model writes nothing.
func TestMergeOpencodeAgent_EmptyLeaderWritesNothing(t *testing.T) {
	dir := t.TempDir()

	written, err := mergeOpencodeAgent(dir, "")
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
