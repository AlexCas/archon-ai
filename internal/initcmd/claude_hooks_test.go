package initcmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteClaudeHooks_WritesScriptAndSettings(t *testing.T) {
	dir := t.TempDir()

	written, err := writeClaudeHooks(dir, io.Discard)
	if err != nil {
		t.Fatalf("writeClaudeHooks() error = %v", err)
	}

	scriptPath := filepath.Join(dir, ".claude", hookScriptSubdir, "commit-attribution.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read hook script: %v", err)
	}
	if string(data) != commitAttributionScript {
		t.Errorf("hook script content does not match commitAttributionScript")
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat hook script: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook script is not executable: mode = %v", info.Mode())
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	var found bool
	for _, p := range written {
		if p == scriptPath || p == settingsPath {
			found = true
		}
	}
	_ = found // both paths are expected; the loop below checks settings content directly

	doc := readSettingsJSON(t, settingsPath)
	preToolUse, ok := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	if !ok || len(preToolUse) != 2 {
		t.Fatalf("hooks.PreToolUse = %v, want exactly two matcher-groups (Bash, Edit|Write)", doc["hooks"])
	}
	group := findMatcherGroup(t, preToolUse, "Bash")
	handlers := group["hooks"].([]any)
	if len(handlers) != 1 {
		t.Fatalf("handlers = %v, want 1", handlers)
	}
	cmd := handlers[0].(map[string]any)["command"].(string)
	if cmd == "" || filepath.Base(cmd) != "commit-attribution.sh" {
		t.Errorf("command = %q, want it to reference commit-attribution.sh", cmd)
	}
}

func TestWriteClaudeHooks_WritesSessionStartEntry(t *testing.T) {
	dir := t.TempDir()

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("writeClaudeHooks() error = %v", err)
	}

	scriptPath := filepath.Join(dir, ".claude", hookScriptSubdir, "session-recovery.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read hook script: %v", err)
	}
	if string(data) != sessionRecoveryScript {
		t.Errorf("hook script content does not match sessionRecoveryScript")
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat hook script: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook script is not executable: mode = %v", info.Mode())
	}

	doc := readSettingsJSON(t, filepath.Join(dir, ".claude", "settings.json"))
	sessionStart, ok := doc["hooks"].(map[string]any)["SessionStart"].([]any)
	if !ok || len(sessionStart) != 1 {
		t.Fatalf("hooks.SessionStart = %v, want exactly one matcher-group", doc["hooks"])
	}
	group := sessionStart[0].(map[string]any)
	if group["matcher"] != "" {
		t.Errorf("matcher = %v, want empty string (fires on every SessionStart sub-event)", group["matcher"])
	}
	handlers := group["hooks"].([]any)
	if len(handlers) != 1 {
		t.Fatalf("handlers = %v, want 1", handlers)
	}
	cmd := handlers[0].(map[string]any)["command"].(string)
	if cmd == "" || filepath.Base(cmd) != "session-recovery.sh" {
		t.Errorf("command = %q, want it to reference session-recovery.sh", cmd)
	}
}

func TestWriteClaudeHooks_WritesStateYAMLGuardEntry(t *testing.T) {
	dir := t.TempDir()

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("writeClaudeHooks() error = %v", err)
	}

	scriptPath := filepath.Join(dir, ".claude", hookScriptSubdir, "state-yaml-guard.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read hook script: %v", err)
	}
	if string(data) != stateYAMLGuardScript {
		t.Errorf("hook script content does not match stateYAMLGuardScript")
	}

	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat hook script: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("hook script is not executable: mode = %v", info.Mode())
	}

	doc := readSettingsJSON(t, filepath.Join(dir, ".claude", "settings.json"))
	preToolUse := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	group := findMatcherGroup(t, preToolUse, "Edit|Write")
	handlers := group["hooks"].([]any)
	if len(handlers) != 1 {
		t.Fatalf("handlers = %v, want 1", handlers)
	}
	cmd := handlers[0].(map[string]any)["command"].(string)
	if cmd == "" || filepath.Base(cmd) != "state-yaml-guard.sh" {
		t.Errorf("command = %q, want it to reference state-yaml-guard.sh", cmd)
	}
}

