package mapgen

import (
	"path/filepath"
	"testing"
)

// seedBackfillFixture creates a vault with 2 archived changes, each holding
// a relative link whose depth still reflects the pre-archive location (as a
// plain directory move would leave it), and returns the root dir.
func seedBackfillFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	openspecDir := filepath.Join(root, "openspec")

	mustMkdirAll(t, filepath.Join(openspecDir, "specs", "foo"))
	writeFile(t, filepath.Join(openspecDir, "specs", "foo", "spec.md"), "## Purpose\n\nFoo capability.\n")

	for _, name := range []string{"alpha", "beta"} {
		archiveDir := filepath.Join(openspecDir, "changes", "archive", "2026-01-01-"+name)
		mustMkdirAll(t, archiveDir)
		writeFile(t, filepath.Join(archiveDir, "state.yaml"), "phase: archived\nstatus: complete\n")
		writeFile(t, filepath.Join(archiveDir, "proposal.md"),
			"Implements [[foo]]. See [spec](../../specs/foo/spec.md).\n")
	}

	return root
}

func TestBackfill_Idempotent(t *testing.T) {
	root := seedBackfillFixture(t)
	paths := []string{
		filepath.Join(root, "openspec", "changes", "archive", "2026-01-01-alpha", "proposal.md"),
		filepath.Join(root, "openspec", "changes", "archive", "2026-01-01-beta", "proposal.md"),
		filepath.Join(root, "openspec", "map.md"),
	}

	if err := Backfill(root); err != nil {
		t.Fatalf("Backfill() first run error = %v", err)
	}
	first := make([]string, len(paths))
	for i, p := range paths {
		first[i] = mustReadFile(t, p)
	}

	if err := Backfill(root); err != nil {
		t.Fatalf("Backfill() second run error = %v", err)
	}
	for i, p := range paths {
		if second := mustReadFile(t, p); first[i] != second {
			t.Errorf("Backfill() not idempotent for %s:\nfirst:\n%s\nsecond:\n%s", p, first[i], second)
		}
	}

	// The link depth must actually have been repaired (one extra ../ level).
	if first[0] == "Implements [[foo]]. See [spec](../../specs/foo/spec.md).\n" {
		t.Errorf("Backfill() did not rewrite the stale link in alpha/proposal.md; got:\n%s", first[0])
	}
}
