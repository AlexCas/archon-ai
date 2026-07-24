package mapgen

import (
	"os"
	"path/filepath"
	"testing"
)

func fixtureGraph() *Graph {
	return &Graph{
		Capabilities: []Capability{
			{Name: "alpha", Purpose: "Handles the alpha capability."},
			{Name: "beta", Purpose: "Handles the beta capability."},
		},
		Changes: []Change{
			{Name: "my-feature", Phase: "design", Status: "in_progress"},
			{Name: "old-feature", Archived: true, Date: "2026-01-05"},
		},
		Edges: []Edge{
			{FromChange: "my-feature", ToCapability: "alpha"},
			{FromChange: "my-feature", ToCapability: "beta"},
			{FromChange: "old-feature", ToCapability: "alpha"},
		},
	}
}

// TestRender_Golden compares Render's output against testdata/map_body.golden.
// Regenerate the fixture by hand if fixtureGraph or the render format changes.
func TestRender_Golden(t *testing.T) {
	got := Render(fixtureGraph())

	want, err := os.ReadFile(filepath.Join("testdata", "map_body.golden"))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if got != string(want) {
		t.Errorf("Render() mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRender_Idempotent(t *testing.T) {
	g := fixtureGraph()
	if Render(g) != Render(g) {
		t.Error("Render(g) is not stable across repeated calls on the same graph")
	}
}
