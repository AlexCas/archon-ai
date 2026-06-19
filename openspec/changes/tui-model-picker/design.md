# Design — tui-model-picker (Slice 3b)

Apply-ready design for the in-tab, per-row provider→model picker that replaces the free-form
`textinput` grid in `internal/tui/models_tab.go`. Mirrors `agent_tab.go`'s hand-rolled cursor
pattern (no `bubbles/list`). Implements the five settled decisions verbatim: in-tab per-row
sub-mode picker, leader uses the picker, free-form always available, legacy preservation, inline
corrupt-cache warning.

Scope of code touched:
- `internal/tui/models_tab.go` — full rewrite (state, construction, update, view, applyToConfig, setWidth).
- `internal/tui/model.go` — catalog wiring at lines 88/93, 182 (feed the provider map, not `models.Resolve()`).
- `internal/tui/models_tab_test.go` — full rewrite.
- `internal/tui/model_test.go` — update the tests that poke `modelsTab` internals (listed in §7).

Out of scope (UNCHANGED): the save path (`saveConfig`→`applyToConfig`→`Save`→`regenerateTemplate`→
`MergeOpencodeAgent`), 3a helpers, the cache reader, `config.ModelRef`/`ParseModelRef`/`FullID`,
`internal/models/resolve.go` (left intact; see §6).

---

## 1. New `modelsTabState` shape

The tab is a list of **rows**. Each row owns its current `ModelRef`, a `changed` flag (drives legacy
preservation), and a reusable `textinput` for free-form. A single tab-level sub-mode enum plus three
cursor indices drive the state machine. The provider catalog is loaded once at construction.

```go
package tui

import (
	"sort"
	"strings"

	"github.com/archon-ai/archon/internal/config"
	"github.com/archon-ai/archon/internal/opencode"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// subMode is the per-tab interaction mode. Only one row can be "open" (in
// providerSelect/modelSelect/freeForm) at a time; that row is m.focusedRow.
type subMode int

const (
	rowNav         subMode = iota // navigating between rows (Up/Down/Tab)
	providerSelect                // choosing a provider for the focused row
	modelSelect                   // choosing a model within the picked provider
	freeForm                      // typing a raw provider/model string
)

// modelRow is one editable assignment line: Default, one per SDD phase, and
// (opencode only) Leader. ref holds the live value seeded from cfg; changed is
// set the moment the user picks/clears/free-forms this row, and gates legacy
// preservation in applyToConfig.
type modelRow struct {
	label   string         // "Default" / phase name / "Leader"
	kind    rowKind        // rowDefault | rowPhase | rowLeader
	phase   string         // phase name when kind==rowPhase, else ""
	ref     config.ModelRef
	changed bool
}

type rowKind int

const (
	rowDefault rowKind = iota
	rowPhase
	rowLeader
)

// modelsTabState is the rewritten Models tab. The leader row, when present, is
// always the final element of rows (kind==rowLeader); leaderEnabled mirrors the
// old behavior (opencode only).
type modelsTabState struct {
	rows          []modelRow
	focusedRow    int
	mode          subMode

	// Provider catalog loaded once at construction from the opencode cache.
	providers map[string]opencode.Provider // provider id -> Provider
	available []string                     // sorted provider ids (DetectAvailableProviders)

	// Picker cursors (valid only while mode != rowNav).
	providerCursor int          // index into m.available
	modelCursor    int          // index into the active provider's FilterModelsForSDD list
	pickedProvider string       // provider chosen in providerSelect, used by modelSelect
	curModels      []opencode.Model // FilterModelsForSDD(providers[pickedProvider]), cached for modelSelect

	// Free-form text entry (shared; reset on each open).
	input textinput.Model

	// cacheWarning is non-empty ONLY when LoadModelsOrEmpty returned an error
	// (cache present but unreadable). Absent cache => "" (silent). Rendered inline.
	cacheWarning string

	leaderEnabled bool
	width         int
}
```

Notes:
- The leader row "folds in" as a normal `modelRow{kind: rowLeader}` appended to `rows` only when
  `cfg.Agent == "opencode"`. The picker treats it identically to any other row; the only difference is
  it is gated behind `leaderEnabled` (matching the old `leaderInputIndex` gate) and `applyToConfig`
  writes it to `cfg.Models.Leader` instead of `Phases`.
- We do NOT keep one `textinput` per row anymore (no auto-fill grid). One shared `input` is reused for
  whichever row is in `freeForm`. This removes `inputs []textinput.Model`, `focusedInput`,
  `autoFillLocks`, `updateAutoFill`, `cycleStaticModel`, `phaseNames`, `catalog []string`.
