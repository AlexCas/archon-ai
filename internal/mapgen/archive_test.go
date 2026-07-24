package mapgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRewriteMove_FixtureArchiveShift seeds a fixture vault where a change
// folder has already been physically moved one level deeper (as a plain
// `mv` would leave it, links untouched) and asserts RewriteMove repairs the
// relative links so they resolve to the same absolute targets, while
// leaving wikilinks and .feature content byte-identical.
//
// Deviation: uses real t.TempDir() fixtures (matching Generate/Check's
// real-root signature) rather than fstest.MapFS, since RewriteMove writes
// via temp+rename and needs a real filesystem — the same deviation already
// recorded for check_test.go in tasks.md 2.13.
func TestRewriteMove_FixtureArchiveShift(t *testing.T) {
	root := t.TempDir()
	openspecDir := filepath.Join(root, "openspec")

	specDir := filepath.Join(openspecDir, "specs", "foo")
	mustMkdirAll(t, specDir)
	writeFile(t, filepath.Join(specDir, "spec.md"), "## Purpose\n\nFoo capability.\n")

	// The folder already lives at its NEW (post-move) location on disk, but
	// its content still has the OLD (pre-move) relative link depth — this
	// is what a plain directory move leaves behind.
	archiveDir := filepath.Join(openspecDir, "changes", "archive", "2026-01-01-my-feature")
	mustMkdirAll(t, archiveDir)
	proposalPath := filepath.Join(archiveDir, "proposal.md")
	writeFile(t, proposalPath, "Implements [[foo]]. See [spec](../../specs/foo/spec.md).\n")

	featureDir := filepath.Join(archiveDir, "specs", "bar")
	mustMkdirAll(t, featureDir)
	featurePath := filepath.Join(featureDir, "spec.feature")
	featureContent := "Feature: Bar\n  Scenario: does a thing\n    Given a thing\n"
	writeFile(t, featurePath, featureContent)

	if err := RewriteMove(root, "changes/my-feature", "changes/archive/2026-01-01-my-feature"); err != nil {
		t.Fatalf("RewriteMove() error = %v", err)
	}

	// (a) rewritten link resolves to the same absolute file.
	assertLinkResolvesTo(t, proposalPath, filepath.Join(specDir, "spec.md"))

	// (b) wikilinks are byte-identical after rewrite.
	proposalAfter := mustReadFile(t, proposalPath)
	if !strings.Contains(proposalAfter, "[[foo]]") {
		t.Errorf("proposal.md wikilink changed; got:\n%s", proposalAfter)
	}

	// (c) .feature content is untouched.
	featureAfter := mustReadFile(t, featurePath)
	if featureAfter != featureContent {
		t.Errorf(".feature content changed; got:\n%s\nwant:\n%s", featureAfter, featureContent)
	}
}

func assertLinkResolvesTo(t *testing.T, srcPath, wantTarget string) {
	t.Helper()
	content := mustReadFile(t, srcPath)
	links := FindRelLinks(content)
	if len(links) != 1 {
		t.Fatalf("FindRelLinks(%s) = %v, want exactly 1 link", srcPath, links)
	}
	target, ok := Resolve(srcPath, links[0].Target)
	if !ok || filepath.Clean(target) != filepath.Clean(wantTarget) {
		t.Errorf("link in %s resolves to %s (ok=%v), want %s", srcPath, target, ok, wantTarget)
	}
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", dir, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(data)
}
