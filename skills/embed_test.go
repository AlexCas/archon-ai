package skills

import (
	"io/fs"
	"testing"
)

func TestFS_ContainsSkills(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(entries) == 0 {
		t.Error("No skills found in embedded FS")
	}

	expectedSkills := []string{
		"sdd-init",
		"sdd-propose",
		"sdd-spec",
		"sdd-design",
		"sdd-tasks",
		"sdd-apply",
		"sdd-verify",
		"sdd-archive",
		"judgment-day",
		"branch-pr",
	}

	foundSkills := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			foundSkills[entry.Name()] = true
		}
	}

	for _, expected := range expectedSkills {
		if !foundSkills[expected] {
			t.Errorf("Expected skill %q not found in embedded FS", expected)
		}
	}
}

func TestFS_SKILLMdAccessible(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(entries) == 0 {
		t.Skip("No skills to test")
	}

	firstSkill := entries[0].Name()
	path := firstSkill + "/SKILL.md"

	data, err := fs.ReadFile(FS, path)
	if err != nil {
		t.Errorf("ReadFile(%s) error = %v", path, err)
	}

	if len(data) == 0 {
		t.Errorf("SKILL.md for %s is empty", firstSkill)
	}
}
