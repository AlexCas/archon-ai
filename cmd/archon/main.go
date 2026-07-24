package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/internal/mapgen"
	"github.com/archon-ai/archon/internal/scaffold"
	"github.com/archon-ai/archon/internal/status"
	"github.com/archon-ai/archon/internal/tui"
	"github.com/archon-ai/archon/internal/version"
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
