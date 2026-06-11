package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/archon-ai/archon/internal/config"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		Agent: "opencode",
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases: map[string]string{
				"explore": "claude-sonnet-4",
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
	if m.activeTab != ModelsTab {
		t.Errorf("activeTab = %d, want %d", m.activeTab, ModelsTab)
	}
	if m.quitting {
		t.Error("quitting should be false")
	}
	if m.modelsTab.inputs[modelInputDefault].Value() != "gpt-4" {
		t.Errorf("default model input = %q, want %q", m.modelsTab.inputs[modelInputDefault].Value(), "gpt-4")
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

	if model.activeTab != MutationTab {
		t.Errorf("activeTab after Tab = %d, want %d", model.activeTab, MutationTab)
	}

	// Test Shift+Tab key
	msg = tea.KeyMsg{Type: tea.KeyShiftTab}
	newModel, _ = model.Update(msg)
	model = newModel.(Model)

	if model.activeTab != ModelsTab {
		t.Errorf("activeTab after Shift+Tab = %d, want %d", model.activeTab, ModelsTab)
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
			Default: "gpt-4",
		},
	}
	m := NewModel(cfg, cfg.HomeDir)
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true}
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

func TestModelsTabState_AutoFill(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
		},
	}
	state := newModelsTabState(cfg)

	// Check that placeholder includes default
	placeholder := state.inputs[modelInputExplore].Placeholder
	if !contains(placeholder, "gpt-4") {
		t.Errorf("placeholder = %q, should contain default model", placeholder)
	}

	// Update default model
	state.inputs[modelInputDefault].SetValue("claude-sonnet-4")
	state.updateAutoFill()

	placeholder = state.inputs[modelInputExplore].Placeholder
	if !contains(placeholder, "claude-sonnet-4") {
		t.Errorf("updated placeholder = %q, should contain new default model", placeholder)
	}
}

func TestModelsTabState_LockOnEdit(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
		},
	}
	state := newModelsTabState(cfg)

	// Simulate user typing in explore field
	state.focusedInput = modelInputExplore
	state.inputs[modelInputExplore].SetValue("custom-model")
	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Should be locked
	if !state.autoFillLocks[modelInputExplore] {
		t.Error("autoFillLocks should be set after editing")
	}

	// Change default model
	state.inputs[modelInputDefault].SetValue("claude-sonnet-4")
	state.updateAutoFill()

	// Explore should keep its value
	if state.inputs[modelInputExplore].Value() != "custom-model" {
		t.Errorf("explore value = %q, want %q", state.inputs[modelInputExplore].Value(), "custom-model")
	}
}

func TestModelsTabState_ApplyToConfig(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases:  make(map[string]string),
		},
	}
	state := newModelsTabState(cfg)

	state.inputs[modelInputDefault].SetValue("claude-sonnet-4")
	state.inputs[modelInputExplore].SetValue("gpt-4o")
	state.inputs[modelInputPropose].SetValue("")

	state.applyToConfig(cfg)

	if cfg.Models.Default != "claude-sonnet-4" {
		t.Errorf("default = %q, want %q", cfg.Models.Default, "claude-sonnet-4")
	}
	if cfg.Models.Phases["explore"] != "gpt-4o" {
		t.Errorf("explore = %q, want %q", cfg.Models.Phases["explore"], "gpt-4o")
	}
	if _, exists := cfg.Models.Phases["propose"]; exists {
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
	m.modelsTab.inputs[modelInputDefault].SetValue("gpt-4")

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
	if m.config.Models.Default != "" {
		t.Errorf("in-memory config was mutated: Models.Default = %q", m.config.Models.Default)
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

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
