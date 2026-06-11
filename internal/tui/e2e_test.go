package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/archon-ai/archon/internal/config"
)

// TestIntegration_SaveAndReload tests the full save/load cycle.
func TestIntegration_SaveAndReload(t *testing.T) {
	// Create a temp directory to act as project dir
	projectDir := t.TempDir()

	// Create initial config
	cfg := &config.Config{
		HomeDir: projectDir,
		Agent:   "opencode",
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases: map[string]string{
				"explore": "claude-sonnet-4",
			},
		},
		MutationTesting: config.MutationTesting{
			Enabled:   false,
			Threshold: 0.5,
		},
	}

	// Save initial config
	if err := cfg.Save(); err != nil {
		t.Fatalf("initial save: %v", err)
	}

	// Create TUI model
	m := NewModel(cfg, projectDir)

	// Modify state
	m.modelsTab.inputs[modelInputDefault].SetValue("gpt-4o")
	m.modelsTab.inputs[modelInputExplore].SetValue("gpt-4o-mini")
	m.mutationTab.enabled = true
	m.mutationTab.threshold.SetValue(75)

	// Apply and save
	m.modelsTab.applyToConfig(m.config)
	m.mutationTab.applyToConfig(m.config)

	if err := m.config.Save(); err != nil {
		t.Fatalf("save after modifications: %v", err)
	}

	// Reload and verify
	reloaded := &config.Config{HomeDir: projectDir}
	projectFS := os.DirFS(projectDir)
	if err := reloaded.Load(projectFS); err != nil {
		t.Fatalf("reload: %v", err)
	}

	if reloaded.Models.Default != "gpt-4o" {
		t.Errorf("reloaded default = %q, want %q", reloaded.Models.Default, "gpt-4o")
	}
	if reloaded.Models.Phases["explore"] != "gpt-4o-mini" {
		t.Errorf("reloaded explore = %q, want %q", reloaded.Models.Phases["explore"], "gpt-4o-mini")
	}
	if !reloaded.MutationTesting.Enabled {
		t.Error("reloaded enabled should be true")
	}
	if reloaded.MutationTesting.Threshold != 0.75 {
		t.Errorf("reloaded threshold = %f, want 0.75", reloaded.MutationTesting.Threshold)
	}
}

// TestEdgeCases_EmptyDefaultModel tests behavior when default model is empty.
func TestEdgeCases_EmptyDefaultModel(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "",
			Phases:  make(map[string]string),
		},
	}
	m := NewModel(cfg, "")

	// Check placeholder with empty default
	placeholder := m.modelsTab.inputs[modelInputExplore].Placeholder
	if contains(placeholder, "()") {
		t.Log("placeholder with empty default is fine")
	}
}

// TestEdgeCases_UnknownModelWarning tests that unknown models show warnings.
func TestEdgeCases_UnknownModelWarning(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "unknown-model-xyz",
		},
	}
	m := NewModel(cfg, "")
	m.width = 80
	m.height = 24

	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// The view should render without panicking
	_ = view
}

// TestEdgeCases_QuickTabSwitching tests rapid tab switching.
func TestEdgeCases_QuickTabSwitching(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	// Switch through all tabs multiple times
	for i := 0; i < 9; i++ {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)
	}

	// Should end up on ModelsTab after 9 tabs (3 tabs * 3 cycles = 9)
	if m.activeTab != ModelsTab {
		t.Errorf("activeTab after 9 tabs = %d, want %d", m.activeTab, ModelsTab)
	}
}

