package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/archon-ai/archon/internal/config"
)

// TestAgentFlow_SwitchAgent tests the full agent switch flow.
func TestAgentFlow_SwitchAgent(t *testing.T) {
	state := newAgentTabState("opencode")

	// Select claude
	state.focusedIndex = 1
	state.update(tea.KeyMsg{Type: tea.KeyEnter})

	if !state.confirmingInit {
		t.Error("should be confirming init")
	}
	if state.selectedAgent != "claude" {
		t.Errorf("selectedAgent = %q, want %q", state.selectedAgent, "claude")
	}

	// Confirm
	cmd := state.triggerInit()
	if cmd == nil {
		t.Error("triggerInit should return a cmd")
	}

	// Execute the command
	if cmd != nil {
		msg := cmd()
		if msg == nil {
			t.Error("init cmd should return a message")
		}
		if initMsg, ok := msg.(agentInitMsg); ok {
			if initMsg.agent != "claude" {
				t.Errorf("initMsg.agent = %q, want %q", initMsg.agent, "claude")
			}
		} else {
			t.Errorf("expected agentInitMsg, got %T", msg)
		}
	}
}

// TestAgentFlow_CancelSwitch tests canceling agent switch.
func TestAgentFlow_CancelSwitch(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if state.confirmingInit {
		t.Error("confirmingInit should be false")
	}
	if state.selectedAgent != "claude" {
		t.Errorf("selectedAgent = %q, want %q", state.selectedAgent, "claude")
	}
}

// TestAgentFlow_SameAgent tests selecting same agent.
func TestAgentFlow_SameAgent(t *testing.T) {
	state := newAgentTabState("opencode")

	// Select opencode (same as current)
	state.focusedIndex = 0
	state.update(tea.KeyMsg{Type: tea.KeyEnter})

	if state.confirmingInit {
		t.Error("should not confirm init for same agent")
	}
}

// TestAgentFlow_ViewWithResult tests view with init result.
func TestAgentFlow_ViewWithResult(t *testing.T) {
	state := newAgentTabState("opencode")
	state.setInitResult("✓ Agent initialized successfully", false)

	view := state.view(80, 24)
	if !contains(view, "✓ Agent initialized successfully") {
		t.Error("view should contain success message")
	}

	state.setInitResult("Error: agent not found", true)
	view = state.view(80, 24)
	if !contains(view, "Error: agent not found") {
		t.Error("view should contain error message")
	}
}

// TestAgentFlow_AvailableAgents tests all agents are listed.
func TestAgentFlow_AvailableAgents(t *testing.T) {
	state := newAgentTabState("opencode")

	view := state.view(80, 24)
	for _, agent := range availableAgents {
		if !contains(view, agent) {
			t.Errorf("view should contain agent %q", agent)
		}
	}
}

// TestAgentFlow_ConfirmDialog tests confirm dialog rendering.
func TestAgentFlow_ConfirmDialog(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	view := state.view(80, 24)
	if !contains(view, "Re-initialization") {
		t.Error("confirm dialog should contain title")
	}
	if !contains(view, "claude") {
		t.Error("confirm dialog should contain agent name")
	}
	if !contains(view, "Proceed?") {
		t.Error("confirm dialog should contain proceed prompt")
	}
}

// TestAgentFlow_ConfirmKeys tests various confirmation keys.
func TestAgentFlow_ConfirmKeys(t *testing.T) {
	tests := []struct {
		key      rune
		shouldInit bool
	}{
		{'y', true},
		{'Y', true},
		{'n', false},
		{'N', false},
	}

	for _, tt := range tests {
		state := newAgentTabState("opencode")
		state.selectedAgent = "claude"
		state.confirmingInit = true

		state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})

		if tt.shouldInit {
			if state.confirmingInit {
				t.Errorf("key %q: should have triggered init", tt.key)
			}
		} else {
			if state.confirmingInit {
				t.Errorf("key %q: should have canceled", tt.key)
			}
		}
	}
}

// TestAgentFlow_ConfirmEnter tests Enter key in confirmation.
func TestAgentFlow_ConfirmEnter(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	cmd, _ := state.update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("Enter should trigger init")
	}
}

