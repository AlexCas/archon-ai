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

		if _, err := fs.Stat(embeddedFS, skillName+"/SKILL.md"); err != nil {
			continue
		}

		destDir := filepath.Join(targetDir, skillName)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return nil, fmt.Errorf("create skill dir %s: %w", skillName, err)
		}

		files, err := fs.ReadDir(embeddedFS, skillName)
		if err != nil {
			return nil, fmt.Errorf("read skill %s: %w", skillName, err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			data, err := fs.ReadFile(embeddedFS, skillName+"/"+file.Name())
			if err != nil {
				return nil, fmt.Errorf("read %s/%s: %w", skillName, file.Name(), err)
			}
			if err := os.WriteFile(filepath.Join(destDir, file.Name()), data, 0o644); err != nil {
				return nil, fmt.Errorf("write %s/%s: %w", skillName, file.Name(), err)
			}
		}

		extracted = append(extracted, skillName)
	}

	return extracted, nil
}
