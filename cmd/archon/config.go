package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/archon-ai/archon/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage archon configuration",
		Long:  "Read and write .archon/config.yaml using dot-delimited keys.",
	}

	cmd.AddCommand(
		newConfigSetCmd(stdout, stderr),
		newConfigGetCmd(stdout, stderr),
		newConfigListCmd(stdout, stderr),
	)

	return cmd
}

func newConfigSetCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg := &config.Config{HomeDir: projectDir}
			projectFS := os.DirFS(projectDir)
			if err := cfg.Load(projectFS); err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if w := config.Validate(value); w != "" {
				fmt.Fprintln(stderr, w)
			}

			if err := setConfigValue(cfg, key, value); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(stdout, "Set %s = %s\n", key, value)
			return nil
		},
	}
}

func newConfigGetCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg := &config.Config{HomeDir: projectDir}
			projectFS := os.DirFS(projectDir)
			if err := cfg.Load(projectFS); err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			value, err := getConfigValue(cfg, key)
			if err != nil {
				return err
			}

			if value != "" {
				fmt.Fprintln(stdout, value)
			}
			return nil
		},
	}
}

func newConfigListCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all model configuration entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			cfg := &config.Config{HomeDir: projectDir}
			projectFS := os.DirFS(projectDir)
			if err := cfg.Load(projectFS); err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if cfg.Models.Default == "" && len(cfg.Models.Phases) == 0 {
				fmt.Fprintln(stdout, "(none configured)")
				return nil
			}

			if cfg.Models.Default != "" {
				fmt.Fprintf(stdout, "models.default = %s\n", cfg.Models.Default)
			}

			if len(cfg.Models.Phases) > 0 {
				phases := make([]string, 0, len(cfg.Models.Phases))
				for k := range cfg.Models.Phases {
					phases = append(phases, k)
				}
				sort.Strings(phases)
				for _, phase := range phases {
					fmt.Fprintf(stdout, "models.phases.%s = %s\n", phase, cfg.Models.Phases[phase])
				}
			}

			return nil
		},
	}
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "models.default":
		cfg.Models.Default = value
		return nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			phase := strings.TrimPrefix(key, "models.phases.")
			if !config.ValidPhases[phase] {
				return fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, archive)", phase)
			}
			if cfg.Models.Phases == nil {
				cfg.Models.Phases = make(map[string]string)
			}
			cfg.Models.Phases[phase] = value
			return nil
		}
		return fmt.Errorf("unknown config key %q (supported: models.default, models.phases.<phase>)", key)
	}
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "models.default":
		return cfg.Models.Default, nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			phase := strings.TrimPrefix(key, "models.phases.")
			if !config.ValidPhases[phase] {
				return "", fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, archive)", phase)
			}
			return cfg.Models.Phases[phase], nil
		}
		return "", fmt.Errorf("unknown config key %q (supported: models.default, models.phases.<phase>)", key)
	}
}
