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

			if strings.HasPrefix(key, "models.") && !strings.HasSuffix(key, ".base_url") {
				if w := config.Validate(value); w != "" {
					fmt.Fprintln(stderr, w)
				}
			}

			if err := setConfigValue(cfg, key, value); err != nil {
				return err
			}

			// REQ-3: advisory BaseURL validation at CLI set-time, after the
			// value has been written so the ref reflects the new BaseURL.
			if strings.HasSuffix(key, ".base_url") {
				if ref, ok := baseURLRefForKey(cfg, key); ok {
					config.ValidateBaseURL(ref, stderr)
				}
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

			if !cfg.Models.HasAny() {
				fmt.Fprintln(stdout, "(none configured)")
				return nil
			}

			if cfg.Models.Default.FullID() != "" {
				fmt.Fprintf(stdout, "models.default = %s\n", cfg.Models.Default.FullID())
			}
			if cfg.Models.Default.BaseURL != "" {
				fmt.Fprintf(stdout, "models.default.base_url = %s\n", cfg.Models.Default.BaseURL)
			}

			if cfg.Models.Leader.FullID() != "" {
				fmt.Fprintf(stdout, "models.leader = %s\n", cfg.Models.Leader.FullID())
			}
			if cfg.Models.Leader.BaseURL != "" {
				fmt.Fprintf(stdout, "models.leader.base_url = %s\n", cfg.Models.Leader.BaseURL)
			}

			if len(cfg.Models.Phases) > 0 {
				phases := make([]string, 0, len(cfg.Models.Phases))
				for k := range cfg.Models.Phases {
					phases = append(phases, k)
				}
				sort.Strings(phases)
				for _, phase := range phases {
					if cfg.Models.Phases[phase].FullID() != "" {
						fmt.Fprintf(stdout, "models.phases.%s = %s\n", phase, cfg.Models.Phases[phase].FullID())
					}
					if baseURL := cfg.Models.Phases[phase].BaseURL; baseURL != "" {
						fmt.Fprintf(stdout, "models.phases.%s.base_url = %s\n", phase, baseURL)
					}
				}
			}

			return nil
		},
	}
}

// baseURLRefForKey returns the ModelRef addressed by a "*.base_url" set key
// (models.default.base_url or models.phases.<phase>.base_url) so its current
// BaseURL can be advisory-validated after the value has been written.
func baseURLRefForKey(cfg *config.Config, key string) (config.ModelRef, bool) {
	switch {
	case key == "models.default.base_url":
		return cfg.Models.Default, true
	case key == "models.leader.base_url":
		return cfg.Models.Leader, true
	case strings.HasPrefix(key, "models.phases.") && strings.HasSuffix(key, ".base_url"):
		phase := strings.TrimSuffix(strings.TrimPrefix(key, "models.phases."), ".base_url")
		return cfg.Models.Phases[phase], true
	default:
		return config.ModelRef{}, false
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
		ref := config.ParseModelRef(value)
		ref.BaseURL = cfg.Models.Default.BaseURL
		cfg.Models.Default = ref
		return nil
	case "models.default.base_url":
		cfg.Models.Default.BaseURL = value
		return nil
	case "models.leader":
		ref := config.ParseModelRef(value)
		ref.BaseURL = cfg.Models.Leader.BaseURL
		cfg.Models.Leader = ref
		return nil
	case "models.leader.base_url":
		cfg.Models.Leader.BaseURL = value
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
	case "security.enabled":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Security.Enabled = b
		return nil
	case "security.profile":
		if value != "cli" && value != "web" {
			return fmt.Errorf("invalid profile %q for security.profile (supported: cli, web)", value)
		}
		cfg.Security.Profile = value
		return nil
	case "impeccable.enabled":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Impeccable.Enabled = b
		return nil
	case "impeccable.auto_install":
		b, err := parseBool(key, value)
		if err != nil {
			return err
		}
		cfg.Impeccable.AutoInstall = b
		return nil
	case "impeccable.severity":
		if err := config.ValidateImpeccableSeverity(value); err != nil {
			return err
		}
		cfg.Impeccable.Severity = value
		return nil
	case "impeccable.product_path":
		cfg.Impeccable.ProductPath = value
		return nil
	case "impeccable.design_path":
		cfg.Impeccable.DesignPath = value
		return nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			rest := strings.TrimPrefix(key, "models.phases.")
			if phase, ok := strings.CutSuffix(rest, ".base_url"); ok {
				if !config.ValidPhases[phase] {
					return fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, judge, archive)", phase)
				}
				if cfg.Models.Phases == nil {
					cfg.Models.Phases = make(map[string]config.ModelRef)
				}
				ref := cfg.Models.Phases[phase]
				ref.BaseURL = value
				cfg.Models.Phases[phase] = ref
				return nil
			}
			phase := rest
			if !config.ValidPhases[phase] {
				return fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, judge, archive)", phase)
			}
			if cfg.Models.Phases == nil {
				cfg.Models.Phases = make(map[string]config.ModelRef)
			}
			existing := cfg.Models.Phases[phase]
			ref := config.ParseModelRef(value)
			ref.BaseURL = existing.BaseURL
			cfg.Models.Phases[phase] = ref
			return nil
		}
		return fmt.Errorf("unknown config key %q (supported: models.default, models.default.base_url, models.leader, models.leader.base_url, models.phases.<phase>, models.phases.<phase>.base_url, playwright.enabled, playwright.test_dir, playwright.base_url, mutation_testing.enabled, security.enabled, security.profile, impeccable.enabled, impeccable.auto_install, impeccable.severity, impeccable.product_path, impeccable.design_path)", key)
	}
}

