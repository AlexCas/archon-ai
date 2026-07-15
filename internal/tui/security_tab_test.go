package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/archon-ai/archon/internal/config"
)

// TestSecurityTabState_DefaultProfile verifies that an empty Profile in the
// config initialises the cycle to "cli" (index 0).
func TestSecurityTabState_DefaultProfile(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false, Profile: ""})
	if state.profile() != "cli" {
		t.Errorf("default profile = %q, want %q", state.profile(), "cli")
	}
}

// TestSecurityTabState_LoadExistingProfile verifies that a pre-set Profile is
// honoured on construction.
func TestSecurityTabState_LoadExistingProfile(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: true, Profile: "web"})
	if state.profile() != "web" {
		t.Errorf("loaded profile = %q, want %q", state.profile(), "web")
	}
	if !state.enabled {
		t.Error("enabled should be true when loaded with Enabled:true")
	}
}

// TestSecurityTabState_ToggleEnabled verifies that pressing Enter on focus 0
// toggles the enabled flag.
func TestSecurityTabState_ToggleEnabled(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false})

	if state.enabled {
		t.Error("enabled should be false initially")
	}

	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if !state.enabled {
		t.Error("enabled should be true after toggle")
	}

	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if state.enabled {
		t.Error("enabled should be false after second toggle")
	}
}

// TestSecurityTabState_SpaceToggle verifies that Space also toggles enabled.
func TestSecurityTabState_SpaceToggle(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false})
	state.update(tea.KeyMsg{Type: tea.KeySpace})
	if !state.enabled {
		t.Error("enabled should be true after Space toggle")
	}
}

// TestSecurityTabState_CycleProfile verifies that pressing Enter on focus 1
// cycles through the two profiles in order: cli → web → cli.
func TestSecurityTabState_CycleProfile(t *testing.T) {
	state := newSecurityTabState(config.Security{Profile: "cli"})
	state.focused = 1 // move focus to profile selector

	if state.profile() != "cli" {
		t.Fatalf("initial profile = %q, want %q", state.profile(), "cli")
	}

	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if state.profile() != "web" {
		t.Errorf("after first cycle = %q, want %q", state.profile(), "web")
	}

	state.update(tea.KeyMsg{Type: tea.KeyEnter})
	if state.profile() != "cli" {
		t.Errorf("after second cycle = %q, want %q", state.profile(), "cli")
	}
}

// TestSecurityTabState_CycleOnlyValidProfiles asserts that repeated cycling
// never produces a value outside {"cli", "web"}.
func TestSecurityTabState_CycleOnlyValidProfiles(t *testing.T) {
	state := newSecurityTabState(config.Security{Profile: "cli"})
	state.focused = 1

	validProfiles := map[string]bool{"cli": true, "web": true}

	for i := 0; i < 20; i++ {
		state.update(tea.KeyMsg{Type: tea.KeyEnter})
		if !validProfiles[state.profile()] {
			t.Errorf("cycle iteration %d: profile = %q is not a valid profile", i, state.profile())
		}
	}
}

// TestSecurityTabState_FocusNavigation verifies Up/Down move between the two
// focusable controls.
func TestSecurityTabState_FocusNavigation(t *testing.T) {
	state := newSecurityTabState(config.Security{})

	if state.focused != 0 {
		t.Errorf("initial focus = %d, want 0", state.focused)
	}

	state.update(tea.KeyMsg{Type: tea.KeyDown})
	if state.focused != 1 {
		t.Errorf("after Down = %d, want 1", state.focused)
	}

	state.update(tea.KeyMsg{Type: tea.KeyDown})
	if state.focused != 0 {
		t.Errorf("after second Down (wrap) = %d, want 0", state.focused)
	}

	state.update(tea.KeyMsg{Type: tea.KeyUp})
	if state.focused != 1 {
		t.Errorf("after Up (wrap) = %d, want 1", state.focused)
	}
}

// TestSecurityTabState_ApplyToConfig verifies that applyToConfig writes
// both Enabled and Profile back into the config struct correctly.
func TestSecurityTabState_ApplyToConfig(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false, Profile: "cli"})

	// Toggle enabled and cycle to "web".
	state.update(tea.KeyMsg{Type: tea.KeyEnter}) // enable
	state.focused = 1
	state.update(tea.KeyMsg{Type: tea.KeyEnter}) // cycle to web

	cfg := &config.Config{}
	state.applyToConfig(cfg)

	if !cfg.Security.Enabled {
		t.Error("config.Security.Enabled should be true")
	}
	if cfg.Security.Profile != "web" {
		t.Errorf("config.Security.Profile = %q, want %q", cfg.Security.Profile, "web")
	}
}

// TestSecurityTabState_ApplyToConfig_Disabled verifies the disabled + cli case.
func TestSecurityTabState_ApplyToConfig_Disabled(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false, Profile: "cli"})

	cfg := &config.Config{}
	state.applyToConfig(cfg)

	if cfg.Security.Enabled {
		t.Error("config.Security.Enabled should be false")
	}
	if cfg.Security.Profile != "cli" {
		t.Errorf("config.Security.Profile = %q, want %q", cfg.Security.Profile, "cli")
	}
}

// TestSecurityTabState_View verifies the view renders without panicking and
// includes the key UI labels.
func TestSecurityTabState_View(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: true, Profile: "web"})
	view := state.view(80, 24)

	if view == "" {
		t.Error("view should not be empty")
	}
	for _, want := range []string{"Security", "Enabled", "Profile", "web"} {
		if !contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

// TestSecurityTabState_InvalidProfileCoercesToCli verifies that a non-empty
// but out-of-set stored profile (e.g. from a hand-edited config) is coerced
// to "cli" (index 0) on construction, so it can never surface in the UI.
func TestSecurityTabState_InvalidProfileCoercesToCli(t *testing.T) {
	state := newSecurityTabState(config.Security{Enabled: false, Profile: "llm"})
	if state.profile() != "cli" {
		t.Errorf("invalid profile %q: got %q, want %q", "llm", state.profile(), "cli")
	}
}

// TestSecurityTabState_SetWidth verifies setWidth does not panic (no-op).
func TestSecurityTabState_SetWidth(t *testing.T) {
	state := newSecurityTabState(config.Security{})
	// Must not panic.
	state.setWidth(80)
	state.setWidth(0)
}
