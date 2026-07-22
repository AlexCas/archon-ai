package mapgen

import (
	"errors"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	wikilinkRe    = regexp.MustCompile(`\[\[([^\]|#]+)(?:[|#][^\]]*)?\]\]`)
	archiveDateRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
)

// Scan walks fsys — rooted at openspec/ — and builds the capability/change
// graph: one Capability per specs/{name}/ dir, one Change per changes/{name}/
// or changes/archive/YYYY-MM-DD-{name}/ dir, and one Edge per distinct
// [[capability]] wikilink found in a change's markdown artifacts.
func Scan(fsys fs.FS) (*Graph, error) {
	caps, err := scanCapabilities(fsys)
	if err != nil {
		return nil, err
	}

	changes, edges, err := scanChanges(fsys)
	if err != nil {
		return nil, err
	}

	return &Graph{Capabilities: caps, Changes: changes, Edges: edges}, nil
}

func scanCapabilities(fsys fs.FS) ([]Capability, error) {
	entries, err := fs.ReadDir(fsys, "specs")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var caps []Capability
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		purpose := readPurpose(fsys, path.Join("specs", e.Name(), "spec.md"))
		caps = append(caps, Capability{Name: e.Name(), Purpose: purpose})
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps, nil
}

// readPurpose extracts the first non-empty paragraph following a "## Purpose"
// heading in a spec.md file. Returns "" if the file or heading is absent.
func readPurpose(fsys fs.FS, p string) string {
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		return ""
	}

	var para []string
	inPurpose := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## Purpose"):
			inPurpose = true
		case !inPurpose:
			continue
		case strings.HasPrefix(trimmed, "#"):
			return strings.Join(para, " ")
		case trimmed == "":
			if len(para) > 0 {
				return strings.Join(para, " ")
			}
		default:
			para = append(para, trimmed)
		}
	}
	return strings.Join(para, " ")
}

func scanChanges(fsys fs.FS) ([]Change, []Edge, error) {
	entries, err := fs.ReadDir(fsys, "changes")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var changes []Change
	var edges []Edge
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == "archive" {
			archived, archEdges, err := scanArchive(fsys)
			if err != nil {
				return nil, nil, err
			}
			changes = append(changes, archived...)
			edges = append(edges, archEdges...)
			continue
		}

		c, ce, err := scanChangeDir(fsys, path.Join("changes", e.Name()), e.Name())
		if err != nil {
			return nil, nil, err
		}
		changes = append(changes, c)
		edges = append(edges, ce...)
	}

	return changes, edges, nil
}

func scanArchive(fsys fs.FS) ([]Change, []Edge, error) {
	entries, err := fs.ReadDir(fsys, "changes/archive")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var changes []Change
	var edges []Edge
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := archiveDateRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		date, name := m[1], m[2]
		c, ce, err := scanChangeDir(fsys, path.Join("changes", "archive", e.Name()), name)
		if err != nil {
			return nil, nil, err
		}
		c.Archived = true
		c.Date = date
		changes = append(changes, c)
		edges = append(edges, ce...)
	}
	return changes, edges, nil
}

// scanChangeDir reads state.yaml for phase/status and extracts distinct
// [[capability]] edges from every markdown artifact under dir.
func scanChangeDir(fsys fs.FS, dir, name string) (Change, []Edge, error) {
	phase, status := readState(fsys, path.Join(dir, "state.yaml"))
	c := Change{Name: name, Phase: phase, Status: status}

	edges, err := scanEdges(fsys, dir, name)
	if err != nil {
		return Change{}, nil, err
	}
	return c, edges, nil
}

type stateYAML struct {
	Phase  string `yaml:"phase"`
	Status string `yaml:"status"`
}

func readState(fsys fs.FS, p string) (phase, status string) {
	data, err := fs.ReadFile(fsys, p)
	if err != nil {
		return "", ""
	}
	var s stateYAML
	if err := yaml.Unmarshal(data, &s); err != nil {
		return "", ""
	}
	return s.Phase, s.Status
}

func scanEdges(fsys fs.FS, dir, changeName string) ([]Edge, error) {
	var edges []Edge
	seen := make(map[string]bool)

	err := fs.WalkDir(fsys, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		for _, m := range wikilinkRe.FindAllStringSubmatch(string(data), -1) {
			cap := strings.TrimSpace(m[1])
			if seen[cap] {
				continue
			}
			seen[cap] = true
			edges = append(edges, Edge{FromChange: changeName, ToCapability: cap})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(edges, func(i, j int) bool { return edges[i].ToCapability < edges[j].ToCapability })
	return edges, nil
}
