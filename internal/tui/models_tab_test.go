package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
	tea "github.com/charmbracelet/bubbletea"
)

// prov builds a minimal opencode.Provider for use in tests.
func prov(id string, models ...opencode.Model) opencode.Provider {
	m := make(map[string]opencode.Model, len(models))
	for _, mod := range models {
		m[mod.ID] = mod
	}
	return opencode.Provider{ID: id, Name: id, Models: m}
}

// toolModel returns a Model with ToolCall=true.
func toolModel(id, name string) opencode.Model {
	return opencode.Model{ID: id, Name: name, ToolCall: true}
}

// noToolModel returns a Model with ToolCall=false (not usable for SDD picks).
func noToolModel(id, name string) opencode.Model {
	return opencode.Model{ID: id, Name: name, ToolCall: false}
}

// keyMsg constructs a tea.KeyMsg for a rune character.
func keyMsg(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// TestModelsTab_PickProviderModelSetsRef covers S1: full provider→model pick flow.
func TestModelsTab_PickProviderModelSetsRef(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			toolModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
			toolModel("anthropic/claude-sonnet-4-6", "Claude Sonnet 4.6"),
		),
	}
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{},
		},
	}
	st := newModelsTabState(cfg, providers, nil)

	// Start at row 0 (Default), navigate down to row 1 (explore)
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	if st.focusedRow != 1 {
		t.Fatalf("focusedRow = %d, want 1", st.focusedRow)
	}

	// Enter → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != providerSelect {
		t.Fatalf("mode = %v, want providerSelect", st.mode)
	}
	// Only one provider "anthropic", cursor at 0. Enter → modelSelect.
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != modelSelect {
		t.Fatalf("mode = %v, want modelSelect", st.mode)
	}

	// Models are sorted by Name: Claude Opus 4.8 < Claude Sonnet 4.6
	// cursor at 0 → "Claude Opus 4.8"; Down → "Claude Sonnet 4.6"
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	// Enter → set ref and return to rowNav
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != rowNav {
		t.Fatalf("mode = %v, want rowNav", st.mode)
	}
	if !st.rows[1].changed {
		t.Error("rows[1].changed should be true after pick")
	}
	got := st.rows[1].ref.FullID()
	if got != "anthropic/claude-sonnet-4-6" {
		t.Errorf("rows[1].ref.FullID() = %q, want %q", got, "anthropic/claude-sonnet-4-6")
	}
}

