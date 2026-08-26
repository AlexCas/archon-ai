# Design: Graphify tab in `archon tui`

<!-- [[graphify-integration]] · [proposal](proposal.md) · [spec](specs/graphify-integration/spec.md) -->

## Technical Approach

Clone the Impeccable tab (`internal/tui/impeccable_tab.go`, 182 lines) into a new
`internal/tui/graphify_tab.go`, adapting the field set to `config.Graphify`'s five
fields (3 bools + 2 textinputs, bools-first). Wire the new `graphifyTabState` into
`model.go` at the nine canonical fan-out sites, fix the two order-sensitive tests
and add one `applyToConfig` test in `model_test.go`, then hand-edit the group-G
docs clause in `templates.go` and `CLAUDE.md`. No new abstractions, no interface
refactor — mechanical parity with every existing tab. Implements R-19 and the
R-05 doc modification from the spec.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Tab shape | Structural clone of `impeccableTabState` | Uniform contract (`new…`/`update`/`refocus`/`view`/`applyToConfig`/`setWidth`); reviewers know the shape; lands ~320 lines |
| Focus order | bools-first: `enabled`(0), `auto_install`(1), `semantic`(2), `version`(3), `output_dir`(4); `graphifyFocusCount = 5` | Clean focus model; matches spec R-19 table |
| Blank text coercion | `applyToConfig` coerces blank `version`→`DefaultGraphifyVersion`, blank `output_dir`→`DefaultGraphifyOutputDir`; never persist `""` | `config.Load` only re-seeds when the block is absent, not for a present-but-empty key; mirrors Impeccable's blank-severity fallback |
| No validation guard | Omit any `Validate…` call in `applyToConfig` | Graphify is advisory-only — no severity enum, no `Load()` validation (config.go L60-63) |
| No install probe | Tab exposes only `auto_install`, no live "installed?" check | Impeccable parity (OQ4) |
| Test placement | Fold into `model_test.go` beside `TestImpeccableTabState_ApplyToConfig` | Matches the file this tab clones |

## Interfaces / Contracts

`graphifyTabState` in `internal/tui/graphify_tab.go`:

```go
type graphifyTabState struct {
    enabled     bool            // focus 0
    autoInstall bool            // focus 1
    semantic    bool            // focus 2
    version     textinput.Model // focus 3
    outputDir   textinput.Model // focus 4
    focused     int
}
const graphifyFocusCount = 5
```

- `newGraphifyTabState(cfg config.Graphify) graphifyTabState` — seed inputs via
  `SetValue(cfg.Version)` / `SetValue(cfg.OutputDir)`, placeholders naming the
  defaults, `Width = 30`, `focused = 0`.
- `update` — Up/Down cycle mod 5 + `refocus()`; Enter/Space toggle bools at focus
  0/1/2; else forward key to the input at focus 3/4.
- `refocus()` — blur `version`+`outputDir`, focus the active one (only 3/4).
- `view(width, height)` — three toggle lines then two labelled inputs, copying
  Impeccable's lipgloss styling; info footer noting advisory/non-blocking behavior.
- `applyToConfig(cfg *config.Config)`:
  ```go
  cfg.Graphify.Enabled = p.enabled
  cfg.Graphify.AutoInstall = p.autoInstall
  cfg.Graphify.Semantic = p.semantic
  v := strings.TrimSpace(p.version.Value())
  if v == "" { v = config.DefaultGraphifyVersion }
  cfg.Graphify.Version = v
  o := strings.TrimSpace(p.outputDir.Value())
  if o == "" { o = config.DefaultGraphifyOutputDir }
  cfg.Graphify.OutputDir = o
  ```
- `setWidth(width)` — `w := width-20; if w<10 {w=10}`; apply to both inputs.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/graphify_tab.go` | Create | ~120-150 lines; the state above |
| `internal/tui/model.go` | Modify | 9 wiring sites (below) |
| `internal/tui/model_test.go` | Modify | Fix 2 tests, add 1 |
| `internal/initcmd/templates.go` | Modify | Group G clause (hand-edit) |
| `CLAUDE.md` | Modify | Same group G clause (hand-edit) |

### The nine `model.go` wiring sites (current line anchors)

1. **Iota** (L22-31): add `GraphifyTab` after `ImpeccableTab`, before `tabCount`
   — makes Graphify the new last tab.
2. **Model field** (L51-52): add `graphifyTab graphifyTabState` after `impeccableTab`.
3. **`NewModel` ctor** (L113): add `graphifyTab: newGraphifyTabState(cfg.Graphify),`.
4. **`WindowSizeMsg` setWidth fan-out** (L134): add `m.graphifyTab.setWidth(m.width)`
   after the impeccable line.
5. **Key-dispatch switch** (L184-188): add `case GraphifyTab:` calling
   `m.graphifyTab.update(msg)` with the same `cmd != nil` append.
6. **`agentInitDoneMsg` rebuild** (L216): add
   `m.graphifyTab = newGraphifyTabState(msg.cfg.Graphify)` after impeccable.
7. **`agentInitDoneMsg` setWidth fan-out** (L224, inside `if m.width > 0`): add
   `m.graphifyTab.setWidth(m.width)`.
8. **`saveConfig` applyToConfig fan-out** (L357): add
   `m.graphifyTab.applyToConfig(cfg)` after impeccable.
9. **`renderTabs` labels** (L286): append `"Graphify"` to the slice **and**
   **`renderTabContent` switch** (L317-318): add `case GraphifyTab:` returning
   `style.Render(m.graphifyTab.view(m.width, m.height))`.

(Sites 4 and 7 are the two distinct setWidth fan-outs; site 9 covers both render
helpers — nine logical joins, matching the spec's "nine canonical sites".)

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `TestModel_Update_ShiftTabWrapsFromAgent` | Change expected wrap target `ImpeccableTab` → `GraphifyTab` (L121-122) + doc comment |
| Unit | `TestModel_renderTabs_Order` | Append `"Graphify"` to the `labels` slice (L222) |
| Unit | `TestGraphifyTabState_ApplyToConfig` (new, beside Impeccable's) | Toggle focus 0/1/2 via Enter; set `version`/`outputDir`; assert all five land on `cfg.Graphify`; a blank-input sub-case asserts coercion to `DefaultGraphifyVersion`/`DefaultGraphifyOutputDir` |

Existing `TestIntegration_TabStateConsistency` and the resize/edge-case tests guard
the fan-out wiring automatically. Run `go test ./internal/tui/...`.

### Docs edits (hand-edit only — never `archon init --force`)

Append the parallel clause "or the Graphify tab in `archon tui`" so group G reads
`The --graphify flag at init time or the Graphify tab in archon tui set the same
value.` in both `internal/initcmd/templates.go` (L100-101, `§`-escaped backticks)
and `CLAUDE.md` (L95-96). `templates-go-drift` constraint: hand-edit both; do not
regenerate.

## Migration / Rollout

No migration required. Additive and self-contained; `git revert` removes the tab
and leaves all Slice A `config.Graphify` plumbing intact.

## Slicing / PR strategy

Session PR strategy = **ask-always**, budget = **800 lines**. Estimate ~320 lines
(new file ~130 + ~15 wiring + ~40 tests + ~4 docs), well under budget — **one PR**
is expected. No chained-PR split needed. Flag for the gate only if implementation
unexpectedly overshoots 800.

## Open Questions

None — all four exploration OQs resolved in the proposal (bools-first, coerce
blank, tests in `model_test.go`, no install probe).
