package initcmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/scaffold"
	"github.com/archon-ai/archon/internal/version"
)

// UpdateOptions configures a single `archon update` invocation.
type UpdateOptions struct {
	HomeDir    string
	ProjectDir string
	// Agent overrides the agent resolved from the existing config. Empty means
	// "use whatever the config already records".
	Agent string
	// Check requests a dry run: classify and report the gap, write nothing.
	Check bool
	// Prune removes orphaned skills (installed but no longer embedded) from the
	// machine-wide global dir and the project link.
	Prune      bool
	EmbeddedFS fs.FS
}

// UpdateResult reports what `Update` found and did. It always carries the gap
// classification and the machine-wide global skills dir so the caller can
// surface scope; Wrote and Pruned reflect whether the filesystem changed.
type UpdateResult struct {
	GapReport       scaffold.GapReport
	GlobalSkillsDir string
	// CopyMode is true when the project's installed skills are real directories
	// rather than symlinks. In that case the global dir is still refreshed but
	// the project is NOT re-linked, and CopyModeWarning explains why.
	CopyMode        bool
	CopyModeWarning string
	// Wrote is true when skills were re-extracted/relinked and config was saved.
	Wrote bool
	// UpToDate is true when there were no gaps and nothing was written.
	UpToDate bool
	// Pruned lists orphaned skills removed (only when Prune is set).
	Pruned []string
}

// Update refreshes installed skills from the embedded set without ever rewriting
// the orchestrator template or resetting user config. It preserves models,
// playwright, mutation_testing, judge, created_at, and agent; only
// harness_version, skill_count, and skill_inventory may change.
func Update(opts UpdateOptions) (*UpdateResult, error) {
	if opts.HomeDir == "" {
		return nil, fmt.Errorf("home directory is required")
	}
	if opts.ProjectDir == "" {
		return nil, fmt.Errorf("project directory is required")
	}
	if opts.EmbeddedFS == nil {
		return nil, fmt.Errorf("embedded filesystem is required")
	}

	// Scenario: "Update before init reports an actionable error" — a missing
	// config must yield an actionable error and write nothing.
	cfg := &config.Config{HomeDir: opts.ProjectDir}
	if err := cfg.Load(os.DirFS(opts.ProjectDir)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no archon configuration found in %s; run 'archon init' first", opts.ProjectDir)
		}
		return nil, fmt.Errorf("load config: %w", err)
	}

	agentName := cfg.Agent
	if opts.Agent != "" {
		if !knownAgent(opts.Agent) {
			return nil, fmt.Errorf("unknown agent %q (valid: opencode, claude, agents, codex)", opts.Agent)
		}
		agentName = opts.Agent
	}
	if agentName == "" {
		return nil, fmt.Errorf("config does not record an agent; run 'archon init' first")
	}

	globalSkillsDir := filepath.Join(opts.HomeDir, ".config", "opencode", "skills")
	projectSkillsDir := resolveProjectSkillsDir(opts.ProjectDir, agentName)

	report, err := scaffold.ClassifyGaps(opts.EmbeddedFS, globalSkillsDir)
	if err != nil {
		return nil, fmt.Errorf("classify gaps: %w", err)
	}

	result := &UpdateResult{
		GapReport:       report,
		GlobalSkillsDir: globalSkillsDir,
	}

	// Scenario: "No gaps reports already up to date" — nothing to do, write nothing.
	if len(report.Added) == 0 && len(report.Changed) == 0 && len(report.Orphaned) == 0 {
		result.UpToDate = true
		return result, nil
	}

	// Scenario: "Copy-mode install warns without re-linking" — if the project
	// link is a real directory (not a symlink), the project keeps its own copy
	// and must update itself; we never re-link over a real directory.
	if cm, warn := detectCopyMode(projectSkillsDir); cm {
		result.CopyMode = true
		result.CopyModeWarning = warn
	}

	// Scenario: "Check reports the diff without writing" — report and return,
	// no filesystem mutation.
	if opts.Check {
		return result, nil
	}

	// Refresh the machine-wide global dir. For copy-mode projects we still
	// refresh the global dir but do NOT relink the project (createSymlinks would
	// overwrite the real directory), so the warning stands on its own.
	if result.CopyMode {
		if _, err := scaffold.Extract(opts.EmbeddedFS, globalSkillsDir); err != nil {
			return nil, fmt.Errorf("extract skills: %w", err)
		}
	} else {
		res, err := refreshSkills(opts.HomeDir, opts.ProjectDir, agentName, opts.EmbeddedFS)
		if err != nil {
			return nil, err
		}
		// Config preservation (spec requirement): clone the loaded config and
		// patch ONLY Version, SkillCount, SkillInventory. models, playwright,
		// mutation_testing, judge, created_at, and agent ride through untouched.
		next := cfg.Clone()
		next.Version = version.Version
		next.SkillCount = len(res.Extracted)
		next.SkillInventory = res.Inventory
		next.HomeDir = opts.ProjectDir
		if err := next.Save(); err != nil {
			return nil, fmt.Errorf("save config: %w", err)
		}
		result.Wrote = true
	}

	// Scenario: "Prune removes orphaned skills" / "Orphans are kept without
	// prune" — orphans are only removed when --prune is set.
	if opts.Prune {
		for _, orphan := range report.Orphaned {
			globalPath := filepath.Join(globalSkillsDir, orphan.Name)
			if err := os.RemoveAll(globalPath); err != nil {
				return nil, fmt.Errorf("prune orphan %s: %w", orphan.Name, err)
			}
			// In copy-mode the project owns a real (non-symlink) copy of its
			// skills, and update promised not to touch it. Removing the project
			// path here would delete that real orphan copy, so we only prune the
			// global orphan and leave the project alone. In the normal (symlink)
			// case we also remove the project link so it stops pointing at a
			// pruned global skill.
			if !result.CopyMode {
				projectPath := filepath.Join(projectSkillsDir, orphan.Name)
				if err := os.RemoveAll(projectPath); err != nil {
					return nil, fmt.Errorf("prune orphan link %s: %w", orphan.Name, err)
				}
			}
			result.Pruned = append(result.Pruned, orphan.Name)
		}
		result.Wrote = true
	}

	return result, nil
}

// detectCopyMode reports whether the project's installed skills are real
// directories rather than symlinks. It scans every installed skill entry and
// treats the project as copy-mode if any entry is a real directory (not a
// symlink), since init links each skill individually and a copy-mode install
// materializes them all as real directories.
func detectCopyMode(projectSkillsDir string) (bool, string) {
	entries, err := os.ReadDir(projectSkillsDir)
	if err != nil {
		return false, ""
	}

	for _, e := range entries {
		link := filepath.Join(projectSkillsDir, e.Name())
		info, err := os.Lstat(link)
		if err != nil {
			continue
		}
		// A symlink is the expected (linked) install; skip it.
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		// A real directory containing a SKILL.md is a copy-mode install.
		if info.IsDir() {
			if _, statErr := os.Stat(filepath.Join(link, "SKILL.md")); statErr == nil {
				return true, "this project's skills are installed as real directories (copy-mode), not symlinks; " +
					"the machine-wide skills were refreshed but this project was NOT re-linked — " +
					"re-run 'archon init' here to refresh the project's own copy"
			}
		}
	}

	return false, ""
}
