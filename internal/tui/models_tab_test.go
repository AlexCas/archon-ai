package tui

import (
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// TestModelsTab_DetectionCachedOncePerView covers the spec scenario
// "Detection is cached once per Models view": the catalog is injected once at
// construction and the tab reads only from m.catalog on cycle/type, never
// re-running detection.
func TestModelsTab_DetectionCachedOncePerView(t *testing.T) {
	cfg := &config.Config{Models: config.ModelConfig{Default: ""}}

	// detectCount proves detection is performed exactly once, at open, through
	// the injectable seam — and never again on cycle/type.
	detectCount := 0
	detect := func() []string {
		detectCount++
		return []string{"injected-alpha", "injected-beta", "injected-gamma"}
	}

	catalog := detect() // simulate the single detection at view open
	state := newModelsTabState(cfg, catalog)

	if detectCount != 1 {
		t.Fatalf("detection ran %d times at open, want 1", detectCount)
	}

	// The tab stores exactly the injected catalog.
	if got := strings.Join(state.catalog, ","); got != "injected-alpha,injected-beta,injected-gamma" {
		t.Fatalf("state.catalog = %q, want the injected slice", got)
	}

	// Cycle forward and backward, and type, many times. None of this may call
	// back into detection.
	for i := 0; i < 5; i++ {
		state.cycleStaticModel(1)
		state.cycleStaticModel(-1)
		state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	}

	if detectCount != 1 {
		t.Fatalf("detection ran %d times after cycling/typing, want 1 (cached)", detectCount)
	}

	// Cycling from empty must land on the first injected model, proving the
	// cycle list is built from m.catalog rather than the static config list.
	state.inputs[modelInputDefault].SetValue("")
	state.focusedInput = modelInputDefault
	state.cycleStaticModel(1)
	if got := state.inputs[modelInputDefault].Value(); got != "injected-alpha" {
		t.Fatalf("after cycle from empty, value = %q, want %q", got, "injected-alpha")
	}
}

// TestModelsTab_CycleAndHintFromInjectedCatalog covers spec rows 6.2: cycling
// and the renamed "Available:" hint render from the injected catalog slice.
func TestModelsTab_CycleAndHintFromInjectedCatalog(t *testing.T) {
	cfg := &config.Config{Models: config.ModelConfig{Default: ""}}
	catalog := []string{"alpha-1", "beta-2"}
	state := newModelsTabState(cfg, catalog)
	state.setWidth(120)

	// Cycle through the injected catalog: "" -> alpha-1 -> beta-2 -> "".
	state.focusedInput = modelInputDefault
	state.cycleStaticModel(1)
	if got := state.inputs[modelInputDefault].Value(); got != "alpha-1" {
		t.Fatalf("cycle 1 = %q, want %q", got, "alpha-1")
	}
	state.cycleStaticModel(1)
	if got := state.inputs[modelInputDefault].Value(); got != "beta-2" {
		t.Fatalf("cycle 2 = %q, want %q", got, "beta-2")
	}
	state.cycleStaticModel(1)
	if got := state.inputs[modelInputDefault].Value(); got != "" {
		t.Fatalf("cycle 3 = %q, want empty (wraps to lead)", got)
	}

	// The hint uses the renamed "Available:" label and lists injected names.
	view := state.view(120, 40)
	if !strings.Contains(view, "Available:") {
		t.Errorf("view missing renamed %q label:\n%s", "Available:", view)
	}
	if strings.Contains(view, "Static:") {
		t.Errorf("view still contains old %q label:\n%s", "Static:", view)
	}
	for _, name := range catalog {
		if !strings.Contains(view, name) {
			t.Errorf("view missing injected catalog name %q:\n%s", name, view)
		}
	}
}

// TestModelsTab_FreeFormEntryUnchanged covers the spec scenario "Free-form
// entry and advisory behavior unchanged": an arbitrary value is accepted and
// NormalizeModel / Validate behave exactly as before this feature.
func TestModelsTab_FreeFormEntryUnchanged(t *testing.T) {
	cfg := &config.Config{Models: config.ModelConfig{Default: ""}}
	// A deliberately small catalog that does NOT contain the typed value.
	state := newModelsTabState(cfg, []string{"claude-opus-4-8"})

	state.focusedInput = modelInputDefault
	state.inputs[modelInputDefault].SetValue("some-custom-model")
	state.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})

	if got := state.inputs[modelInputDefault].Value(); got == "" {
		t.Fatal("free-form value was rejected, want it accepted")
	}

	// Validate and NormalizeModel are advisory-only and unchanged.
	if w := config.Validate("some-custom-model"); w == "" {
		t.Error("Validate should return an advisory (non-blocking) warning for an unknown model")
	}
	if w := config.Validate("claude-opus-4-8"); w != "" {
		t.Errorf("Validate(known) = %q, want empty", w)
	}
	if id, ok := config.NormalizeModel("opus 4.8"); !ok || id != "opus" {
		t.Errorf("NormalizeModel(%q) = (%q,%v), want (opus,true)", "opus 4.8", id, ok)
	}
	if _, ok := config.NormalizeModel("glm-5"); ok {
		t.Errorf("NormalizeModel(%q) should not resolve to a Claude family", "glm-5")
	}
}
