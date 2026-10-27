package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/internal/mapgen"
	"github.com/archon-ai/archon/internal/route"
	"github.com/archon-ai/archon/internal/scaffold"
	"github.com/archon-ai/archon/internal/skillregistry"
	"github.com/archon-ai/archon/internal/status"
	"github.com/archon-ai/archon/internal/tui"
	"github.com/archon-ai/archon/internal/version"
	"github.com/archon-ai/archon/internal/workflow"
	"github.com/archon-ai/archon/skills"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd(os.Stdout, os.Stderr).Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "archon",
		Short:         "Archon AI orchestration harness",
		Long:          "Bootstrap and manage the SDD orchestration harness for AI-assisted development.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.SetOut(stdout)
	root.SetErr(stderr)

	root.AddCommand(
		newInitCmd(stdout, stderr),
		newUpdateCmd(stdout, stderr),
		newRollbackCmd(stdout, stderr),
		newVersionCmd(stdout),
		newStatusCmd(stdout, stderr),
		newConfigCmd(stdout, stderr),
		newTuiCmd(stdout, stderr),
		newMapCmd(stdout, stderr),
		newRouteCmd(stdout, stderr),
		newSkillRegistryCmd(stdout, stderr),
		newWorkflowCmd(stdout, stderr),
	)

	return root
}

// embeddedSkillCount counts the embedded skill directories (those carrying a
// SKILL.md), so dry-run output and hints reflect the real shipped set rather
// than a hardcoded number.
func embeddedSkillCount() int {
	entries, err := fs.ReadDir(skills.FS, ".")
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := fs.Stat(skills.FS, entry.Name()+"/SKILL.md"); err == nil {
			count++
		}
	}
	return count
}

func newInitCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		agentFlag        string
		forceFlag        bool
		dryRunFlag       bool
		playwrightFlag   bool
		securityFlag     bool
		modelFlag        string
		modelExploreFlag string
		modelProposeFlag string
		modelSpecFlag    string
		modelDesignFlag  string
		modelTasksFlag   string
		modelApplyFlag   string
		modelVerifyFlag  string
		modelJudgeFlag   string
		modelArchiveFlag string
		modelLeaderFlag  string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the orchestration harness in the current project",
		Long:  "Extract embedded skills, scaffold agent config, and write orchestrator templates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}

			if dryRunFlag {
				fmt.Fprintln(stdout, "Dry run — no changes will be made.")
				fmt.Fprintf(stdout, "  Project dir: %s\n", projectDir)
				fmt.Fprintf(stdout, "  Home dir:    %s\n", homeDir)
				if agentFlag != "" {
					fmt.Fprintf(stdout, "  Agent:       %s (override)\n", agentFlag)
				} else {
					fmt.Fprintln(stdout, "  Agent:       (auto-detect)")
				}
				fmt.Fprintf(stdout, "  Force:       %t\n", forceFlag)
				fmt.Fprintf(stdout, "  Skills:      %d embedded skills would be extracted\n", embeddedSkillCount())
				return nil
			}

			modelFlags := map[string]string{
				"explore": modelExploreFlag,
				"propose": modelProposeFlag,
				"spec":    modelSpecFlag,
				"design":  modelDesignFlag,
				"tasks":   modelTasksFlag,
				"apply":   modelApplyFlag,
				"verify":  modelVerifyFlag,
				"archive": modelArchiveFlag,
			}
			if modelJudgeFlag != "" {
				modelFlags["judge"] = modelJudgeFlag
			} else {
				modelFlags["judge"] = modelVerifyFlag
			}

			for _, v := range modelFlags {
				if w := config.Validate(v); w != "" {
					fmt.Fprintln(stderr, w)
				}
			}
			if w := config.Validate(modelFlag); w != "" {
				fmt.Fprintln(stderr, w)
			}
			// The leader model is an opencode provider/model-id (e.g.
			// "anthropic/claude-sonnet-4-...") rather than a Claude family alias,
			// so only run the Claude-oriented advisory check when the value is not
			// already in provider/model-id form (contains "/").
			if !strings.Contains(modelLeaderFlag, "/") {
				if w := config.Validate(modelLeaderFlag); w != "" {
					fmt.Fprintln(stderr, w)
				}
			}

			opts := initcmd.Options{
				HomeDir:           homeDir,
				ProjectDir:        projectDir,
				Agent:             agentFlag,
				Force:             forceFlag,
				EmbeddedFS:        skills.FS,
				ModelDefault:      modelFlag,
				ModelPhases:       modelFlags,
				ModelLeader:       modelLeaderFlag,
				Playwright:        playwrightFlag,
				Security:          securityFlag,
				OverwriteTemplate: forceFlag,
			}

			result, err := initcmd.Run(opts)
			if errors.Is(err, initcmd.ErrTemplateExists) {
				// An orchestrator file already exists. Ask before replacing it;
				// if the user declines, do not initialize at all.
				if !confirmOverwrite(cmd.InOrStdin(), stdout, err) {
					fmt.Fprintln(stdout, "Init cancelled — existing orchestrator file kept unchanged.")
					return nil
				}
				opts.OverwriteTemplate = true
				result, err = initcmd.Run(opts)
			}
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			fmt.Fprintf(stdout, "Archon harness initialized successfully.\n")
			fmt.Fprintf(stdout, "  Agent:    %s\n", result.Agent)
			fmt.Fprintf(stdout, "  Skills:   %d extracted\n", result.ExtractedCount)
			fmt.Fprintf(stdout, "  Config:   %s\n", result.ConfigPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&agentFlag, "agent", "", "Override agent detection (opencode, claude, agents, codex)")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force re-initialization, replacing an existing orchestrator file without prompting")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would happen without making changes")
	cmd.Flags().BoolVar(&playwrightFlag, "playwright", false, "Enable Playwright E2E test generation and execution for web projects")
	cmd.Flags().BoolVar(&securityFlag, "security", false, "Enable the security-baseline gate (propose/spec/tasks/verify/judge hooks)")
	cmd.Flags().StringVar(&modelFlag, "model", "", "Default AI model for all SDD phases")
	cmd.Flags().StringVar(&modelExploreFlag, "model-explore", "", "Model for the explore phase")
	cmd.Flags().StringVar(&modelProposeFlag, "model-propose", "", "Model for the propose phase")
	cmd.Flags().StringVar(&modelSpecFlag, "model-spec", "", "Model for the spec phase")
	cmd.Flags().StringVar(&modelDesignFlag, "model-design", "", "Model for the design phase")
	cmd.Flags().StringVar(&modelTasksFlag, "model-tasks", "", "Model for the tasks phase")
	cmd.Flags().StringVar(&modelApplyFlag, "model-apply", "", "Model for the apply phase")
	cmd.Flags().StringVar(&modelVerifyFlag, "model-verify", "", "Model for the verify phase")
	cmd.Flags().StringVar(&modelJudgeFlag, "model-judge", "", "Model for the judge phase")
	cmd.Flags().StringVar(&modelArchiveFlag, "model-archive", "", "Model for the archive phase")
	cmd.Flags().StringVar(&modelLeaderFlag, "leader", "", "Opencode leader model (provider/model-id) for the archon-leader primary agent")

	return cmd
}

