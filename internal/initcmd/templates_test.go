package initcmd

import (
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
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

func TestTemplates_ContainSDDSessionPreflight(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	tests := []struct {
		name   string
		render func(TemplateData) (string, error)
	}{
		{"AGENTS.md", RenderAgentsMD},
		{"CLAUDE.md", RenderClaudeMD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tt.render(data)
			if err != nil {
				t.Fatalf("render error = %v", err)
			}

			required := []string{
				"## SDD Session Preflight (HARD GATE)",
				"## Vague Request Guard (MANDATORY)",
				"## Human Review Gate (MANDATORY)",
				"Antes de continuar con SDD",
				"¿Quieres ajustar algo en esta fase antes de continuar?",
			}

			for _, req := range required {
				if !strings.Contains(content, req) {
					t.Errorf("%s missing %q", tt.name, req)
				}
			}
		})
	}
}

func TestTemplates_FiveRules(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	tests := []struct {
		name   string
		render func(TemplateData) (string, error)
	}{
		{"AGENTS.md", RenderAgentsMD},
		{"CLAUDE.md", RenderClaudeMD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tt.render(data)
			if err != nil {
				t.Fatalf("render error = %v", err)
			}

			// Check all orchestrator rules are present
			rules := []string{
				"1. Check harness-workflow before any phase transition",
				"2. Delegate each phase to sdd-* sub-agent",
				"3. Write/update SESSION_STATUS.md at the root on every phase transition",
				"4. After every phase that produces an editable artifact, run the Human Review Gate",
				"5. After verify, invoke harness-judge",
				"6. When playwright.enabled, run the generated Playwright tests after verify and judge pass",
				"7. On judge fail: re-apply with feedback (max 3 retries)",
				"8. Commits carry ONLY the user's authorship — no Co-Authored-By or tool attribution",
			}

			for _, rule := range rules {
				if !strings.Contains(content, rule) {
					t.Errorf("%s missing rule %q", tt.name, rule)
				}
			}

			// Ensure there is no rule 9 (exactly 8 rules)
			if strings.Contains(content, "9. ") {
				t.Errorf("%s should have exactly 8 rules, found rule 9", tt.name)
			}
		})
	}
}

func TestTemplates_BacktickRendering(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	// Verify backtick placeholder was replaced with actual backticks
	backtickChecks := []string{
		"`interactive`",
		"`auto`",
		"`openspec`",
		"`engram`",
		"`sdd-explore`",
		"`sdd-propose`",
		"`internal/billing`",
	}

	for _, check := range backtickChecks {
		if !strings.Contains(content, check) {
			t.Errorf("RenderAgentsMD() missing backtick-wrapped text %q", check)
		}
	}

	// Verify no § placeholder remains
	if strings.Contains(content, "§") {
		t.Error("RenderAgentsMD() still contains § placeholder — backtick replacement failed")
	}
}

func TestTemplates_CodeBlockRendering(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	// Verify the Spanish prompt code block is properly rendered with triple backticks
	if !strings.Contains(content, "```text") {
		t.Error("RenderAgentsMD() missing ```text code block opening")
	}
}

