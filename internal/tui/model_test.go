package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/initcmd"
	"github.com/archon-ai/archon/internal/opencode"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		Agent: "opencode",
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
			Phases: map[string]config.ModelRef{
				"explore": {Model: "claude-sonnet-4"},
			},
		},
		MutationTesting: config.MutationTesting{
			Enabled:   true,
			Threshold: 0.75,
		},
	}
	m := NewModel(cfg, "/tmp/test")

	if m.config != cfg {
		t.Error("config should be set")
	}
	if m.projectDir != "/tmp/test" {
		t.Errorf("projectDir = %q, want %q", m.projectDir, "/tmp/test")
	}
	if m.activeTab != AgentTab {
		t.Errorf("activeTab = %d, want %d", m.activeTab, AgentTab)
	}
	if m.quitting {
		t.Error("quitting should be false")
	}
	if m.modelsTab.rows[0].ref.FullID() != "gpt-4" {
		t.Errorf("default model row ref = %q, want %q", m.modelsTab.rows[0].ref.FullID(), "gpt-4")
	}
	if !m.mutationTab.enabled {
		t.Error("mutation enabled should be true")
	}
	if m.mutationTab.threshold.Value != 75 {
		t.Errorf("mutation threshold = %d, want 75", m.mutationTab.threshold.Value)
	}
	if m.agentTab.selectedAgent != "opencode" {
		t.Errorf("agent = %q, want %q", m.agentTab.selectedAgent, "opencode")
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	newModel, cmd := m.Update(msg)
	model := newModel.(Model)

	if model.width != 80 {
		t.Errorf("width = %d, want 80", model.width)
	}
	if model.height != 24 {
		t.Errorf("height = %d, want 24", model.height)
	}
	if cmd != nil {
		t.Error("Update(WindowSize) should return nil cmd")
	}
}

func TestModel_Update_TabNavigation(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	// Test Tab key
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.activeTab != ModelsTab {
		t.Errorf("activeTab after Tab = %d, want %d", model.activeTab, ModelsTab)
	}

	// Test Shift+Tab key
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.activeTab != AgentTab {
		t.Errorf("activeTab after Shift+Tab = %d, want %d", model.activeTab, AgentTab)
	}
}

// TestModel_Update_ShiftTabWrapsFromAgent verifies that a single Shift+Tab
// from a freshly-constructed model (default AgentTab) wraps around to the
// last tab, SecurityTab.
func TestModel_Update_ShiftTabWrapsFromAgent(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	if m.activeTab != AgentTab {
		t.Fatalf("precondition failed: activeTab = %d, want %d", m.activeTab, AgentTab)
	}

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ := m.Update(msg)
	model := newModel.(Model)

	if model.activeTab != SecurityTab {
		t.Errorf("activeTab after Shift+Tab from AgentTab = %d, want %d", model.activeTab, SecurityTab)
	}
}

func TestModel_Update_Quit(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: false}

	newModel, cmd := m.Update(msg)
	model := newModel.(Model)

	if !model.quitting {
		t.Error("quitting should be true")
	}
	if cmd == nil {
		t.Error("Update(Quit) should return a cmd")
	}
}

func TestModel_Update_Save(t *testing.T) {
	cfg := &config.Config{
		HomeDir: t.TempDir(),
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
		},
	}
	m := NewModel(cfg, cfg.HomeDir)
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, cmd := m.Update(msg)
	model := newModel.(Model)

	if cmd == nil {
		t.Error("Update(Save) should return a cmd")
	}

	// Execute the save command
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("save cmd should return a message")
		}
		if str, ok := result.(string); ok {
			if str != "✓ Configuration saved" {
				t.Errorf("save result = %q, want %q", str, "✓ Configuration saved")
			}
		}
	}

	_ = model
}

func TestModel_View(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	view := m.View()
	if view != "Loading..." {
		t.Errorf("View() without size = %q, want %q", view, "Loading...")
	}

	// With size
	m.width = 80
	m.height = 24
	view = m.View()
	if view == "" {
		t.Error("View() should not be empty")
	}
	if view == "Loading..." {
		t.Error("View() should not be Loading... with size set")
	}
}

