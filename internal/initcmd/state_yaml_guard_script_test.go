package initcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestStateYAMLGuardScript_Execution runs the actual generated shell script
// (not just its string content) against crafted PreToolUse stdin payloads
// for Edit/Write calls, skipping if jq isn't on PATH — the script fails open
// in that case (see stateYAMLGuardScript), so there is nothing to assert.
func TestStateYAMLGuardScript_Execution(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH; the script fails open without it — nothing to test")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "state-yaml-guard.sh")
	if err := os.WriteFile(scriptPath, []byte(stateYAMLGuardScript), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	run := func(t *testing.T, filePath string) (exitCode int, stderr string) {
		t.Helper()
		stdin, err := json.Marshal(map[string]any{
			"hook_event_name": "PreToolUse",
			"tool_name":       "Edit",
			"tool_input":      map[string]any{"file_path": filePath},
		})
		if err != nil {
			t.Fatalf("marshal stdin: %v", err)
		}

		cmd := exec.Command("sh", scriptPath)
		cmd.Stdin = bytes.NewReader(stdin)
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf

		err = cmd.Run()
		if err == nil {
			return 0, errBuf.String()
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), errBuf.String()
		}
		t.Fatalf("run script: %v", err)
		return -1, ""
	}

	tests := []struct {
		name      string
		filePath  string
		wantBlock bool
	}{
		{"direct edit of state.yaml is blocked", "/repo/openspec/changes/my-feature/state.yaml", true},
		{"a different change's state.yaml is also blocked", "/repo/openspec/changes/other-change/state.yaml", true},
		{"editing a spec.md is allowed", "/repo/openspec/changes/my-feature/specs/foo/spec.md", false},
		{"a file that merely contains \"state.yaml\" in its name is allowed", "/repo/openspec/changes/my-feature/not-state.yaml.bak", false},
		{"an unrelated file is allowed", "/repo/README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stderr := run(t, tt.filePath)
			blocked := code == 2
			if blocked != tt.wantBlock {
				t.Errorf("file_path %q: exit code = %d (blocked=%v), want blocked=%v; stderr=%q", tt.filePath, code, blocked, tt.wantBlock, stderr)
			}
			if blocked && stderr == "" {
				t.Errorf("file_path %q: blocked but no reason was written to stderr", tt.filePath)
			}
		})
	}
}
