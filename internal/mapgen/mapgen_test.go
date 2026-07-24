package mapgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_Idempotent(t *testing.T) {
	root := t.TempDir()
	openspecDir := filepath.Join(root, "openspec")
	if err := os.MkdirAll(filepath.Join(openspecDir, "specs", "alpha"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "specs", "alpha", "spec.md"),
		[]byte("## Purpose\n\nAlpha does things.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := Generate(root); err != nil {
		t.Fatalf("Generate() first run error = %v", err)
	}
	first, err := os.ReadFile(filepath.Join(openspecDir, "map.md"))
	if err != nil {
		t.Fatalf("read map.md after first run: %v", err)
	}

	if err := Generate(root); err != nil {
		t.Fatalf("Generate() second run error = %v", err)
	}
	second, err := os.ReadFile(filepath.Join(openspecDir, "map.md"))
	if err != nil {
		t.Fatalf("read map.md after second run: %v", err)
	}

	if string(first) != string(second) {
		t.Errorf("Generate() is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestGenerate_PreservesProseOutsideMarkers(t *testing.T) {
	root := t.TempDir()
	openspecDir := filepath.Join(root, "openspec")
	if err := os.MkdirAll(openspecDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	authored := "# My Vault\n\nAuthored preamble.\n\n<!-- MAP:START -->\nstale\n<!-- MAP:END -->\n\nTrailing prose.\n"
	if err := os.WriteFile(filepath.Join(openspecDir, "map.md"), []byte(authored), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := Generate(root); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(openspecDir, "map.md"))
	if err != nil {
		t.Fatalf("read map.md: %v", err)
	}

	gotStr := string(got)
	for _, want := range []string{"# My Vault", "Authored preamble.", "Trailing prose."} {
		if !strings.Contains(gotStr, want) {
			t.Errorf("map.md missing authored prose %q; got:\n%s", want, gotStr)
		}
	}
	if strings.Contains(gotStr, "stale") {
		t.Errorf("map.md still contains stale managed content; got:\n%s", gotStr)
	}
}