func TestTemplates_AgentsAndClaudeIdentical(t *testing.T) {
	data := TemplateData{
		Agent:          "test-agent",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	agents, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	claude, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	if agents != claude {
		t.Error("agentsTemplate and claudeTemplate are not identical — they may have diverged silently")
	}
}

func TestTemplates_PhaseModelsBlock(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels: config.ResolvePhaseModels(config.ModelConfig{
			Phases: map[string]string{
				"explore": "Opus 4.8",
				"propose": "sonnet",
				"design":  "claude-opus-4-8",
			},
		}),
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	mustContain := []string{
		"## Phase Models",
		"- explore: opus",
		"- propose: sonnet",
		"- design: opus",
	}
	for _, want := range mustContain {
		if !strings.Contains(content, want) {
			t.Errorf("rendered content missing %q", want)
		}
	}

	// No raw display strings and no unresolved placeholder leak through.
	if strings.Contains(content, "Opus 4.8") {
		t.Error("rendered content contains raw display string \"Opus 4.8\"")
	}
	if strings.Contains(content, backtickPlaceholder) {
		t.Error("rendered content still contains § placeholder")
	}
}

func TestTemplates_PhaseModelsOmittedWhenEmpty(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels:    nil,
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	if strings.Contains(content, "## Phase Models") {
		t.Error("rendered content should omit ## Phase Models header when PhaseModels is nil")
	}
}

func TestTemplates_PhaseModelsBlockMatchesAcrossPaths(t *testing.T) {
	mc := config.ModelConfig{
		Default: "sonnet",
		Phases: map[string]string{
			"explore": "Opus 4.8",
			"verify":  "haiku",
		},
	}

	// Mirror init's writeTemplate data construction.
	initData := TemplateData{
		ProjectName:    "proj",
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels:    config.ResolvePhaseModels(mc),
	}
	// Mirror the TUI regenerateTemplate data construction for the same config.
	tuiData := TemplateData{
		ProjectName:    "proj",
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels:    config.ResolvePhaseModels(mc),
	}

	initContent, err := RenderClaudeMD(initData)
	if err != nil {
		t.Fatalf("init render error = %v", err)
	}
	tuiContent, err := RenderClaudeMD(tuiData)
	if err != nil {
		t.Fatalf("tui render error = %v", err)
	}

	initBlock := phaseModelsBlock(t, initContent)
	tuiBlock := phaseModelsBlock(t, tuiContent)
	if initBlock != tuiBlock {
		t.Errorf("phase model blocks differ across paths:\ninit:\n%s\ntui:\n%s", initBlock, tuiBlock)
	}
}

// phaseModelsBlock extracts the "## Phase Models" section up to the next "##"
// header so the two render paths can be compared byte-for-byte.
func phaseModelsBlock(t *testing.T, content string) string {
	t.Helper()
	start := strings.Index(content, "## Phase Models")
	if start == -1 {
		t.Fatal("content missing ## Phase Models block")
	}
	rest := content[start+len("## Phase Models"):]
	if end := strings.Index(rest, "\n## "); end != -1 {
		return content[start : start+len("## Phase Models")+end]
	}
	return content[start:]
}

func TestTemplates_LeaderPersona(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	tests := []struct {
		name   string
		render func(TemplateData) (string, error)
	}{
		{"AGENTS.md", RenderAgentsMD},
		{"CLAUDE.md", RenderClaudeMD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := tt.render(data)
			if err != nil {
				t.Fatalf("render error = %v", err)
			}

			// Verify persona section header is present
			if !strings.Contains(content, "## Leader Persona") {
				t.Errorf("%s missing ## Leader Persona section", tt.name)
			}

			// Verify ordering: persona must come before Phase Order
			personaIdx := strings.Index(content, "## Leader Persona")
			phaseOrderIdx := strings.Index(content, "## Phase Order")
			if personaIdx == -1 || phaseOrderIdx == -1 {
				t.Errorf("%s missing persona or phase order section", tt.name)
			} else if personaIdx >= phaseOrderIdx {
				t.Errorf("%s persona section should come before Phase Order", tt.name)
			}

			// Verify all 4 domains are covered
			requiredDomains := []string{
				"**Scope**:",
				"**Language**:",
				"**Tone**:",
				"**Behavior**:",
			}
			for _, domain := range requiredDomains {
				if !strings.Contains(content, domain) {
					t.Errorf("%s missing persona domain %q", tt.name, domain)
				}
			}

			// Verify key rules are present
			keyRules := []string{
				"ALWAYS reply in the user's current language",
				"neutral/professional Spanish",
				"Do NOT use voseo",
				"Warm and direct",
				"avoid CAPS",
				"Seek clarification",
				"Never say",
			}
			for _, rule := range keyRules {
				if !strings.Contains(content, rule) {
					t.Errorf("%s missing key rule %q", tt.name, rule)
				}
			}
		})
	}
}