func TestModel_View_Quitting(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.quitting = true
	view := m.View()
	if view != "" {
		t.Errorf("View() when quitting = %q, want empty", view)
	}
}

func TestModel_renderTabs(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.width = 80
	rendered := m.renderTabs()
	if rendered == "" {
		t.Error("renderTabs() should not be empty")
	}
}

// TestModel_renderTabs_Order verifies the tab header renders the labels in
// the order Agent, Models, Judge, Mutation Testing, Playwright — i.e. that
// Agent is now the first tab. Robust to styling/ANSI: it only asserts the
// relative order of the substrings within the rendered string.
func TestModel_renderTabs_Order(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.width = 80
	rendered := m.renderTabs()

	labels := []string{"Agent", "Models", "Judge", "Mutation Testing", "Playwright"}
	lastIndex := -1
	for _, label := range labels {
		idx := strings.Index(rendered, label)
		if idx == -1 {
			t.Fatalf("renderTabs() output missing label %q; rendered:\n%s", label, rendered)
		}
		if idx <= lastIndex {
			t.Errorf("label %q at index %d is not after previous label (index %d); order broken; rendered:\n%s", label, idx, lastIndex, rendered)
		}
		lastIndex = idx
	}
}

func TestModel_renderTabContent(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.width = 80
	m.height = 24

	m.activeTab = ModelsTab
	content := m.renderTabContent()
	if content == "" {
		t.Error("renderTabContent() for ModelsTab should not be empty")
	}

	m.activeTab = MutationTab
	content = m.renderTabContent()
	if content == "" {
		t.Error("renderTabContent() for MutationTab should not be empty")
	}

	m.activeTab = AgentTab
	content = m.renderTabContent()
	if content == "" {
		t.Error("renderTabContent() for AgentTab should not be empty")
	}
}

func TestModel_renderHelp(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	help := m.renderHelp()
	if help == "" {
		t.Error("renderHelp() should not be empty")
	}
}

func TestModel_renderStatus(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.statusMsg = "Test message"
	m.statusErr = false
	rendered := m.renderStatus()
	if rendered == "" {
		t.Error("renderStatus() should not be empty")
	}

	m.statusErr = true
	rendered = m.renderStatus()
	if rendered == "" {
		t.Error("renderStatus() with error should not be empty")
	}
}

func TestCheckTerminal(t *testing.T) {
	// In a test environment, stdin is typically not a terminal
	err := CheckTerminal()
	if err == nil {
		// This might pass in some environments, but typically fails in tests
		t.Log("CheckTerminal() passed — stdin is a terminal (unexpected in tests)")
	}
	// We don't assert failure because it depends on the test environment
}

func TestModelsTabState_ApplyToConfig(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
			Phases: map[string]config.ModelRef{
				"explore": config.ParseModelRef("gpt-4o"),
				"propose": config.ParseModelRef("claude-opus-4-8"),
			},
		},
	}
	state := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Set default via picker (direct ref mutation, as if user picked/free-formed)
	state.rows[0].ref = config.ParseModelRef("claude-sonnet-4")
	state.rows[0].changed = true

	// Find and set explore row
	for i, r := range state.rows {
		if r.phase == "explore" {
			state.rows[i].ref = config.ParseModelRef("gpt-4o")
			state.rows[i].changed = true
		}
		if r.phase == "propose" {
			// Clear propose → should be deleted
			state.rows[i].ref = config.ModelRef{}
			state.rows[i].changed = true
		}
	}

	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	state.applyToConfig(outCfg)

	if outCfg.Models.Default.FullID() != "claude-sonnet-4" {
		t.Errorf("default = %q, want %q", outCfg.Models.Default.FullID(), "claude-sonnet-4")
	}
	if outCfg.Models.Phases["explore"].FullID() != "gpt-4o" {
		t.Errorf("explore = %q, want %q", outCfg.Models.Phases["explore"].FullID(), "gpt-4o")
	}
	if _, exists := outCfg.Models.Phases["propose"]; exists {
		t.Error("propose should be deleted when empty")
	}
}

