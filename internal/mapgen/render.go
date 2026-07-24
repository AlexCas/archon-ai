package mapgen

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// Render produces the deterministic managed-region body for g: the same
// graph always renders to the same bytes (idempotency, see mapgen_test.go).
func Render(g *Graph) string {
	var b strings.Builder
	b.WriteString(renderCapabilities(g.Capabilities))
	b.WriteString("\n")
	b.WriteString(renderActiveChanges(g.Changes))
	b.WriteString("\n")
	b.WriteString(renderArchive(g.Changes))
	b.WriteString("\n")
	b.WriteString(renderBacklinks(g))
	return b.String()
}

func renderCapabilities(caps []Capability) string {
	sorted := append([]Capability(nil), caps...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var b strings.Builder
	b.WriteString("## Capabilities\n")
	for _, c := range sorted {
		if c.Purpose != "" {
			fmt.Fprintf(&b, "- [[%s]] — %s\n", c.Name, c.Purpose)
		} else {
			fmt.Fprintf(&b, "- [[%s]]\n", c.Name)
		}
	}
	return b.String()
}

func renderActiveChanges(changes []Change) string {
	active := filterChanges(changes, false)
	sort.Slice(active, func(i, j int) bool { return active[i].Name < active[j].Name })

	var b strings.Builder
	b.WriteString("## Active Changes\n")
	b.WriteString("| Change | Phase | Status |\n")
	b.WriteString("|--------|-------|--------|\n")
	for _, c := range active {
		link := path.Join("changes", c.Name, "proposal.md")
		fmt.Fprintf(&b, "| [%s](%s) | %s | %s |\n", c.Name, link, c.Phase, c.Status)
	}
	return b.String()
}

func renderArchive(changes []Change) string {
	archived := filterChanges(changes, true)

	byDate := make(map[string][]Change)
	var dates []string
	for _, c := range archived {
		if _, ok := byDate[c.Date]; !ok {
			dates = append(dates, c.Date)
		}
		byDate[c.Date] = append(byDate[c.Date], c)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	var b strings.Builder
	b.WriteString("## Archive\n")
	for _, date := range dates {
		group := byDate[date]
		sort.Slice(group, func(i, j int) bool { return group[i].Name < group[j].Name })
		fmt.Fprintf(&b, "### %s\n", date)
		for _, c := range group {
			dirName := fmt.Sprintf("%s-%s", c.Date, c.Name)
			link := path.Join("changes", "archive", dirName, "proposal.md")
			fmt.Fprintf(&b, "- [%s](%s)\n", c.Name, link)
		}
	}
	return b.String()
}

func renderBacklinks(g *Graph) string {
	backlinks := g.Backlinks()
	var caps []string
	for cap := range backlinks {
		caps = append(caps, cap)
	}
	sort.Strings(caps)

	var b strings.Builder
	b.WriteString("## Backlinks\n")
	for _, cap := range caps {
		fmt.Fprintf(&b, "- [[%s]] ← %s\n", cap, strings.Join(backlinks[cap], ", "))
	}
	return b.String()
}

func filterChanges(changes []Change, archived bool) []Change {
	var out []Change
	for _, c := range changes {
		if c.Archived == archived {
			out = append(out, c)
		}
	}
	return out
}
