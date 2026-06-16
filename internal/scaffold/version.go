package scaffold

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type VersionInfo struct {
	Name         string
	EmbeddedVer  string
	InstalledVer string
	NeedsUpdate  bool
}

// SkillChange describes a single skill's version delta between the embedded set
// and what is installed on disk.
type SkillChange struct {
	Name         string
	EmbeddedVer  string
	InstalledVer string
}

// GapReport classifies every skill into one of three buckets:
//   - Added:    embedded skill with no installed SKILL.md.
//   - Changed:  both present and the embedded SKILL.md content differs from the
//     installed one (see ClassifyGaps for the content-based rule).
//   - Orphaned: installed skill directory with a SKILL.md but no embedded
//     counterpart (no longer shipped).
type GapReport struct {
	Added    []SkillChange
	Changed  []SkillChange
	Orphaned []SkillChange
}

// ClassifyGaps walks both the embedded skill set and the installed directory and
// classifies each skill as added, changed, or orphaned. installedDir is the
// machine-wide global skills directory.
func ClassifyGaps(embeddedFS fs.FS, installedDir string) (GapReport, error) {
	var report GapReport

	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return report, fmt.Errorf("read embedded skills: %w", err)
	}

	embeddedSkills := make(map[string]struct{})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		embeddedPath := skillName + "/SKILL.md"

		embeddedData, err := fs.ReadFile(embeddedFS, embeddedPath)
		if err != nil {
			// A directory without a SKILL.md is not a skill, so skip it.
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return report, fmt.Errorf("read embedded %s: %w", skillName, err)
		}

		embeddedSkills[skillName] = struct{}{}
		embeddedVer := extractVersion(string(embeddedData))

		installedPath := filepath.Join(installedDir, skillName, "SKILL.md")
		installedData, err := os.ReadFile(installedPath)
		if err != nil {
			if os.IsNotExist(err) {
				report.Added = append(report.Added, SkillChange{
					Name:        skillName,
					EmbeddedVer: embeddedVer,
				})
				continue
			}
			return report, fmt.Errorf("read installed %s: %w", skillName, err)
		}

		installedVer := extractVersion(string(installedData))

		// Content is the source of truth for "changed" (decision D4). Many
		// embedded skills genuinely ship version "1.0", so a version-only diff
		// would silently miss content updates whenever the version string is
		// unchanged. We therefore compare the SKILL.md bytes directly and report
		// Changed when they differ, regardless of version equality. The version
		// strings are still recorded on SkillChange for reporting.
		if !bytes.Equal(embeddedData, installedData) {
			report.Changed = append(report.Changed, SkillChange{
				Name:         skillName,
				EmbeddedVer:  embeddedVer,
				InstalledVer: installedVer,
			})
		}
	}

	// Walk the installed directory for orphans: installed skills (dir with a
	// SKILL.md) that have no embedded counterpart.
	installedEntries, err := os.ReadDir(installedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return report, fmt.Errorf("read installed dir: %w", err)
	}

	for _, entry := range installedEntries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		if _, ok := embeddedSkills[skillName]; ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(installedDir, skillName, "SKILL.md")); err != nil {
			continue
		}
		report.Orphaned = append(report.Orphaned, SkillChange{Name: skillName})
	}

	return report, nil
}

// DetectVersionGaps reports skills that need updating (added or changed),
// preserving its original contract by delegating to ClassifyGaps. Orphans are
// intentionally excluded — this function answers "what is out of date", not
// "what is no longer shipped".
func DetectVersionGaps(embeddedFS fs.FS, installedDir string) ([]VersionInfo, error) {
	report, err := ClassifyGaps(embeddedFS, installedDir)
	if err != nil {
		return nil, err
	}

	var gaps []VersionInfo
	for _, c := range report.Added {
		gaps = append(gaps, VersionInfo{
			Name:        c.Name,
			EmbeddedVer: c.EmbeddedVer,
			NeedsUpdate: true,
		})
	}
	for _, c := range report.Changed {
		gaps = append(gaps, VersionInfo{
			Name:         c.Name,
			EmbeddedVer:  c.EmbeddedVer,
			InstalledVer: c.InstalledVer,
			NeedsUpdate:  true,
		})
	}

	return gaps, nil
}

// SkillVersion reads name/SKILL.md from the embedded filesystem and returns its
// metadata.version frontmatter value, or "" if the skill or version is absent.
func SkillVersion(embeddedFS fs.FS, name string) string {
	data, err := fs.ReadFile(embeddedFS, name+"/SKILL.md")
	if err != nil {
		return ""
	}
	return ExtractVersion(string(data))
}

// ExtractVersion parses the metadata.version frontmatter value from a SKILL.md
// document, returning "" when no version is declared.
func ExtractVersion(content string) string {
	return extractVersion(content)
}

func extractVersion(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inFrontmatter := false
	inMetadata := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}

		if !inFrontmatter {
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}

		if inMetadata && strings.HasPrefix(trimmed, "version:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				ver := strings.TrimSpace(parts[1])
				ver = strings.Trim(ver, `"'`)
				return ver
			}
		}

		if inMetadata && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inMetadata = false
		}
	}

	return ""
}