func TestMutationTabState_Toggle(t *testing.T) {
	state := newMutationTabState(config.MutationTesting{Enabled: false})

	if state.enabled {
		t.Error("enabled should be false")
	}

	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !state.enabled {
		t.Error("enabled should be true after toggle")
	}
}

func TestMutationTabState_Slider(t *testing.T) {
	state := newMutationTabState(config.MutationTesting{Enabled: false, Threshold: 0.5})

	state.focused = 1
	state.update(tea.KeyMsg{Type: tea.KeyRight})
	if state.threshold.Value != 51 {
		t.Errorf("threshold = %d, want 51", state.threshold.Value)
	}

	state.update(tea.KeyMsg{Type: tea.KeyLeft})
	if state.threshold.Value != 50 {
		t.Errorf("threshold = %d, want 50", state.threshold.Value)
	}
}

func TestMutationTabState_ApplyToConfig(t *testing.T) {
	cfg := &config.Config{}
	state := newMutationTabState(config.MutationTesting{Enabled: true, Threshold: 0.75})
	state.enabled = false
	state.threshold.SetValue(25)

	state.applyToConfig(cfg)

	if cfg.MutationTesting.Enabled {
		t.Error("enabled should be false")
	}
	if cfg.MutationTesting.Threshold != 0.25 {
		t.Errorf("threshold = %f, want 0.25", cfg.MutationTesting.Threshold)
	}
}

func TestAgentTabState_Navigation(t *testing.T) {
	state := newAgentTabState("opencode")

	state.update(tea.KeyMsg{Type: tea.KeyDown})
	if state.focusedIndex != 0 {
		t.Errorf("focusedIndex = %d, want 0", state.focusedIndex)
	}

	state.update(tea.KeyMsg{Type: tea.KeyDown})
	if state.focusedIndex != 1 {
		t.Errorf("focusedIndex = %d, want 1", state.focusedIndex)
	}

	state.update(tea.KeyMsg{Type: tea.KeyUp})
	if state.focusedIndex != 0 {
		t.Errorf("focusedIndex = %d, want 0", state.focusedIndex)
	}
}

func TestAgentTabState_Select(t *testing.T) {
	state := newAgentTabState("opencode")

	state.focusedIndex = 1 // claude
	state.update(tea.KeyMsg{Type: tea.KeyEnter})

	if state.selectedAgent != "claude" {
		t.Errorf("selectedAgent = %q, want %q", state.selectedAgent, "claude")
	}
	if !state.confirmingInit {
		t.Error("confirmingInit should be true")
	}
}

func TestAgentTabState_ConfirmCancel(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if state.confirmingInit {
		t.Error("confirmingInit should be false after cancel")
	}
}

func TestAgentTabState_SetInitResult(t *testing.T) {
	state := newAgentTabState("opencode")
	state.setInitResult("Success", false)

	if state.initResult != "Success" {
		t.Errorf("initResult = %q, want %q", state.initResult, "Success")
	}
	if state.initError {
		t.Error("initError should be false")
	}

	state.setInitResult("Error", true)
	if !state.initError {
		t.Error("initError should be true")
	}
}

func TestSaveConfig_RegeneratesClaudeMD(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "claude",
		Version:    "1.0.0",
		SkillCount: 10,
	}
	m := NewModel(cfg, tmpDir)

	cmd := m.saveConfig()
	result := cmd()

	if err, ok := result.(error); ok {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	// Verify CLAUDE.md was created with gate sections
	claudeMD := filepath.Join(tmpDir, "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}

	content := string(data)
	requiredSections := []string{
		"SDD Session Preflight",
		"Vague Request Guard",
		"Human Review Gate",
		"Antes de continuar con SDD",
	}
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("CLAUDE.md missing section %q", section)
		}
	}
}