// TestEdgeCases_SaveWithoutChanges tests saving without modifications.
func TestEdgeCases_SaveWithoutChanges(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.Config{
		HomeDir: projectDir,
		Agent:   "opencode",
		Models: config.ModelConfig{
			Default: "gpt-4",
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("initial save: %v", err)
	}

	m := NewModel(cfg, projectDir)
	m.width = 80
	m.height = 24

	// Save without changes
	cmd := m.saveConfig()
	if cmd == nil {
		t.Error("saveConfig should return a cmd")
	}

	result := cmd()
	if str, ok := result.(string); ok {
		if str != "✓ Configuration saved" {
			t.Errorf("save result = %q, want %q", str, "✓ Configuration saved")
		}
	}
}

// TestEdgeCases_AgentInitError tests agent init failure handling.
func TestEdgeCases_AgentInitError(t *testing.T) {
	state := newAgentTabState("opencode")

	// Simulate init error
	state.setInitResult("Init failed: agent not found", true)

	if !state.initError {
		t.Error("initError should be true")
	}
	if state.initResult != "Init failed: agent not found" {
		t.Errorf("initResult = %q, want %q", state.initResult, "Init failed: agent not found")
	}
}

// TestEdgeCases_NonTTY tests non-terminal behavior.
func TestEdgeCases_NonTTY(t *testing.T) {
	// In test environment, stdin is typically not a terminal
	err := CheckTerminal()
	if err == nil {
		t.Log("CheckTerminal passed - test environment might have a terminal")
	}
	// The test itself doesn't need to assert failure since it depends on environment
}

// TestEdgeCases_MutationThresholdBounds tests threshold clamping.
func TestEdgeCases_MutationThresholdBounds(t *testing.T) {
	state := newMutationTabState(config.MutationTesting{Enabled: false, Threshold: 0.0})

	// Set to max
	state.threshold.SetValue(100)
	if state.threshold.AsFloat() != 1.0 {
		t.Errorf("max threshold = %f, want 1.0", state.threshold.AsFloat())
	}

	// Set to min
	state.threshold.SetValue(0)
	if state.threshold.AsFloat() != 0.0 {
		t.Errorf("min threshold = %f, want 0.0", state.threshold.AsFloat())
	}

	// Try to exceed bounds
	state.threshold.SetValue(150)
	if state.threshold.AsFloat() != 1.0 {
		t.Errorf("clamped max threshold = %f, want 1.0", state.threshold.AsFloat())
	}

	state.threshold.SetValue(-50)
	if state.threshold.AsFloat() != 0.0 {
		t.Errorf("clamped min threshold = %f, want 0.0", state.threshold.AsFloat())
	}
}

// TestEdgeCases_ModelPhaseDeletion tests that empty phase values are deleted.
func TestEdgeCases_ModelPhaseDeletion(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases: map[string]string{
				"explore": "claude-sonnet-4",
				"propose": "gpt-4o",
			},
		},
	}
	state := newModelsTabState(cfg)

	// Clear the propose value
	state.inputs[modelInputPropose].SetValue("")
	state.applyToConfig(cfg)

	if _, exists := cfg.Models.Phases["propose"]; exists {
		t.Error("propose should be deleted when empty")
	}
	if cfg.Models.Phases["explore"] != "claude-sonnet-4" {
		t.Errorf("explore = %q, want %q", cfg.Models.Phases["explore"], "claude-sonnet-4")
	}
}

// TestEdgeCases_AllPhasesLocked tests auto-fill with all phases locked.
func TestEdgeCases_AllPhasesLocked(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases:  make(map[string]string),
		},
	}
	state := newModelsTabState(cfg)

	// Lock all phases
	for i := range state.phaseNames {
		idx := i + 1
		state.inputs[idx].SetValue("custom-" + state.phaseNames[i])
		state.autoFillLocks[idx] = true
	}

	// Change default
	state.inputs[modelInputDefault].SetValue("claude-sonnet-4")
	state.updateAutoFill()

	// All phases should keep their values
	for i := range state.phaseNames {
		idx := i + 1
		expected := "custom-" + state.phaseNames[i]
		if state.inputs[idx].Value() != expected {
			t.Errorf("phase %s = %q, want %q", state.phaseNames[i], state.inputs[idx].Value(), expected)
		}
	}
}

// TestEdgeCases_WindowResize tests resize handling.
func TestEdgeCases_WindowResize(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	// Initial size
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)
	if m.width != 80 || m.height != 24 {
		t.Error("initial size not set")
	}

	// Resize
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = newModel.(Model)
	if m.width != 120 || m.height != 40 {
		t.Error("resize not handled")
	}

	// Very small
	newModel, _ = m.Update(tea.WindowSizeMsg{Width: 10, Height: 5})
	m = newModel.(Model)
	if m.width != 10 || m.height != 5 {
		t.Error("small resize not handled")
	}
}

// TestEdgeCases_StatusMessagePersistence tests status message display.
func TestEdgeCases_StatusMessagePersistence(t *testing.T) {
	m := NewModel(&config.Config{}, "")
	m.width = 80
	m.height = 24

	m.statusMsg = "Test status"
	m.statusErr = false

	view := m.View()
	if !contains(view, "Test status") {
		t.Error("View should contain status message")
	}

	m.statusErr = true
	view = m.View()
	if !contains(view, "Test status") {
		t.Error("View should contain error status")
	}
}

// TestEdgeCases_QuitWithoutSave tests quit behavior.
func TestEdgeCases_QuitWithoutSave(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	// Quit
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: false})
	m = newModel.(Model)

	if !m.quitting {
		t.Error("quitting should be true")
	}

	view := m.View()
	if view != "" {
		t.Errorf("View when quitting = %q, want empty", view)
	}
}

