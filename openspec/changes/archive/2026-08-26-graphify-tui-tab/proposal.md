# Proposal: Graphify tab in `archon tui`

## Intent

`config.Graphify` is fully plumbed (Slice A, PR #103) — struct, defaults, `Load`/`Clone`/`Save`, `archon config get/set`, and `archon status` — but there is **no interactive editing surface**. Every other opt-in gate (Judge, Mutation, Playwright, Security, Impeccable) has a `archon tui` tab; Graphify is the odd gate out, forcing users to hand-edit YAML or memorize `config set` keys. This change adds the missing tab, at parity with the Impeccable precedent.

## Scope

### In Scope
- New `internal/tui/graphify_tab.go` mirroring `impeccable_tab.go`: `graphifyTabState`, `newGraphifyTabState`, `update`/`refocus`/`view`/`applyToConfig`/`setWidth`.
- Expose all 5 `config.Graphify` fields — 3 toggles (`enabled`, `auto_install`, `semantic`) + 2 textinputs (`version`, `output_dir`); `graphifyFocusCount = 5`, **bools-first** focus order (OQ1).
- Wire the ~9 `model.go` sites (iota, field, ctor, two `setWidth` fan-outs, key dispatch, `agentInitDoneMsg` rebuild, `saveConfig` fan-out, `renderTabs`, `renderTabContent`).
- Fix the 2 tests that break (`TestModel_Update_ShiftTabWrapsFromAgent`, `TestModel_renderTabs_Order`) + add `TestGraphifyTabState_ApplyToConfig`.
- Hand-edit docs: group G clause in `internal/initcmd/templates.go` and root `CLAUDE.md`.

### Out of Scope
- The `tabModel` interface refactor to kill the 9-site fan-out (deferred; Approach 2).
- Any live "graphify installed / not installed" probe — parity with Impeccable, no runtime check (OQ4).
- Config-struct, CLI, status, or `--graphify` flag changes (all done in Slice A).

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `graphify-integration`: add a requirement for the `archon tui` Graphify tab editing surface (the config schema and advisory-gate behavior are unchanged).

## Approach

Approach 1 — clone the Impeccable tab. Proven, uniform with every existing tab, no framework decisions, lands near the ~320-line estimate. Two resolved design decisions: blank `version`/`output_dir` **coerce to package defaults** in `applyToConfig` (never persist `""`; mirrors Impeccable's blank-severity fallback, OQ2); focus/toggle order is bools-first (OQ1). Test placement (OQ3): **fold unit tests into `model_test.go`** next to `TestImpeccableTabState_ApplyToConfig` — matches the direct structural precedent this tab clones, keeping the new tab's tests beside its twin rather than in a separate file.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/graphify_tab.go` | New | ~120-150 lines, Impeccable clone |
| `internal/tui/model.go` | Modified | ~9 wiring sites |
| `internal/tui/model_test.go` | Modified | Fix 2 tests, add 1 applyToConfig test |
| `internal/initcmd/templates.go` | Modified | Group G "or the Graphify tab" clause (hand-edit) |
| `CLAUDE.md` | Modified | Same group G clause (hand-edit) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Under-wire a `model.go` site (silent reset / no resize) | Med | Grep every `impeccableTab` occurrence as checklist; lean on tab-consistency + resize e2e tests |
| Blank textinput persists `""` (Load won't re-seed present-but-empty key) | Med | `applyToConfig` coerces blank → package default |
| Existing tab tests drift (new last tab shifts wrap + order) | High | Fix both in same slice |
| `archon init --force` reverts merged docs (templates-go-drift) | Med | Hand-edit `templates.go` + `CLAUDE.md`; never regenerate |
| 8th header label overflows narrow terminals | Low | Verify via `TestEdgeCases_WindowResize` |

## Rollback Plan

Self-contained and additive: `git revert` the change. Deleting `graphify_tab.go` and reverting the `model.go`/test/docs edits removes the tab; `config.Graphify` and all Slice A plumbing remain untouched and functional.

## Dependencies

- Slice A (`config.Graphify` plumbing, PR #103) — already merged.

## Success Criteria

- [ ] `archon tui` shows a "Graphify" tab; all 5 fields are viewable and editable.
- [ ] Toggling/editing then saving round-trips through `config.Graphify` (defaults coerced on blank).
- [ ] `go test ./internal/tui/...` passes (2 fixed + 1 new test).
- [ ] Group G docs in `templates.go` and `CLAUDE.md` carry the tab clause; no `init --force` used.
