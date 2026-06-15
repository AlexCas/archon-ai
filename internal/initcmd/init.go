package initcmd

import (
	"errors"
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
	HomeDir      string
	ProjectDir   string
	Agent        string
	Force        bool
	EmbeddedFS   fs.FS
	ModelDefault string
	ModelPhases  map[string]string
	// Playwright enables generation and execution of Playwright E2E tests.
	Playwright bool
	// OverwriteTemplate, when true, replaces an existing orchestrator file
	// (CLAUDE.md / AGENTS.md) without prompting. When false and the file
	// already exists, Run aborts with ErrTemplateExists so the caller can
	// ask the user whether to replace it.
	OverwriteTemplate bool
}

// ErrTemplateExists is returned by Run when the orchestrator template file
// (CLAUDE.md or AGENTS.md) already exists and OverwriteTemplate is false.
var ErrTemplateExists = errors.New("orchestrator file already exists")

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

	// Guard an existing orchestrator file BEFORE doing any work so that a
	// declined overwrite leaves the project completely untouched.
	templatePath := templateFilePath(opts.ProjectDir, agentName)
	if !opts.OverwriteTemplate {
		if _, statErr := os.Stat(templatePath); statErr == nil {
			return nil, fmt.Errorf("%w: %s", ErrTemplateExists, templatePath)
		}
	}

	// Ensure the agent directory exists. init no longer depends on the folder
	// having been created beforehand — selecting an agent creates its folder.
	if err := ensureAgentDir(opts.ProjectDir, agentName); err != nil {
		return nil, fmt.Errorf("create agent dir: %w", err)
	}

	res, err := refreshSkills(opts.HomeDir, opts.ProjectDir, agentName, opts.EmbeddedFS)
	if err != nil {
		return nil, err
	}
	globalSkillsDir := res.GlobalSkillsDir
	projectSkillsDir := res.ProjectSkillsDir
	extracted := res.Extracted

	cfg := buildConfig(agentName, extracted, res.Inventory, opts.ModelDefault, opts.ModelPhases, opts.Playwright)
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
	// An explicit agent selection always wins and does NOT require the folder
	// to already exist — init will create it.
	if opts.Agent != "" {
		if !knownAgent(opts.Agent) {
			return "", fmt.Errorf("unknown agent %q (valid: opencode, claude, agents, codex)", opts.Agent)
		}
		return opts.Agent, nil
	}

	projectFS := os.DirFS(opts.ProjectDir)
	result, err := agent.Detect(projectFS)
	if err != nil {
		return "", fmt.Errorf("%w; pass --agent to select one (opencode, claude, agents, codex)", err)
	}

	return result.Agent, nil
}

func knownAgent(name string) bool {
	switch name {
	case "opencode", "claude", "agents", "codex":
		return true
	default:
		return false
	}
}

// agentBaseDir maps an agent name to its top-level directory inside the project.
func agentBaseDir(projectDir, agentName string) string {
	switch agentName {
	case "claude":
		return filepath.Join(projectDir, ".claude")
	case "agents":
		return filepath.Join(projectDir, ".agents")
	case "codex":
		return filepath.Join(projectDir, ".codex")
	default:
		return filepath.Join(projectDir, ".opencode")
	}
}

// ensureAgentDir creates the agent's top-level directory if it does not exist.
func ensureAgentDir(projectDir, agentName string) error {
	return os.MkdirAll(agentBaseDir(projectDir, agentName), 0o755)
}

// templateFilePath returns the orchestrator file path for the given agent:
// CLAUDE.md for the claude agent, AGENTS.md otherwise.
func templateFilePath(projectDir, agentName string) string {
	if agentName == "claude" {
		return filepath.Join(projectDir, "CLAUDE.md")
	}
	return filepath.Join(projectDir, "AGENTS.md")
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

func buildConfig(agentName string, extracted []string, inventory []config.SkillInventory, modelDefault string, modelPhases map[string]string, playwright bool) *config.Config {
	var phases map[string]string
	for k, v := range modelPhases {
		if v != "" {
			if phases == nil {
				phases = make(map[string]string)
			}
			phases[k] = v
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
		Playwright: config.Playwright{
			Enabled: playwright,
		},
		Models: config.ModelConfig{
			Default: modelDefault,
			Phases:  phases,
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

	switch agentName {
	case "claude":
		content, _ = RenderClaudeMD(data)
	default:
		content, _ = RenderAgentsMD(data)
	}

	path := templateFilePath(projectDir, agentName)
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
