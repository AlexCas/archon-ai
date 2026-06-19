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
				"¿Querés ajustar algo en esta fase antes de continuar?",
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

			// Check all 5 rules are present
			rules := []string{
				"1. Check harness-workflow before any phase transition",
				"2. Delegate each phase to sdd-* sub-agent",
				"3. After every phase that produces an editable artifact, run the Human Review Gate",
				"4. After verify, invoke harness-judge",
				"5. On judge fail: re-apply with feedback (max 3 retries)",
			}

			for _, rule := range rules {
				if !strings.Contains(content, rule) {
					t.Errorf("%s missing rule %q", tt.name, rule)
				}
			}

			// Ensure rule 6 does NOT exist (exactly 5 rules)
			if strings.Contains(content, "6. ") {
				t.Errorf("%s should have exactly 5 rules, found rule 6", tt.name)
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

// TestTemplates_AgentsAndClaudeShareCommonCore verifies that both templates
// produce the same orchestrator core sections (persona, preflight, vague guard,
// human review gate) even though their Phase Models sections intentionally differ.
func TestTemplates_AgentsAndClaudeShareCommonCore(t *testing.T) {
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

	// The two templates intentionally differ in their Phase Models section:
	// agentsTemplate describes the opencode.json mechanism; claudeTemplate
	// describes the advisory model: parameter mechanism for the claude agent.
	// Verify they differ (a strict equality check would hide future divergence
	// in the wrong direction).
	if agents == claude {
		t.Error("agentsTemplate and claudeTemplate are identical — they should differ in the Phase Models section")
	}

	// Both must share the common orchestrator core sections.
	commonSections := []string{
		"## Leader Persona",
		"## Phase Order",
		"## SDD Session Preflight (HARD GATE)",
		"## Vague Request Guard (MANDATORY)",
		"## Human Review Gate (MANDATORY)",
		"## Phase Models",
	}
	for _, section := range commonSections {
		if !strings.Contains(agents, section) {
			t.Errorf("AGENTS.md missing common section %q", section)
		}
		if !strings.Contains(claude, section) {
			t.Errorf("CLAUDE.md missing common section %q", section)
		}
	}
}

func TestTemplates_PhaseModelsSection(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	// AGENTS.md (opencode): must have Phase Models and must reference opencode.json
	// because per-phase models are wired via the opencode.json agent definitions.
	t.Run("AGENTS.md", func(t *testing.T) {
		content, err := RenderAgentsMD(data)
		if err != nil {
			t.Fatalf("render error = %v", err)
		}
		if !strings.Contains(content, "## Phase Models") {
			t.Error("AGENTS.md missing ## Phase Models section")
		}
		if !strings.Contains(content, "opencode.json") {
			t.Error("AGENTS.md Phase Models section must mention opencode.json (opencode wiring mechanism)")
		}
	})

	// CLAUDE.md (claude agent): must have Phase Models but must NOT reference
	// opencode.json — claude uses the advisory §model: <id>§ delegation parameter,
	// not opencode.json agent definitions.
	t.Run("CLAUDE.md", func(t *testing.T) {
		content, err := RenderClaudeMD(data)
		if err != nil {
			t.Fatalf("render error = %v", err)
		}
		if !strings.Contains(content, "## Phase Models") {
			t.Error("CLAUDE.md missing ## Phase Models section")
		}
		if strings.Contains(content, "opencode.json") {
			t.Error("CLAUDE.md Phase Models section must NOT mention opencode.json — that is the opencode mechanism, not the claude mechanism")
		}
		// Must describe the advisory delegation mechanism instead.
		if !strings.Contains(content, "model:") {
			t.Error("CLAUDE.md Phase Models section must describe the model: delegation parameter")
		}
	})
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