func TestSaveConfig_RegeneratesAgentsMD(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "opencode",
		Version:    "1.0.0",
		SkillCount: 10,
	}
	m := NewModel(cfg, tmpDir)

	cmd := m.saveConfig()
	result := cmd()

	if err, ok := result.(error); ok {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	// Verify AGENTS.md was created with gate sections
	agentsMD := filepath.Join(tmpDir, "AGENTS.md")
	data, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("AGENTS.md not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "SDD Session Preflight") {
		t.Error("AGENTS.md missing SDD Session Preflight section")
	}
	if !strings.Contains(content, "Human Review Gate") {
		t.Error("AGENTS.md missing Human Review Gate section")
	}
}

func TestSaveConfig_DoesNotCreateClaudeDir(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "claude",
		Version:    "1.0.0",
		SkillCount: 10,
	}
	m := NewModel(cfg, tmpDir)

	cmd := m.saveConfig()
	result := cmd()

	if err, ok := result.(error); ok {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	// Verify .claude/ directory was NOT created (regenerateTemplate writes to root)
	claudeDir := filepath.Join(tmpDir, ".claude")
	_, err := os.Stat(claudeDir)
	if err == nil {
		t.Fatal(".claude/ directory should NOT be created by regenerateTemplate")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("unexpected error checking .claude/: %v", err)
	}
}

func TestSaveConfig_FailureDoesNotMutateInMemoryConfig(t *testing.T) {
	// Use a read-only directory so Save() will fail
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "opencode",
		Version:    "0.9.0",
		SkillCount: 5,
	}
	m := NewModel(cfg, tmpDir)
	m.width = 80
	m.height = 24

	// Set some tab values that would be applied
	m.modelsTab.rows[0].ref = config.ParseModelRef("gpt-4")
	m.modelsTab.rows[0].changed = true

	// Make the directory read-only so Save fails
	if err := os.Chmod(tmpDir, 0o555); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	defer os.Chmod(tmpDir, 0o755) // restore for cleanup

	cmd := m.saveConfig()
	result := cmd()

	// Should return error
	if _, ok := result.(error); !ok {
		t.Fatalf("expected error, got: %v", result)
	}

	// Verify in-memory config was NOT mutated
	if m.config.Models.Default.FullID() != "" {
		t.Errorf("in-memory config was mutated: Models.Default = %q", m.config.Models.Default.FullID())
	}
	if m.config.Agent != "opencode" {
		t.Errorf("in-memory config was mutated: Agent = %q", m.config.Agent)
	}
}

func TestSaveConfig_RegenerateFailureKeepsOriginalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "opencode",
		Version:    "0.9.0",
		SkillCount: 5,
	}
	m := NewModel(cfg, tmpDir)
	m.width = 80
	m.height = 24

	// First save succeeds
	cmd := m.saveConfig()
	result := cmd()
	if err, ok := result.(error); ok {
		t.Fatalf("first save failed: %v", err)
	}

	// Verify first save worked
	agentsMD := filepath.Join(tmpDir, "AGENTS.md")
	data, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}
	originalContent := string(data)

	// Make project dir read-only so regenerateTemplate fails
	if err := os.Chmod(tmpDir, 0o555); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	defer os.Chmod(tmpDir, 0o755)

	// Try to save again (should fail at regenerateTemplate)
	cmd = m.saveConfig()
	result = cmd()
	if _, ok := result.(error); !ok {
		t.Fatalf("expected error on regenerate failure, got: %v", result)
	}

	// Verify original AGENTS.md is unchanged
	data, err = os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("AGENTS.md missing: %v", err)
	}
	if string(data) != originalContent {
		t.Error("AGENTS.md was modified when regenerateTemplate failed")
	}

	// Verify in-memory config is unchanged
	if m.config.Version != "0.9.0" {
		t.Errorf("in-memory config was mutated: Version = %q", m.config.Version)
	}
}

