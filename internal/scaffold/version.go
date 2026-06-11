package scaffold

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type VersionInfo struct {
	Name           string
	EmbeddedVer    string
	InstalledVer   string
	NeedsUpdate    bool
}

func DetectVersionGaps(embeddedFS fs.FS, installedDir string) ([]VersionInfo, error) {
	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded skills: %w", err)
	}

	var gaps []VersionInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		embeddedPath := skillName + "/SKILL.md"
		
		embeddedData, err := fs.ReadFile(embeddedFS, embeddedPath)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", skillName, err)
		}

		embeddedVer := extractVersion(string(embeddedData))
		if embeddedVer == "" {
			continue
		}

		installedPath := filepath.Join(installedDir, skillName, "SKILL.md")
		installedData, err := os.ReadFile(installedPath)
		if err != nil {
			if os.IsNotExist(err) {
				gaps = append(gaps, VersionInfo{
					Name:        skillName,
					EmbeddedVer: embeddedVer,
					NeedsUpdate: true,
				})
				continue
			}
			return nil, fmt.Errorf("read installed %s: %w", skillName, err)
		}

		installedVer := extractVersion(string(installedData))
		if installedVer != embeddedVer {
			gaps = append(gaps, VersionInfo{
				Name:         skillName,
				EmbeddedVer:  embeddedVer,
				InstalledVer: installedVer,
				NeedsUpdate:  true,
			})
		}
	}

	return gaps, nil
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