// TestAgentFlow_ConfirmEsc tests ESC key in confirmation.
func TestAgentFlow_ConfirmEsc(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	state.update(tea.KeyMsg{Type: tea.KeyEsc})

	if state.confirmingInit {
		t.Error("ESC should cancel")
	}
}

// TestAgentFlow_NavigationBounds tests navigation bounds.
func TestAgentFlow_NavigationBounds(t *testing.T) {
	state := newAgentTabState("opencode")

	// Focus first item
	state.focusedIndex = 0

	// Navigate up from top
	state.update(tea.KeyMsg{Type: tea.KeyUp})
	if state.focusedIndex != 0 {
		t.Errorf("up from top = %d, want 0", state.focusedIndex)
	}

	// Navigate down to bottom
	for i := 0; i < len(availableAgents)+5; i++ {
		state.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if state.focusedIndex >= len(availableAgents) {
		t.Errorf("focusedIndex = %d, should be < %d", state.focusedIndex, len(availableAgents))
	}
}

// TestAgentFlow_ApplyToConfig tests applying to config.
func TestAgentFlow_ApplyToConfig(t *testing.T) {
	cfg := &config.Config{Agent: "opencode"}
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"

	state.applyToConfig(cfg)

	if cfg.Agent != "claude" {
		t.Errorf("agent = %q, want %q", cfg.Agent, "claude")
	}
}

// TestAgentFlow_ConfirmDialogRendering tests confirm dialog rendering.
func TestAgentFlow_ConfirmDialogRendering(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	dialog := state.renderConfirmDialog()
	if dialog == "" {
		t.Error("confirm dialog should not be empty")
	}
	if !contains(dialog, "claude") {
		t.Error("confirm dialog should contain agent name")
	}
}

// TestAgentFlow_MarkerCurrentAgent tests current agent marker.
func TestAgentFlow_MarkerCurrentAgent(t *testing.T) {
	state := newAgentTabState("opencode")
	state.focusedIndex = 0

	view := state.view(80, 24)
	if !contains(view, "▸") {
		t.Error("view should contain marker for current agent")
	}
}

// TestAgentFlow_InvalidKeyDuringConfirm tests invalid key during confirm.
func TestAgentFlow_InvalidKeyDuringConfirm(t *testing.T) {
	state := newAgentTabState("opencode")
	state.selectedAgent = "claude"
	state.confirmingInit = true

	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if !state.confirmingInit {
		t.Error("confirmingInit should still be true for invalid key")
	}
}

// TestAgentFlow_ConfirmDialogWithEmptyAgent tests confirm dialog with empty agent.
func TestAgentFlow_ConfirmDialogWithEmptyAgent(t *testing.T) {
	state := newAgentTabState("")
	state.selectedAgent = "opencode"
	state.confirmingInit = true

	dialog := state.renderConfirmDialog()
	if !contains(dialog, "opencode") {
		t.Error("confirm dialog should contain selected agent")
	}
}

// TestAgentFlow_MultipleSwitches tests multiple agent switches.
func TestAgentFlow_MultipleSwitches(t *testing.T) {
	state := newAgentTabState("opencode")

	// Switch to claude
	state.focusedIndex = 1
	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if state.selectedAgent != "claude" {
		t.Error("should select claude")
	}

	// Cancel
	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if state.confirmingInit {
		t.Error("should cancel")
	}

	// Switch to codex
	state.focusedIndex = 2
	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if state.selectedAgent != "codex" {
		t.Error("should select codex")
	}
}

// TestAgentFlow_StatusMessageClears tests that status messages can be cleared.
func TestAgentFlow_StatusMessageClears(t *testing.T) {
	state := newAgentTabState("opencode")
	state.setInitResult("Error", true)

	if state.initResult != "Error" {
		t.Error("initResult should be set")
	}

	state.setInitResult("", false)
	if state.initResult != "" {
		t.Error("initResult should be cleared")
	}
	if state.initError {
		t.Error("initError should be false")
	}
}