- `curModels`/`pickedProvider` are recomputed when entering `modelSelect` (deterministic via
  `FilterModelsForSDD`, already sorted by Name in 3a).

---

## 2. Construction (`newModelsTabState` replacement)

The parent now passes the **provider map** (loaded once when the TUI opens) plus a flag distinguishing
absent vs corrupt cache. Two viable signatures; **chosen**: pass the map and the load error so the tab
owns the warning copy and stays the single source of truth.

```go
// newModelsTabState builds the Models tab from cfg and the opencode provider
// catalog. providers is the map from opencode.LoadModelsOrEmpty (empty when the
// cache is absent); loadErr is the error it returned (non-nil only when the
// cache is present but unreadable — the corrupt-cache seam from 3a). The tab
// seeds each row's ModelRef from cfg.Models and never re-reads the cache.
func newModelsTabState(cfg *config.Config, providers map[string]opencode.Provider, loadErr error) modelsTabState {
	leaderEnabled := cfg.Agent == "opencode"

	rows := make([]modelRow, 0, 1+len(config.PhaseOrder)+1)
	rows = append(rows, modelRow{label: "Default", kind: rowDefault, ref: cfg.Models.Default})
	for _, phase := range config.PhaseOrder { // deterministic order, 8 phases
		var ref config.ModelRef
		if cfg.Models.Phases != nil {
			ref = cfg.Models.Phases[phase] // zero ModelRef when unset
		}
		rows = append(rows, modelRow{label: phase, kind: rowPhase, phase: phase, ref: ref})
	}
	if leaderEnabled {
		rows = append(rows, modelRow{label: "Leader", kind: rowLeader, ref: cfg.Models.Leader})
	}

	ti := textinput.New()
	ti.Placeholder = "provider/model"
	ti.Width = 30

	st := modelsTabState{
		rows:          rows,
		focusedRow:    0,
		mode:          rowNav,
		providers:     providers,
		available:     opencode.DetectAvailableProviders(providers), // sorted, may be empty
		input:         ti,
		leaderEnabled: leaderEnabled,
	}
	if loadErr != nil {
		st.cacheWarning = "⚠ opencode model cache is unreadable — pick is unavailable; type a model name (e)"
	}
	return st
}
```

Construction details:
- Rows are seeded from `cfg.Models.{Default,Phases[phase],Leader}` so an untouched row carries the exact
  loaded `ModelRef` (including a legacy bare alias `{Provider:"", Model:"opus"}`). `changed` defaults
  `false`.
- `available` may be empty (absent cache, or no qualifying providers) — that is the free-form-only path.
- `cacheWarning` is set ONLY when `loadErr != nil`. Absent cache → `providers` empty, `loadErr` nil →
  no warning (settled decision 5).

Caller wiring (see §6): `model.go:88` loads the map + error once; `model.go:93` and `model.go:182`
pass them to `newModelsTabState`.

---

## 3. `update()` — key state machine per sub-mode

Signature unchanged: `func (m *modelsTabState) update(msg tea.Msg) (tea.Cmd, bool)`. The parent
(`model.go:140-144`) routes Models-tab keys here AFTER global keys (`tab`/`shift+tab`/`ctrl+s`/`ctrl+q`/`q`)
are matched in `Model.Update`. So global keys never reach this method — but in `rowNav` any key this
method does not consume is simply ignored (returns `nil, true`), and tab/save/quit are already handled
upstream. The return `bool` (handled) is currently ignored by the parent; keep returning `true`.

Free-form trigger key: **`e`** (edit). Verified non-colliding: parent globals are
`tab/shift+tab/ctrl+s/ctrl+q/q`; this tab no longer uses `ctrl+n/ctrl+p`. `e` only fires in `rowNav`;
inside `freeForm` it is a literal character routed to the textinput.

```go
func (m *modelsTabState) update(msg tea.Msg) (tea.Cmd, bool) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		// Non-key msgs only matter to the textinput while in freeForm.
		if m.mode == freeForm {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return cmd, true
		}
		return nil, true
	}

	switch m.mode {
	case rowNav:
		return m.updateRowNav(key)
	case providerSelect:
		return m.updateProviderSelect(key)
	case modelSelect:
		return m.updateModelSelect(key)
	case freeForm:
		return m.updateFreeForm(key)
	}
	return nil, true
}
```

### 3a. `rowNav`

