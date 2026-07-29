package tui

import (
	"errors"
	"fmt"
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

// reasoningModel returns a Model with ToolCall=true and Reasoning=true.
func reasoningModel(id, name string) opencode.Model {
	return opencode.Model{ID: id, Name: name, ToolCall: true, Reasoning: true}
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

// E3-5a: picking a Reasoning=true model enters effortSelect mode.
func TestModelsTab_ReasoningModelEntersEffortSelect(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			reasoningModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect (anthropic)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // pick reasoning model → effortSelect

	if st.mode != effortSelect {
		t.Errorf("mode = %v, want effortSelect after picking reasoning model", st.mode)
	}
	if st.effortCursor != 0 {
		t.Errorf("effortCursor = %d, want 0 (reset to top)", st.effortCursor)
	}
}

// E3-5b: choosing "high" in effortSelect sets ref.Effort = "high" and returns to rowNav.
func TestModelsTab_EffortSelectHighSetsEffort(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			reasoningModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → effortSelect (cursor 0 = "default")

	// Navigate to "high" (index 3): Down x3
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	if st.effortCursor != 3 {
		t.Fatalf("effortCursor = %d, want 3 (high)", st.effortCursor)
	}

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm "high"

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after effort selection", st.mode)
	}
	if st.rows[0].ref.Effort != "high" {
		t.Errorf("ref.Effort = %q, want %q", st.rows[0].ref.Effort, "high")
	}
}

// E3-5c: choosing "default" in effortSelect maps to empty Effort.
func TestModelsTab_EffortSelectDefaultMapsToEmpty(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			reasoningModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → effortSelect (cursor 0 = "default")
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm "default"

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav", st.mode)
	}
	if st.rows[0].ref.Effort != "" {
		t.Errorf("ref.Effort = %q, want empty (default maps to empty)", st.rows[0].ref.Effort)
	}
}

