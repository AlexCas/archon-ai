package initcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCommitAttributionScript_Execution runs the actual generated shell
// script (not just its string content) against crafted PreToolUse stdin
// payloads, skipping if jq isn't on PATH — the script itself fails open in
// that case (see commitAttributionScript), so there is nothing to assert.
func TestCommitAttributionScript_Execution(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH; the script fails open without it — nothing to test")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "commit-attribution.sh")
	if err := os.WriteFile(scriptPath, []byte(commitAttributionScript), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	run := func(t *testing.T, bashCommand string) (exitCode int, stderr string) {
		t.Helper()
		stdin, err := json.Marshal(map[string]any{
			"hook_event_name": "PreToolUse",
			"tool_name":       "Bash",
			"tool_input":      map[string]any{"command": bashCommand},
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
		command   string
		wantBlock bool
	}{
		{"plain commit is allowed", `git commit -m "fix: thing"`, false},
		{"commit with Co-Authored-By is blocked", "git commit -m \"fix: thing\n\nCo-Authored-By: Claude <noreply@anthropic.com>\"", true},
		{"commit with mixed-case Co-Authored-By is blocked", "git commit -m \"fix: thing\n\nCO-AUTHORED-BY: Claude\"", true},
		{"gh pr create with Generated with is blocked", `gh pr create --title "x" --body "Generated with Claude Code"`, true},
		{"unrelated bash command is allowed", `npm test`, false},
		{"commit mentioning co-authored-by only in an unrelated string is still blocked (substring match)", `git commit -m "docs: explain co-authored-by convention"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stderr := run(t, tt.command)
			blocked := code == 2
			if blocked != tt.wantBlock {
				t.Errorf("command %q: exit code = %d (blocked=%v), want blocked=%v; stderr=%q", tt.command, code, blocked, tt.wantBlock, stderr)
			}
			if blocked && stderr == "" {
				t.Errorf("command %q: blocked but no reason was written to stderr", tt.command)
			}
		})
	}
}
