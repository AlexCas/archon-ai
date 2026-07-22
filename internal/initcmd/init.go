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
	ModelLeader  string
	ModelPhases  map[string]string
	// Playwright enables generation and execution of Playwright E2E tests.
	Playwright bool
	// Security enables the security-baseline gate across the SDD phases.
	Security bool
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

	cfg := buildConfig(agentName, extracted, res.Inventory, opts.ModelDefault, opts.ModelLeader, opts.ModelPhases, opts.Playwright, opts.Security)
	cfg.HomeDir = opts.ProjectDir
	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	rollback := buildRollbackManifest(cfg, extracted, globalSkillsDir, projectSkillsDir)
	rollback.HomeDir = opts.ProjectDir

	if err := writeTemplate(opts.ProjectDir, agentName, len(extracted), config.ResolvePhaseModels(cfg.Models)); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	// Merge the archon-leader agent into opencode.json for opencode projects.
	// This must run before WriteManifest so the written path is registered for
	// rollback alongside everything else init created.
	if agentName == "opencode" {
		mergedPath, err := mergeOpencodeAgent(opts.ProjectDir, cfg.Models)
		if err != nil {
			return nil, fmt.Errorf("merge opencode agent: %w", err)
		}
		if mergedPath != "" {
			rollback.CreatedPaths = append(rollback.CreatedPaths, mergedPath)
		}
	}

	// Write one .claude/agents/archon-<phase>.md per resolvable phase for
	// claude projects. Must run before WriteManifest so all written paths are
	// registered for rollback.
	if agentName == "claude" {
		paths, err := writeClaudeAgents(opts.ProjectDir, cfg.Models)
		if err != nil {
			return nil, fmt.Errorf("write claude agents: %w", err)
		}
		rollback.CreatedPaths = append(rollback.CreatedPaths, paths...)
	}

	if err := rollback.WriteManifest(); err != nil {
		return nil, fmt.Errorf("write rollback manifest: %w", err)
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

func buildConfig(agentName string, extracted []string, inventory []config.SkillInventory, modelDefault string, modelLeader string, modelPhases map[string]string, playwright bool, security bool) *config.Config {
	var phases map[string]config.ModelRef
	for k, v := range modelPhases {
		if v != "" {
			if phases == nil {
				phases = make(map[string]config.ModelRef)
			}
			phases[k] = config.ParseModelRef(v)
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
		Judge: config.Judge{
			Enabled: true,
		},
		Playwright: config.Playwright{
			Enabled: playwright,
		},
		Security: config.Security{
			Enabled: security,
		},
		Models: config.ModelConfig{
			Default: config.ParseModelRef(modelDefault),
			Leader:  config.ParseModelRef(modelLeader),
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

func writeTemplate(projectDir, agentName string, skillCount int, phaseModels []config.PhaseModel) error {
	data := TemplateData{
		ProjectName:    filepath.Base(projectDir),
		Agent:          agentName,
		HarnessVersion: version.Version,
		SkillCount:     skillCount,
		PhaseModels:    phaseModels,
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

// mapMDPreamble seeds openspec/map.md when it does not already exist. The
// authored prose above the markers is never touched again; `archon map`
// (internal/mapgen) regenerates only the content between the markers. See
// skills/_shared/spec-vault.md for the full vault convention.
const mapMDPreamble = `# openspec Map

This is the vault's entry node: a generated overview of every capability and
change under openspec/. The managed region below is regenerated by ` + "`archon map`" + `
and MUST NOT be hand-edited. See ` + "`skills/_shared/spec-vault.md`" + ` for the full
vault convention.

<!-- MAP:START -->
<!-- MAP:END -->
`

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

	if err := seedMapMD(projectDir); err != nil {
		return err
	}

	return nil
}

// seedMapMD creates openspec/map.md with the authored preamble and empty
// managed markers if the file does not already exist. It never overwrites an
// existing map.md — regeneration of the managed region is `archon map`'s job.
func seedMapMD(projectDir string) error {
	mapPath := filepath.Join(projectDir, "openspec", "map.md")

	if _, err := os.Stat(mapPath); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", mapPath, err)
	}

	tmp := mapPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(mapMDPreamble), 0o644); err != nil {
		return fmt.Errorf("write map.md: %w", err)
	}
	if err := os.Rename(tmp, mapPath); err != nil {
		return fmt.Errorf("rename map.md: %w", err)
	}

	return nil
}
