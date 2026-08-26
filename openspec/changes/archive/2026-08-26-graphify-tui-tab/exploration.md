# Exploration: Graphify tab in the Archon TUI (`archon tui`)

## Project Type
**Web testing**: not-web

Archon is a Go CLI/TUI orchestrator (Bubbletea). No browser surface, no web
framework, no E2E web tooling. Playwright (preflight group E) MUST stay off, and
the Impeccable recommendation (group F) does not apply.

## Current State

`archon tui` is a Bubbletea program in `internal/tui`. `Model` (in `model.go`)
owns a fixed set of tabs enumerated by the `Tab` iota:

```
AgentTab, ModelsTab, JudgeTab, MutationTab, PlaywrightTab, SecurityTab, ImpeccableTab, tabCount
```

Each tab is a self-contained `<name>TabState` struct stored on `Model`, with a
uniform contract:
- `new<Name>TabState(cfg.<Section>) <name>TabState` — build from config.
- `update(msg tea.Msg) (tea.Cmd, bool)` — handle keys for the focused tab.
- `view(width, height int) string` — render.
- `applyToConfig(cfg *config.Config)` — write the tab's fields into a `*config.Config`.
- `setWidth(width int)` — resize inputs.

`Model` wires each tab in **seven parallel places**, all of which the new tab must
join:
1. `Tab` iota constant (`model.go` ~L22-31) — add `GraphifyTab` before `tabCount`.
2. Struct field on `Model` (~L45-53) — `graphifyTab graphifyTabState`.
3. `NewModel` construction (~L108-115) — `newGraphifyTabState(cfg.Graphify)`.
4. `Update` → `WindowSizeMsg` setWidth fan-out (~L129-135).
5. `Update` → `tea.KeyMsg` tab-specific dispatch switch (~L158-194).
6. `Update` → `agentInitDoneMsg` rebuild + setWidth fan-out (~L211-226).
7. `saveConfig()` applyToConfig fan-out (~L352-358).

Plus the two render helpers:
8. `renderTabs()` label slice (~L286) — append `"Graphify"`.
9. `renderTabContent()` switch (~L306-323) — add a `case GraphifyTab`.

The **Impeccable tab is the exact structural precedent** (`impeccable_tab.go`,
182 lines): another opt-in external gate with an `enabled` toggle, an
`auto_install` toggle, and free-text config fields backed by
`textinput.Model`. It uses an integer `focused` index cycled by Up/Down over a
`impeccableFocusCount = 5` modulus; Enter/Space toggles the two bools at focus 0
and 1; other focus indices forward the key to the focused `textinput`. `refocus()`
blurs all inputs then focuses the active one. This is the pattern to copy.

The **Security tab** (`security_tab.go`) is a secondary reference for the
"enum-cycle a value with Enter" idiom (not needed here — Graphify has no enum
field, but useful if `version` ever becomes a picker).

## Config surface to expose

`config.Graphify` (`internal/config/config.go` L64-70) — five fields, all
already round-tripped by `Config.Load`, `Clone`, `Save`, the `archon config
get/set` CLI, and `archon status`:

| Field         | YAML          | Type   | TUI control (proposed)        | Default |
|---------------|---------------|--------|-------------------------------|---------|
| `Enabled`     | `enabled`     | bool   | toggle (Enter/Space)          | `false` |
| `AutoInstall` | `auto_install`| bool   | toggle (Enter/Space)          | `false` |
| `Version`     | `version`     | string | `textinput`                   | `v0.9.45` (`DefaultGraphifyVersion`) |
| `OutputDir`   | `output_dir`  | string | `textinput`                   | `.archon/graphify` (`DefaultGraphifyOutputDir`) |
| `Semantic`    | `semantic`    | bool   | toggle (Enter/Space)          | `false` |

Key difference vs Impeccable: Graphify has **three** bool toggles and **two**
text fields (Impeccable has two toggles + three text fields), and — critically —
**no validation and no severity enum**. `applyToConfig` is therefore simpler than
Impeccable's: there is no `ValidateImpeccableSeverity`-style guard and no
blocking-verdict semantics (Graphify is advisory-only, never blocks). The one
defaulting concern is the blank-input case for `Version`/`OutputDir` — see Risks.

## Affected Areas

- `internal/tui/graphify_tab.go` — **new file**, ~120-150 lines. Mirror
  `impeccable_tab.go`: `graphifyTabState` struct, `graphifyFocusCount`,
  `newGraphifyTabState`, `update`, `refocus`, `view`, `applyToConfig`, `setWidth`.
- `internal/tui/model.go` — **~9 edit sites** listed above (iota, field, ctor,
  two setWidth fan-outs, key dispatch, agentInitDoneMsg rebuild, applyToConfig
  fan-out, renderTabs labels, renderTabContent switch).
- `internal/tui/model_test.go` — update `TestModel_Update_ShiftTabWrapsFromAgent`
  (Shift+Tab from AgentTab must now land on `GraphifyTab`, the new last tab, not
  `ImpeccableTab`) and `TestModel_renderTabs_Order` (append `"Graphify"` to the
  expected label list). Add `TestGraphifyTabState_ApplyToConfig` next to the
  existing `TestImpeccableTabState_ApplyToConfig` (note: Impeccable's unit tests
  live in `model_test.go`, not a dedicated file).
