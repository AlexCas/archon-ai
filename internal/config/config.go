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
	Models          ModelConfig      `yaml:"models,omitempty"`
	SkillInventory  []SkillInventory `yaml:"skill_inventory"`
	HomeDir         string           `yaml:"-"`
}

func (c *Config) configPath() string {
	return filepath.Join(c.HomeDir, ".archon", "config.yaml")
}

func (c *Config) Load(fsys fs.FS) error {
	path := filepath.Join(".archon", "config.yaml")
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
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