func TestWriteClaudeHooks_IdempotentReRun(t *testing.T) {
	dir := t.TempDir()

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("first writeClaudeHooks() error = %v", err)
	}
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	first, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("second writeClaudeHooks() error = %v", err)
	}
	second, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	if string(first) != string(second) {
		t.Errorf("re-running writeClaudeHooks produced a different settings.json:\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	doc := readSettingsJSON(t, settingsPath)
	preToolUse := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	if len(preToolUse) != 2 {
		t.Fatalf("hooks.PreToolUse has %d groups after re-run, want 2 (no duplicate entries)", len(preToolUse))
	}
}

func TestWriteClaudeHooks_PreservesUnrelatedSettings(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	preexisting := map[string]any{
		"env": map[string]any{"MY_VAR": "1"},
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": "Edit",
					"hooks": []any{
						map[string]any{"type": "command", "command": "/home/user/my-own-hook.sh"},
					},
				},
			},
		},
	}
	data, err := json.MarshalIndent(preexisting, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("writeClaudeHooks() error = %v", err)
	}

	doc := readSettingsJSON(t, filepath.Join(settingsDir, "settings.json"))

	if env, ok := doc["env"].(map[string]any); !ok || env["MY_VAR"] != "1" {
		t.Errorf("env key was not preserved: %v", doc["env"])
	}

	preToolUse := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	if len(preToolUse) != 3 {
		t.Fatalf("hooks.PreToolUse has %d groups, want 3 (user's Edit group + archon's Bash and Edit|Write groups)", len(preToolUse))
	}

	var sawUserEdit, sawArchonBash, sawArchonEditWrite bool
	for _, g := range preToolUse {
		group := g.(map[string]any)
		switch group["matcher"] {
		case "Edit":
			sawUserEdit = true
			cmd := group["hooks"].([]any)[0].(map[string]any)["command"].(string)
			if cmd != "/home/user/my-own-hook.sh" {
				t.Errorf("user's Edit hook command was altered: %q", cmd)
			}
		case "Bash":
			sawArchonBash = true
		case "Edit|Write":
			sawArchonEditWrite = true
		}
	}
	if !sawUserEdit {
		t.Error("user's pre-existing Edit matcher-group was dropped")
	}
	if !sawArchonBash {
		t.Error("archon's Bash matcher-group was not added")
	}
	if !sawArchonEditWrite {
		t.Error("archon's Edit|Write matcher-group was not added")
	}
}

func TestWriteClaudeHooks_ReplacesStaleArchonEntryOnUpdate(t *testing.T) {
	dir := t.TempDir()

	// Simulate a stale archon-owned entry from an older version (different
	// command path) that a re-run must replace, not duplicate.
	settingsDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	stale := map[string]any{
		"hooks": map[string]any{
			"PreToolUse": []any{
				map[string]any{
					"matcher": "Bash",
					"hooks": []any{
						map[string]any{"type": "command", "command": "${CLAUDE_PROJECT_DIR}/.claude/hooks/archon/old-name.sh"},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(stale, "", "  ")
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if _, err := writeClaudeHooks(dir, io.Discard); err != nil {
		t.Fatalf("writeClaudeHooks() error = %v", err)
	}

	doc := readSettingsJSON(t, filepath.Join(settingsDir, "settings.json"))
	preToolUse := doc["hooks"].(map[string]any)["PreToolUse"].([]any)
	// The stale Bash entry is replaced in place (not duplicated); the
	// Edit|Write entry is simply new since the fixture had no stale
	// equivalent for it.
	if len(preToolUse) != 2 {
		t.Fatalf("hooks.PreToolUse has %d groups, want 2 (stale Bash entry replaced, Edit|Write newly added)", len(preToolUse))
	}
	bashGroup := findMatcherGroup(t, preToolUse, "Bash")
	cmd := bashGroup["hooks"].([]any)[0].(map[string]any)["command"].(string)
	if filepath.Base(cmd) != "commit-attribution.sh" {
		t.Errorf("command = %q, want the current script name, not the stale one", cmd)
	}
}

// findMatcherGroup returns the single matcher-group in groups whose
// "matcher" field equals matcher, failing the test if there isn't exactly
// one.
func findMatcherGroup(t *testing.T, groups []any, matcher string) map[string]any {
	t.Helper()
	var found map[string]any
	count := 0
	for _, g := range groups {
		group := g.(map[string]any)
		if group["matcher"] == matcher {
			found = group
			count++
		}
	}
	if count != 1 {
		t.Fatalf("found %d matcher-groups with matcher %q in %v, want exactly 1", count, matcher, groups)
	}
	return found
}

func readSettingsJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return doc
}