func getConfigValue(cfg *config.Config, key string) (string, error) {
	switch key {
	case "models.default":
		return cfg.Models.Default.FullID(), nil
	case "models.default.base_url":
		return cfg.Models.Default.BaseURL, nil
	case "models.leader":
		return cfg.Models.Leader.FullID(), nil
	case "models.leader.base_url":
		return cfg.Models.Leader.BaseURL, nil
	case "playwright.enabled":
		return strconv.FormatBool(cfg.Playwright.Enabled), nil
	case "playwright.test_dir":
		return cfg.Playwright.TestDir, nil
	case "playwright.base_url":
		return cfg.Playwright.BaseURL, nil
	case "mutation_testing.enabled":
		return strconv.FormatBool(cfg.MutationTesting.Enabled), nil
	case "security.enabled":
		return strconv.FormatBool(cfg.Security.Enabled), nil
	case "security.profile":
		return cfg.Security.Profile, nil
	case "impeccable.enabled":
		return strconv.FormatBool(cfg.Impeccable.Enabled), nil
	case "impeccable.auto_install":
		return strconv.FormatBool(cfg.Impeccable.AutoInstall), nil
	case "impeccable.severity":
		return cfg.Impeccable.Severity, nil
	case "impeccable.product_path":
		return cfg.Impeccable.ProductPath, nil
	case "impeccable.design_path":
		return cfg.Impeccable.DesignPath, nil
	default:
		if strings.HasPrefix(key, "models.phases.") {
			rest := strings.TrimPrefix(key, "models.phases.")
			if phase, ok := strings.CutSuffix(rest, ".base_url"); ok {
				if !config.ValidPhases[phase] {
					return "", fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, judge, archive)", phase)
				}
				return cfg.Models.Phases[phase].BaseURL, nil
			}
			phase := rest
			if !config.ValidPhases[phase] {
				return "", fmt.Errorf("unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, judge, archive)", phase)
			}
			return cfg.Models.Phases[phase].FullID(), nil
		}
		return "", fmt.Errorf("unknown config key %q (supported: models.default, models.default.base_url, models.leader, models.leader.base_url, models.phases.<phase>, models.phases.<phase>.base_url, playwright.enabled, playwright.test_dir, playwright.base_url, mutation_testing.enabled, security.enabled, security.profile, impeccable.enabled, impeccable.auto_install, impeccable.severity, impeccable.product_path, impeccable.design_path)", key)
	}
}
