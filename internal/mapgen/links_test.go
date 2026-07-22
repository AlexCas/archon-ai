package mapgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRelLinks(t *testing.T) {
	md := "See [design](design.md), [[archon-map]], " +
		"[proposal](../../proposal.md), [site](https://example.com), " +
		"[anchor](#section), and [[harness-workflow|Workflow]].\n"

	got := FindRelLinks(md)
	want := []string{"design.md", "../../proposal.md"}

	if len(got) != len(want) {
		t.Fatalf("FindRelLinks() = %+v, want %d relative links", got, len(want))
	}
	for i, l := range got {
		if l.Target != want[i] {
			t.Errorf("FindRelLinks()[%d].Target = %q, want %q", i, l.Target, want[i])
		}
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	changeDir := filepath.Join(root, "changes", "foo")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "design.md"), []byte("design\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	srcPath := filepath.Join(changeDir, "proposal.md")

	if _, ok := Resolve(srcPath, "design.md"); !ok {
		t.Error("Resolve() ok = false for existing target, want true")
	}
	if _, ok := Resolve(srcPath, "missing.md"); ok {
		t.Error("Resolve() ok = true for missing target, want false")
	}
}

func TestRewrite_OneLevelDeeper(t *testing.T) {
	md := "See [spec](../../specs/bar/spec.md) and [[archon-map]].\n"

	got := Rewrite(md, "changes/foo", "changes/archive/2026-01-01-foo")
	want := "See [spec](../../../specs/bar/spec.md) and [[archon-map]].\n"

	if got != want {
		t.Errorf("Rewrite() = %q, want %q", got, want)
	}
}

func TestRewrite_LeavesWikilinksByteIdentical(t *testing.T) {
	md := "Implements [[alpha]] and [[beta|Beta]]; see [proposal](../proposal.md).\n"

	got := Rewrite(md, "changes/foo/specs/x", "changes/archive/2026-01-01-foo/specs/x")

	for _, want := range []string{"[[alpha]]", "[[beta|Beta]]"} {
		if !strings.Contains(got, want) {
			t.Errorf("Rewrite() = %q, wikilink %q not byte-identical", got, want)
		}
	}
}
