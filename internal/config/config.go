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

// Playwright controls generation and execution of Playwright end-to-end tests
// derived from Gherkin scenarios. When enabled, the harness generates Playwright
// specs from the feature files during the apply phase and runs them after the
// verify and judge phases.
type Playwright struct {
	Enabled bool   `yaml:"enabled"`
	TestDir string `yaml:"test_dir,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`
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
	Playwright      Playwright       `yaml:"playwright"`
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

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		Version:         c.Version,
		Agent:           c.Agent,
		SkillCount:      c.SkillCount,
		CreatedAt:       c.CreatedAt,
		HomeDir:         c.HomeDir,
		MutationTesting: c.MutationTesting,
		Playwright:      c.Playwright,
		Models:          ModelConfig{Default: c.Models.Default, Phases: make(map[string]string, len(c.Models.Phases))},
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
