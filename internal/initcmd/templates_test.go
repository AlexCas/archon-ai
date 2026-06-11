package initcmd

import (
	"strings"
	"testing"
)

func TestRenderAgentsMD(t *testing.T) {
	data := TemplateData{
		ProjectName:    "test-project",
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     23,
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	checks := []string{
		"ARCHON AI Orchestrator",
		"explore → propose → spec → design → tasks → apply → verify → judge → archive",
		"Agent: opencode",
		"Harness Version: 1.0.0",
		"Skills: 23",
		".archon/config.yaml",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("RenderAgentsMD() missing %q", check)
		}
	}
}

func TestRenderClaudeMD(t *testing.T) {
	data := TemplateData{
		ProjectName:    "test-project",
		Agent:          "claude",
		HarnessVersion: "2.0.0",
		SkillCount:     15,
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	checks := []string{
		"ARCHON AI Orchestrator",
		"Agent: claude",
		"Harness Version: 2.0.0",
		"Skills: 15",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("RenderClaudeMD() missing %q", check)
		}
	}
}

func TestRenderAgentsMD_EmptyData(t *testing.T) {
	data := TemplateData{}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	if content == "" {
		t.Error("RenderAgentsMD() returned empty content")
	}

	if !strings.Contains(content, "ARCHON AI Orchestrator") {
		t.Error("RenderAgentsMD() missing title")
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	data := TemplateData{}

	_, err := renderTemplate("test", "{{.Invalid", data)
	if err == nil {
		t.Error("renderTemplate() should fail with invalid template")
	}
}

func TestRenderAgentsMD_AllFieldsPopulated(t *testing.T) {
	data := TemplateData{
		ProjectName:    "my-project",
		Agent:          "codex",
		HarnessVersion: "3.1.4",
		SkillCount:     42,
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	if !strings.Contains(content, "codex") {
		t.Error("Agent name not rendered")
	}
	if !strings.Contains(content, "3.1.4") {
		t.Error("Harness version not rendered")
	}
	if !strings.Contains(content, "42") {
		t.Error("Skill count not rendered")
	}
}
