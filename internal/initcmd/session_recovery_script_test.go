package initcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSessionRecoveryScript_Execution runs the actual generated shell script
// (not just its string content) against crafted SessionStart stdin
// payloads, skipping if jq isn't on PATH — the script fails open in that
// case (see sessionRecoveryScript), so there is nothing to assert.
func TestSessionRecoveryScript_Execution(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not on PATH; the script fails open without it — nothing to test")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "session-recovery.sh")
	if err := os.WriteFile(scriptPath, []byte(sessionRecoveryScript), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	run := func(t *testing.T, cwd string) (stdout string, exitCode int) {
		t.Helper()
		stdin, err := json.Marshal(map[string]any{
			"hook_event_name": "SessionStart",
			"cwd":             cwd,
		})
		if err != nil {
			t.Fatalf("marshal stdin: %v", err)
		}

		cmd := exec.Command("sh", scriptPath)
		cmd.Stdin = bytes.NewReader(stdin)
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		err = cmd.Run()
		if err == nil {
			return outBuf.String(), 0
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return outBuf.String(), exitErr.ExitCode()
		}
		t.Fatalf("run script: %v", err)
		return "", -1
	}

	t.Run("prints SESSION_STATUS.md content when it exists", func(t *testing.T) {
		projectDir := t.TempDir()
		const statusContent = "- **Active change**: my-feature\n- **Phase**: apply (in_progress)\n"
		if err := os.WriteFile(filepath.Join(projectDir, "SESSION_STATUS.md"), []byte(statusContent), 0o644); err != nil {
			t.Fatalf("write SESSION_STATUS.md: %v", err)
		}

		stdout, code := run(t, projectDir)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if !bytes.Contains([]byte(stdout), []byte(statusContent)) {
			t.Errorf("stdout = %q, want it to contain the SESSION_STATUS.md content", stdout)
		}
	})

	t.Run("prints nothing when there is no SESSION_STATUS.md", func(t *testing.T) {
		projectDir := t.TempDir() // empty — no SESSION_STATUS.md

		stdout, code := run(t, projectDir)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0", code)
		}
		if stdout != "" {
			t.Errorf("stdout = %q, want empty output when there is nothing to resume", stdout)
		}
	})
}
