package mapgen

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IssueKind classifies a single Check finding.
type IssueKind string

const (
	// IssueDangling flags a relative link whose target does not exist.
	IssueDangling IssueKind = "dangling"
	// IssueStale flags a managed region whose on-disk content no longer
	// matches a fresh regeneration.
	IssueStale IssueKind = "stale"
	// IssueMissingRegion flags openspec/map.md (the vault's entry node) when
	// it has no managed MAP:START/END region at all, or only one of the two
	// markers (a partial region) — both states mean the vault's map can no
	// longer be regenerated safely.
	IssueMissingRegion IssueKind = "missing_region"
)

// errNoManagedRegion signals that content has neither MAP:START nor
// MAP:END. Unlike ErrPartialMarker/ErrNestedRegion, this is not a mapgen
// package-level error: for most .md files it is the normal, expected state
// (they simply have no managed region), and Check must not flag any file
// for it — except map.md itself, the vault's entry node, which is required
// to carry a managed region.
var errNoManagedRegion = errors.New("mapgen: no MAP:START/MAP:END markers found")

// Issue is a single problem found by Check.
type Issue struct {
	File   string // path relative to root
	Kind   IssueKind
	Detail string
}

// Check walks every .md file under {root}/openspec, flags relative links
// that do not resolve (IssueDangling), and flags managed regions whose
// on-disk content no longer matches a fresh regeneration (IssueStale).
// Check never writes to disk.
func Check(root string) ([]Issue, error) {
	openspecDir := filepath.Join(root, "openspec")
	mapPath := filepath.Join(openspecDir, "map.md")

	if _, err := os.Stat(openspecDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat openspec: %w", err)
	}

	var issues []Issue
	err := filepath.WalkDir(openspecDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		content := string(data)
		rel, err := filepath.Rel(root, p)
		if err != nil {
			rel = p
		}

		for _, l := range FindRelLinks(content) {
			if _, ok := Resolve(p, l.Target); !ok {
				issues = append(issues, Issue{
					File:   rel,
					Kind:   IssueDangling,
					Detail: fmt.Sprintf("link %q does not resolve", l.Target),
				})
			}
		}

		if p == mapPath {
			body, extractErr := extractManagedBody(content)
			switch {
			case extractErr == nil:
				g, scanErr := Scan(os.DirFS(openspecDir))
				if scanErr != nil {
					return fmt.Errorf("scan openspec: %w", scanErr)
				}
				if body != Render(g) {
					issues = append(issues, Issue{
						File:   rel,
						Kind:   IssueStale,
						Detail: "managed region does not match a fresh regeneration",
					})
				}
			case errors.Is(extractErr, errNoManagedRegion):
				issues = append(issues, Issue{
					File:   rel,
					Kind:   IssueMissingRegion,
					Detail: "map.md has no MAP:START/MAP:END managed region",
				})
			case errors.Is(extractErr, ErrPartialMarker):
				issues = append(issues, Issue{
					File:   rel,
					Kind:   IssueMissingRegion,
					Detail: "map.md has only one of the MAP:START/MAP:END markers (partial region)",
				})
			// ErrNestedRegion (and any other unexpected error) on map.md is
			// left unreported here, matching prior behavior — Generate/Check
			// callers that need the managed body already surface
			// ErrNestedRegion directly when they call Splice.
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// extractManagedBody returns the content between the MAP:START/END markers
// (mirroring Splice's insertion format). It returns errNoManagedRegion if
// neither marker is present, ErrPartialMarker if exactly one is present, and
// ErrNestedRegion if either marker repeats or the pair is out of order.
func extractManagedBody(content string) (string, error) {
	start, end, err := markerSpan(content)
	if err != nil {
		return "", err
	}
	if start == -1 {
		return "", errNoManagedRegion
	}

	body := content[start+len(mapStart) : end]
	body = strings.TrimPrefix(body, "\n")
	return body, nil
}
