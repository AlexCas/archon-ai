package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/internal/status"
	"github.com/archon-ai/archon/internal/version"
	"github.com/archon-ai/archon/skills"
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
		newRollbackCmd(stdout, stderr),
		newVersionCmd(stdout),
		newStatusCmd(stdout, stderr),
		newConfigCmd(stdout, stderr),
	)

	return root
}

func newInitCmd(stdout, stderr io.Writer) *cobra.Command {
	var (
		agentFlag        string
		forceFlag        bool
		dryRunFlag       bool
		modelFlag        string
		modelExploreFlag string
		modelProposeFlag string
		modelSpecFlag    string
		modelDesignFlag  string
		modelTasksFlag   string
		modelApplyFlag   string
		modelVerifyFlag  string
		modelArchiveFlag string
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
				fmt.Fprintln(stdout, "  Skills:      21 embedded skills would be extracted")
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

			for _, v := range modelFlags {
				if w := config.Validate(v); w != "" {
					fmt.Fprintln(stderr, w)
				}
			}
			if w := config.Validate(modelFlag); w != "" {
				fmt.Fprintln(stderr, w)
			}

			opts := initcmd.Options{
				HomeDir:      homeDir,
				ProjectDir:   projectDir,
				Agent:        agentFlag,
				Force:        forceFlag,
				EmbeddedFS:   skills.FS,
				ModelDefault: modelFlag,
				ModelPhases:  modelFlags,
			}

			result, err := initcmd.Run(opts)
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
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Force re-initialization even if already initialized")
	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what would happen without making changes")
	cmd.Flags().StringVar(&modelFlag, "model", "", "Default AI model for all SDD phases")
	cmd.Flags().StringVar(&modelExploreFlag, "model-explore", "", "Model for the explore phase")
	cmd.Flags().StringVar(&modelProposeFlag, "model-propose", "", "Model for the propose phase")
	cmd.Flags().StringVar(&modelSpecFlag, "model-spec", "", "Model for the spec phase")
	cmd.Flags().StringVar(&modelDesignFlag, "model-design", "", "Model for the design phase")
	cmd.Flags().StringVar(&modelTasksFlag, "model-tasks", "", "Model for the tasks phase")
	cmd.Flags().StringVar(&modelApplyFlag, "model-apply", "", "Model for the apply phase")
	cmd.Flags().StringVar(&modelVerifyFlag, "model-verify", "", "Model for the verify phase")
	cmd.Flags().StringVar(&modelArchiveFlag, "model-archive", "", "Model for the archive phase")

	return cmd
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

			status.Display(stdout, cfg)
			return nil
		},
	}
}