| Key            | Action |
|----------------|--------|
| Up             | `focusedRow = max(0, focusedRow-1)` |
| Down / Tab     | `focusedRow = min(len(rows)-1, focusedRow+1)` |
| Enter          | open picker: if `len(m.available)==0` → open free-form on this row instead (no providers); else `mode=providerSelect`, `providerCursor=0` |
| `e`            | open free-form: `mode=freeForm`, seed `m.input` with `m.rows[focusedRow].ref.FullID()`, `m.input.Focus()` |
| (anything else)| `return nil, true` (ignored; globals handled by parent) |

Note: do NOT wrap Up/Down (clamp), matching `agent_tab`'s bounded cursor. Tab here advances rows; the
parent's global Tab switches tabs and is matched first, so row-internal Tab is only reachable if the
parent's keymap misses — it won't for `tab`. To stay safe and match agent_tab semantics, bind row
advance to **Down only** for the cursor and let Tab fall through to the parent. (Implementer choice
below — see Risk R1.) Chosen: bind Down only to advance; do not consume Tab. This keeps tab-to-switch
working and avoids ambiguity.

```go
func (m *modelsTabState) updateRowNav(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.focusedRow > 0 {
			m.focusedRow--
		}
		return nil, true
	case tea.KeyDown:
		if m.focusedRow < len(m.rows)-1 {
			m.focusedRow++
		}
		return nil, true
	case tea.KeyEnter:
		if len(m.available) == 0 {
			m.openFreeForm()
			return nil, true
		}
		m.mode = providerSelect
		m.providerCursor = 0
		return nil, true
	case tea.KeyRunes:
		if len(key.Runes) == 1 && key.Runes[0] == 'e' {
			m.openFreeForm()
			return nil, true
		}
	}
	return nil, true
}

func (m *modelsTabState) openFreeForm() {
	m.mode = freeForm
	m.input.SetValue(m.rows[m.focusedRow].ref.FullID())
	m.input.CursorEnd()
	m.input.Focus()
}
```

### 3b. `providerSelect`

| Key   | Action |
|-------|--------|
| Up    | `providerCursor = max(0, providerCursor-1)` |
| Down  | `providerCursor = min(len(available)-1, providerCursor+1)` |
| Enter | `pickedProvider = available[providerCursor]`; `curModels = FilterModelsForSDD(providers[pickedProvider])`; if `len(curModels)==0` → fall back to free-form (provider has no SDD models); else `mode=modelSelect`, `modelCursor=0` |
| Esc   | `mode=rowNav` (cancel) |

```go
func (m *modelsTabState) updateProviderSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.providerCursor > 0 {
			m.providerCursor--
		}
	case tea.KeyDown:
		if m.providerCursor < len(m.available)-1 {
			m.providerCursor++
		}
	case tea.KeyEnter:
		m.pickedProvider = m.available[m.providerCursor]
		m.curModels = opencode.FilterModelsForSDD(m.providers[m.pickedProvider])
		if len(m.curModels) == 0 {
			m.openFreeForm() // provider qualified (built-in) but has no SDD models
			return nil, true
		}
		m.mode = modelSelect
		m.modelCursor = 0
	case tea.KeyEsc:
		m.mode = rowNav
	}
	return nil, true
}
```

### 3c. `modelSelect`

| Key   | Action |
|-------|--------|
| Up    | `modelCursor = max(0, modelCursor-1)` |
| Down  | `modelCursor = min(len(curModels)-1, modelCursor+1)` |
| Enter | set the row's `ModelRef` from `pickedProvider` + selected model id; `changed=true`; `mode=rowNav` |
| Esc   | `mode=providerSelect` (back one step) |

The provider→ModelRef mapping handles the bare-vs-slashed cache key asymmetry (settled edge case):

```go
func (m *modelsTabState) updateModelSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.modelCursor > 0 {
			m.modelCursor--
		}
	case tea.KeyDown:
		if m.modelCursor < len(m.curModels)-1 {
			m.modelCursor++
		}
	case tea.KeyEnter:
		m.rows[m.focusedRow].ref = refFromCacheKey(m.pickedProvider, m.curModels[m.modelCursor].ID)
		m.rows[m.focusedRow].changed = true
		m.mode = rowNav
	case tea.KeyEsc:
		m.mode = providerSelect
	}
	return nil, true
}

// refFromCacheKey builds a ModelRef from a provider id and a cache model key.
// Under the "opencode" provider keys are BARE (e.g. "deepseek-v4-pro") so we set
// Provider+Model and FullID yields "opencode/deepseek-v4-pro". Under other
// providers the key is ALREADY-SLASHED (e.g. "xai/grok-4"); putting it in Model
// alone lets FullID's "/" short-circuit return it as-is (no double-prefix).
func refFromCacheKey(provider, key string) config.ModelRef {
	if strings.Contains(key, "/") {
		return config.ModelRef{Model: key} // already provider-qualified; FullID returns as-is
	}
	return config.ModelRef{Provider: provider, Model: key}
}
```