// E3-5d: picking a Reasoning=false model skips effortSelect → rowNav, Effort empty.
func TestModelsTab_NonReasoningModelSkipsEffortSelect(t *testing.T) {
	providers := map[string]opencode.Provider{
		"openai": prov("openai",
			toolModel("openai/gpt-4o", "GPT-4o"), // ToolCall=true, Reasoning=false
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect (openai)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // pick non-reasoning model → rowNav directly

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav (effort step skipped for non-reasoning model)", st.mode)
	}
	if st.rows[0].ref.Effort != "" {
		t.Errorf("ref.Effort = %q, want empty for non-reasoning model", st.rows[0].ref.Effort)
	}
}

// E3-5e: Esc from effortSelect returns to modelSelect (model already set, not cleared).
func TestModelsTab_EscFromEffortSelectGoesBackToModelSelect(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			reasoningModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → effortSelect
	if st.mode != effortSelect {
		t.Fatalf("expected effortSelect, got %v", st.mode)
	}

	// The model ref is already set when we entered effortSelect.
	modelAlreadySet := st.rows[0].ref.FullID()
	if modelAlreadySet == "" {
		t.Fatal("model should already be set before Esc from effortSelect")
	}

	st.update(tea.KeyMsg{Type: tea.KeyEsc}) // back to modelSelect

	if st.mode != modelSelect {
		t.Errorf("mode = %v, want modelSelect after Esc from effortSelect", st.mode)
	}
	// Model ref stays set (Esc does not clear it).
	if st.rows[0].ref.FullID() != modelAlreadySet {
		t.Errorf("ref.FullID() = %q after Esc, want %q (model should stay)", st.rows[0].ref.FullID(), modelAlreadySet)
	}
}

// scrollUpText returns true when the rendered view contains the "↑ N more"
// indicator (not the hint line "↑/↓"). We check for the dim style prefix.
func scrollUpText(v string) bool {
	return strings.Contains(v, "↑ ") && !strings.Contains(v, "↑/↓")
}

// E3-6a: provider picker scrolls when there are more providers than maxVisible.
func TestModelsTab_ProviderSelectScrolls(t *testing.T) {
	// Create 15 providers (more than maxVisible at height=24: 24-13=11).
	providers := make(map[string]opencode.Provider)
	for i := 0; i < 15; i++ {
		id := fmt.Sprintf("provider-%02d", i)
		providers[id] = prov(id, toolModel(id+"/model-1", "Model 1"))
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.mode = providerSelect
	st.providerCursor = 0

	// view() computes maxVisible = height - 13 = 24 - 13 = 11.
	view := st.view(80, 24)

	// Should NOT show "↑ N more" at offset 0 (the hint line "↑/↓" is not the indicator).
	if strings.Contains(view, "↑ ") && !strings.Contains(view, "↑/↓") {
		t.Errorf("should not show scroll-up at offset 0; view:\n%s", view)
	}
	// Should show "↓ N more" since 15 > 11.
	if !strings.Contains(view, "↓ 4 more") {
		t.Errorf("should show '↓ 4 more' (15 total - 11 visible); view:\n%s", view)
	}
	// Should show provider-00 (the first one).
	if !strings.Contains(view, "provider-00") {
		t.Errorf("should show the first provider; view:\n%s", view)
	}
}

// E3-6b: scrolling down in provider picker adjusts offset and shows scroll-up indicator.
func TestModelsTab_ProviderSelectScrollDown(t *testing.T) {
	providers := make(map[string]opencode.Provider)
	for i := 0; i < 15; i++ {
		id := fmt.Sprintf("provider-%02d", i)
		providers[id] = prov(id, toolModel(id+"/model-1", "Model 1"))
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 5
	st.mode = providerSelect

	// Navigate far enough that cursor passes the visible window edge.
	for i := 0; i < 6; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if st.providerCursor != 6 {
		t.Fatalf("providerCursor = %d, want 6", st.providerCursor)
	}
	if st.pickerOffset != 2 {
		t.Errorf("pickerOffset = %d, want 2 (cursor 6 - maxVisible 5 + 1)", st.pickerOffset)
	}

	// view() recomputes pickerMaxVisible, so set it back after the call.
	view := st.view(80, 24)
	st.pickerMaxVisible = 5
	if !strings.Contains(view, "↑ 2 more") {
		t.Errorf("should show '↑ 2 more'; view:\n%s", view)
	}

	// Navigate all the way to the bottom.
	for i := 0; i < 10; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if st.providerCursor != 14 {
		t.Fatalf("providerCursor = %d, want 14 (last)", st.providerCursor)
	}
	// maxVisible was restored to 5 after view(), so offset at bottom: 14-5+1=10.
	if st.pickerOffset != 10 {
		t.Errorf("pickerOffset = %d, want 10 (last visible window at bottom)", st.pickerOffset)
	}

	view = st.view(80, 24)
	if !strings.Contains(view, "↑ 10 more") {
		t.Errorf("should show '↑ 10 more' at bottom; view:\n%s", view)
	}
	if !strings.Contains(view, "provider-14") {
		t.Errorf("last provider should be visible; view:\n%s", view)
	}
}

// E3-6c: scrolling up in provider picker adjusts offset back.
func TestModelsTab_ProviderSelectScrollUp(t *testing.T) {
	providers := make(map[string]opencode.Provider)
	for i := 0; i < 15; i++ {
		id := fmt.Sprintf("provider-%02d", i)
		providers[id] = prov(id, toolModel(id+"/model-1", "Model 1"))
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 5
	st.mode = providerSelect
	st.providerCursor = 14
	st.pickerOffset = 10

	// Press Up to go to cursor 13 (still in visible window, no offset change).
	st.update(tea.KeyMsg{Type: tea.KeyUp})
	if st.pickerOffset != 10 {
		t.Errorf("pickerOffset = %d, want 10 (still within window)", st.pickerOffset)
	}

	// 10 Ups from cursor 13 → cursor 3, offset follows cursor when it enters range.
	// cursor: 13→12→11→10→9→8→7→6→5→4→3
	// offset triggers at cursor=9 (9 < 10 → offset=9), then 8,7,6,5,4,3
	for i := 0; i < 10; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyUp})
	}
	if st.providerCursor != 3 {
		t.Fatalf("providerCursor = %d, want 3", st.providerCursor)
	}
	if st.pickerOffset != 3 {
		t.Errorf("pickerOffset = %d, want 3 (cursor at start of window)", st.pickerOffset)
	}

	// Press Up once more → cursor 2, offset should become 2.
	st.update(tea.KeyMsg{Type: tea.KeyUp})
	if st.pickerOffset != 2 {
		t.Errorf("pickerOffset = %d, want 2 (cursor above window)", st.pickerOffset)
	}

	st.pickerMaxVisible = 5
	view := st.view(80, 24)
	if scrollUpText(view) {
		t.Errorf("should not show scroll-up at top; view:\n%s", view)
	}
}

// E3-6d: model picker scrolls when there are many models.
func TestModelsTab_ModelSelectScrolls(t *testing.T) {
	// One provider with 15 models.
	models := make([]opencode.Model, 0, 15)
	for i := 0; i < 15; i++ {
		models = append(models, toolModel(fmt.Sprintf("p/m-%02d", i), fmt.Sprintf("Model %02d", i)))
	}
	providers := map[string]opencode.Provider{
		"test-prov": prov("test-prov", models...),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 5

	// Enter providerSelect → enter modelSelect.
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect (test-prov at 0)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect

	if st.mode != modelSelect {
		t.Fatalf("mode = %v, want modelSelect", st.mode)
	}
	if st.pickerOffset != 0 {
		t.Errorf("pickerOffset = %d, want 0 on entry", st.pickerOffset)
	}

	// Navigate to item 6.
	for i := 0; i < 6; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	if st.modelCursor != 6 {
		t.Fatalf("modelCursor = %d, want 6", st.modelCursor)
	}
	if st.pickerOffset != 2 {
		t.Errorf("pickerOffset = %d, want 2 (cursor 6 - maxVisible 5 + 1)", st.pickerOffset)
	}

	st.pickerMaxVisible = 5
	view := st.view(80, 24)
	if !strings.Contains(view, "↑ 2 more") {
		t.Errorf("should show '↑ 2 more'; view:\n%s", view)
	}
	if !strings.Contains(view, "Model 06") {
		t.Errorf("should show Model 06 (cursor); view:\n%s", view)
	}
}

// E3-6e: model picker resets scroll offset when entering from providerSelect.
func TestModelsTab_ModelSelectResetsOffset(t *testing.T) {
	models := make([]opencode.Model, 0, 10)
	for i := 0; i < 10; i++ {
		models = append(models, toolModel(fmt.Sprintf("p/m-%02d", i), fmt.Sprintf("M%02d", i)))
	}
	providers := map[string]opencode.Provider{
		"p1": prov("p1", models...),
		"p2": prov("p2", toolModel("p2/m", "M")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 3

	// Enter providerSelect, scroll down, then enter modelSelect.
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyDown})  // cursor to p2

	// This should not affect modelSelect's offset since pickerOffset gets reset.
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect

	if st.mode != modelSelect {
		t.Fatalf("mode = %v, want modelSelect", st.mode)
	}
	if st.pickerOffset != 0 {
		t.Errorf("pickerOffset = %d, want 0 after entering modelSelect", st.pickerOffset)
	}
}

// E3-7a: typing filters providers in providerSelect.
func TestModelsTab_ProviderSelectFilter(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic", toolModel("anthropic/claude-opus-4-8", "Claude Opus 4.8")),
		"openai":    prov("openai", toolModel("openai/gpt-4o", "GPT-4o")),
		"opencode":  prov("opencode", toolModel("deepseek-v4-pro", "DeepSeek V4 Pro")),
		"google":    prov("google", toolModel("google/gemini-2", "Gemini 2")),
		"xai":       prov("xai", toolModel("xai/grok-4", "Grok 4")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.mode = providerSelect
	st.pickerMaxVisible = 20 // show all for simplicity

	// Type "open" → should match "openai" and "opencode".
	st.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	st.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	st.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	st.update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if st.filter != "open" {
		t.Fatalf("filter = %q, want %q", st.filter, "open")
	}
	// filteredLen should be 2 (openai, opencode).
	if n := st.filteredLen(); n != 2 {
		t.Fatalf("filteredLen = %d, want 2", n)
	}
	// Cursor should be reset to 0.
	if st.providerCursor != 0 {
		t.Errorf("providerCursor = %d, want 0 after filter", st.providerCursor)
	}

	// First visible filtered item should be "openai" (sorted: openai < opencode).
	idx0 := st.filteredIndices[0]
	if st.available[idx0] != "openai" {
		t.Errorf("first filtered provider = %q, want %q", st.available[idx0], "openai")
	}

	// View should show filter text.
	view := st.view(80, 24)
	if !strings.Contains(view, "filter: open") {
		t.Errorf("view should show filter text; view:\n%s", view)
	}

	// Backspace should remove last character.
	st.update(tea.KeyMsg{Type: tea.KeyBackspace})
	if st.filter != "ope" {
		t.Errorf("filter after backspace = %q, want %q", st.filter, "ope")
	}
	// After "ope": matches "openai" and "opencode" still (both contain "ope").
	if n := st.filteredLen(); n != 2 {
		t.Errorf("filteredLen after backspace = %d, want 2", n)
	}
}

// E3-7b: Esc clears the filter but stays in providerSelect; second Esc exits to rowNav.
func TestModelsTab_ProviderSelectFilterEsc(t *testing.T) {
	providers := map[string]opencode.Provider{
		"openai":   prov("openai", toolModel("openai/gpt-4o", "GPT-4o")),
		"opencode": prov("opencode", toolModel("deepseek-v4-pro", "DeepSeek V4 Pro")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.mode = providerSelect

	// Type a filter.
	st.update(keyMsg('o'))
	if st.filter != "o" {
		t.Fatalf("filter = %q, want %q", st.filter, "o")
	}

	// First Esc → clears filter, stays in providerSelect.
	st.update(tea.KeyMsg{Type: tea.KeyEsc})
	if st.filter != "" {
		t.Errorf("filter should be empty after first Esc, got %q", st.filter)
	}
	if st.mode != providerSelect {
		t.Errorf("mode = %v, want providerSelect after first Esc", st.mode)
	}

	// Second Esc → exits to rowNav.
	st.update(tea.KeyMsg{Type: tea.KeyEsc})
	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after second Esc", st.mode)
	}
}

// E3-7c: typing filters models in modelSelect.
func TestModelsTab_ModelSelectFilter(t *testing.T) {
	providers := map[string]opencode.Provider{
		"test": prov("test",
			toolModel("test/claude-opus", "Claude Opus"),
			toolModel("test/claude-sonnet", "Claude Sonnet"),
			toolModel("test/gpt-4o", "GPT-4o"),
			toolModel("test/gpt-4o-mini", "GPT-4o Mini"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 20

	// Enter providerSelect → modelSelect.
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect (test at 0)
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect

	// Type "claude".
	st.update(keyMsg('c'))
	st.update(keyMsg('l'))
	st.update(keyMsg('a'))
	st.update(keyMsg('u'))
	st.update(keyMsg('d'))
	st.update(keyMsg('e'))

	if st.filter != "claude" {
		t.Fatalf("filter = %q, want %q", st.filter, "claude")
	}
	if n := st.filteredLen(); n != 2 {
		t.Fatalf("filteredLen = %d, want 2 (Claude Opus, Claude Sonnet)", n)
	}

	// First visible filtered item should be "Claude Opus" (sorted by Name).
	idx0 := st.filteredIndices[0]
	if st.curModels[idx0].Name != "Claude Opus" {
		t.Errorf("first filtered model = %q, want %q", st.curModels[idx0].Name, "Claude Opus")
	}

	// Cursor at 0; navigate down and select.
	st.update(tea.KeyMsg{Type: tea.KeyDown})
	if st.modelCursor != 1 {
		t.Fatalf("modelCursor = %d, want 1", st.modelCursor)
	}
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after selection", st.mode)
	}
	// Filter should be cleared after selection.
	if st.filter != "" {
		t.Errorf("filter should be empty after selection, got %q", st.filter)
	}
	// Should have selected "Claude Sonnet" (index 1 in filtered = second Claude model).
	if st.rows[0].ref.FullID() != "test/claude-sonnet" {
		t.Errorf("selected ref = %q, want %q", st.rows[0].ref.FullID(), "test/claude-sonnet")
	}
}

// E3-7d: typing a filter that matches nothing shows "(no matches)".
func TestModelsTab_FilterNoMatches(t *testing.T) {
	providers := map[string]opencode.Provider{
		"openai": prov("openai", toolModel("openai/gpt-4o", "GPT-4o")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.mode = providerSelect

	st.update(keyMsg('z'))
	st.update(keyMsg('z'))
	st.update(keyMsg('z'))

	if n := st.filteredLen(); n != 0 {
		t.Fatalf("filteredLen = %d, want 0 (no matches)", n)
	}

	view := st.view(80, 24)
	if !strings.Contains(view, "(no matches)") {
		t.Errorf("view should show '(no matches)'; view:\n%s", view)
	}

	// Enter should be a no-op (no crash).
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != providerSelect {
		t.Errorf("mode = %v, want providerSelect (no-op on Enter with no matches)", st.mode)
	}
}

// E3-7e: filter is cleared when entering modelSelect from providerSelect.
func TestModelsTab_FilterClearedOnModeTransition(t *testing.T) {
	providers := map[string]opencode.Provider{
		"openai":   prov("openai", toolModel("openai/gpt-4o", "GPT-4o")),
		"opencode": prov("opencode", toolModel("deepseek-v4-pro", "DeepSeek V4 Pro")),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.pickerMaxVisible = 20

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect

	// Type filter "open".
	st.update(keyMsg('o'))
	st.update(keyMsg('p'))
	st.update(keyMsg('e'))
	st.update(keyMsg('n'))
	if st.filter != "open" {
		t.Fatalf("filter = %q, want %q", st.filter, "open")
	}
	if st.filteredLen() != 2 {
		t.Fatalf("filteredLen = %d, want 2", st.filteredLen())
	}

	// Press Enter to select "openai" (cursor at 0).
	st.update(tea.KeyMsg{Type: tea.KeyEnter})
	if st.mode != modelSelect {
		t.Fatalf("mode = %v, want modelSelect", st.mode)
	}
	// Filter should be cleared for modelSelect.
	if st.filter != "" {
		t.Errorf("filter should be empty in modelSelect, got %q", st.filter)
	}
	if st.filteredIndices != nil {
		t.Errorf("filteredIndices should be nil in modelSelect")
	}
}

// B4.1 [REQ-7]: "User can set BaseURL via TUI sub-mode" — pressing 'u' opens
// baseURLEdit, typing a URL and pressing Enter commits it to ref.BaseURL.
func TestModelsTab_SetBaseURLViaSubMode(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"apply": {Provider: "ollama", Model: "llama3"},
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	applyIdx := -1
	for i, r := range st.rows {
		if r.phase == "apply" {
			applyIdx = i
			break
		}
	}
	if applyIdx < 0 {
		t.Fatal("apply row not found")
	}
	for i := 0; i < applyIdx; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	st.update(keyMsg('u'))
	if st.mode != baseURLEdit {
		t.Fatalf("mode = %v, want baseURLEdit after 'u'", st.mode)
	}

	url := "http://localhost:11434/v1"
	for _, r := range url {
		st.update(keyMsg(r))
	}
	st.update(tea.KeyMsg{Type: tea.KeyEnter})

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Enter", st.mode)
	}
	if st.rows[applyIdx].ref.BaseURL != url {
		t.Errorf("ref.BaseURL = %q, want %q", st.rows[applyIdx].ref.BaseURL, url)
	}
	if !st.rows[applyIdx].changed {
		t.Error("rows[applyIdx].changed should be true after setting BaseURL")
	}

	// Saving persists the BaseURL into ModelConfig.
	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	st.applyToConfig(outCfg)
	if outCfg.Models.Phases["apply"].BaseURL != url {
		t.Errorf("saved config BaseURL = %q, want %q", outCfg.Models.Phases["apply"].BaseURL, url)
	}
}

// B4.2 [REQ-7]: "User can clear BaseURL via TUI sub-mode" — clearing the
// input and pressing Enter sets BaseURL back to "".
func TestModelsTab_ClearBaseURLViaSubMode(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	applyIdx := -1
	for i, r := range st.rows {
		if r.phase == "apply" {
			applyIdx = i
			break
		}
	}
	for i := 0; i < applyIdx; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	st.update(keyMsg('u'))
	if st.mode != baseURLEdit {
		t.Fatalf("mode = %v, want baseURLEdit after 'u'", st.mode)
	}
	if st.input.Value() != "http://localhost:11434/v1" {
		t.Errorf("input seeded with %q, want the row's current BaseURL", st.input.Value())
	}

	st.input.SetValue("") // simulate clearing the field
	st.update(tea.KeyMsg{Type: tea.KeyEnter})

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Enter", st.mode)
	}
	if st.rows[applyIdx].ref.BaseURL != "" {
		t.Errorf("ref.BaseURL = %q, want empty after clearing", st.rows[applyIdx].ref.BaseURL)
	}

	// The config YAML for that ref reverts to scalar form (Effort=="" && BaseURL=="").
	outCfg := &config.Config{Models: config.ModelConfig{Phases: map[string]config.ModelRef{}}}
	st.applyToConfig(outCfg)
	ref := outCfg.Models.Phases["apply"]
	if ref.BaseURL != "" {
		t.Errorf("saved ref.BaseURL = %q, want empty", ref.BaseURL)
	}
}

// B4.3 [REQ-7]: "Escape cancels BaseURL edit" — typing a new value then
// pressing Escape leaves the row's BaseURL unchanged.
func TestModelsTab_EscapeCancelsBaseURLEdit(t *testing.T) {
	const original = "http://localhost:11434/v1"
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"apply": {Provider: "ollama", Model: "llama3", BaseURL: original},
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)

	applyIdx := -1
	for i, r := range st.rows {
		if r.phase == "apply" {
			applyIdx = i
			break
		}
	}
	for i := 0; i < applyIdx; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}

	st.update(keyMsg('u'))
	// Clear the seeded value so typing below replaces the original URL rather
	// than appending to it — simulating the real user flow of deleting and
	// retyping a replacement endpoint before cancelling with Escape.
	st.input.SetValue("")
	for _, r := range "http://something-else:9999/v1" {
		st.update(keyMsg(r))
	}
	st.update(tea.KeyMsg{Type: tea.KeyEsc})

	if st.mode != rowNav {
		t.Errorf("mode = %v, want rowNav after Esc", st.mode)
	}
	if st.rows[applyIdx].ref.BaseURL != original {
		t.Errorf("ref.BaseURL = %q after Esc, want unchanged %q", st.rows[applyIdx].ref.BaseURL, original)
	}
	if st.rows[applyIdx].changed {
		t.Error("changed should be false after Esc cancel")
	}
}

// B4.4 [REQ-7]: "Row with BaseURL shows endpoint in display" — a row with a
// BaseURL set renders the URL alongside the provider/model value.
func TestModelsTab_RowWithBaseURLShowsEndpointInDisplay(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)
	st.setWidth(80)

	view := st.view(80, 24)
	if !strings.Contains(view, "http://localhost:11434/v1") {
		t.Errorf("view should show the BaseURL for the apply row; view:\n%s", view)
	}
}

// B4.4b [REQ-7]: "BaseURL-only phase row renders correctly (no model set)" —
// a phase row where Model is empty but BaseURL is set (preserved by
// applyToConfig's guard) must display "(default) @ <url>" in both the
// unfocused and focused states, with no ANSI contamination of focus.Render.
func TestModelsTab_BaseURLOnlyPhaseRowRendersDefaultPlusEndpoint(t *testing.T) {
	cfg := &config.Config{
		Models: config.ModelConfig{
			Phases: map[string]config.ModelRef{
				// No model set, only a BaseURL — the state produced when a user
				// presses 'u' on an empty phase row before picking a model.
				"apply": {Provider: "ollama", BaseURL: "http://localhost:11434/v1"},
			},
		},
	}
	st := newModelsTabState(cfg, map[string]opencode.Provider{}, nil)
	st.setWidth(80)

	// Unfocused: navigate away from the apply row so another row is focused.
	// The default focused row is 0 (Default); apply is at a later index.
	view := st.view(80, 24)
	if !strings.Contains(view, "http://localhost:11434/v1") {
		t.Errorf("unfocused view should contain the BaseURL; view:\n%s", view)
	}
	if !strings.Contains(view, "(default)") {
		t.Errorf("unfocused view should contain the (default) placeholder; view:\n%s", view)
	}

	// Focused: navigate to the apply row and render.
	applyIdx := -1
	for i, r := range st.rows {
		if r.phase == "apply" {
			applyIdx = i
			break
		}
	}
	if applyIdx < 0 {
		t.Fatal("apply row not found")
	}
	for i := 0; i < applyIdx; i++ {
		st.update(tea.KeyMsg{Type: tea.KeyDown})
	}
	focusedView := st.view(80, 24)
	if !strings.Contains(focusedView, "http://localhost:11434/v1") {
		t.Errorf("focused view should contain the BaseURL; view:\n%s", focusedView)
	}
	if !strings.Contains(focusedView, "(default)") {
		t.Errorf("focused view should contain the (default) placeholder; view:\n%s", focusedView)
	}
}

// E3-5f: effortSelect renders a cursor list with "Effort:" header and hint.
func TestModelsTab_EffortSelectView(t *testing.T) {
	providers := map[string]opencode.Provider{
		"anthropic": prov("anthropic",
			reasoningModel("anthropic/claude-opus-4-8", "Claude Opus 4.8"),
		),
	}
	cfg := &config.Config{}
	st := newModelsTabState(cfg, providers, nil)
	st.setWidth(80)

	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → providerSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → modelSelect
	st.update(tea.KeyMsg{Type: tea.KeyEnter}) // → effortSelect

	view := st.view(80, 24)
	for _, opt := range effortOptions {
		if !strings.Contains(view, opt) {
			t.Errorf("view missing effort option %q; view:\n%s", opt, view)
		}
	}
	if !strings.Contains(view, "Effort:") {
		t.Errorf("view missing \"Effort:\" header; view:\n%s", view)
	}
	if !strings.Contains(view, "choose effort") {
		t.Errorf("view missing effort hint; view:\n%s", view)
	}
}
