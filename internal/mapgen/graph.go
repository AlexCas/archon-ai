// Package mapgen builds and renders the openspec vault map: a deterministic
// index of capabilities and changes, plus a backlink graph derived from
// [[capability]] wikilinks. See skills/_shared/spec-vault.md for the
// convention this package implements.
package mapgen

import "sort"

// Capability is a node under specs/{name}/spec.md.
type Capability struct {
	Name    string
	Purpose string
}

// Change is a node under changes/{name}/ (active) or
// changes/archive/YYYY-MM-DD-{name}/ (archived).
type Change struct {
	Name     string
	Phase    string
	Status   string
	Archived bool
	Date     string // YYYY-MM-DD; only set when Archived.
}

// Edge is a [[capability]] reference found in one of a change's markdown
// artifacts.
type Edge struct {
	FromChange   string
	ToCapability string
}

// Graph is the full in-memory model produced by Scan.
type Graph struct {
	Capabilities []Capability
	Changes      []Change
	Edges        []Edge
}

// Backlinks returns, for each referenced capability, the sorted list of
// distinct change names that reference it.
func (g *Graph) Backlinks() map[string][]string {
	out := make(map[string][]string)
	for _, e := range g.Edges {
		out[e.ToCapability] = append(out[e.ToCapability], e.FromChange)
	}
	for cap, changes := range out {
		sort.Strings(changes)
		out[cap] = changes
	}
	return out
}
