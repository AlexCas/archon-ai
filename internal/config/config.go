package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type MutationTesting struct {
	Enabled   bool    `yaml:"enabled"`
	Tool      string  `yaml:"tool,omitempty"`
	Threshold float64 `yaml:"threshold,omitempty"`
}

// Judge controls the judge phase: dual adversarial review via judgment-day plus
// any enabled quality gates (mutation testing, Playwright E2E). When disabled,
// the orchestrator skips the entire judge phase and advances from verify
// straight to archive. Defaults to enabled when the section is absent.
type Judge struct {
	Enabled bool `yaml:"enabled"`
}

// Playwright controls generation and execution of Playwright end-to-end tests
// derived from Gherkin scenarios. When enabled, the harness generates Playwright
// specs from the feature files during the apply phase and runs them after the
// verify and judge phases.
type Playwright struct {
	Enabled bool   `yaml:"enabled"`
	TestDir string `yaml:"test_dir,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`
}

// Security controls the security-baseline opt-in gate. When enabled, the
// harness injects security-review hooks across the propose, spec, tasks,
// verify, and judge phases. Profiles are restricted to "cli" and "web".
// Defaults to disabled (Enabled:false) when the block is absent.
type Security struct {
	Enabled bool   `yaml:"enabled"`
	Profile string `yaml:"profile,omitempty"` // "cli" | "web"
}

// Impeccable controls the opt-in design-language quality gate backed by the
// external npm tool `npx impeccable`. When enabled, the harness references the
// target project's Impeccable design docs during design, runs Impeccable design
// verbs during apply, and executes the `npx impeccable detect` gate after the
// judge phase. Severity governs which finding categories block the judge gate.
// Defaults to disabled (Enabled:false) when the block is absent.
type Impeccable struct {
	Enabled     bool   `yaml:"enabled"`
	AutoInstall bool   `yaml:"auto_install"`
	Severity    string `yaml:"severity,omitempty"`
	ProductPath string `yaml:"product_path,omitempty"`
	DesignPath  string `yaml:"design_path,omitempty"`
}

// ValidImpeccableSeverities is the fixed set of allowed impeccable.severity values.
var ValidImpeccableSeverities = []string{"block-deterministic", "block-all", "advisory"}

// ValidateImpeccableSeverity rejects any impeccable.severity value outside the
// fixed set. Exported so both config.Load() and the CLI `config set` path share
// a single source of truth for the three valid values.
func ValidateImpeccableSeverity(s string) error {
	switch s {
	case "block-deterministic", "block-all", "advisory":
		return nil
	default:
		return fmt.Errorf("invalid impeccable.severity %q (valid: block-deterministic, block-all, advisory)", s)
	}
}

type SkillInventory struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Source  string `yaml:"source"`
}

type Config struct {
	Version         string           `yaml:"harness_version"`
	Agent           string           `yaml:"agent"`
	SkillCount      int              `yaml:"skill_count"`
	CreatedAt       time.Time        `yaml:"created_at"`
	MutationTesting MutationTesting  `yaml:"mutation_testing"`
	Judge           Judge            `yaml:"judge"`
	Playwright      Playwright       `yaml:"playwright"`
	Security        Security         `yaml:"security"`
	Impeccable      Impeccable       `yaml:"impeccable"`
	Models          ModelConfig      `yaml:"models,omitempty"`
	SkillInventory  []SkillInventory `yaml:"skill_inventory"`
	HomeDir         string           `yaml:"-"`
}

func (c *Config) configPath() string {
	return filepath.Join(c.HomeDir, ".archon", "config.yaml")
}

func (c *Config) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, ".archon/config.yaml")
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// The judge phase runs by default; an absent judge section means "enabled".
	// Pre-seed the default so unmarshal only overrides it when the YAML sets it
	// explicitly (e.g. `judge: {enabled: false}`).
	c.Judge.Enabled = true

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	// Impeccable severity defaults to the safe "block-deterministic" mode; an
	// absent or empty value is normalized here so downstream consumers never
	// see "". Normalize BEFORE validate so an absent block is not rejected.
	if c.Impeccable.Severity == "" {
		c.Impeccable.Severity = "block-deterministic"
	}
	if err := ValidateImpeccableSeverity(c.Impeccable.Severity); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

// Clone returns a deep copy of the Config: maps and slices are copied so the
// clone shares no mutable state with the original. It is a hand-rolled,
// field-by-field copy.
//
// IMPORTANT: every new Config field MUST be added here. A field that is added to
// the struct but not copied below is silently dropped on every `archon update`
// (which clones the loaded config before patching it). The round-trip test in
// config_test.go (TestConfig_CloneRoundtrip) fails loudly if a field is missed.
func (c *Config) Clone() *Config {
	clone := &Config{
		Version:         c.Version,
		Agent:           c.Agent,
		SkillCount:      c.SkillCount,
		CreatedAt:       c.CreatedAt,
		HomeDir:         c.HomeDir,
		MutationTesting: c.MutationTesting,
		Judge:           c.Judge,
		Playwright:      c.Playwright,
		Security:        c.Security,
		Impeccable:      c.Impeccable, // value copy — no maps/slices inside
		Models:          ModelConfig{Default: c.Models.Default, Leader: c.Models.Leader, Phases: make(map[string]ModelRef, len(c.Models.Phases))},
		SkillInventory:  make([]SkillInventory, len(c.SkillInventory)),
	}
	for k, v := range c.Models.Phases {
		clone.Models.Phases[k] = v
	}
	copy(clone.SkillInventory, c.SkillInventory)
	return clone
}

func (c *Config) Save() error {
	path := c.configPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename config: %w", err)
	}

	return nil
}
