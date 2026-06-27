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
	// Rules shared across both harnesses (rules 1, 3-8).
	sharedRules := []string{
		"1. Check harness-workflow before any phase transition",
		"3. Write/update SESSION_STATUS.md at the root on every phase transition",
		"4. After every phase that produces an editable artifact, run the Human Review Gate",
		"5. After verify, invoke harness-judge",
		"6. When playwright.enabled, run the generated Playwright tests after verify and judge pass",
		"7. On judge fail: re-apply with feedback (max 3 retries)",
		"8. Commits carry ONLY the user's authorship — no Co-Authored-By or tool attribution",
	}

	tests := []struct {
		name      string
		render    func(TemplateData) (string, error)
		rule2Want string
	}{
		{
			name:      "AGENTS.md",
			render:    RenderAgentsMD,
			rule2Want: "2. Delegate each phase to the `archon-<phase>` subagent",
		},
		{
			name:      "CLAUDE.md",
			render:    RenderClaudeMD,
			rule2Want: "2. Delegate each phase to the `archon-<phase>` subagent; do not pass a per-call model parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				Agent:          "opencode",
				HarnessVersion: "1.0.0",
				SkillCount:     10,
			}
			content, err := tt.render(data)
			if err != nil {
				t.Fatalf("render error = %v", err)
			}

			// Check shared rules are present in both harness docs.
			for _, rule := range sharedRules {
				if !strings.Contains(content, rule) {
					t.Errorf("%s missing rule %q", tt.name, rule)
				}
			}

			// Check the per-harness Rule 2 wording.
			if !strings.Contains(content, tt.rule2Want) {
				t.Errorf("%s missing rule 2 %q", tt.name, tt.rule2Want)
			}

			// Ensure there is no rule 9 (exactly 8 rules).
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

// TestTemplates_AgentsAndClaudeSharedSections verifies that both harness docs
// share the common orchestratorSections content while allowing per-harness
// differences in the Rules and Phase Models blocks.
func TestTemplates_AgentsAndClaudeSharedSections(t *testing.T) {
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

	// Both docs must share the SDD Session Preflight, Vague Request Guard, and
	// Human Review Gate sections from orchestratorSections.
	sharedSections := []string{
		"## SDD Session Preflight (HARD GATE)",
		"## Vague Request Guard (MANDATORY)",
		"## Human Review Gate (MANDATORY)",
		"## Leader Persona",
		"## Session Status (SESSION_STATUS.md) — MANDATORY",
		"## Commit Attribution (HARD RULE)",
	}
	for _, section := range sharedSections {
		if !strings.Contains(agents, section) {
			t.Errorf("AGENTS.md missing shared section %q", section)
		}
		if !strings.Contains(claude, section) {
			t.Errorf("CLAUDE.md missing shared section %q", section)
		}
	}

	// Templates must differ in their Rule 2 wording (by design).
	if agents == claude {
		t.Error("agentsTemplate and claudeTemplate are identical — per-harness Rule 2 divergence is missing")
	}
}

func TestTemplates_PhaseModelsBlock(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels: config.ResolvePhaseModels(config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"explore": {Model: "opus"},
				"propose": {Model: "sonnet"},
				"design":  {Model: "opus"},
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

func TestTemplates_PhaseModelsNonClaudeDefault(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels:    config.ResolvePhaseModels(config.ModelConfig{Default: config.ModelRef{Model: "gemini-2.5-pro"}}),
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	if !strings.Contains(content, "## Phase Models") {
		t.Error("rendered content missing ## Phase Models block for non-Claude default")
	}
	if !strings.Contains(content, "gemini-2.5-pro") {
		t.Error("rendered content missing catalog id \"gemini-2.5-pro\"")
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
	// Non-Claude default mixed with per-phase Claude and OpenAI values, so the
	// across-paths byte-identity check exercises a non-Claude default too
	// (spec scenario: "Non-Claude default renders an identical block across paths").
	mc := config.ModelConfig{
		Default: config.ModelRef{Model: "gemini-2.5-pro"},
		Phases: map[string]config.ModelRef{
			"explore": {Model: "opus"},
			"tasks":   {Model: "gpt-4o"},
			"verify":  {Model: "haiku"},
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

// 5.8a: CLAUDE.md names subagents as the hard gate (not advisory).
func TestTemplates_ClaudePhaseModelsIsHardGate(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels: config.ResolvePhaseModels(config.ModelConfig{
			Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		}),
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	if !strings.Contains(content, "## Phase Models") {
		t.Error("CLAUDE.md missing ## Phase Models block")
	}

	// Must reference archon-<phase> subagents as the binding.
	if !strings.Contains(content, "archon-<phase>") {
		t.Error("CLAUDE.md Phase Models block does not name archon-<phase> subagents")
	}

	// Must not call model selection "advisory".
	if strings.Contains(content, "advisory") {
		t.Error("CLAUDE.md Phase Models block must not use the word \"advisory\"")
	}

	// Must not leak platform precedence noise into the generated doc.
	if strings.Contains(content, "CLAUDE_CODE_SUBAGENT_MODEL") {
		t.Error("CLAUDE.md must not mention CLAUDE_CODE_SUBAGENT_MODEL (implementation noise)")
	}
}

// 5.8b: AGENTS.md Phase Models block states the binding lives in opencode.json.
func TestTemplates_AgentsPhaseModelsPointsAtOpencodeJSON(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
		PhaseModels: config.ResolvePhaseModels(config.ModelConfig{
			Default: config.ParseModelRef("anthropic/claude-sonnet-4-6"),
		}),
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	if !strings.Contains(content, "## Phase Models") {
		t.Error("AGENTS.md missing ## Phase Models block")
	}

	if !strings.Contains(content, "opencode.json") {
		t.Error("AGENTS.md Phase Models block must state that the binding lives in opencode.json")
	}
}

// 5.8c: CLAUDE.md routes delegation to the named subagent and instructs no
// per-call model parameter.
func TestTemplates_ClaudeDelegationRuleNamesSubagentAndNoPerCallModel(t *testing.T) {
	data := TemplateData{
		Agent:          "claude",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	content, err := RenderClaudeMD(data)
	if err != nil {
		t.Fatalf("RenderClaudeMD() error = %v", err)
	}

	// Rule 2 must name archon-<phase> as the delegation target.
	if !strings.Contains(content, "archon-<phase>") {
		t.Error("CLAUDE.md delegation rule does not name archon-<phase> as the delegation target")
	}

	// Rule 2 must instruct the leader not to pass a per-call model parameter.
	if !strings.Contains(content, "per-call model parameter") {
		t.Error("CLAUDE.md delegation rule must instruct the leader not to pass a per-call model parameter")
	}
}

// 5.8d: AGENTS.md routes delegation to the named subagent.
func TestTemplates_AgentsDelegationRuleNamesSubagent(t *testing.T) {
	data := TemplateData{
		Agent:          "opencode",
		HarnessVersion: "1.0.0",
		SkillCount:     10,
	}

	content, err := RenderAgentsMD(data)
	if err != nil {
		t.Fatalf("RenderAgentsMD() error = %v", err)
	}

	// Rule 2 must name archon-<phase> as the delegation target.
	if !strings.Contains(content, "archon-<phase>") {
		t.Error("AGENTS.md delegation rule does not name archon-<phase> as the delegation target")
	}
}
