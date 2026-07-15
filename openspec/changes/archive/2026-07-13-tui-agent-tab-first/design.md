# Design: TUI Opens on the Agent Tab

## Technical Approach

Approach 1 from the exploration (minimal reorder). Tab order lives in three
index-aligned literals in `internal/tui/model.go`: the `iota` enum, the header
label slice, and the `NewModel` default-tab constant. We edit all three together
so `AgentTab = 0` becomes the source of truth, the header reads Agent-first, and
the TUI opens on Agent. `tabCount` stays `5`. The two `switch m.activeTab` blocks
are keyed by named constant, so they are order-agnostic and keep working
unchanged. Navigation is modulo arithmetic over `tabCount`, also order-agnostic.
Four order/default-coupled test assertions are updated to prove the reorder
landed. Satisfies the four ADDED requirements in `specs/tui-tabs/spec.md`.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Ordering mechanism | Reorder the three parallel literals in place | Single ordered `[]tabDef` descriptor table (explore Approach 2) | Smallest diff; PR 1 of 3 under a 400-line budget — scope discipline. Descriptor refactor deferred as follow-up. |
| Default landing tab | Explicit `activeTab: AgentTab` in `NewModel` | Rely on enum reorder alone | Default is a *named constant*, not `0`. Reordering the enum alone leaves the default on `ModelsTab` (now second). Must set it explicitly to honor "pick the agent first". |
| Switch blocks | Leave both `switch` bodies untouched | Reorder cases for cosmetic alignment | Name-keyed cases are correct regardless of enum value; reordering adds diff noise with no behavior change. |

## Data Flow

    enum (iota)  ──index──▶  label slice  ──renderTabs Tab(i)──▶  header strip
        │                                                              │
        └── NewModel default: AgentTab ──▶ activeTab ──▶ renderTabContent

The enum value is the single ordering authority. `renderTabs` (`model.go:256-262`)
matches `Tab(i)` against `m.activeTab`, so the label slice MUST stay index-aligned
with the enum — the one true coupling this change must preserve.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | Enum reorder (`21-28`), default tab (`92`), label slice (`253`) |
| `internal/tui/model_test.go` | Modify | Update assertions at `37-38`, `91-92`, `100-101` |
| `internal/tui/e2e_test.go` | Modify | Update full-cycle default assertion at `126-127` |

## Concrete Before / After

### Enum (`model.go:21-28`)

Before: `ModelsTab=0, JudgeTab=1, MutationTab=2, PlaywrightTab=3, AgentTab=4, tabCount=5`

After (order of the `const` block):

    AgentTab Tab = iota  // 0
    ModelsTab            // 1
    JudgeTab             // 2
    MutationTab          // 3
    PlaywrightTab        // 4
    tabCount             // 5

### Label slice (`model.go:253`)

Before: `[]string{"Models", "Judge", "Mutation Testing", "Playwright", "Agent"}`
After:  `[]string{"Agent", "Models", "Judge", "Mutation Testing", "Playwright"}`

### Default landing (`model.go:92`)

Before: `activeTab: ModelsTab,`
After:  `activeTab: AgentTab,`

### Coupled test assertions

| # | Location | Meaning | Before asserts | After asserts |
|---|----------|---------|----------------|---------------|
| 1 | `model_test.go:37-38` | `NewModel` default tab | `ModelsTab` | `AgentTab` |
| 2 | `model_test.go:91-92` | after one `Tab` press from default | `JudgeTab` | `ModelsTab` |
| 3 | `model_test.go:100-101` | after `Shift+Tab` back to default | `ModelsTab` | `AgentTab` |
| 4 | `e2e_test.go:126-127` | after `tabCount*3` presses (returns to default) | `ModelsTab` | `AgentTab` |

Assertions 1, 3, 4 all track the new default (`AgentTab`). Assertion 2 tracks the
new *neighbor* of the default: with Agent first, one `Tab` press goes
`Agent → Models`, so it becomes `ModelsTab` (not `JudgeTab`).

## Interfaces / Contracts

No new interfaces, types, or signatures. `Tab`, `tabCount`, `NewModel`, `Update`,
`renderTabs`, `renderTabContent` keep their existing signatures. Per-tab state
structs and their `newXxxTabState`/`update`/`view`/`applyToConfig` methods are
field-keyed and untouched.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Default tab, Tab/Shift+Tab neighbors | Updated assertions 1-3 in `model_test.go` |
| Integration (Bubbletea) | Full-cycle return to default | Updated assertion 4 in `e2e_test.go` |
| Regression | Per-tab content non-empty | `model_test.go:197-212` — name-keyed, no change needed |

Run `go test ./internal/tui/...`; all four updated assertions must pass and no
other tests regress.

## Migration / Rollout

No migration required. No persisted active tab in `.archon/config.yaml`; no
serialization change. Single-commit `git revert` restores prior order.

**Chained-PR isolation**: This is PR 1 of 3. Keep the diff confined to the three
`internal/tui` files and the enum/label/default/test edits above — no adjacent
refactors — so PRs 2-3 (models config flow) rebase cleanly. No other package
references the Tab constants (verified: `grep` for `ModelsTab|JudgeTab|MutationTab|PlaywrightTab|AgentTab|tabCount` outside `internal/tui` returns nothing), so the
new index values cannot leak into later PRs.

## Open Questions

- None. Enum order, label order, and default tab are locked by the proposal and
  spec; source verified against the explore notes.
