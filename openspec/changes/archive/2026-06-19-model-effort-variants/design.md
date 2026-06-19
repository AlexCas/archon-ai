# Design — model-effort-variants (Slice 4, option b)

Apply-ready. Adds effort selection for reasoning models + writes `variant` to opencode.json.
Settled by `specs/model-effort-variants/spec.md`. No plugin/cache/embed.

## A. config/model.go — carry Effort through resolution

`PhaseModel` gains an `Effort` field; `ResolvePhaseModels` sets it from the resolved ref.

```go
type PhaseModel struct {
	Phase  string
	Model  string
	Effort string // resolved ModelRef.Effort (variant); "" = provider default
}
```
In `ResolvePhaseModels`, the emit line becomes:
```go
out = append(out, PhaseModel{Phase: p, Model: ref.FullID(), Effort: ref.Effort})
```
(`ref` is the resolved phase-or-default ref; phase/fallback/omit logic unchanged.)

## B. initcmd/opencode_mode.go — write variant

Add `Variant` AFTER `Model` in both structs (declaration order = JSON order; `omitempty` so an
empty effort omits the key, keeping effortless output byte-identical to today):

```go
type archonLeaderAgent struct {
	Mode        string `json:"mode"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Variant     string `json:"variant,omitempty"` // effort/reasoning level; omitted when empty
	Prompt      string `json:"prompt"`
}

type archonPhaseAgent struct {
	Mode        string `json:"mode"`
	Hidden      bool   `json:"hidden"`
	Model       string `json:"model"`
	Variant     string `json:"variant,omitempty"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}
```
Populate:
- leader (line ~84): add `Variant: models.Leader.Effort,`
- phase loop (line ~92): add `Variant: pm.Effort,`

Idempotency: archon rewrites the whole `archon-*` entry each run, so an emptied effort re-serializes
WITHOUT the variant key (omitempty) — no stale variant survives. Verified by a re-run test.

## C. tui/models_tab.go — effortSelect sub-mode

Enum: add `effortSelect` after `freeForm`:
```go
const (
	rowNav subMode = iota
	providerSelect
	modelSelect
	freeForm
	effortSelect
)
```
State: add `effortCursor int` to `modelsTabState`. Fixed options (package-level or const):
```go
// effortOptions[0] "default" maps to empty Effort (provider default).
var effortOptions = []string{"default", "low", "medium", "high"}
```

`updateModelSelect` Enter (currently sets ref + `mode=rowNav` at lines 214-216): after building the
ref and marking `changed`, branch on the picked model's Reasoning:
```go
case tea.KeyEnter:
	picked := m.curModels[m.modelCursor]
	m.rows[m.focusedRow].ref = refFromCacheKey(m.pickedProvider, picked.ID)
	m.rows[m.focusedRow].changed = true
	if picked.Reasoning {
		m.mode = effortSelect
		m.effortCursor = 0
	} else {
		m.mode = rowNav // Effort stays empty (ref freshly built)
	}
```

New handler:
```go
func (m *modelsTabState) updateEffortSelect(key tea.KeyMsg) (tea.Cmd, bool) {
	switch key.Type {
	case tea.KeyUp:
		if m.effortCursor > 0 { m.effortCursor-- }
	case tea.KeyDown:
		if m.effortCursor < len(effortOptions)-1 { m.effortCursor++ }
	case tea.KeyEnter:
		opt := effortOptions[m.effortCursor]
		if opt == "default" { opt = "" }
		m.rows[m.focusedRow].ref.Effort = opt
		m.mode = rowNav
	case tea.KeyEsc:
		m.mode = modelSelect // step back; model already set
	}
	return nil, true
}
```
Wire into `update()` dispatch (`switch m.mode`, ~line 129) and `view()` (`switch m.mode` in renderRow,
~line 302): render an indented cursor list of `effortOptions` with the focus style on `effortCursor`,
header `Effort:`; hint line `↑/↓: choose effort · Enter: set · Esc: back`. Non-key msgs in effortSelect
need no textinput handling (only freeForm uses the input).

Free-form path: unchanged — `ParseModelRef` does not set Effort, so free-form rows have empty Effort.

## D. config round-trip — already done

`MarshalYAML` emits a mapping `{provider, model, effort}` when `Effort != ""`, scalar otherwise
(merged in the foundation, tested). No change. Confirm the picker→applyToConfig path carries
`row.ref.Effort` (it does — applyToConfig writes `row.ref` verbatim).

## E. Test plan
- `config/model_test.go`: `ResolvePhaseModels` sets `PhaseModel.Effort` from the resolved ref
  (phase ref with effort; fallback-to-default ref's effort).
- `initcmd/opencode_mode_test.go`: variant present when a phase/leader ref has Effort; variant key
  ABSENT when Effort empty; re-run byte-identical with mixed effort/empty (idempotency). Reuse
  `phaseAgentFrom`/`leaderAgentFrom`.
- `tui/models_tab_test.go`: picking a `Reasoning:true` model enters effortSelect; choosing "high"
  sets `rows[i].ref.Effort=="high"`; choosing "default" → ""; a `Reasoning:false` model skips
  effortSelect (mode→rowNav, Effort empty); Esc from effortSelect → modelSelect. Use the existing
  `toolModel` helper plus a reasoning variant `reasoningModel(id,name)` (ToolCall+Reasoning true).

## Determinism / size
Deterministic: fixed effortOptions order; JSON keys sorted; omitempty variant. Existing effortless
configs + opencode.json unchanged. Size ~90-130 prod + ~90-130 test ≈ 200. Under D1 400.
