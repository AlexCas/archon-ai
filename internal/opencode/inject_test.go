package opencode

import (
	"encoding/json"
	"strings"
	"testing"
)

// minimalOverlay is a small overlay for injection tests (avoids embedding
// the full sdd-overlay.json which has 10 subagents).
var minimalOverlay = []byte(`{
	"agent": {
		"archon-orchestrator": {
			"mode": "primary",
			"prompt": "{file:./AGENTS.md}"
		},
		"sdd-apply": {
			"mode": "subagent",
			"hidden": true,
			"model": "",
			"tools": {}
		},
		"sdd-spec": {
			"mode": "subagent",
			"hidden": true,
			"model": "",
			"tools": {}
		},
		"sdd-explore": {
			"mode": "subagent",
			"hidden": true,
			"model": "",
			"tools": {}
		}
	}
}`)

func parseAgentModel(t *testing.T, data []byte, agentKey string) (string, bool) {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("unmarshal inject output: %v", err)
	}
	agents, ok := root["agent"].(map[string]any)
	if !ok {
		t.Fatal("agent is not a map")
	}
	agent, exists := agents[agentKey]
	if !exists {
		return "", false
	}
	agentMap, ok := agent.(map[string]any)
	if !ok {
		return "", true
	}
	model, _ := agentMap["model"].(string)
	return model, true
}

// agentHasModelKey returns (modelValue, keyPresent, agentPresent).
// keyPresent distinguishes "key absent" from "key present with empty value".
func agentHasModelKey(t *testing.T, data []byte, agentKey string) (string, bool, bool) {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("unmarshal inject output: %v", err)
	}
	agents, ok := root["agent"].(map[string]any)
	if !ok {
		t.Fatal("agent is not a map")
	}
	agentRaw, agentPresent := agents[agentKey]
	if !agentPresent {
		return "", false, false
	}
	agentMap, ok := agentRaw.(map[string]any)
	if !ok {
		return "", false, true
	}
	val, keyPresent := agentMap["model"]
	if !keyPresent {
		return "", false, true
	}
	modelStr, _ := val.(string)
	return modelStr, true, true
}

func TestInject_ExplicitPhaseWins(t *testing.T) {
	phases := map[string]string{
		"apply": "openai/gpt-4o",
	}
	result, err := Inject(minimalOverlay, "anthropic/claude-sonnet-4", phases, map[string]bool{})
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	model, exists := parseAgentModel(t, result, "sdd-apply")
	if !exists {
		t.Fatal("sdd-apply not found in inject output")
	}
	if model != "openai/gpt-4o" {
		t.Errorf("sdd-apply.model = %q, want openai/gpt-4o", model)
	}
}

func TestInject_ExistingUserAgentPreserved(t *testing.T) {
	// sdd-spec already exists in user's opencode.json — it should be removed
	// from the overlay so the deep-merge preserves the user's entry.
	existingAgentKeys := map[string]bool{"sdd-spec": true}
	result, err := Inject(minimalOverlay, "anthropic/claude-sonnet-4", map[string]string{}, existingAgentKeys)
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	_, exists := parseAgentModel(t, result, "sdd-spec")
	if exists {
		t.Error("sdd-spec should be absent from overlay when user already has it (merge will preserve user entry)")
	}
}

func TestInject_DefaultFallback(t *testing.T) {
	// No explicit phase assignment, no existing user key — default model should be injected.
	result, err := Inject(minimalOverlay, "anthropic/claude-sonnet-4", map[string]string{}, map[string]bool{})
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	model, exists := parseAgentModel(t, result, "sdd-explore")
	if !exists {
		t.Fatal("sdd-explore not found in inject output")
	}
	if model != "anthropic/claude-sonnet-4" {
		t.Errorf("sdd-explore.model = %q, want anthropic/claude-sonnet-4", model)
	}
}

func TestInject_OrchestratorNotModified(t *testing.T) {
	result, err := Inject(minimalOverlay, "anthropic/claude-sonnet-4", map[string]string{}, map[string]bool{})
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(result, &root); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	agents := root["agent"].(map[string]any)
	orch := agents["archon-orchestrator"].(map[string]any)
	if _, hasModel := orch["model"]; hasModel {
		t.Error("archon-orchestrator should not have model injected")
	}
}

// TestInject_EmptyDefaultModel_NoModelKey verifies Fix 1: when defaultModel is ""
// and no per-phase assignment applies, the resulting agent must NOT have a "model"
// key at all (not even "model": ""). opencode must fall back to its own default.
func TestInject_EmptyDefaultModel_NoModelKey(t *testing.T) {
	result, err := Inject(minimalOverlay, "", map[string]string{}, map[string]bool{})
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	for _, agentKey := range []string{"sdd-apply", "sdd-spec", "sdd-explore"} {
		_, keyPresent, agentPresent := agentHasModelKey(t, result, agentKey)
		if !agentPresent {
			t.Errorf("%s not found in inject output", agentKey)
			continue
		}
		if keyPresent {
			t.Errorf("%s has a model key when defaultModel is empty — key must be absent, not set to \"\"", agentKey)
		}
	}

	// Also assert the marshaled JSON contains zero `"model": ""` occurrences.
	if strings.Contains(string(result), `"model": ""`) {
		t.Error(`marshaled overlay contains "model": "" — empty model must be omitted entirely`)
	}
}

// TestInject_PerPhaseExplicit_OnlyThatPhaseHasModel verifies Fix 1 (mixed case):
// when only one phase has an explicit model and defaultModel is empty, only that
// phase agent has a model key; all others must have the key absent.
func TestInject_PerPhaseExplicit_OnlyThatPhaseHasModel(t *testing.T) {
	phases := map[string]string{
		"apply": "openai/gpt-4o",
	}
	result, err := Inject(minimalOverlay, "", phases, map[string]bool{})
	if err != nil {
		t.Fatalf("Inject error = %v", err)
	}

	// sdd-apply must have the explicit model.
	applyVal, applyKeyPresent, applyPresent := agentHasModelKey(t, result, "sdd-apply")
	if !applyPresent {
		t.Fatal("sdd-apply not found in inject output")
	}
	if !applyKeyPresent {
		t.Error("sdd-apply must have a model key when an explicit phase model is set")
	}
	if applyVal != "openai/gpt-4o" {
		t.Errorf("sdd-apply.model = %q, want openai/gpt-4o", applyVal)
	}

	// sdd-explore has no explicit model and default is empty: model key must be absent.
	_, exploreKeyPresent, explorePresent := agentHasModelKey(t, result, "sdd-explore")
	if !explorePresent {
		t.Fatal("sdd-explore not found in inject output")
	}
	if exploreKeyPresent {
		t.Error("sdd-explore must NOT have a model key when defaultModel is empty and no explicit phase set")
	}

	// No "model": "" anywhere in the marshaled output.
	if strings.Contains(string(result), `"model": ""`) {
		t.Error(`marshaled overlay contains "model": "" — empty model must be omitted entirely`)
	}
}
