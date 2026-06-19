package opencode

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOverlaySubagentKeysMatchSkillFolders asserts that the subagent keys in
// assets/sdd-overlay.json (minus archon-orchestrator) exactly equal the set of
// sdd-* directories found under the repo's skills/ folder. This fails the build
// if a new sdd-* skill is added without a matching overlay entry (or vice versa).
func TestOverlaySubagentKeysMatchSkillFolders(t *testing.T) {
	// --- collect overlay agent keys ---
	var overlay struct {
		Agent map[string]json.RawMessage `json:"agent"`
	}
	if err := json.Unmarshal(overlayJSON, &overlay); err != nil {
		t.Fatalf("unmarshal overlay: %v", err)
	}

	overlayKeys := make(map[string]bool)
	for k := range overlay.Agent {
		if k == "archon-orchestrator" {
			continue
		}
		overlayKeys[k] = true
	}

	// --- collect sdd-* skill folders from the repo ---
	// The test binary is run from the package directory; we walk up to find the
	// repo root (contains go.mod) and then read skills/.
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Skipf("cannot locate repo root (may be running outside the repo): %v", err)
	}

	skillsDir := filepath.Join(repoRoot, "skills")
	skillKeys := make(map[string]bool)

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "sdd-") {
			continue
		}
		// Only count dirs that contain a SKILL.md.
		if _, statErr := os.Stat(filepath.Join(skillsDir, name, "SKILL.md")); statErr != nil {
			continue
		}
		skillKeys[name] = true
	}

	if len(skillKeys) == 0 {
		t.Fatal("no sdd-* skill folders found — check that the test can reach the repo skills/ directory")
	}

	// --- assert sets are equal ---
	for k := range overlayKeys {
		if !skillKeys[k] {
			t.Errorf("overlay has agent %q but no matching skills/%s/SKILL.md", k, k)
		}
	}
	for k := range skillKeys {
		if !overlayKeys[k] {
			t.Errorf("skills/%s exists but overlay has no matching agent entry", k)
		}
	}
}

// findRepoRoot walks up from the current directory until it finds go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fs.ErrNotExist
}