- `internal/tui/graphify_tab_test.go` — **optional new file** for focus/toggle
  unit tests, mirroring `security_tab_test.go` (that tab keeps its own test file;
  either placement is consistent with existing precedent).
- `internal/initcmd/templates.go` — **docs drift touch-up**. Group G text
  (L100-102) currently says only "The `--graphify` flag at init time sets the
  same value" — unlike group F which reads "The `--impeccable` flag at init time
  **or the Impeccable tab in `archon tui`** set the same value." Once the tab
  exists, group G should gain the parallel "or the Graphify tab in `archon tui`"
  clause. **HARD CONSTRAINT from `templates-go-drift` memory**: `templates.go` is
  already behind root `CLAUDE.md`; do NOT run `archon init --force` to regenerate
  it (that reverts merged archive-before-PR content). Hand-edit only, and mirror
  the identical clause into root `CLAUDE.md` group G by hand.

**Not affected** (already complete from Slice A, no change needed): `config.go`
(struct + defaults exist), `cmd/archon/config.go` (get/set keys exist),
`internal/status/display.go` (Graphify block exists), `skills/graphify/SKILL.md`,
`--graphify` init flag.

## Approaches

1. **Clone the Impeccable tab (recommended)** — copy `impeccable_tab.go` to
   `graphify_tab.go`, adapt the struct to the five Graphify fields (3 toggles +
   2 textinputs), wire the ~9 `model.go` sites, fix the two affected tests, add
   one `applyToConfig` test, and hand-edit the two docs files.
   - Pros: proven pattern, uniform with every other tab, reviewers already know
     the shape, no framework decisions; lands near the ~320-line estimate.
   - Cons: repeats the known "wire in nine places" boilerplate (a latent
     refactor smell across all tabs, but out of scope here).
   - Effort: Low-Medium.

2. **Introduce a `tabModel` interface + slice to kill the fan-out** — refactor
   all tabs behind a common interface so new tabs register once.
   - Pros: removes the 9-site wiring; future tabs become trivial.
   - Cons: touches all seven existing tabs and their tests; balloons the diff far
     past the ~320-line budget; scope creep unrelated to Graphify.
   - Effort: High.

## Recommendation

**Approach 1.** It matches the Slice A proposal's own deferral note
(`internal/tui/graphify_tab.go` + `model.go` wiring + test, ~320 lines) and the
Impeccable precedent the task calls out. Keep the interface refactor (Approach 2)
as a separate future change if tab-wiring boilerplate becomes painful.

## Risks

- **Nine-site wiring is easy to under-wire.** Missing the `agentInitDoneMsg`
  rebuild or a `setWidth` fan-out yields a tab that silently resets or fails to
  resize. Use `TestIntegration_TabStateConsistency` / the e2e resize test as a
  guard and grep every place `impeccableTab` appears in `model.go` as the
  checklist.
- **Blank textinput handling.** Impeccable falls back to a safe default on blank
  severity. For Graphify, `Version`/`OutputDir` have package defaults
  (`DefaultGraphifyVersion`, `DefaultGraphifyOutputDir`); `applyToConfig` should
  decide whether a blank field writes `""` (config.Load re-seeds defaults only
  when the block is absent, NOT when a key is present-but-empty) or coerces to
  the default. Coercing to the default in `applyToConfig` is safer and matches
  Impeccable's blank-severity behavior. This is the one real design decision.
- **Test drift beyond the new tab.** Adding `GraphifyTab` as the new last tab
  shifts Shift+Tab wrap target and the `renderTabs` order assertion; both
  existing tests WILL fail until updated. Expected, but must be done in the same
  slice.
- **Docs drift trap.** Editing `templates.go` risks tempting a regenerate;
  `templates-go-drift` memory forbids `archon init --force`. Hand-edit both
  `templates.go` and root `CLAUDE.md`.
- **Tab-header width.** Adding an eighth label ("Graphify") to `renderTabs`
  lengthens the header row; verify it still renders on narrow terminals (the
  `TestEdgeCases_WindowResize` path).

## Open Questions

1. Toggle ordering in `view`/focus: put all three bools first
   (`enabled`, `auto_install`, `semantic` at focus 0/1/2) then the two textinputs
   (3/4)? Or group `semantic` after the paths like the status display orders it?
   Recommend bools-first for a clean focus model (`graphifyFocusCount = 5`).
2. Should a blank `Version`/`OutputDir` in the tab coerce to the package default
   or persist empty? (Recommend coerce — see Risks.)
3. Add a dedicated `graphify_tab_test.go` (Security-tab style) or keep unit tests
   in `model_test.go` (Impeccable style)? Both are precedented; pick one.
4. Any "installed / not installed" visual for the `graphify` binary, or leave
   install state out of the TUI (Impeccable's tab shows no install probe — it
   only exposes `auto_install`)? Recommend parity with Impeccable: no live probe.

## Ready for Proposal
**Yes.** Scope is well-bounded (~320 lines), the precedent file and the exact
wiring sites are identified, config plumbing already exists from Slice A, and the
only genuine decisions are the four open questions above (all with a recommended
default). The orchestrator should confirm the blank-field coercion and toggle
ordering with the user, then proceed to `sdd-propose`.
