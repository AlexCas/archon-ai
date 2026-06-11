package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RollbackManifest struct {
	Version      string   `json:"version"`
	CreatedPaths []string `json:"paths"`
	BackupPath   string   `json:"original_agents_md_backup,omitempty"`
	HomeDir      string   `json:"-"`
}

func (m *RollbackManifest) manifestPath() string {
	return filepath.Join(m.HomeDir, ".archon", "rollback.json")
}

func (m *RollbackManifest) WriteManifest() error {
	path := m.manifestPath()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp manifest: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename manifest: %w", err)
	}

	return nil
}

func (m *RollbackManifest) Cleanup() error {
	for i := len(m.CreatedPaths) - 1; i >= 0; i-- {
		path := m.CreatedPaths[i]
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}

	if m.BackupPath != "" {
		agentsPath := filepath.Join(m.HomeDir, "AGENTS.md")
		if _, err := os.Stat(m.BackupPath); err == nil {
			if err := os.Remove(agentsPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove current AGENTS.md: %w", err)
			}
			if err := os.Rename(m.BackupPath, agentsPath); err != nil {
				return fmt.Errorf("restore AGENTS.md: %w", err)
			}
		}
	}

	manifestPath := m.manifestPath()
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove manifest: %w", err)
	}

	return nil
}

func LoadManifest(homeDir string) (*RollbackManifest, error) {
	path := filepath.Join(homeDir, ".archon", "rollback.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m RollbackManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}
	m.HomeDir = homeDir

	return &m, nil
}

func (m *RollbackManifest) BackupAgentsMD() error {
	agentsPath := filepath.Join(m.HomeDir, "AGENTS.md")
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return nil
	}

	backupName := fmt.Sprintf("AGENTS.md.backup.%s", time.Now().Format("20060102"))
	backupPath := filepath.Join(m.HomeDir, backupName)

	if err := os.Rename(agentsPath, backupPath); err != nil {
		return fmt.Errorf("backup AGENTS.md: %w", err)
	}

	m.BackupPath = backupPath

	return nil
}
