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

func TestFindRelLinks_IgnoresFencedCodeBlock(t *testing.T) {
	md := "See [design](design.md).\n\n" +
		"```\n" +
		"Example: [proposal](../../proposal.md)\n" +
		"```\n"

	got := FindRelLinks(md)
	want := []string{"design.md"}

	if len(got) != len(want) {
		t.Fatalf("FindRelLinks() = %+v, want %d relative links", got, len(want))
	}
	for i, l := range got {
		if l.Target != want[i] {
			t.Errorf("FindRelLinks()[%d].Target = %q, want %q", i, l.Target, want[i])
		}
	}
}

func TestFindRelLinks_IgnoresInlineCodeSpan(t *testing.T) {
	md := "Use syntax like `[text](../specs/x/spec.md)` and then " +
		"see [design](design.md) for details.\n"

	got := FindRelLinks(md)
	want := []string{"design.md"}

	if len(got) != len(want) {
		t.Fatalf("FindRelLinks() = %+v, want %d relative links", got, len(want))
	}
	for i, l := range got {
		if l.Target != want[i] {
			t.Errorf("FindRelLinks()[%d].Target = %q, want %q", i, l.Target, want[i])
		}
	}
}

func TestFindRelLinks_IgnoresDoubleBacktickCodeSpan(t *testing.T) {
	// A `` `...` `` span (double-backtick delimiters wrapping content that
	// itself contains single backticks) must be masked as one unit; a naive
	// scan can stop at the first single backtick and leak the link inside.
	md := "e.g. `` `[text](../specs/x/spec.md)` ``, then see " +
		"[design](design.md).\n"

	got := FindRelLinks(md)
	want := []string{"design.md"}

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

func TestRewrite_PreservesLinksInCodeRegions(t *testing.T) {
	md := "See [spec](../../specs/bar/spec.md) for details.\n\n" +
		"```\n" +
		"Example: [link](../../specs/foo/spec.md)\n" +
		"```\n\n" +
		"Inline: `[link](../../specs/baz/spec.md)`.\n"

	got := Rewrite(md, "changes/foo", "changes/archive/2026-01-01-foo")

	if !strings.Contains(got, "[spec](../../../specs/bar/spec.md)") {
		t.Errorf("Rewrite() = %q, want prose link depth-shifted", got)
	}
	if !strings.Contains(got, "Example: [link](../../specs/foo/spec.md)") {
		t.Errorf("Rewrite() = %q, want fenced-code link byte-identical", got)
	}
	if !strings.Contains(got, "Inline: `[link](../../specs/baz/spec.md)`.") {
		t.Errorf("Rewrite() = %q, want inline-code link byte-identical", got)
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