### 3d. `freeForm`

| Key   | Action |
|-------|--------|
| Enter | `m.rows[focusedRow].ref = ParseModelRef(strings.TrimSpace(m.input.Value()))`; `changed=true`; `m.input.Blur()`; `mode=rowNav` |
| Esc   | cancel: `m.input.Blur()`; `mode=rowNav` (row ref unchanged, `changed` untouched) |
| other | forward to `m.input.Update(msg)` (typing) |

```go
func (m *modelsTabState) updateFreeForm(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyEnter:
		m.rows[m.focusedRow].ref = config.ParseModelRef(strings.TrimSpace(m.input.Value()))
		m.rows[m.focusedRow].changed = true
		m.input.Blur()
		m.mode = rowNav
		return nil, true
	case tea.KeyEsc:
		m.input.Blur()
		m.mode = rowNav
		return nil, true
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	return cmd, true
}
```

Empty free-form on Enter → `ParseModelRef("")` = `ModelRef{}` (cleared), `changed=true`. `applyToConfig`
treats a changed+empty phase row as a delete (§5).

---

## 4. `view()`

Signature unchanged: `func (m *modelsTabState) view(width, height int) string`. Deterministic: rows in
fixed order, provider list sorted (`available`), model list sorted by Name (`FilterModelsForSDD`).

Layout:

```
Model Configuration
↑/↓: move · Enter: pick provider/model · e: type a model

⚠ opencode model cache is unreadable — ...        (ONLY when cacheWarning != "")

  Default:   anthropic/claude-opus-4-8
▸ explore:   opencode/deepseek-v4-pro
  propose:   (default)
  ...
  archive:   haiku

Leader:      openai/gpt-4o                         (opencode only)
```

Rendering rules per row, by mode:

- **rowNav (this row not focused)**: `label` + the current value `row.ref.FullID()`. If the ref is empty
  AND it's a phase row, render `(default)` in dim style (the phase falls back to Default at resolve time).
  The focused row gets the agent_tab focus style (`Background 63 / Foreground 0`) and a `▸` marker.
- **providerSelect (only on the focused row)**: render the row label, then an indented cursor list of
  `m.available` with the agent_tab focus style on `providerCursor`. Header line `Provider:`.
- **modelSelect (only on the focused row)**: render `Provider: <pickedProvider>`, then an indented cursor
  list of `m.curModels` (display `Name` or `ID`; use `m.curModels[i].Name` falling back to `ID` if Name
  empty) with focus style on `modelCursor`. Header `Model:`.
- **freeForm (only on the focused row)**: render `m.input.View()` inline after the label, plus a dim hint
  `Enter to set · Esc to cancel`.

Other rows always render in their plain rowNav form while one row is open (only one row can be in a
non-rowNav mode).

Help/hint line (dim, fixed):
- rowNav: `↑/↓: move · Enter: pick · e: type a model · (no providers → e only when cache empty)`
- providerSelect: `↑/↓: choose provider · Enter: next · Esc: cancel`
- modelSelect: `↑/↓: choose model · Enter: set · Esc: back`
- freeForm: `type provider/model · Enter: set · Esc: cancel`

Empty-providers note: when `len(m.available)==0`, the rowNav hint reads
`No detected models — press e to type a provider/model` and Enter opens free-form (§3a). This is the
absent-cache path.

The old per-row `config.Validate` advisory warning and the leader `/`-guard warning are DROPPED from the
picker rows (picked values come from the catalog and are inherently valid; free-form values are the user's
explicit escape hatch). KEEP a single advisory only on the leader row in free-form/legacy display IF
desired — but to keep this slice lean and deterministic, drop it. (This changes `TestModelsTab_LeaderWarningGuard`
expectations — see §7.) Styling reuse: copy the `labelStyle` (Width 15, right align) and the focus
`Background(63)/Foreground(0)` style from the current files; no new color constants.

```go
func (m *modelsTabState) view(width, height int) string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).MarginBottom(1)
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	label := lipgloss.NewStyle().Width(15).Align(lipgloss.Right).MarginRight(1)
	focus := lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("63")).Foreground(lipgloss.Color("0"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString(title.Render("Model Configuration"))
	b.WriteString("\n")
	b.WriteString(hint.Render(m.hintLine()))
	b.WriteString("\n")
	if m.cacheWarning != "" {
		warn := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		b.WriteString(warn.Render(m.cacheWarning))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	for i := range m.rows {
		b.WriteString(m.renderRow(i, label, focus, dim))
		b.WriteString("\n")
	}
	return b.String()
}
```

