package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
	tea "github.com/charmbracelet/bubbletea"
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
			Default: config.ModelRef{Model: "gpt-4"},
			Phases: map[string]config.ModelRef{
				"explore": {Model: "claude-sonnet-4"},
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

	// Set values via the new rows API — directly set refs and mark changed.
	m.modelsTab.rows[0].ref = config.ParseModelRef("gpt-4o")
	m.modelsTab.rows[0].changed = true
	// explore is rows[1] (PhaseOrder[0])
	m.modelsTab.rows[1].ref = config.ParseModelRef("gpt-4o-mini")
	m.modelsTab.rows[1].changed = true

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

	if reloaded.Models.Default.FullID() != "gpt-4o" {
		t.Errorf("reloaded default = %q, want %q", reloaded.Models.Default.FullID(), "gpt-4o")
	}
	if reloaded.Models.Phases["explore"].FullID() != "gpt-4o-mini" {
		t.Errorf("reloaded explore = %q, want %q", reloaded.Models.Phases["explore"].FullID(), "gpt-4o-mini")
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
			Default: config.ModelRef{},
			Phases:  make(map[string]config.ModelRef),
		},
	}
	m := NewModel(cfg, "")

	// Default row should have empty ref
	if m.modelsTab.rows[0].ref.FullID() != "" {
		t.Errorf("default row ref = %q, want empty", m.modelsTab.rows[0].ref.FullID())
	}
}

// TestEdgeCases_LegacyModelRendersWithoutPanic asserts the Models tab renders a
// legacy/unknown model value (not in the catalog) without panicking. The picker
// keeps such values verbatim; there is no advisory warning (dropped per D2).
func TestEdgeCases_LegacyModelRendersWithoutPanic(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "unknown-model-xyz"},
		},
	}
	m := NewModel(cfg, "")
	m.width = 80
	m.height = 24

	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

// TestEdgeCases_QuickTabSwitching tests rapid tab switching.
func TestEdgeCases_QuickTabSwitching(t *testing.T) {
	m := NewModel(&config.Config{}, "")

	// Switch through all tabs an exact number of full cycles.
	cycles := int(tabCount) * 3
	for i := 0; i < cycles; i++ {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)
	}

	// A whole number of full cycles must land back on AgentTab.
	if m.activeTab != AgentTab {
		t.Errorf("activeTab after %d tabs = %d, want %d", cycles, m.activeTab, AgentTab)
	}
}

// TestEdgeCases_SaveWithoutChanges tests saving without modifications.
func TestEdgeCases_SaveWithoutChanges(t *testing.T) {
	projectDir := t.TempDir()

	cfg := &config.Config{
		HomeDir: projectDir,
		Agent:   "opencode",
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
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
			Default: config.ModelRef{Model: "gpt-4"},
			Phases: map[string]config.ModelRef{
				"explore": {Model: "claude-sonnet-4"},
				"propose": {Model: "gpt-4o"},
			},
		},
	}
	state := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Clear the propose value via free-form (rows[2] = propose, PhaseOrder index 1)
	// Find the propose row index
	proposeIdx := -1
	for i, r := range state.rows {
		if r.phase == "propose" {
			proposeIdx = i
			break
		}
	}
	if proposeIdx < 0 {
		t.Fatal("propose row not found")
	}
	state.rows[proposeIdx].ref = config.ModelRef{} // clear
	state.rows[proposeIdx].changed = true
	state.applyToConfig(cfg)

	if _, exists := cfg.Models.Phases["propose"]; exists {
		t.Error("propose should be deleted when empty")
	}
	if cfg.Models.Phases["explore"].FullID() != "claude-sonnet-4" {
		t.Errorf("explore = %q, want %q", cfg.Models.Phases["explore"].FullID(), "claude-sonnet-4")
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
	if !strings.Contains(view, "Test status") {
		t.Error("View should contain status message")
	}

	m.statusErr = true
	view = m.View()
	if !strings.Contains(view, "Test status") {
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

// TestEdgeCases_ModelsTabNavigation tests navigation in models tab using Down key.
func TestEdgeCases_ModelsTabNavigation(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
			Phases:  make(map[string]config.ModelRef),
		},
	}
	state := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Navigate down through all rows (clamped — will stop at the last row)
	total := len(state.rows)
	for i := 0; i < total+2; i++ {
		state.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Should clamp at the last row index
	if state.focusedRow != total-1 {
		t.Errorf("after clamped navigation, focusedRow = %d, want %d", state.focusedRow, total-1)
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
			Default: config.ModelRef{Model: "claude-sonnet-4"},
			Phases: map[string]config.ModelRef{
				"explore": {Model: "gpt-4"},
				"apply":   {Model: "gpt-4o"},
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
	if !strings.Contains(content, "agent:") {
		t.Error("config should contain agent field")
	}
	if !strings.Contains(content, "mutation_testing:") {
		t.Error("config should contain mutation_testing field")
	}
	if !strings.Contains(content, "models:") {
		t.Error("config should contain models field")
	}
}

// TestIntegration_TabStateConsistency tests that all tabs maintain state.
func TestIntegration_TabStateConsistency(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "gpt-4"},
			Phases: map[string]config.ModelRef{
				"explore": {Model: "claude-sonnet-4"},
			},
		},
		MutationTesting: config.MutationTesting{
			Enabled:   true,
			Threshold: 0.6,
		},
		Agent: "opencode",
	}
	m := NewModel(cfg, "")

	// Verify initial state — rows[0] is Default
	if m.modelsTab.rows[0].ref.FullID() != "gpt-4" {
		t.Errorf("default model not loaded: %q", m.modelsTab.rows[0].ref.FullID())
	}
	if !m.mutationTab.enabled {
		t.Error("mutation enabled not loaded")
	}
	if m.agentTab.selectedAgent != "opencode" {
		t.Error("agent not loaded")
	}

	// Modify state
	m.modelsTab.rows[0].ref = config.ParseModelRef("gpt-4o")
	m.modelsTab.rows[0].changed = true
	m.mutationTab.enabled = false
	m.agentTab.selectedAgent = "claude"

	// Verify modifications
	if m.modelsTab.rows[0].ref.FullID() != "gpt-4o" {
		t.Error("default model not updated")
	}
	if m.mutationTab.enabled {
		t.Error("mutation enabled not updated")
	}
	if m.agentTab.selectedAgent != "claude" {
		t.Error("agent not updated")
	}
}
