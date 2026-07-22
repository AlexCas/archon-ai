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
)

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
			if body, ok := extractManagedBody(content); ok {
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
// (mirroring Splice's insertion format) and whether exactly one marker pair
// was found. Nested or absent markers return ok=false.
func extractManagedBody(content string) (string, bool) {
	if strings.Count(content, mapStart) != 1 || strings.Count(content, mapEnd) != 1 {
		return "", false
	}

	start := strings.Index(content, mapStart)
	end := strings.Index(content, mapEnd)
	if start == -1 || end == -1 || end < start {
		return "", false
	}

	body := content[start+len(mapStart) : end]
	body = strings.TrimPrefix(body, "\n")
	return body, true
}
