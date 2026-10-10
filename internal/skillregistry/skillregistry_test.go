package skillregistry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeSkill creates <root>/<name>/SKILL.md with the given frontmatter name
// and description, so tests can build fake project/user skill trees.
func writeSkill(t *testing.T, root, dirName, name, description string) {
	t.Helper()
	dir := filepath.Join(root, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	content := "---\nname: " + name + "\ndescription: \"" + description + "\"\n---\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

func TestRefresh_ProjectSkillWinsOverUser(t *testing.T) {
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	writeSkill(t, projectDir+"/skills", "foo", "foo", "project version")
	writeSkill(t, homeDir+"/.claude/skills", "foo", "foo", "user version")

	res, err := Refresh(projectDir, homeDir)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if len(res.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(res.Entries))
	}
	got := res.Entries[0]
	if got.Scope != "project" || got.Description != "project version" {
		t.Errorf("got %+v, want scope=project description=%q", got, "project version")
	}
}

func TestRefresh_FirstUserRootWinsAmongDuplicates(t *testing.T) {
	projectDir := t.TempDir()
	homeDir := t.TempDir()

	// .pi/agent/skills is scanned before .config/opencode/skills.
	writeSkill(t, homeDir+"/.pi/agent/skills", "bar", "bar", "pi version")
	writeSkill(t, homeDir+"/.config/opencode/skills", "bar", "bar", "opencode version")

	res, err := Refresh(projectDir, homeDir)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if len(res.Entries) != 1 || res.Entries[0].Description != "pi version" {
		t.Fatalf("got %+v, want single entry with description %q", res.Entries, "pi version")
	}
}

func TestRefresh_SkipsSddSharedAndSelf(t *testing.T) {
	projectDir := t.TempDir()

	writeSkill(t, projectDir+"/skills", "sdd-apply", "sdd-apply", "should be skipped")
	writeSkill(t, projectDir+"/skills", "_shared", "_shared", "should be skipped")
	writeSkill(t, projectDir+"/skills", "skill-registry", "skill-registry", "should be skipped")
	writeSkill(t, projectDir+"/skills", "harness-judge", "harness-judge", "should be kept")

	res, err := Refresh(projectDir, "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if len(res.Entries) != 1 || res.Entries[0].Name != "harness-judge" {
		t.Fatalf("got %+v, want only harness-judge", res.Entries)
	}
}

func TestRefresh_EmptyWhenNothingFound(t *testing.T) {
	projectDir := t.TempDir()

	res, err := Refresh(projectDir, "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if len(res.Entries) != 0 {
		t.Errorf("Entries = %v, want empty", res.Entries)
	}
}

func TestRefresh_EmptyHomeDirSkipsUserScan(t *testing.T) {
	projectDir := t.TempDir()
	writeSkill(t, projectDir+"/skills", "foo", "foo", "project only")

	res, err := Refresh(projectDir, "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if len(res.Sources) != 0 {
		t.Errorf("Sources = %v, want empty when homeDir is \"\"", res.Sources)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("Entries = %v, want 1", res.Entries)
	}
}

func TestRender_ContainsExpectedSections(t *testing.T) {
	res := &Result{
		Sources: []string{"/home/u/.claude/skills"},
		Entries: []Entry{{Name: "foo", Description: "does foo things", Scope: "project", Path: "skills/foo/SKILL.md"}},
	}
	out := Render("my-project", res, time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC))

	for _, want := range []string{
		"# Skill Registry — my-project",
		"archon skill-registry refresh --force",
		"Last updated: 2026-10-10",
		"## Sources scanned",
		"/home/u/.claude/skills",
		"## Contract",
		"## Skills",
		"| `foo` | does foo things | project | `skills/foo/SKILL.md` |",
		"## Loading protocol",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestWrite_CacheHitSkipsRewrite(t *testing.T) {
	projectDir := t.TempDir()
	writeSkill(t, projectDir+"/skills", "foo", "foo", "v1")

	res, err := Refresh(projectDir, "")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	path1, hit1, err := Write(projectDir, res, false, time.Now())
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if hit1 {
		t.Fatalf("first Write() reported cache hit, want a real write")
	}

	data1, err := os.ReadFile(path1)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}

	// Second refresh of the same unchanged skill set, one second later.
	time.Sleep(10 * time.Millisecond)
	res2, _ := Refresh(projectDir, "")
	path2, hit2, err := Write(projectDir, res2, false, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !hit2 {
		t.Fatalf("second Write() with unchanged skill set should be a cache hit")
	}

	data2, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if string(data1) != string(data2) {
		t.Errorf("registry file changed on a cache hit")
	}
}

func TestWrite_ForceAlwaysRewrites(t *testing.T) {
	projectDir := t.TempDir()
	writeSkill(t, projectDir+"/skills", "foo", "foo", "v1")

	res, _ := Refresh(projectDir, "")
	if _, hit, err := Write(projectDir, res, false, time.Now()); err != nil || hit {
		t.Fatalf("first Write() error=%v hit=%v, want a real write", err, hit)
	}

	if _, hit, err := Write(projectDir, res, true, time.Now().Add(time.Hour)); err != nil || hit {
		t.Fatalf("forced Write() error=%v hit=%v, want force to bypass the cache", err, hit)
	}
}

func TestWrite_RewritesWhenSkillSetChanges(t *testing.T) {
	projectDir := t.TempDir()
	writeSkill(t, projectDir+"/skills", "foo", "foo", "v1")

	res, _ := Refresh(projectDir, "")
	if _, _, err := Write(projectDir, res, false, time.Now()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	writeSkill(t, projectDir+"/skills", "bar", "bar", "new skill")
	res2, _ := Refresh(projectDir, "")
	_, hit, err := Write(projectDir, res2, false, time.Now())
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if hit {
		t.Errorf("adding a skill should invalidate the cache, got a cache hit")
	}
}
