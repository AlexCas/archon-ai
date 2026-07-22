package mapgen

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func fixtureFS() fstest.MapFS {
	return fstest.MapFS{
		"specs/alpha/spec.md": &fstest.MapFile{Data: []byte(
			"# alpha Specification\n\n## Purpose\n\nHandles the alpha capability.\n\n## Requirements\n",
		)},
		"specs/beta/spec.md": &fstest.MapFile{Data: []byte(
			"# beta Specification\n\n## Purpose\n\nHandles the beta capability.\n",
		)},
		"changes/my-feature/state.yaml": &fstest.MapFile{Data: []byte(
			"phase: design\nstatus: in_progress\n",
		)},
		"changes/my-feature/proposal.md": &fstest.MapFile{Data: []byte(
			"Implements [[alpha]] and touches [[beta]].\n",
		)},
		"changes/archive/2026-01-05-old-feature/state.yaml": &fstest.MapFile{Data: []byte(
			"phase: archive\nstatus: completed\n",
		)},
		"changes/archive/2026-01-05-old-feature/proposal.md": &fstest.MapFile{Data: []byte(
			"Implements [[alpha]].\n",
		)},
	}
}

func TestScan(t *testing.T) {
	g, err := Scan(fixtureFS())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	wantCaps := []Capability{
		{Name: "alpha", Purpose: "Handles the alpha capability."},
		{Name: "beta", Purpose: "Handles the beta capability."},
	}
	if !reflect.DeepEqual(g.Capabilities, wantCaps) {
		t.Errorf("Capabilities = %+v, want %+v", g.Capabilities, wantCaps)
	}

	if len(g.Changes) != 2 {
		t.Fatalf("len(Changes) = %d, want 2", len(g.Changes))
	}

	var active, archived *Change
	for i := range g.Changes {
		c := &g.Changes[i]
		if c.Archived {
			archived = c
		} else {
			active = c
		}
	}

	if active == nil || active.Name != "my-feature" || active.Phase != "design" || active.Status != "in_progress" {
		t.Errorf("active change = %+v, want my-feature/design/in_progress", active)
	}
	if archived == nil || archived.Name != "old-feature" || archived.Date != "2026-01-05" {
		t.Errorf("archived change = %+v, want old-feature/2026-01-05", archived)
	}

	backlinks := g.Backlinks()
	if !reflect.DeepEqual(backlinks["alpha"], []string{"my-feature", "old-feature"}) {
		t.Errorf("Backlinks()[alpha] = %v, want [my-feature old-feature]", backlinks["alpha"])
	}
	if !reflect.DeepEqual(backlinks["beta"], []string{"my-feature"}) {
		t.Errorf("Backlinks()[beta] = %v, want [my-feature]", backlinks["beta"])
	}
}

func TestScan_WikilinkMaskingAndCapabilityFilter(t *testing.T) {
	fsys := fstest.MapFS{
		"specs/realcap/spec.md": &fstest.MapFile{Data: []byte(
			"# realcap Specification\n\n## Purpose\n\nHandles the real capability.\n",
		)},
		"changes/my-feature/state.yaml": &fstest.MapFile{Data: []byte(
			"phase: design\nstatus: in_progress\n",
		)},
		"changes/my-feature/proposal.md": &fstest.MapFile{Data: []byte(
			"Implements [[realcap]] end to end.\n\n" +
				"Use inline code like `[[not-a-cap]]` for illustration.\n\n" +
				"```\n[[not-a-cap]]\n```\n",
		)},
	}

	g, err := Scan(fsys)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(g.Edges) != 1 || g.Edges[0].ToCapability != "realcap" || g.Edges[0].FromChange != "my-feature" {
		t.Fatalf("Edges = %+v, want exactly one edge my-feature -> realcap", g.Edges)
	}
}

func TestScan_EmptyVault(t *testing.T) {
	g, err := Scan(fstest.MapFS{})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(g.Capabilities) != 0 || len(g.Changes) != 0 || len(g.Edges) != 0 {
		t.Errorf("Scan(empty) = %+v, want all-empty graph", g)
	}
}
