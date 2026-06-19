package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
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

			if strings.HasPrefix(key, "models.") {
				if w := config.Validate(value); w != "" {
					fmt.Fprintln(stderr, w)
				}
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

			if cfg.Models.Default.FullID() == "" && len(cfg.Models.Phases) == 0 {
				fmt.Fprintln(stdout, "(none configured)")
				return nil
			}

			if cfg.Models.Default.FullID() != "" {
				fmt.Fprintf(stdout, "models.default = %s\n", cfg.Models.Default.FullID())
			}

			if len(cfg.Models.Phases) > 0 {
				phases := make([]string, 0, len(cfg.Models.Phases))
				for k := range cfg.Models.Phases {
					phases = append(phases, k)
				}
				sort.Strings(phases)
				for _, phase := range phases {
					fmt.Fprintf(stdout, "models.phases.%s = %s\n", phase, cfg.Models.Phases[phase].FullID())
				}
			}

			return nil
		},
	}
}

func parseBool(key, value string) (bool, error) {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean for %q: %q (use true/false)", key, value)
	}
	return b, nil
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "models.default":
		cfg.Models.Default = config.ParseModelRef(value)
		return nil
	case "playwright.enabled":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Playwright.Enabled = b
		return nil
	case "playwright.test_dir":
		cfg.Playwright.TestDir = value
		return nil
	case "playwright.base_url":
		cfg.Playwright.BaseURL = value
		return nil
	case "mutation_testing.enabled":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.MutationTesting.Enabled = b
		return nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			phase := strings.TrimPrefix(key, "models.phases.")
			if !config.ValidPhases[phase] {
				return fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, archive)", phase)
			}
			if cfg.Models.Phases == nil {
				cfg.Models.Phases = make(map[string]config.ModelRef)
			}
			cfg.Models.Phases[phase] = config.ParseModelRef(value)
			return nil
		}
		return fmt.Errorf("unknown config key %q (supported: models.default, models.phases.<phase>, playwright.enabled, playwright.test_dir, playwright.base_url, mutation_testing.enabled)", key)
	}
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "models.default":
		return cfg.Models.Default.FullID(), nil
	case "playwright.enabled":
		return strconv.FormatBool(cfg.Playwright.Enabled), nil
	case "playwright.test_dir":
		return cfg.Playwright.TestDir, nil
	case "playwright.base_url":
		return cfg.Playwright.BaseURL, nil
	case "mutation_testing.enabled":
		return strconv.FormatBool(cfg.MutationTesting.Enabled), nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			phase := strings.TrimPrefix(key, "models.phases.")
			if !config.ValidPhases[phase] {
				return "", fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, archive)", phase)
			}
			return cfg.Models.Phases[phase].FullID(), nil
		}
		return "", fmt.Errorf("unknown config key %q (supported: models.default, models.phases.<phase>, playwright.enabled, playwright.test_dir, playwright.base_url, mutation_testing.enabled)", key)
	}
}