// confirmOverwrite asks the user whether to replace an existing orchestrator
// file. It returns true only on an explicit affirmative answer.
func confirmOverwrite(in io.Reader, stdout io.Writer, cause error) bool {
	fmt.Fprintf(stdout, "An orchestrator file already exists (%v).\n", cause)
	fmt.Fprint(stdout, "Replace it? [y/N]: ")

	reader := bufio.NewReader(in)
	line, _ := reader.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func newUpdateCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		checkFlag bool
		pruneFlag bool
		agentFlag string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh installed skills from the embedded set",
		Long: "Refresh installed skills from the embedded set without rewriting the orchestrator " +
			"template or resetting user config. Skills live in a machine-wide directory, so a refresh " +
			"affects every project symlinked to it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("get home directory: %w", err)
			}

			result, err := initcmd.Update(initcmd.UpdateOptions{
				HomeDir:    homeDir,
				ProjectDir: projectDir,
				Agent:      agentFlag,
				Check:      checkFlag,
				Prune:      pruneFlag,
				EmbeddedFS: skills.FS,
			})
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if result.UpToDate {
				fmt.Fprintln(stdout, "Skills are already up to date — nothing to do.")
				return nil
			}

			rep := result.GapReport
			switch {
			case checkFlag:
				fmt.Fprintln(stdout, "Update check — no changes will be made.")
			case result.CopyMode:
				// In copy-mode the machine-wide skills were re-extracted, but this
				// project keeps its own real copy: it was NOT re-linked, its config
				// was NOT changed, and only `archon init` here can refresh it.
				fmt.Fprintln(stdout, "Machine-wide skills refreshed from the embedded set; this project keeps its own copy and was NOT updated (config unchanged).")
			case result.Wrote:
				fmt.Fprintln(stdout, "Skills refreshed from the embedded set.")
			default:
				fmt.Fprintln(stdout, "No changes were written.")
			}
			fmt.Fprintf(stdout, "  Added:    %d\n", len(rep.Added))
			fmt.Fprintf(stdout, "  Changed:  %d\n", len(rep.Changed))
			fmt.Fprintf(stdout, "  Orphaned: %d\n", len(rep.Orphaned))
			printSkillNames(stdout, "Added", rep.Added)
			printSkillNames(stdout, "Changed", rep.Changed)
			printSkillNames(stdout, "Orphaned", rep.Orphaned)

			if len(rep.Orphaned) > 0 && !pruneFlag && !checkFlag {
				fmt.Fprintln(stdout, "  Orphaned skills were kept. Re-run with --prune to remove them.")
			}
			if len(result.Pruned) > 0 {
				fmt.Fprintf(stdout, "  Pruned:   %d orphaned skill(s) removed\n", len(result.Pruned))
			}

			fmt.Fprintf(stdout, "\nScope: skills are machine-wide (%s) — this refresh affects all symlinked projects.\n", result.GlobalSkillsDir)

			if result.CopyMode {
				fmt.Fprintf(stderr, "Warning: %s\n", result.CopyModeWarning)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Report the diff (added/changed/orphaned) without writing anything")
	cmd.Flags().BoolVar(&pruneFlag, "prune", false, "Remove orphaned skills (installed but no longer embedded)")
	cmd.Flags().StringVar(&agentFlag, "agent", "", "Override the agent recorded in config (opencode, claude, agents, codex)")

	return cmd
}

func printSkillNames(w io.Writer, label string, changes []scaffold.SkillChange) {
	for _, c := range changes {
		fmt.Fprintf(w, "    %-9s %s\n", label+":", c.Name)
	}
}

func newRollbackCmd(stdout, stderr io.Writer) *cobra.Command {
	var dryRunFlag bool

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Remove all files created by archon init",
		Long:  "Read the rollback manifest and remove all created paths in reverse order.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			manifest, err := config.LoadManifest(projectDir)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Fprintln(stdout, "Nothing to rollback — no archon initialization found.")
					return nil
				}
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if dryRunFlag {
				fmt.Fprintln(stdout, "Dry run — the following paths would be removed:")
				for _, p := range manifest.CreatedPaths {
					fmt.Fprintf(stdout, "  %s\n", p)
				}
				if manifest.BackupPath != "" {
					fmt.Fprintf(stdout, "  AGENTS.md would be restored from: %s\n", manifest.BackupPath)
				}
				return nil
			}

			if err := manifest.Cleanup(); err != nil {
				fmt.Fprintf(stderr, "Error during rollback: %v\n", err)
				return err
			}

			fmt.Fprintln(stdout, "Rollback complete — all archon files removed.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would be removed without making changes")

	return cmd
}

func newVersionCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the archon version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(stdout, version.Print())
		},
	}
}

func newStatusCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current harness status",
		Long:  "Read .archon/config.yaml and display agent, harness version, and skill inventory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg := &config.Config{HomeDir: projectDir}
			projectFS := os.DirFS(projectDir)
			if err := cfg.Load(projectFS); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Fprintln(stderr, "No archon configuration found. Run 'archon init' first.")
					return fmt.Errorf("not initialized")
				}
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			// Compute the update hint count. The status command must never fail
			// because of the hint, so any detection error degrades to 0.
			n := 0
			homeDir, err := os.UserHomeDir()
			if err == nil {
				globalSkillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
				if report, err := scaffold.ClassifyGaps(skills.FS, globalSkillsDir); err == nil {
					n = len(report.Added) + len(report.Changed)
				}
			}

			status.DisplayWithUpdate(stdout, cfg, n)
			return nil
		},
	}
}

