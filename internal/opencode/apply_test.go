package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// applyResult parses the merged opencode.json and returns the root object.
func parseOpenCodeJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("Unmarshal %s: %v", path, err)
	}
	return root
}

func getAgents(t *testing.T, root map[string]any) map[string]any {
	t.Helper()
	agents, ok := root["agent"].(map[string]any)
	if !ok {
		t.Fatal("agent key missing or not an object")
	}
	return agents
}

func agentModel(t *testing.T, agents map[string]any, key string) string {
	t.Helper()
	a, ok := agents[key].(map[string]any)
	if !ok {
		t.Fatalf("agent %q not found", key)
	}
	return a["model"].(string)
}

func TestApply_NoExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".config", "opencode", "opencode.json")

	backupPath, warnings, err := Apply(ApplyOptions{
		SettingsPath: settingsPath,
		CachePath:    filepath.Join(tmpDir, "no-cache.json"),
		DefaultModel: "claude-sonnet-4",
		Phases: map[string]string{
			"apply": "gpt-4o",
		},
	})

	if err != nil {
		t.Fatalf("Apply error = %v", err)
	}
	if backupPath != "" {
		t.Errorf("backupPath should be empty when no prior file, got %q", backupPath)
	}
	// No warnings for known model names.
	for _, w := range warnings {
		if strings.Contains(w, "cannot resolve") {
			t.Errorf("unexpected resolution warning: %s", w)
		}
	}

	root := parseOpenCodeJSON(t, settingsPath)
	agents := getAgents(t, root)

	// archon-orchestrator must be present with mode primary.
	orch, ok := agents["archon-orchestrator"].(map[string]any)
	if !ok {
		t.Fatal("archon-orchestrator not found")
	}
	if orch["mode"] != "primary" {
		t.Errorf("archon-orchestrator.mode = %v, want primary", orch["mode"])
	}

	// sdd-apply should have the explicit phase model resolved.
	applyModel := agentModel(t, agents, "sdd-apply")
	if applyModel != "openai/gpt-4o" {
		t.Errorf("sdd-apply.model = %q, want openai/gpt-4o", applyModel)
	}

	// sdd-explore should have the default model (no explicit phase for explore).
	exploreModel := agentModel(t, agents, "sdd-explore")
	if exploreModel != "anthropic/claude-sonnet-4" {
		t.Errorf("sdd-explore.model = %q, want anthropic/claude-sonnet-4", exploreModel)
	}
}

func TestApply_PrePopulatedFile_PreservesUserKeys(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".config", "opencode")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "opencode.json")

	// Write a pre-existing opencode.json with user content.
	original := map[string]any{
		"provider": map[string]any{
			"my-provider": map[string]any{
				"name": "My Custom Provider",
			},
		},
		"agent": map[string]any{
			"my-custom-agent": map[string]any{
				"mode":   "primary",
				"prompt": "{file:./MY_AGENTS.md}",
			},
		},
	}
	originalData, _ := json.MarshalIndent(original, "", "  ")
	if err := os.WriteFile(settingsPath, originalData, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	backupPath, _, err := Apply(ApplyOptions{
		SettingsPath: settingsPath,
		CachePath:    filepath.Join(tmpDir, "no-cache.json"),
		DefaultModel: "claude-sonnet-4",
	})

	if err != nil {
		t.Fatalf("Apply error = %v", err)
	}
	if backupPath == "" {
		t.Error("backupPath should be non-empty when prior file existed")
	}

	// Verify backup was recorded correctly.
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("ReadFile backup: %v", err)
	}
	if string(backupData) != string(originalData) {
		t.Error("backup content does not match original")
	}

	// Verify user keys preserved in the merged file.
	root := parseOpenCodeJSON(t, settingsPath)

	// User provider should still be there.
	providers, ok := root["provider"].(map[string]any)
	if !ok {
		t.Fatal("provider key missing after merge")
	}
	if _, exists := providers["my-provider"]; !exists {
		t.Error("user provider my-provider should be preserved")
	}

	// User agent should still be there.
	agents := getAgents(t, root)
	if _, exists := agents["my-custom-agent"]; !exists {
		t.Error("user agent my-custom-agent should be preserved")
	}

	// archon-orchestrator should be added.
	if _, exists := agents["archon-orchestrator"]; !exists {
		t.Error("archon-orchestrator should be added by merge")
	}

	// Rollback: copy backup → target and verify original content restored.
	if err := atomicCopy(backupPath, settingsPath); err != nil {
		t.Fatalf("restore backup: %v", err)
	}
	restored, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile restored: %v", err)
	}
	if string(restored) != string(originalData) {
		t.Error("restored content does not match original")
	}
}

func TestApply_EmptySettingsPath_ReturnsWarning(t *testing.T) {
	backupPath, warnings, err := Apply(ApplyOptions{
		SettingsPath: "",
		DefaultModel: "claude-sonnet-4",
	})
	if err != nil {
		t.Fatalf("Apply should not error with empty SettingsPath, got: %v", err)
	}
	if backupPath != "" {
		t.Errorf("backupPath should be empty, got %q", backupPath)
	}
	if len(warnings) == 0 {
		t.Error("expected a warning when SettingsPath is empty")
	}
}

func TestApply_UnresolvableModel_DoesNotFail(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, ".config", "opencode", "opencode.json")

	_, warnings, err := Apply(ApplyOptions{
		SettingsPath: settingsPath,
		CachePath:    filepath.Join(tmpDir, "no-cache.json"),
		DefaultModel: "totally-unknown-model-xyz",
	})
	if err != nil {
		t.Fatalf("Apply should not fail on unresolvable model, got: %v", err)
	}
	// Should warn about the unresolvable model.
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "cannot resolve") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a warning about unresolvable model")
	}
}
