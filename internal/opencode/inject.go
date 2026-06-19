package opencode

import (
	"encoding/json"
	"errors"
	"os"
)

// Inject injects the resolved model into each sdd-<phase> agent in the overlay
// using a 3-case decision tree per agent:
//
//  1. phases[phase] is a non-empty resolved qualified ID → set that model.
//  2. existingAgentKeys[phase] is true (user already has this agent defined) →
//     omit the agent from the overlay so the deep-merge preserves the user's entry.
//  3. Neither → set defaultModel as the fallback (prevents silent inheritance of
//     the orchestrator's model).
//
// The "archon-orchestrator" entry is never touched by model injection.
// If defaultModel is empty and no explicit phase model applies, the "model" key
// is removed from the agent so opencode falls back to its own default (rather than
// shipping an empty model string); the caller may warn.
func Inject(overlay []byte, defaultModel string, phases map[string]string, existingAgentKeys map[string]bool) ([]byte, error) {
	var root map[string]any
	if err := json.Unmarshal(overlay, &root); err != nil {
		return nil, err
	}

	agentsRaw, ok := root["agent"]
	if !ok {
		return overlay, nil
	}
	agents, ok := agentsRaw.(map[string]any)
	if !ok {
		return overlay, nil
	}

	// Collect agents to remove (case 2: existing user agent — preserve via merge).
	toRemove := []string{}

	for agentKey, agentDef := range agents {
		if agentKey == "archon-orchestrator" {
			continue
		}

		agentMap, ok := agentDef.(map[string]any)
		if !ok {
			continue
		}

		// Strip "sdd-" prefix to get the phase name used in phases map.
		phase := agentKey
		if len(agentKey) > 4 && agentKey[:4] == "sdd-" {
			phase = agentKey[4:]
		}

		explicitModel := phases[phase]
		switch {
		case explicitModel != "":
			// Case 1: explicit per-phase assignment wins.
			agentMap["model"] = explicitModel
		case existingAgentKeys[agentKey]:
			// Case 2: user already has this agent — remove from overlay so the
			// deep-merge preserves whatever the user already has.
			toRemove = append(toRemove, agentKey)
		default:
			// Case 3: default model fallback. When defaultModel is empty, omit the
			// key entirely so opencode falls back to its own default rather than
			// receiving an explicit empty string.
			if defaultModel != "" {
				agentMap["model"] = defaultModel
			} else {
				delete(agentMap, "model")
			}
		}
	}

	for _, key := range toRemove {
		delete(agents, key)
	}

	return json.MarshalIndent(root, "", "  ")
}

// readExistingAgentKeys reads the agent keys from an existing opencode.json file.
// Returns an empty map (and nil error) when the file is absent or has no agent section.
func readExistingAgentKeys(settingsPath string) (map[string]bool, error) {
	if settingsPath == "" {
		return map[string]bool{}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]bool{}, nil
		}
		return map[string]bool{}, nil
	}

	var root struct {
		Agent map[string]json.RawMessage `json:"agent"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		// Malformed file — treat as empty.
		return map[string]bool{}, nil
	}

	keys := make(map[string]bool, len(root.Agent))
	for k := range root.Agent {
		keys[k] = true
	}
	return keys, nil
}