// newMapCmd regenerates or checks the openspec vault map (openspec/map.md).
func newMapCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		checkFlag    bool
		backfillFlag bool
	)

	cmd := &cobra.Command{
		Use:   "map",
		Short: "Regenerate the openspec vault map (openspec/map.md)",
		Long:  "Walk openspec/specs and openspec/changes, then regenerate the managed region of openspec/map.md.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			switch {
			case backfillFlag:
				changes, _ := mapgen.ArchivedChangeNames(projectDir)
				for _, c := range changes {
					fmt.Fprintf(stdout, "Backfilling %s...\n", c.Name)
				}
				if err := mapgen.Backfill(projectDir); err != nil {
					fmt.Fprintf(stderr, "Error: %v\n", err)
					return err
				}
				fmt.Fprintln(stdout, "Backfill complete.")
				return nil
			case checkFlag:
				issues, err := mapgen.Check(projectDir)
				if err != nil {
					fmt.Fprintf(stderr, "Error: %v\n", err)
					return err
				}
				if len(issues) == 0 {
					fmt.Fprintln(stdout, "openspec/map.md is up to date; no issues found.")
					return nil
				}
				for _, issue := range issues {
					fmt.Fprintf(stderr, "%s: %s: %s\n", issue.File, issue.Kind, issue.Detail)
				}
				return fmt.Errorf("map --check: %d issue(s) found", len(issues))
			default:
				if err := mapgen.Generate(projectDir); err != nil {
					fmt.Fprintf(stderr, "Error: %v\n", err)
					return err
				}
				fmt.Fprintln(stdout, "openspec/map.md regenerated.")
				return nil
			}
		},
	}

	cmd.Flags().BoolVar(&checkFlag, "check", false, "Check map.md and links for staleness without writing")
	cmd.Flags().BoolVar(&backfillFlag, "backfill", false, "Rewrite boundary-crossing links in archived changes and regenerate map.md")

	return cmd
}

// newSkillRegistryCmd groups skill-registry subcommands.
func newSkillRegistryCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill-registry",
		Short: "Manage the cross-tool skill registry (.atl/skill-registry.md)",
	}
	cmd.AddCommand(newSkillRegistryRefreshCmd(stdout, stderr))
	return cmd
}

// newSkillRegistryRefreshCmd rescans project and user-level skill
// directories and rewrites .atl/skill-registry.md. This is the delegator-
// facing index skills/_shared/skill-resolver.md and skills/judgment-day
// depend on to hand fresh subagents exact skill paths — subagents have no
// native visibility into a project's skills/ directory, so without this
// index they start every delegation blind.
func newSkillRegistryRefreshCmd(stdout, stderr io.Writer) *cobra.Command {
	var forceFlag bool

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Rescan project and user skill directories and rewrite .atl/skill-registry.md",
		Long:  "Scans <project>/skills plus the fixed user-level skill roots (~/.pi/agent/skills, ~/.agents/skills, ~/.config/opencode/skills, ~/.claude/skills, ~/.codex/skills) for */SKILL.md, deduplicates by skill name (project wins, then first-scanned user root), and rewrites .atl/skill-registry.md. Skips the rewrite when the scanned content fingerprint matches the cache, unless --force is set.",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			home, err := os.UserHomeDir()
			if err != nil {
				home = ""
			}

			res, err := skillregistry.Refresh(projectDir, home)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			path, cacheHit, err := skillregistry.Write(projectDir, res, forceFlag, time.Now())
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if cacheHit {
				fmt.Fprintf(stdout, "%s: cache hit, %d skill(s) indexed, unchanged.\n", path, len(res.Entries))
				return nil
			}
			fmt.Fprintf(stdout, "%s: regenerated, %d skill(s) indexed.\n", path, len(res.Entries))
			return nil
		},
	}

	cmd.Flags().BoolVar(&forceFlag, "force", false, "Regenerate even if the scanned content fingerprint matches the cache")

	return cmd
}