func TestSaveConfig_UpdatesExistingTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tmpDir,
		Agent:      "opencode",
		Version:    "0.9.0",
		SkillCount: 5,
	}
	m := NewModel(cfg, tmpDir)

	// First save
	cmd := m.saveConfig()
	if result := cmd(); result != nil {
		if _, ok := result.(error); ok {
			t.Fatalf("first saveConfig() error: %v", result)
		}
	}

	// Update config
	cfg.Version = "1.0.0"
	cfg.SkillCount = 20

	// Second save
	cmd = m.saveConfig()
	if result := cmd(); result != nil {
		if _, ok := result.(error); ok {
			t.Fatalf("second saveConfig() error: %v", result)
		}
	}

	// Verify template was updated with new values
	agentsMD := filepath.Join(tmpDir, "AGENTS.md")
	data, err := os.ReadFile(agentsMD)
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "1.0.0") {
		t.Error("AGENTS.md should contain updated version 1.0.0")
	}
	if !strings.Contains(content, "Skills: 20") {
		t.Error("AGENTS.md should contain updated skill count 20")
	}
}

// S7: the TUI save path must produce the same opencode.json agent.archon-leader
// as a direct initcmd merge with the same leader model, byte-for-byte.
func TestSaveConfig_OpencodeLeaderMatchesInitMerge(t *testing.T) {
	const leader = "anthropic/claude-sonnet-4-20250514"

	// TUI save path: drive saveConfig for an opencode project whose Models.Leader
	// is set, then read the resulting opencode.json.
	tuiDir := t.TempDir()
	cfg := &config.Config{
		HomeDir:    tuiDir,
		Agent:      "opencode",
		Version:    "1.0.0",
		SkillCount: 10,
		Models:     config.ModelConfig{Leader: config.ParseModelRef(leader)},
	}
	m := NewModel(cfg, tuiDir)

	result := m.saveConfig()()
	if err, ok := result.(error); ok {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	tuiBytes, err := os.ReadFile(filepath.Join(tuiDir, "opencode.json"))
	if err != nil {
		t.Fatalf("TUI opencode.json not written: %v", err)
	}

	// Reference path: a direct merge with the same models config into a fresh dir.
	initDir := t.TempDir()
	if _, err := initcmd.MergeOpencodeAgent(initDir, config.ModelConfig{Leader: config.ParseModelRef(leader)}); err != nil {
		t.Fatalf("MergeOpencodeAgent() error = %v", err)
	}
	initBytes, err := os.ReadFile(filepath.Join(initDir, "opencode.json"))
	if err != nil {
		t.Fatalf("init opencode.json not written: %v", err)
	}

	if !bytes.Equal(tuiBytes, initBytes) {
		t.Errorf("TUI save != init merge:\nTUI:\n%s\ninit:\n%s", tuiBytes, initBytes)
	}
}

// TestModelsTab_LeaderWarningGuard verifies the leader row is present for opencode
// agents and absent for non-opencode agents. The per-row Validate advisory is
// dropped (D2); this test asserts leader-row presence/absence, not the ⚠ guard.
func TestModelsTab_LeaderWarningGuard(t *testing.T) {
	// opencode agent: leader row renders.
	cfg := &config.Config{Agent: "opencode", Models: config.ModelConfig{Leader: config.ParseModelRef("openai/gpt-4o")}}
	state := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)
	view := state.view(80, 24)
	if !contains(view, "Leader") {
		t.Fatalf("leader row should render for opencode; view:\n%s", view)
	}
	// No per-row advisory ⚠ in the new picker (D2 drop).
	if contains(view, "⚠") {
		t.Errorf("new picker should not emit per-row ⚠ advisory; view:\n%s", view)
	}

	// Non-opencode agent: no leader row.
	cfg2 := &config.Config{Agent: "claude", Models: config.ModelConfig{Leader: config.ParseModelRef("openai/gpt-4o")}}
	state2 := newModelsTabState(cfg2, map[string]opencode.Provider{}, nil)
	view2 := state2.view(80, 24)
	if contains(view2, "Leader") {
		t.Errorf("leader row should NOT render for non-opencode; view:\n%s", view2)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
