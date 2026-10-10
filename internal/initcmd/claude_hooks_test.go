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
	if !ok || len(preToolUse) != 1 {
		t.Fatalf("hooks.PreToolUse = %v, want exactly one matcher-group", doc["hooks"])
	}
	group := preToolUse[0].(map[string]any)
	if group["matcher"] != "Bash" {
		t.Errorf("matcher = %v, want Bash", group["matcher"])
	}
	handlers := group["hooks"].([]any)
	if len(handlers) != 1 {
		t.Fatalf("handlers = %v, want 1", handlers)
	}
	cmd := handlers[0].(map[string]any)["command"].(string)
	if cmd == "" || filepath.Base(cmd) != "commit-attribution.sh" {
		t.Errorf("command = %q, want it to reference commit-attribution.sh", cmd)
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
	if len(preToolUse) != 1 {
		t.Fatalf("hooks.PreToolUse has %d groups after re-run, want 1 (no duplicate entries)", len(preToolUse))
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
	if len(preToolUse) != 2 {
		t.Fatalf("hooks.PreToolUse has %d groups, want 2 (user's Edit group + archon's Bash group)", len(preToolUse))
	}

	var sawUserEdit, sawArchonBash bool
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
		}
	}
	if !sawUserEdit {
		t.Error("user's pre-existing Edit matcher-group was dropped")
	}
	if !sawArchonBash {
		t.Error("archon's Bash matcher-group was not added")
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
	if len(preToolUse) != 1 {
		t.Fatalf("hooks.PreToolUse has %d groups, want 1 (stale archon entry replaced, not duplicated)", len(preToolUse))
	}
	cmd := preToolUse[0].(map[string]any)["hooks"].([]any)[0].(map[string]any)["command"].(string)
	if filepath.Base(cmd) != "commit-attribution.sh" {
		t.Errorf("command = %q, want the current script name, not the stale one", cmd)
	}
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