// newRouteCmd resolves which SDD phase a natural-language message targets:
// the deterministic code pre-router (internal/route). It NEVER executes a
// phase, NEVER writes state, and NEVER invokes an LLM — it emits JSON on
// stdout (the machine-readable contract) plus a human echo line on stderr.
func newRouteCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		changeFlag string
		phaseFlag  string
		statusFlag string
		jsonFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "route \"<message>\"",
		Short: "Resolve which SDD phase a natural-language message targets",
		Long:  "Deterministic code pre-router: applies explicit-agent, control, implicit, dual-action (D3), and keyword rules against MSG plus discovered/overridden state, emitting {phase,rule,path,active_change} JSON on stdout and a human echo line on stderr. Falls through to \"CLASSIFY\" when no code rule fires.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			activeChange := route.ActiveChange(projectDir, changeFlag)

			phase, status := route.ReadState(projectDir, activeChange)
			if phaseFlag != "" {
				phase = phaseFlag
			}
			if statusFlag != "" {
				status = statusFlag
			}

			result := route.Resolve(route.Input{
				Message:      args[0],
				Phase:        phase,
				Status:       status,
				ActiveChange: activeChange,
			})

			out, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("marshal route result: %w", err)
			}
			fmt.Fprintln(stdout, string(out))
			fmt.Fprintf(stderr, "→ Router: archon-%s  (rule: %s, active-change: %s)\n", result.Phase, result.Rule, result.ActiveChange)
			return nil
		},
	}

	cmd.Flags().StringVar(&changeFlag, "change", "", "Override active-change discovery")
	cmd.Flags().StringVar(&phaseFlag, "phase", "", "Override the discovered current phase (testability)")
	cmd.Flags().StringVar(&statusFlag, "status", "", "Override the discovered current status (testability)")
	cmd.Flags().BoolVar(&jsonFlag, "json", true, "Emit JSON to stdout (always on; flag is a no-op kept for CLI-contract compatibility)")

	return cmd
}

// newWorkflowCmd groups the state.yaml-mutating and read-only subcommands
// that implement skills/harness-workflow/SKILL.md's Decision Gates and
// skills/harness-judge/SKILL.md's retry cap in code (internal/workflow),
// instead of trusting an orchestrating LLM to reimplement that algorithm
// correctly from prose every time.
func newWorkflowCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Validate and mutate openspec/changes/{name}/state.yaml (the SDD phase state machine)",
	}
	cmd.AddCommand(
		newWorkflowStatusCmd(stdout, stderr),
		newWorkflowAdvanceCmd(stdout, stderr),
		newWorkflowCompleteCmd(stdout, stderr),
		newWorkflowRetryCmd(stdout, stderr),
	)
	return cmd
}

// resolveWorkflowChange applies the same active-change precedence as `archon
// route` (flag > SESSION_STATUS.md > sole folder > none), so every workflow
// subcommand resolves "which change" identically to the router.
func resolveWorkflowChange(projectDir, changeFlag string) (string, error) {
	activeChange := route.ActiveChange(projectDir, changeFlag)
	if activeChange == "" || activeChange == "none" {
		return "", fmt.Errorf("no active change found (pass --change, or set SESSION_STATUS.md's \"Active change\", or leave exactly one folder under openspec/changes/)")
	}
	return activeChange, nil
}

func newWorkflowStatusCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		changeFlag string
		jsonFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report the current phase/status/retry-count for a change, and render its Phase History block",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			changeName, err := resolveWorkflowChange(projectDir, changeFlag)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			state, err := workflow.ReadState(workflow.StatePath(projectDir, changeName))
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if jsonFlag {
				out, err := json.Marshal(map[string]any{
					"change":      changeName,
					"phase":       state.Phase,
					"status":      state.Status,
					"track":       state.Track,
					"size":        state.Size,
					"retry_count": state.RetryCount,
				})
				if err != nil {
					return fmt.Errorf("marshal status: %w", err)
				}
				fmt.Fprintln(stdout, string(out))
				return nil
			}

			history, err := state.PhaseHistoryMarkdown()
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			fmt.Fprintf(stdout, "## Workflow State\n\n")
			fmt.Fprintf(stdout, "**Change**: %s\n", changeName)
			fmt.Fprintf(stdout, "%s\n", state.StatusLine())
			if state.RetryCount > 0 {
				fmt.Fprintf(stdout, "**Judge retry count**: %d/%d\n", state.RetryCount, workflow.MaxJudgeRetries)
			}
			fmt.Fprintf(stdout, "\n")
			fmt.Fprintf(stdout, "Paste the block below into SESSION_STATUS.md's \"Current phase\" line and \"Phase History\" section as-is — both are generated from state.yaml, not hand-typed:\n\n")
			fmt.Fprintf(stdout, "%s\n%s", state.StatusLine(), history)
			return nil
		},
	}

	cmd.Flags().StringVar(&changeFlag, "change", "", "Override active-change discovery")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Emit JSON instead of the human report")

	return cmd
}

func newWorkflowAdvanceCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		changeFlag string
		toFlag     string
		trackFlag  string
		jsonFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "advance",
		Short: "Validate and, if allowed, perform a phase transition (writes state.yaml atomically)",
		Long:  "Implements skills/harness-workflow/SKILL.md's Decision Gates table. On an allowed transition, writes state.yaml (phase, status=in_progress, appended history entry) and best-effort regenerates openspec/map.md -- a map regen failure is a warning, never a rollback of the state.yaml write. On a blocked transition, state.yaml is left untouched and the command exits non-zero. If state.yaml does not exist yet for this change, --to must name the selected track's first phase; a fresh state.yaml is bootstrapped with --track (default \"full\") instead of erroring.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if toFlag == "" {
				return fmt.Errorf("--to is required")
			}
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			changeName, err := resolveWorkflowChange(projectDir, changeFlag)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			statePath := workflow.StatePath(projectDir, changeName)
			state, err := workflow.ReadState(statePath)
			if err != nil {
				if !errors.Is(err, workflow.ErrStateNotFound) {
					fmt.Fprintf(stderr, "Error: %v\n", err)
					return err
				}
				// No state.yaml yet for this change: bootstrap a fresh one
				// on the selected track. CheckTransition only allows this
				// when --to names that track's first phase.
				state = &workflow.State{Track: trackFlag}
			}

			result, err := state.Advance(toFlag, time.Now())
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if !result.Allowed {
				if jsonFlag {
					out, _ := json.Marshal(map[string]any{"allowed": false, "reason": result.Reason, "missing_phases": result.MissingPhases})
					fmt.Fprintln(stdout, string(out))
				}
				fmt.Fprintf(stderr, "blocked: %s\n", result.Reason)
				if len(result.MissingPhases) > 0 {
					fmt.Fprintf(stderr, "missing phases: %s\n", strings.Join(result.MissingPhases, ", "))
				}
				return fmt.Errorf("transition blocked: %s", result.Reason)
			}

			if err := state.Write(statePath); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if err := mapgen.Generate(projectDir); err != nil {
				fmt.Fprintf(stderr, "Warning: openspec/map.md regen failed (state.yaml write stands): %v\n", err)
			}

			if jsonFlag {
				out, _ := json.Marshal(map[string]any{"allowed": true, "resuming": result.Resuming, "phase": state.Phase, "status": state.Status})
				fmt.Fprintln(stdout, string(out))
				return nil
			}
			fmt.Fprintf(stdout, "%s advanced to %s (in_progress)\n", changeName, state.Phase)
			return nil
		},
	}

	cmd.Flags().StringVar(&changeFlag, "change", "", "Override active-change discovery")
	cmd.Flags().StringVar(&toFlag, "to", "", "The requested phase")
	cmd.Flags().StringVar(&trackFlag, "track", "", "Track to bootstrap a new change on (full|bugfix|fast; default full). Ignored once state.yaml exists.")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Emit JSON instead of the human report")

	return cmd
}

// validWorkflowSizes are the only values accepted by --size on `workflow
// complete`, per harness-fast-track's sizing classification (sdd-explore
// Step 3c): small|substantial|unknown.
var validWorkflowSizes = map[string]bool{
	"small":       true,
	"substantial": true,
	"unknown":     true,
}

// validWorkflowTracks are the only values accepted by --track on `workflow
// complete`'s post-explore switch (see newWorkflowCompleteCmd's --track flag).
var validWorkflowTracks = map[string]bool{
	"full":   true,
	"bugfix": true,
	"fast":   true,
}

func newWorkflowCompleteCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		changeFlag string
		sizeFlag   string
		trackFlag  string
	)

	cmd := &cobra.Command{
		Use:   "complete",
		Short: "Mark the change's current phase completed (writes state.yaml atomically)",
		Long:  "Marks the current phase completed and appends a history entry. --size optionally persists sdd-explore's size classification (small|substantial|unknown) onto state.yaml; an empty/absent flag leaves any existing Size untouched. --track optionally switches the change's track (full|bugfix|fast) -- the only point this is accepted is while completing the explore phase (the suggestion-only track-selection moment per harness-fast-track's spec); completing any other phase with --track is rejected. Invalid --size or --track values, or --track on a non-explore phase, are rejected before any write, leaving state.yaml unchanged.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if sizeFlag != "" && !validWorkflowSizes[sizeFlag] {
				err := fmt.Errorf("invalid --size: %s (supported: small, substantial, unknown)", sizeFlag)
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}
			if trackFlag != "" && !validWorkflowTracks[trackFlag] {
				err := fmt.Errorf("invalid --track: %s (supported: full, bugfix, fast)", trackFlag)
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			changeName, err := resolveWorkflowChange(projectDir, changeFlag)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			statePath := workflow.StatePath(projectDir, changeName)
			state, err := workflow.ReadState(statePath)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}
			if state.Status == "completed" {
				err := fmt.Errorf("%s is already completed — run archon workflow advance to start the next phase, not complete again", state.Phase)
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}
			if trackFlag != "" && state.Phase != "explore" {
				err := fmt.Errorf("--track is only accepted while completing the explore phase (current phase: %s)", state.Phase)
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			state.Complete(time.Now())
			if sizeFlag != "" {
				state.Size = sizeFlag
			}
			if trackFlag != "" {
				state.Track = trackFlag
			}

			if err := state.Write(statePath); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			fmt.Fprintf(stdout, "%s: %s marked completed\n", changeName, state.Phase)
			return nil
		},
	}

	cmd.Flags().StringVar(&changeFlag, "change", "", "Override active-change discovery")
	cmd.Flags().StringVar(&sizeFlag, "size", "", "Persist sdd-explore's size classification (small|substantial|unknown); omit to leave Size untouched")
	cmd.Flags().StringVar(&trackFlag, "track", "", "Switch the change's track (full|bugfix|fast); only accepted while completing the explore phase")

	return cmd
}

func newWorkflowRetryCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retry",
		Short: "Manage the judge phase's persisted retry count",
	}

	var incChangeFlag string
	incCmd := &cobra.Command{
		Use:   "increment",
		Short: "Increment the judge retry count, or fail if the cap (3) is already reached",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			changeName, err := resolveWorkflowChange(projectDir, incChangeFlag)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			statePath := workflow.StatePath(projectDir, changeName)
			state, err := workflow.ReadState(statePath)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			count, err := state.IncrementRetry()
			if err != nil {
				fmt.Fprintf(stdout, "max_retries_exceeded: true\n")
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			if err := state.Write(statePath); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			fmt.Fprintf(stdout, "retry_count: %d/%d\n", count, workflow.MaxJudgeRetries)
			return nil
		},
	}
	incCmd.Flags().StringVar(&incChangeFlag, "change", "", "Override active-change discovery")

	var resetChangeFlag string
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the judge retry count to 0 (call once the single focused review passes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			changeName, err := resolveWorkflowChange(projectDir, resetChangeFlag)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			statePath := workflow.StatePath(projectDir, changeName)
			state, err := workflow.ReadState(statePath)
			if err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			state.ResetRetry()

			if err := state.Write(statePath); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			fmt.Fprintf(stdout, "retry_count: 0/%d\n", workflow.MaxJudgeRetries)
			return nil
		},
	}
	resetCmd.Flags().StringVar(&resetChangeFlag, "change", "", "Override active-change discovery")

	cmd.AddCommand(incCmd, resetCmd)
	return cmd
}

func newTuiCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Interactive terminal UI for configuration",
		Long:  "Launch a TUI to configure models, mutation testing, and agent settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tui.CheckTerminal(); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return fmt.Errorf("tui requires a terminal")
			}

			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg := &config.Config{HomeDir: projectDir}
			projectFS := os.DirFS(projectDir)
			if err := cfg.Load(projectFS); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// Not initialized yet: launch the TUI with a default config
					// so the user can pick an agent and init from the Agent tab.
					fmt.Fprintln(stdout, "No archon configuration found — open the Agent tab to initialize.")
					cfg = &config.Config{HomeDir: projectDir}
				} else {
					fmt.Fprintf(stderr, "Error: %v\n", err)
					return err
				}
			}

			model := tui.NewModel(cfg, projectDir)
			p := tea.NewProgram(model, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(stderr, "Error: %v\n", err)
				return err
			}

			return nil
		},
	}
}
