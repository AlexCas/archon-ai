package initcmd

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/scaffold"
)

// RefreshResult captures the outcome of a skill refresh: where skills were
// materialized, which skills were extracted, and a truthful inventory built from
// each skill's real frontmatter version.
type RefreshResult struct {
	GlobalSkillsDir  string
	ProjectSkillsDir string
	Extracted        []string
	Inventory        []config.SkillInventory
}

// refreshSkills extracts the embedded skills into the machine-wide global skills
// directory, links them into the project's skills directory, and builds a
// truthful inventory keyed off each skill's real metadata.version frontmatter.
//
// It performs the skill-side work only: it does NOT write the orchestrator
// template, the config file, the rollback manifest, or the openspec scaffold.
// Those concerns stay with init's Run so that update can reuse this routine
// without ever touching CLAUDE.md/AGENTS.md or resetting user config.
func refreshSkills(homeDir, projectDir, agentName string, embeddedFS fs.FS) (*RefreshResult, error) {
	globalSkillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
	extracted, err := scaffold.Extract(embeddedFS, globalSkillsDir)
	if err != nil {
		return nil, fmt.Errorf("extract skills: %w", err)
	}

	projectSkillsDir := resolveProjectSkillsDir(projectDir, agentName)
	if err := createSymlinks(globalSkillsDir, projectSkillsDir, extracted); err != nil {
		return nil, fmt.Errorf("create symlinks: %w", err)
	}

	inventory := make([]config.SkillInventory, len(extracted))
	for i, name := range extracted {
		inventory[i] = config.SkillInventory{
			Name:    name,
			Version: scaffold.SkillVersion(embeddedFS, name),
			Source:  "embedded",
		}
	}

	return &RefreshResult{
		GlobalSkillsDir:  globalSkillsDir,
		ProjectSkillsDir: projectSkillsDir,
		Extracted:        extracted,
		Inventory:        inventory,
	}, nil
}
