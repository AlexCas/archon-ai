package mapgen

import (
	"os"
	"path/filepath"
	"testing"
)

// seedFixtureVault creates a minimal openspec/ tree (one capability, one
// active change with a valid intra-change link) and returns the root dir.
func seedFixtureVault(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, "openspec", "specs", "alpha")
	changeDir := filepath.Join(root, "openspec", "changes", "my-feature")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFile(t, filepath.Join(specDir, "spec.md"), "## Purpose\n\nAlpha capability.\n")
	writeFile(t, filepath.Join(changeDir, "state.yaml"), "phase: design\nstatus: in_progress\n")
	writeFile(t, filepath.Join(changeDir, "proposal.md"), "Implements [[alpha]]. See [design](design.md).\n")
	writeFile(t, filepath.Join(changeDir, "design.md"), "Design details.\n")
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func TestCheck_FreshGenerate_NoIssues(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Check() = %+v, want no issues", issues)
	}
}

func TestCheck_StaleManagedRegion(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	mapPath := filepath.Join(root, "openspec", "map.md")
	data, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatalf("read map.md: %v", err)
	}
	stale, err := Splice(string(data), "## Capabilities\nhand-edited and wrong\n")
	if err != nil {
		t.Fatalf("Splice() error = %v", err)
	}
	writeFile(t, mapPath, stale)

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !hasIssue(issues, "openspec/map.md", IssueStale) {
		t.Errorf("Check() = %+v, want a stale issue for openspec/map.md", issues)
	}
}

func TestCheck_DanglingLink_ReadOnly(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	proposalPath := filepath.Join(root, "openspec", "changes", "my-feature", "proposal.md")
	writeFile(t, proposalPath, "Implements [[alpha]]. See [missing](does-not-exist.md).\n")
	before, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal.md before Check: %v", err)
	}

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	wantFile := "openspec/changes/my-feature/proposal.md"
	if !hasIssue(issues, wantFile, IssueDangling) {
		t.Errorf("Check() = %+v, want a dangling issue for %s", issues, wantFile)
	}

	after, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal.md after Check: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("Check() modified %s; want read-only", proposalPath)
	}
}

func TestCheck_MapMissingRegion(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	mapPath := filepath.Join(root, "openspec", "map.md")
	writeFile(t, mapPath, "# Vault Map\n\nHand written, no managed region.\n")

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !hasIssue(issues, "openspec/map.md", IssueMissingRegion) {
		t.Errorf("Check() = %+v, want IssueMissingRegion for openspec/map.md", issues)
	}
}

func TestCheck_MapPartialMarker_ReportedNotCrashed(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	mapPath := filepath.Join(root, "openspec", "map.md")
	writeFile(t, mapPath, "# Vault Map\n\n"+mapStart+"\norphaned body\n")

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v, want a reported issue instead of a hard error", err)
	}
	if !hasIssue(issues, "openspec/map.md", IssueMissingRegion) {
		t.Errorf("Check() = %+v, want IssueMissingRegion for partial-marker map.md", issues)
	}
}

func TestCheck_OrdinaryFileWithoutRegion_NotFlagged(t *testing.T) {
	root := seedFixtureVault(t)
	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	issues, err := Check(root)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if hasIssue(issues, "openspec/changes/my-feature/design.md", IssueMissingRegion) {
		t.Errorf("Check() = %+v, want design.md (no managed region, not map.md) not flagged", issues)
	}
}

func hasIssue(issues []Issue, file string, kind IssueKind) bool {
	for _, i := range issues {
		if filepath.ToSlash(i.File) == file && i.Kind == kind {
			return true
		}
	}
	return false
}
