package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Extract(embeddedFS fs.FS, targetDir string) ([]string, error) {
	entries, err := fs.ReadDir(embeddedFS, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded skills: %w", err)
	}

	var extracted []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(skillName, "SKILL.md")

		data, err := fs.ReadFile(embeddedFS, skillPath)
		if err != nil {
			return nil, fmt.Errorf("read skill %s: %w", skillName, err)
		}

		destDir := filepath.Join(targetDir, skillName)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return nil, fmt.Errorf("create skill dir %s: %w", skillName, err)
		}

		destPath := filepath.Join(destDir, "SKILL.md")
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return nil, fmt.Errorf("write skill %s: %w", skillName, err)
		}

		extracted = append(extracted, skillName)
	}

	return extracted, nil
}
