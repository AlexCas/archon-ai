package initcmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/archon-ai/archon/internal/agent"
	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/scaffold"
	"github.com/archon-ai/archon/internal/version"
)

type Options struct {
	HomeDir     string
	ProjectDir  string
	Agent       string
	Force       bool
	EmbeddedFS  fs.FS
}

type Result struct {
	Agent          string
	SkillsDir      string
	ExtractedCount int
	ConfigPath     string
}

func Run(opts Options) (*Result, error) {
	if opts.HomeDir == "" {
		return nil, fmt.Errorf("home directory is required")
	}
	if opts.ProjectDir == "" {
		return nil, fmt.Errorf("project directory is required")
	}
	if opts.EmbeddedFS == nil {
		return nil, fmt.Errorf("embedded filesystem is required")
	}

	agentName, err := detectAgent(opts)
	if err != nil {
		return nil, fmt.Errorf("detect agent: %w", err)
	}

	globalSkillsDir := filepath.Join(opts.HomeDir, ".config", "opencode", "skills")
	extracted, err := scaffold.Extract(opts.EmbeddedFS, globalSkillsDir)
	if err != nil {
		return nil, fmt.Errorf("extract skills: %w", err)
	}

	projectSkillsDir := resolveProjectSkillsDir(opts.ProjectDir, agentName)
	if err := createSymlinks(globalSkillsDir, projectSkillsDir, extracted); err != nil {
		return nil, fmt.Errorf("create symlinks: %w", err)
	}

	cfg := buildConfig(agentName, extracted)
	cfg.HomeDir = opts.ProjectDir
	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	rollback := buildRollbackManifest(cfg, extracted, globalSkillsDir, projectSkillsDir)
	rollback.HomeDir = opts.ProjectDir
	if err := rollback.WriteManifest(); err != nil {
		return nil, fmt.Errorf("write rollback manifest: %w", err)
	}

	if err := writeTemplate(opts.ProjectDir, agentName, len(extracted)); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	if err := createOpenSpecDir(opts.ProjectDir); err != nil {
		return nil, fmt.Errorf("create openspec dir: %w", err)
	}

	return &Result{
		Agent:          agentName,
		SkillsDir:      globalSkillsDir,
		ExtractedCount: len(extracted),
		ConfigPath:     filepath.Join(opts.ProjectDir, ".archon", "config.yaml"),
	}, nil
}

func detectAgent(opts Options) (string, error) {
	projectFS := os.DirFS(opts.ProjectDir)
	result, err := agent.Detect(projectFS)
	if err != nil {
		if opts.Agent != "" {
			return opts.Agent, nil
		}
		return "", err
	}

	if opts.Agent != "" {
		for _, d := range result.Dirs {
			if d == opts.Agent {
				return opts.Agent, nil
			}
		}
		return "", fmt.Errorf("specified agent %q not found in project", opts.Agent)
	}

	if len(result.Dirs) == 1 {
		return result.Agent, nil
	}

	return result.Agent, nil
}

func resolveProjectSkillsDir(projectDir, agentName string) string {
	switch agentName {
	case "opencode":
		return filepath.Join(projectDir, ".opencode", "skills")
	case "claude":
		return filepath.Join(projectDir, ".claude", "skills")
	case "agents":
		return filepath.Join(projectDir, ".agents", "skills")
	case "codex":
		return filepath.Join(projectDir, ".codex", "skills")
	default:
		return filepath.Join(projectDir, ".opencode", "skills")
	}
}

func createSymlinks(globalDir, projectDir string, skills []string) error {
	for _, skill := range skills {
		if err := scaffold.SymlinkOrCopy(globalDir, projectDir, skill); err != nil {
			return fmt.Errorf("symlink %s: %w", skill, err)
		}
	}
	return nil
}

func buildConfig(agentName string, extracted []string) *config.Config {
	inventory := make([]config.SkillInventory, len(extracted))
	for i, name := range extracted {
		inventory[i] = config.SkillInventory{
			Name:    name,
			Version: "1.0",
			Source:  "embedded",
		}
	}

	return &config.Config{
		Version:    version.Version,
		Agent:      agentName,
		SkillCount: len(extracted),
		CreatedAt:  time.Now().UTC(),
		MutationTesting: config.MutationTesting{
			Enabled: false,
		},
		SkillInventory: inventory,
	}
}

func buildRollbackManifest(cfg *config.Config, extracted []string, globalDir, projectDir string) *config.RollbackManifest {
	var paths []string

	paths = append(paths, filepath.Join(cfg.HomeDir, ".archon", "config.yaml"))
	paths = append(paths, filepath.Join(cfg.HomeDir, ".archon", "rollback.json"))

	for _, skill := range extracted {
		paths = append(paths, filepath.Join(projectDir, skill))
	}

	return &config.RollbackManifest{
		Version:      version.Version,
		CreatedPaths: paths,
	}
}

func writeTemplate(projectDir, agentName string, skillCount int) error {
	data := TemplateData{
		ProjectName:    filepath.Base(projectDir),
		Agent:          agentName,
		HarnessVersion: version.Version,
		SkillCount:     skillCount,
	}

	var content string
	var filename string

	switch agentName {
	case "claude":
		content, _ = RenderClaudeMD(data)
		filename = "CLAUDE.md"
	default:
		content, _ = RenderAgentsMD(data)
		filename = "AGENTS.md"
	}

	path := filepath.Join(projectDir, filename)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write template: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename template: %w", err)
	}

	return nil
}

func createOpenSpecDir(projectDir string) error {
	dirs := []string{
		filepath.Join(projectDir, "openspec"),
		filepath.Join(projectDir, "openspec", "changes"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}

	return nil
}