`renderRow(i, ...)` switches on `m.mode` when `i == m.focusedRow` (drawing the inline picker/free-form
list), else draws the plain `label + ref.FullID()` (or `(default)` for an empty phase row), with the
focus style applied when `i == m.focusedRow && m.mode == rowNav`.

---

## 5. `applyToConfig()` — legacy preservation

Signature unchanged: `func (m *modelsTabState) applyToConfig(cfg *config.Config)`. Rule: a row the user
did NOT change (`changed == false`) is written back **verbatim from its seeded `ref`** (which IS the
loaded value), preserving legacy bare aliases. A changed phase row that is now empty is deleted; a changed
default/leader that is empty zeroes the ref (matching today's behavior for cleared fields).

Because rows are seeded from `cfg` and we write every row back, an untouched legacy ref round-trips
identically. The `changed` flag exists only to make the intent explicit and to drive the empty-phase
delete-vs-keep decision: an untouched empty phase row must NOT create an empty entry.

```go
func (m *modelsTabState) applyToConfig(cfg *config.Config) {
	if cfg.Models.Phases == nil {
		cfg.Models.Phases = make(map[string]config.ModelRef)
	}
	for _, row := range m.rows {
		switch row.kind {
		case rowDefault:
			cfg.Models.Default = row.ref // seeded verbatim when unchanged
		case rowLeader:
			cfg.Models.Leader = row.ref
		case rowPhase:
			if row.ref.Model == "" {
				// Empty: delete the phase entry (whether seeded-empty or cleared).
				delete(cfg.Models.Phases, row.phase)
			} else {
				cfg.Models.Phases[row.phase] = row.ref
			}
		}
	}
}
```

Legacy preservation proof: a row seeded with `{Provider:"", Model:"opus"}` and never touched writes
`cfg.Models.Default = {Provider:"", Model:"opus"}` — byte-identical on `Save` (scalar marshal, per
foundation `MarshalYAML`). A row seeded with `{Provider:"anthropic", Model:"claude-opus-4-8"}` likewise
round-trips. We never run a seeded value through `ParseModelRef` (which would be a no-op here anyway, but
the rule is: don't re-parse, just keep the ref).

Leader: only present in `rows` when `leaderEnabled`; for non-opencode agents there is no leader row, so
`cfg.Models.Leader` is left exactly as loaded (matching the old `leaderInputIndex() < 0` skip).

---

## 6. Catalog wiring in `model.go`

Replace the flat `models.Resolve()` feed with a single cache load that yields the provider map AND the
load error (corrupt-vs-absent seam). The `Model` struct field `modelsTab.catalog` (referenced at
`model.go:182` for reuse) is removed; cache the provider map + error on the parent instead.

**`NewModel` (lines 84-99):**

```go
func NewModel(cfg *config.Config, projectDir string) Model {
	// Load the opencode provider catalog once when the TUI opens. Absent cache =>
	// empty map + nil err (silent); corrupt cache => err (drives an inline warning).
	var providers map[string]opencode.Provider
	var cacheErr error
	if path, err := opencode.DefaultCachePath(); err == nil {
		providers, cacheErr = opencode.LoadModelsOrEmpty(path)
	} else {
		cacheErr = err
	}
	return Model{
		config:        cfg,
		projectDir:    projectDir,
		activeTab:     ModelsTab,
		modelsTab:     newModelsTabState(cfg, providers, cacheErr),
		// ... unchanged ...
	}
}
```

Add `providers`/`cacheErr` fields to `Model` (or recompute is fine since it's only read here +
`agentInitDoneMsg`), then at the `agentInitDoneMsg` rebuild (**line 182**):

```go
m.modelsTab = newModelsTabState(msg.cfg, m.providers, m.cacheErr)
```

So store the loaded map + err on `Model` (two new fields: `providers map[string]opencode.Provider`,
`cacheErr error`) to reuse without re-reading (matches the "init does not re-run detection" comment).

Imports: add `"github.com/archon-ai/archon/internal/opencode"` to `model.go`; the `"github.com/archon-ai/archon/internal/models"`
import becomes unused IN model.go (only `models.Resolve()` used it) — **remove it from model.go**.

**Save path (lines 311-355): UNCHANGED.** `saveConfig` calls `applyToConfig`→`Save`→`regenerateTemplate`
(uses `ResolvePhaseModels` FullIDs) →`MergeOpencodeAgent`. None of these care how the tab gathered its
refs; they read `cfg.Models`. Confirmed no edits needed there.

**`internal/models/resolve.go` (`Resolve`/`cacheModelNames`/`ResolveModels`): LEAVE AS-IS.** After this
slice the TUI no longer calls `models.Resolve()`, but it is still part of the public package and exercised
by `resolve_test.go`; it may be used elsewhere (status/CLI). Out of scope to delete (would look like
reverting prior work). Note for archive: `models.Resolve()` becomes unused by the TUI specifically; flag
as a possible future cleanup, do NOT remove here.

---

## 7. Test plan

### Rewrite `internal/tui/models_tab_test.go`

Drive `update()` with synthetic `tea.KeyMsg` through each sub-mode; assert `rows[i].ref` and `view()`
substrings. Construct with `newModelsTabState(cfg, providers, loadErr)` where `providers` is a hand-built
`map[string]opencode.Provider` (no real cache). Helper to build a provider:

```go
func prov(id string, models ...opencode.Model) opencode.Provider { /* id + map keyed by m.ID */ }
func toolModel(id, name string) opencode.Model { return opencode.Model{ID: id, Name: name, ToolCall: true} }
```

| Test | Drives | Asserts | Spec scenario |
|------|--------|---------|---------------|
| `TestModelsTab_PickProviderModelSetsRef` | focus row, Enter, Down, Enter, Down, Enter | `rows[1].ref == {Provider:"anthropic", Model:"claude-..."}`, `changed==true`, `mode==rowNav` | S1 |
| `TestModelsTab_OpencodeBareKeyMapsToRef` | pick `opencode` provider + bare-key model | `ref == {Provider:"opencode", Model:"deepseek-v4-pro"}`, `FullID()=="opencode/deepseek-v4-pro"` | S2 |
| `TestModelsTab_SlashedKeyNoDoublePrefix` | pick a provider whose model key already contains "/" | `ref.FullID()` equals the key verbatim (no `provider/provider/`) | S2 |
| `TestModelsTab_EscCancelsPicker` | Enter→providerSelect, Esc | `mode==rowNav`, `rows[i].ref` unchanged, `changed==false` | S3 |
| `TestModelsTab_EscFromModelSelectGoesBack` | into modelSelect, Esc | `mode==providerSelect` | S3 |
| `TestModelsTab_FreeFormTogglesAndParses` | `e`, type "x/y", Enter | `ref=={Provider:"x",Model:"y"}`, `changed==true`, `mode==rowNav` | S4 |
| `TestModelsTab_FreeFormEscCancels` | `e`, type, Esc | `ref` unchanged, `mode==rowNav` | S4 |
| `TestModelsTab_LegacyUntouchedPreserved` | seed Default `{Model:"opus"}`, navigate only, `applyToConfig` | `cfg.Models.Default=={Model:"opus"}` (Provider=="") | S5 |
| `TestModelsTab_ApplyDeletesClearedPhase` | seed phase, `e`, clear, Enter, apply | phase key absent from `cfg.Models.Phases` | S6 |
| `TestModelsTab_CorruptCacheWarningShown` | construct with `loadErr != nil` | `view()` contains the warning substring | S7 |
| `TestModelsTab_AbsentCacheFreeFormOnly` | empty providers, nil err; Enter on a row | `mode==freeForm` (no provider list), no warning in `view()` | S8 |
| `TestModelsTab_LeaderUsesPicker` | `cfg.Agent=="opencode"`; last row kind==rowLeader; pick → `cfg.Models.Leader` set | leader row present + picker works; non-opencode → no leader row | S9 |
| `TestModelsTab_DeterministicSortedLists` | two providers / multiple models; assert `available` and rendered list order | sorted by id / Name | S10 |
| `TestModelsTab_NavigationClamps` | Up at top, Down at bottom | `focusedRow` clamps at `[0, len-1]` | (nav) |

### Update `internal/tui/model_test.go`

These tests poke removed internals (`inputs`, `modelInputDefault`, `modelInputExplore`, `cycleStaticModel`,
`updateAutoFill`, `autoFillLocks`, `catalog`) and the dropped advisory warning. Rewrite each against the
new `rows` API:

- **`TestNewModel`** (line 43-44): replace `inputs[modelInputDefault].Value()=="gpt-4"` with
  `modelsTab.rows[0].ref.FullID()=="gpt-4"` (seed `cfg.Models.Default={Model:"gpt-4"}`).
- **`TestModelsTabState_AutoFill`** (250): DELETE — auto-fill is removed (no placeholder grid).
- **`TestModelsTabState_LockOnEdit`** (274): DELETE — locks removed.
- **`TestModelsTabState_ApplyToConfig`** (302): rewrite to set rows via the picker/free-form path then
  assert `cfg.Models.Default/Phases` (cover the empty-phase delete).
- **`TestSaveConfig_FailureDoesNotMutateInMemoryConfig`** (533, line 547): replace
  `inputs[modelInputDefault].SetValue("gpt-4")` with a free-form set on `rows[0]` (e.g. enter freeForm,
  set value, Enter) OR directly `m.modelsTab.rows[0].ref = config.ParseModelRef("gpt-4"); rows[0].changed=true`.
  The save-failure assertion is unchanged.
- **`TestModelsTab_LeaderWarningGuard`** (719): the picker drops the per-row advisory warning. Rewrite to
  assert the leader ROW renders for opencode (`view` contains "Leader") and is ABSENT for non-opencode,
  rather than asserting the `⚠` guard. (Decision D2 — confirm at review.)
- Any test calling `newModelsTabState(cfg, config.StaticModels())` must switch to the new 3-arg signature
  `newModelsTabState(cfg, providers, nil)` (pass `map[string]opencode.Provider{}` for the absent-cache case).

Follow `go-testing` conventions: direct `update()` with synthetic `tea.KeyMsg`, assert state +
`view()` substrings; no teatest needed (state machine is synchronous).

---

## 8. Determinism + edge cases

1. **Empty available providers (absent cache)**: `available == []`. rowNav Enter opens free-form directly;
   the hint says "No detected models — press e to type a provider/model". No warning (decision 5).
2. **Corrupt cache**: `loadErr != nil`; `cacheWarning` rendered inline (yellow `214`), never stderr. Picker
   still works only if `available` non-empty (LoadModelsOrEmpty returns nil map on corrupt → empty
   available → free-form path). So corrupt cache effectively = free-form + a warning.
3. **Legacy ref not in the catalog** (e.g. `{Model:"opus"}` or a provider/model no longer cached): displayed
   verbatim via `FullID()`; kept unless the user re-picks/free-forms (preservation §5). The picker does not
   try to pre-select it.
4. **opencode bare key vs already-slashed key**: handled by `refFromCacheKey` (§3c) — bare under the
   built-in `opencode` provider → `{Provider, Model}`; a key containing "/" → `{Model: key}` so `FullID`'s
   "/" short-circuit avoids double-prefix. Pinned by `TestModelsTab_SlashedKeyNoDoublePrefix`.
5. **Leader row only when `agent==opencode`**: `leaderEnabled` gates appending the `rowLeader` row;
   non-opencode never shows/edits it and `applyToConfig` leaves `cfg.Models.Leader` untouched.
6. **Determinism**: `available` from `DetectAvailableProviders` (sorted ids), model lists from
   `FilterModelsForSDD` (sorted by Name); rows in fixed `[Default, PhaseOrder..., Leader]` order. `view()`
   has no map iteration. Reproducible output for golden/substring assertions.
7. **Single open row invariant**: only `m.focusedRow` can be in a non-rowNav mode; entering a picker/free-form
   does not change `focusedRow`, and all transitions return to `rowNav` before another row can open.
8. **Empty free-form confirm**: `ParseModelRef("")` = zero ref; phase row → delete on apply; default/leader →
   zeroed (cleared). Marked `changed=true` so the delete intent is honored.

---

## 9. Recommended spec scenarios (for the spec phase)

S1. **Pick provider then model sets the ModelRef** — Given the Models tab with a non-empty provider
catalog, When the user focuses a row, presses Enter, selects a provider, presses Enter, selects a model,
and presses Enter, Then the row's ModelRef is `{Provider, Model}` for the picked pair and the tab returns
to row navigation.

S2. **opencode bare key and already-slashed key map without double-prefix** — Given the built-in
`opencode` provider with a bare model key and another provider with an already-slashed key, When each is
picked, Then the opencode ref is `{Provider:"opencode", Model:<key>}` (FullID `opencode/<key>`) and the
slashed-key ref's FullID equals the key verbatim (no `provider/provider/`).

S3. **Esc cancels the picker without changing the row** — Given a row in provider or model selection, When
the user presses Esc, Then model selection returns to provider selection (one step back) and provider
selection returns to row navigation, and in both cases the row's ModelRef and changed flag are unchanged.

S4. **Free-form toggles on any row and parses on confirm** — Given any focused row, When the user presses
`e`, types a `provider/model` string, and presses Enter, Then the row's ModelRef equals
`ParseModelRef(value)` and changed is true; When the user presses Esc instead, Then the row is unchanged.

S5. **Untouched legacy ModelRef is preserved verbatim** — Given a row seeded from config with a legacy bare
alias (`Provider==""`), When the user navigates without re-picking or editing that row and saves, Then
`applyToConfig` writes that row's ModelRef byte-identically (provider stays empty; never guessed).

S6. **Clearing a phase row deletes the phase entry** — Given a phase row with a value, When the user
free-forms it to empty and confirms and saves, Then the phase key is absent from `cfg.Models.Phases`.

S7. **Corrupt cache shows an inline warning, never stderr** — Given the opencode cache is present but
unreadable (LoadModelsOrEmpty returns an error), When the Models tab renders, Then an inline warning is
shown in the view and nothing is written to stderr.

S8. **Absent cache yields free-form only with no warning** — Given no opencode cache (empty provider map,
no error), When the user presses Enter on a row, Then free-form entry opens directly (no provider list) and
no warning is shown.

S9. **Leader row uses the picker only for opencode** — Given `agent == "opencode"`, When the tab renders,
Then a Leader row is present and supports the same provider→model picker and free-form; Given a non-opencode
agent, Then no Leader row is shown and `cfg.Models.Leader` is left as loaded on save.

S10. **Provider and model lists are deterministic and sorted** — Given multiple providers and models, When
the picker renders, Then providers appear in sorted id order and models in sorted Name order, identically
across runs.

S11. **A provider with no SDD models falls back to free-form** — Given a qualifying provider with no
tool_call models, When the user selects it, Then free-form entry opens (no empty model list).

S12. **Navigation clamps at the row bounds** — Given the first/last row focused, When the user presses
Up/Down past the edge, Then `focusedRow` stays within `[0, len(rows)-1]`.

---

## 10. Size estimate

| Area | LOC (approx) |
|------|--------------|
| `models_tab.go` rewrite (state + construct + update state machine + view + apply + helpers) | ~210 prod |
| `model.go` wiring (NewModel load, two fields, rebuild line, import swap) | ~20 prod |
| `models_tab_test.go` rewrite (~13 tests + helpers) | ~170 test |
| `model_test.go` edits (delete 2, rewrite 4, signature fixes) | ~40 net test |
| **Total** | **~230 prod + ~210 test ≈ 440** |

This is at/just over the **D1 400-line** budget — consistent with the proposal's ~300-400 forecast trending
to the high end. **Flag at PR (C1)**: the overage is dominated by the test rewrite (mechanical), prod is
~230. Mitigations if a tighter PR is required: (a) keep the legacy `config.Validate` advisory drop (already
planned) to avoid carrying warning code, (b) the `model_test.go` deletes (AutoFill/LockOnEdit) reduce churn.
No further splitting is sensible — the state machine, view, and tests are one cohesive unit. Recommend
shipping as one PR with the size flagged in the body per C1.

---

## Design decisions made (confirm at review)

- **D1 — Free-form trigger key = `e`.** Non-colliding with parent globals (`tab/shift+tab/ctrl+s/ctrl+q/q`)
  and the removed `ctrl+n/ctrl+p`. Alternative `/` or `i` also work; `e` (edit) reads clearest.
- **D2 — Drop the per-row `config.Validate` advisory + the leader `/`-guard warning.** Picked values are
  catalog-valid; free-form is the explicit escape hatch. This simplifies `view()` and removes the
  `TestModelsTab_LeaderWarningGuard` `⚠` expectation (rewritten to assert leader-row presence). If you want
  to keep a free-form advisory, it adds ~15 LOC + a test.
- **D3 — Row advance bound to Down only (not Tab) in `rowNav`.** Keeps the parent's global Tab = switch-tab
  working unambiguously and mirrors agent_tab (which uses Up/Down only). The proposal mentioned "Tab between
  rows"; Down is the safe equivalent.
- **D4 — One shared free-form `textinput`** (not one per row) since auto-fill is removed. Reduces state and
  matches the single-open-row invariant.
- **D5 — `newModelsTabState` takes `(cfg, providers map, loadErr error)`** so the tab owns the warning copy
  and the absent-vs-corrupt decision lives in one place.

## Risks to confirm

- **R1 — Tab-vs-Down for row navigation (D3).** If you specifically want in-tab Tab to advance rows, we must
  intercept it before the parent — that means parent routing changes (out of this slice's stated scope).
  Recommend Down-only.
- **R2 — Size at/over D1 400.** Flag at PR; prod is comfortably under (~230), test rewrite drives the total.
- **R3 — Dropping the advisory warning (D2)** changes one existing test's intent. Low risk (advisory was
  cosmetic), but it is a visible behavior change — confirm.
