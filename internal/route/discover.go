package route

import (
	"bufio"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// stateYAML mirrors the subset of openspec/changes/{name}/state.yaml this
// package needs, following the same shape as mapgen.readState.
type stateYAML struct {
	Phase  string `yaml:"phase"`
	Status string `yaml:"status"`
}

// readState reads phase/status from a state.yaml file at p within fsys. It
// is tolerant of a missing or malformed file — both return ("", "") rather
// than an error, since the router must never hard-fail on state drift.
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

// ReadState reads phase/status for changeName's state.yaml under root. It
// returns ("", "") when changeName is "" or "none", or when the file is
// missing/corrupt — tolerant by design, never a hard failure.
func ReadState(root, changeName string) (phase, status string) {
	if changeName == "" || changeName == "none" {
		return "", ""
	}
	return readState(os.DirFS(root), path.Join("openspec/changes", changeName, "state.yaml"))
}

// activeChangeRe parses a "Active change:" field from SESSION_STATUS.md,
// tolerant of markdown bold/spacing drift (e.g. "**Active change**: foo").
var activeChangeRe = regexp.MustCompile(`(?i)\*{0,2}active change\*{0,2}\s*:\**\s*([A-Za-z0-9._-]+)`)

// readActiveChangeFromSessionStatus parses the "Active change:" field from
// SESSION_STATUS.md at the repo root (fsys root). Returns ("", false) when
// the file is absent or carries no recognizable field.
func readActiveChangeFromSessionStatus(fsys fs.FS) (string, bool) {
	data, err := fs.ReadFile(fsys, "SESSION_STATUS.md")
	if err != nil {
		return "", false
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		if m := activeChangeRe.FindStringSubmatch(scanner.Text()); m != nil {
			if name := strings.TrimSpace(m[1]); name != "" {
				return name, true
			}
		}
	}
	return "", false
}

// soleChangeDir returns the single non-archive directory under
// openspec/changes/, or ("", false) when there are zero or more than one
// candidates.
func soleChangeDir(fsys fs.FS) (string, bool) {
	entries, err := fs.ReadDir(fsys, "openspec/changes")
	if err != nil {
		return "", false
	}
	var found string
	count := 0
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "archive" {
			continue
		}
		found = e.Name()
		count++
	}
	if count == 1 {
		return found, true
	}
	return "", false
}

// activeChangeFS implements the D2 active-change fallback chain against an
// fs.FS, so it is unit-testable with testing/fstest.MapFS without touching
// the real filesystem. ActiveChange (below) is the CLI-facing wrapper over
// the real project root.
func activeChangeFS(fsys fs.FS, flagOverride string) string {
	if flagOverride != "" {
		return flagOverride
	}
	if name, ok := readActiveChangeFromSessionStatus(fsys); ok {
		return name
	}
	if name, ok := soleChangeDir(fsys); ok {
		return name
	}
	return "none"
}

// ActiveChange resolves the active SDD change under root using the D2
// precedence: --change flag > SESSION_STATUS.md "Active change:" field >
// sole non-archive folder under openspec/changes/ > "none". It never
// hard-fails — any read/parse error falls through to the next step.
func ActiveChange(root, flagOverride string) string {
	return activeChangeFS(os.DirFS(root), flagOverride)
}
