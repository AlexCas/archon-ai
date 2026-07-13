# Proposal: TUI Opens on the Agent Tab

## Intent

`archon tui` currently opens on the **Models** tab, but the intended workflow is
"pick the agent first, then configure the models." The landing tab and left-to-right
tab order contradict that flow. This change makes the TUI open on **Agent** and puts
Agent first in the header, so the screen matches the mental model on launch.

## Scope

### In Scope
- Reorder the `Tab` enum so `AgentTab` is first: `Agent, Models, Judge, Mutation Testing, Playwright`.
- Reorder the header label slice to match the new enum order.
- Set the default landing tab in `NewModel` to `AgentTab`.
- Update the 4 order/default-coupled test assertions to the new layout.

### Out of Scope
- The single ordered tab-descriptor table refactor (Approach 2) — deferred as a follow-up.
- Any change to tab *content*, per-tab state structs, or `applyToConfig` behavior.
- Persisting the active tab to `.archon/config.yaml` (none exists; not needed).

## Capabilities

### New Capabilities
- None. Tab ordering/landing has no `openspec/specs/` capability; this is UI behavior within `internal/tui`.

### Modified Capabilities
- None. No spec-level requirement changes; behavior is enforced by the TUI tests.

## Approach

Approach 1 from exploration (minimal reorder). Tab order lives in three index-aligned
places, all edited together: the `iota` enum, the header label slice, and the default-tab
constant. `tabCount` stays 5. Navigation is modulo arithmetic, so it is order-agnostic.
The two `switch m.activeTab` blocks are name-keyed and keep working untouched (optionally
reordered for readability). Tests that hard-assert the default tab and its Tab-key neighbor
are updated to prove the reorder landed.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go:21-28` | Modified | Reorder enum: `AgentTab=0 … PlaywrightTab=4` |
| `internal/tui/model.go:92` | Modified | `NewModel` default `activeTab: AgentTab` |
| `internal/tui/model.go:253` | Modified | Header label slice reordered Agent-first |
| `internal/tui/model_test.go:37-38,91-92,100-101` | Modified | Default + Tab/Shift+Tab neighbor assertions |
| `internal/tui/e2e_test.go:118-128` | Modified | Full-cycle default-tab assertion → `AgentTab` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Enum and label slice desync | Med | Edit both together; test assertions guard alignment |
| Default not moving (named constant, not `0`) | Med | Explicitly set `activeTab: AgentTab` |
| Stale test assertions fail | High (expected) | They are the safety net; update all 4 in the same change |
| Chained-PR index coupling (PR 1 of 3) | Low | Keep enum change isolated and documented for clean rebase |

## Rollback Plan

Single-commit revert. All edits are confined to three files in `internal/tui`; `git revert`
restores the prior enum order, label slice, default tab, and test assertions. No data,
config, or serialization migration to undo.

## Dependencies

- None. No other package references the `Tab` constants (confined to `internal/tui`).

## Size Estimate

~30-40 changed lines across 3 files (1 production, 2 test). Well under the 400-line budget.

## Success Criteria

- [ ] `archon tui` opens on the Agent tab.
- [ ] Header order reads `Agent, Models, Judge, Mutation Testing, Playwright`.
- [ ] `go test ./internal/tui/...` passes with updated assertions.
- [ ] Diff touches only `internal/tui/model.go`, `model_test.go`, `e2e_test.go`.