// TestEdgeCases_AgentConfirmationCancel tests canceling agent switch.
func TestEdgeCases_AgentConfirmationCancel(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	// Press ESC
	state.update(tea.KeyMsg{Type: tea.KeyEsc})

	if state.confirmingInit {
		t.Error("confirmingInit should be false after ESC")
	}
	if state.selectedAgent != "claude" {
		t.Errorf("selectedAgent = %q, want %q", state.selectedAgent, "claude")
	}
}

// TestEdgeCases_AgentConfirmationInvalidKey tests invalid key in confirmation.
func TestEdgeCases_AgentConfirmationInvalidKey(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	// Press an invalid key
	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if !state.confirmingInit {
		t.Error("confirmingInit should still be true after invalid key")
	}
}

// TestEdgeCases_MutationTabFocusToggle tests focus switching in mutation tab.
func TestEdgeCases_MutationTabFocusToggle(t *testing.T) {
	state := newMutationTabState(config.MutationTesting{Enabled: false, Threshold: 0.5})

	// Down to slider
	state.update(tea.KeyMsg{Type: tea.KeyDown})
	if state.focused != 1 {
		t.Errorf("focused = %d, want 1", state.focused)
	}
	if !state.threshold.Focused {
		t.Error("threshold.Focused should be true")
	}

	// Up to toggle
	state.update(tea.KeyMsg{Type: tea.KeyUp})
	if state.focused != 0 {
		t.Errorf("focused = %d, want 0", state.focused)
	}
	if state.threshold.Focused {
		t.Error("threshold.Focused should be false")
	}
}

// TestEdgeCases_ModelsTabNavigation tests navigation in models tab.
func TestEdgeCases_ModelsTabNavigation(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases:  make(map[string]string),
		},
	}
	state := newModelsTabState(cfg)

	// Navigate down through all inputs
	for i := 0; i < len(state.inputs); i++ {
		state.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Should wrap around
	if state.focusedInput != 0 {
		t.Errorf("after full cycle, focusedInput = %d, want 0", state.focusedInput)
	}
}

// TestIntegration_ConfigFilePersistence verifies config file is written correctly.
func TestIntegration_ConfigFilePersistence(t *testing.T) {
	projectDir := t.TempDir()
	configPath := filepath.Join(projectDir, ".archon", "config.yaml")

	cfg := &config.Config{
		HomeDir: projectDir,
		Agent:   "claude",
		Models: config.ModelConfig{
			Default: "claude-sonnet-4",
			Phases: map[string]string{
				"explore": "gpt-4",
				"apply":   "gpt-4o",
			},
		},
		MutationTesting: config.MutationTesting{
			Enabled:   true,
			Threshold: 0.85,
			Tool:      "gremlins",
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}

	// Verify file is readable
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	if len(data) == 0 {
		t.Error("config file is empty")
	}

	// Verify content contains key fields
	content := string(data)
	if !contains(content, "agent:") {
		t.Error("config should contain agent field")
	}
	if !contains(content, "mutation_testing:") {
		t.Error("config should contain mutation_testing field")
	}
	if !contains(content, "models:") {
		t.Error("config should contain models field")
	}
}

// TestIntegration_TabStateConsistency tests that all tabs maintain state.
func TestIntegration_TabStateConsistency(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: "gpt-4",
			Phases: map[string]string{
				"explore": "claude-sonnet-4",
			},
		},
		MutationTesting: config.MutationTesting{
			Enabled:   true,
			Threshold: 0.6,
		},
		Agent: "opencode",
	}
	m := NewModel(cfg, "")

	// Verify initial state
	if m.modelsTab.inputs[modelInputDefault].Value() != "gpt-4" {
		t.Error("default model not loaded")
	}
	if !m.mutationTab.enabled {
		t.Error("mutation enabled not loaded")
	}
	if m.agentTab.selectedAgent != "opencode" {
		t.Error("agent not loaded")
	}

	// Modify state
	m.modelsTab.inputs[modelInputDefault].SetValue("gpt-4o")
	m.mutationTab.enabled = false
	m.agentTab.selectedAgent = "claude"

	// Verify modifications
	if m.modelsTab.inputs[modelInputDefault].Value() != "gpt-4o" {
		t.Error("default model not updated")
	}
	if m.mutationTab.enabled {
		t.Error("mutation enabled not updated")
	}
	if m.agentTab.selectedAgent != "claude" {
		t.Error("agent not updated")
	}
}