// TestModelsTab_OpencodeBareKeyMapsToRef covers S2: opencode bare key.
func TestModelsTab_OpencodeBareKeyMapsToRef(t *testing.T) {
	providers := map[string]opencode.Provider{
		"opencode": prov("opencode",
			toolModel("deepseek-v4-pro", "DeepSeek V4 Pro"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	// Enter → providerSelect (one provider: opencode at index 0)
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	// Enter → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	// Enter → pick deepseek-v4-pro
	st.update(tea.KeyMsg{Type: tea.KeyEnter})

	row := st.rows[0]
	if row.ref.Provider != "opencode" {
		t.Errorf("ref.Provider = %q, want %q", row.ref.Provider, "opencode")
	}
	if row.ref.Model != "deepseek-v4-pro" {
		t.Errorf("ref.Model = %q, want %q", row.ref.Model, "deepseek-v4-pro")
	}
	if row.ref.FullID() != "opencode/deepseek-v4-pro" {
		t.Errorf("FullID() = %q, want %q", row.ref.FullID(), "opencode/deepseek-v4-pro")
	}
}

// TestModelsTab_SlashedKeyNoDoublePrefix covers S2: already-slashed key.
func TestModelsTab_SlashedKeyNoDoublePrefix(t *testing.T) {
	providers := map[string]opencode.Provider{
		"xai": prov("xai",
			toolModel("xai/grok-4", "Grok 4"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect (xai)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → pick xai/grok-4

	row := st.rows[0]
	if row.ref.FullID() != "xai/grok-4" {
		t.Errorf("FullID() = %q, want %q (no double-prefix)", row.ref.FullID(), "xai/grok-4")
	}
}

// TestModelsTab_EscCancelsPicker covers S3: Esc from providerSelect → rowNav.
func TestModelsTab_EscCancelsPicker(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic", toolModel("anthropic/claude-opus-4-8", "Claude Opus 4.8")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	origRef := st.rows[0].ref

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	if st.mode != providerSelect {
		t.Fatalf("expected providerSelect, got %v", st.mode)
	}
	st.update(tea.KeyMsg{Type: tea.KeyEsc}) // cancel
	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Esc", st.mode)
	}
	if st.rows[0].changed {
		t.Error("changed should be false after cancel")
	}
	if st.rows[0].ref != origRef {
		t.Error("ref should be unchanged after cancel")
	}
}

// TestModelsTab_EscFromModelSelectGoesBack covers S3: Esc from modelSelect → providerSelect.
func TestModelsTab_EscFromModelSelectGoesBack(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic", toolModel("anthropic/claude-opus-4-8", "Claude Opus 4.8")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	if st.mode != modelSelect {
		t.Fatalf("expected modelSelect, got %v", st.mode)
	}
	st.update(tea.KeyMsg{Type: tea.KeyEsc}) // back
	if st.mode != providerSelect {
		t.Errorf("mode = %v, want providerSelect after Esc", st.mode)
	}
}

// TestModelsTab_FreeFormTogglesAndParses covers S4.
func TestModelsTab_FreeFormTogglesAndParses(t *testing.T) {
	cfg := &config.Config{}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	origRef := st.rows[0].ref

	// Press 'e' → freeForm
	st.update(keyMsg('e'))
	if st.mode != freeForm {
		t.Fatalf("mode = %v, want freeForm after 'e'", st.mode)
	}

	// Type "x/y"
	st.update(keyMsg('x'))
	st.update(keyMsg('/'))
	st.update(keyMsg('y'))

	// Enter → confirm
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Enter", st.mode)
	}
	if !st.rows[0].changed {
		t.Error("rows[0].changed should be true")
	}
	want := config.ParseModelRef("x/y")
	if st.rows[0].ref != want {
		t.Errorf("ref = %+v, want %+v", st.rows[0].ref, want)
	}
	_ = origRef
}

// TestModelsTab_FreeFormEscCancels covers S4: Esc from freeForm.
func TestModelsTab_FreeFormEscCancels(t *testing.T) {
	cfg := &config.Config{Models: config.ModelConfig{Default: config.ParseModelRef("openai/gpt-4o")}}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)
	origRef := st.rows[0].ref

	st.update(keyMsg('e'))
	// Type something
	st.update(keyMsg('z'))
	// Esc → cancel
	st.update(tea.KeyMsg{Type: tea.KeyEsc})

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Esc", st.mode)
	}
	if st.rows[0].changed {
		t.Error("changed should be false after Esc cancel")
	}
	if st.rows[0].ref != origRef {
		t.Errorf("ref changed on cancel: got %+v, want %+v", st.rows[0].ref, origRef)
	}
}

// TestModelsTab_LegacyUntouchedPreserved covers S5.
func TestModelsTab_LegacyUntouchedPreserved(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Default: config.ModelRef{Model: "opus"}, // bare alias, Provider=""
			Phases:  map[string]config.ModelRef{},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Only navigate, do not edit
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	st.update(tea.KeyMsg{Type: tea.KeyUp})

	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	st.applyToConfig(outCfg)

	if outCfg.Models.Default.Provider != "" {
		t.Errorf("Default.Provider = %q, want empty (legacy preserved)", outCfg.Models.Default.Provider)
	}
	if outCfg.Models.Default.Model != "opus" {
		t.Errorf("Default.Model = %q, want %q", outCfg.Models.Default.Model, "opus")
	}
}

// TestModelsTab_ApplyDeletesClearedPhase covers S6.
func TestModelsTab_ApplyDeletesClearedPhase(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"explore": config.ParseModelRef("anthropic/claude-opus-4-8"),
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Find the explore row (rows[1] = PhaseOrder[0])
	exploreIdx := -1
	for i, r := range st.rows {
		if r.phase == "explore" {
			exploreIdx = i
			break
		}
	}
	if exploreIdx < 0 {
		t.Fatal("explore row not found")
	}
	// Navigate to explore row
	for i := 0; i < exploreIdx; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Free-form: press 'e', clear value, Enter
	st.update(keyMsg('e'))
	// Clear the input value by setting it directly (simulates user deleting)
	st.input.SetValue("")
	st.update(tea.KeyMsg{Type: tea.KeyEnter})

	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	st.applyToConfig(outCfg)

	if _, exists := outCfg.Models.Phases["explore"]; exists {
		t.Error("explore should be deleted after clearing")
	}
}

// TestModelsTab_CorruptCacheWarningShown covers S7.
func TestModelsTab_CorruptCacheWarningShown(t *testing.T) {
	cfg := &config.Config{}
	loadErr := errors.New("json: cannot unmarshal")
	st := newModelsTabState(cfg, nil, loadErr)
	st.setWidth(80)

	view := st.view(80, 24)
	if !strings.Contains(view, "⚠") {
		t.Errorf("view should contain ⚠ warning when cache is corrupt; view:\n%s", view)
	}
	if !strings.Contains(view, "unreadable") {
		t.Errorf("view should mention unreadable; view:\n%s", view)
	}
}

// TestModelsTab_AbsentCacheFreeFormOnly covers S8.
func TestModelsTab_AbsentCacheFreeFormOnly(t *testing.T) {
	cfg := &config.Config{}
	// Empty providers, nil err = absent cache
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	view := st.view(80, 24)
	if strings.Contains(view, "⚠") {
		t.Errorf("view should NOT contain ⚠ for absent cache; view:\n%s", view)
	}

	// Enter on a row should open freeForm directly (no provider list)
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != freeForm {
		t.Errorf("mode = %v, want freeForm when no providers", st.mode)
	}
}

// TestModelsTab_LeaderUsesPicker covers S9.
func TestModelsTab_LeaderUsesPicker(t *testing.T) {
	providers := map[string]opencode.Provider{
		"openai": prov("openai", toolModel("openai/gpt-4o", "GPT-4o")),
	}

	// opencode agent: leader row present
	cfgOC := &config.Config{Agent: "opencode"}
	st := newModelsTabState(cfgOC, providers, nil)

	lastRow := st.rows[len(st.rows)-1]
	if lastRow.kind != rowLeader {
		t.Errorf("last row kind = %v, want rowLeader", lastRow.kind)
	}

	view := st.view(80, 24)
	if !strings.Contains(view, "Leader") {
		t.Errorf("view should contain Leader for opencode agent; view:\n%s", view)
	}

	// Navigate to leader row and pick a model
	for i := 0; i < len(st.rows)-1; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect (openai)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → pick gpt-4o

	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	st.applyToConfig(outCfg)
	if outCfg.Models.Leader.FullID() != "openai/gpt-4o" {
		t.Errorf("Leader.FullID() = %q, want %q", outCfg.Models.Leader.FullID(), "openai/gpt-4o")
	}

	// Non-opencode agent: no leader row
	cfgClaude := &config.Config{Agent: "claude"}
	st2 := newModelsTabState(cfgClaude, providers, nil)
	for _, r := range st2.rows {
		if r.kind == rowLeader {
			t.Error("non-opencode agent should not have leader row")
		}
	}
	view2 := st2.view(80, 24)
	if strings.Contains(view2, "Leader") {
		t.Errorf("non-opencode view should not contain Leader; view:\n%s", view2)
	}
}

// TestModelsTab_DeterministicSortedLists covers S10.
func TestModelsTab_DeterministicSortedLists(t *testing.T) {
	providers := map[string]opencode.Provider{
		"zz-provider": prov("zz-provider",
			toolModel("zz-provider/m-b", "B Model"),
			toolModel("zz-provider/m-a", "A Model"),
		),
		"aa-provider": prov("aa-provider",
			toolModel("aa-provider/m-x", "X Model"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	// Providers must be sorted: aa-provider < zz-provider
	if len(st.available) < 2 {
		t.Fatalf("expected 2 providers, got %d", len(st.available))
	}
	if st.available[0] != "aa-provider" || st.available[1] != "zz-provider" {
		t.Errorf("providers not sorted: %v", st.available)
	}

	// Enter → providerSelect → pick zz-provider (index 1)
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	st.update(tea.KeyMsg{Type: tea.KeyDown})  // cursor to zz-provider
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect

	// Models sorted by Name: "A Model" < "B Model"
	if len(st.curModels) < 2 {
		t.Fatalf("expected 2 models, got %d", len(st.curModels))
	}
	if st.curModels[0].Name != "A Model" || st.curModels[1].Name != "B Model" {
		t.Errorf("models not sorted by Name: %v", st.curModels)
	}
}

// TestModelsTab_NavigationClamps covers the nav edge cases.
func TestModelsTab_NavigationClamps(t *testing.T) {
	cfg := &config.Config{}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	// Up at top — should stay at 0
	st.update(tea.KeyMsg{Type: tea.KeyUp})
	if st.focusedRow != 0 {
		t.Errorf("focusedRow = %d, want 0 after Up at top", st.focusedRow)
	}

	// Navigate to last row
	for i := 0; i < len(st.rows)+5; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	last := len(st.rows) - 1
	if st.focusedRow != last {
		t.Errorf("focusedRow = %d, want %d after Down past end", st.focusedRow, last)
	}

	// Down at bottom — should stay at last
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	if st.focusedRow != last {
		t.Errorf("focusedRow = %d, want %d after Down at bottom", st.focusedRow, last)
	}
}

// TestModelsTab_ProviderNoSDDModelsFallsBackToFreeForm covers S11.
func TestModelsTab_ProviderNoSDDModelsFallsBackToFreeForm(t *testing.T) {
	providers := map[string]opencode.Provider{
		"opencode": prov("opencode",
			noToolModel("some-model", "Some Model"), // no tool_call
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	// opencode always qualifies (builtinProviderID) even with no tool_call models
	// available[0] == "opencode"
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // pick opencode → no SDD models → freeForm
	if st.mode != freeForm {
		t.Errorf("mode = %v, want freeForm when provider has no SDD models", st.mode)
	}
}
