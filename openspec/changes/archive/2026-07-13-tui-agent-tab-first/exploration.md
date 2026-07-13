# Exploration: tui-agent-tab-first

## Project Type
**Web testing**: not-web

Signals confirmed: Go module `github.com/archon-ai/archon` (go 1.25.0), CLI via
`spf13/cobra`, terminal UI via `charmbracelet/bubbletea`/`bubbles`/`lipgloss`.
No `package.json`, no web framework, no HTTP server routes, no Playwright/Cypress/
chromedp. Playwright generation MUST stay disabled for this change.

## Current State

The `archon tui` screen is a Bubbletea program in `internal/tui`. Tab identity is
an `iota` enum (`Tab int`) and tab order is expressed in THREE parallel places that
must stay index-aligned:

1. The enum (`internal/tui/model.go:21-28`):
   `ModelsTab=0, JudgeTab=1, MutationTab=2, PlaywrightTab=3, AgentTab=4, tabCount=5`.
2. The header label slice (`internal/tui/model.go:253`):
   `[]string{"Models","Judge","Mutation Testing","Playwright","Agent"}` — rendered by
   index in `renderTabs()` (`model.go:256-262`), where `Tab(i) == m.activeTab`
   ties the label position directly to the enum value.
3. The content-routing `switch m.activeTab` (`renderTabContent`, `model.go:273-286`)
   and the key-routing `switch m.activeTab` (`Update`, `model.go:139-165`).

Navigation is modulo arithmetic over `tabCount` (`model.go:125-131`), so it is
order-agnostic and needs no change. `tabCount` stays 5.

The default active tab is set in `NewModel` (`model.go:92`): `activeTab: ModelsTab`.
This is a **named-constant** assignment, not a hardcoded `0`, so once the enum is
reordered the default will follow whatever `ModelsTab` resolves to — which is NOT
what we want. To honor "pick the agent first", the default should become
`AgentTab`.

Each tab's state struct and its `newXxxTabState`, `setWidth`, `update`, `view`,
`applyToConfig` methods live in their own files (`agent_tab.go`, `models_tab.go`,
`judge_tab.go`, `mutation_tab.go`, `playwright_tab.go`). These are keyed by struct
field name, not by tab index, so they are unaffected by reordering. `saveConfig`
(`model.go:311-344`) calls each tab's `applyToConfig` unconditionally — order there
is irrelevant.

## Index-Dependency Map (every place coupled to tab ordering)

Production code (`internal/tui/model.go`):
- `21-28` enum block — the source of truth for order; reorder here.
- `92` `NewModel` default `activeTab: ModelsTab` — change to `AgentTab` so agent
  is the landing tab.
- `253` header label slice — reorder to match the new enum order.
- `139-165` `Update` key-routing switch — `case` bodies are keyed by name, so they
  do NOT need reordering, but the switch as a whole should stay readable.
- `273-286` `renderTabContent` switch — same: name-keyed, order-independent, but
  worth ordering for readability.
- `256-262` `renderTabs` loop — relies on `Tab(i)` matching enum; correct as long
  as the label slice matches the enum order.

Tests coupled to order/index:
- `internal/tui/model_test.go:37-38` — asserts `NewModel` default `activeTab == ModelsTab`.
  MUST change to `AgentTab` (default landing tab changes).
- `internal/tui/model_test.go:91-92` — after one `Tab` press expects `JudgeTab`.
  With agent first the sequence becomes `Agent → Models → …`; this assertion MUST
  be updated to the new neighbor of the default tab.
- `internal/tui/model_test.go:100-101` — after `Shift+Tab` expects to return to
  `ModelsTab`; MUST update to the new default (`AgentTab`).
- `internal/tui/model_test.go:197-212` — `renderTabContent` per-tab non-empty
  checks are name-keyed (`ModelsTab`, `MutationTab`, `AgentTab`); order-agnostic,
  no change needed.
- `internal/tui/e2e_test.go:118-128` (`TestEdgeCases_QuickTabSwitching`) — cycles
  `tabCount*3` and expects to land back on `ModelsTab`. This asserts the DEFAULT
  tab, so with a new default it MUST become `AgentTab`.

Tests NOT coupled to order (safe): all `newAgentTabState`/`newModelsTabState`/
`newMutationTabState` unit tests in `model_test.go`, `e2e_test.go`,
`models_tab_test.go`, `agent_flow_test.go` construct tab states directly and never
reference the enum ordering.

No other package references the `Tab` constants — `grep` across `*.go` shows all
usages are confined to `internal/tui`. There is no persisted "active tab" in
`.archon/config.yaml`, so no config/serialization migration is required.

## Affected Areas
- `internal/tui/model.go` — enum reorder (`21-28`), default tab (`92`), header
  label slice (`253`); optionally reorder the two `switch` blocks for readability.
- `internal/tui/model_test.go` — update the three order/default assertions
  (`37-38`, `91-92`, `100-101`).
- `internal/tui/e2e_test.go` — update the full-cycle default assertion (`118-128`).

## Approaches

1. **Reorder the enum + label slice, set default to AgentTab** — Move `AgentTab`
   to the first `iota` position so the enum becomes
   `AgentTab=0, ModelsTab=1, JudgeTab=2, MutationTab=3, PlaywrightTab=4`; reorder
   the header slice to `{"Agent","Models","Judge","Mutation Testing","Playwright"}`;
   set `NewModel` default to `AgentTab`. Update the coupled assertions in the two
   test files.
   - Pros: matches the desired mental model (enum order == visual order == landing
     tab); smallest conceptual surface; the two `switch` bodies keep working
     untouched because they are name-keyed.
   - Cons: three source spots plus tests must stay aligned; a future contributor
     could still desync the label slice from the enum.
   - Effort: Low.

2. **Introduce a single ordered tab-descriptor table** — Replace the parallel enum
   + label slice with one `[]tabDef{name, render, update}` slice iterated by index,
   removing the two `switch` blocks.
   - Pros: eliminates the parallel-list desync risk permanently; order lives in one
     place.
   - Cons: larger refactor touching routing and rendering; expands the diff well
     beyond the reorder ask; higher review cost for PR 1 of a chained series where
     scope discipline matters.
   - Effort: Medium.

## Recommendation

**Approach 1.** The request is a focused reorder, and this is PR 1 of 3 chained
changes under a 400-line review budget, so scope discipline matters. Approach 1
keeps the diff small (one production file, two test files) and low-risk. The
name-keyed `switch` blocks mean the only true order-dependencies are the enum, the
label slice, and the default-tab constant — all edited together. Capture the
"single-descriptor-table" refactor (Approach 2) as a follow-up note rather than
bundling it here.

One decision the proposal must make explicit: **the default landing tab.** The
literal request ("pick the agent, then configure the models") implies the TUI
should OPEN on the Agent tab, i.e. `NewModel` default becomes `AgentTab`. If we
only reorder the enum but leave `activeTab: ModelsTab`, the TUI would still open on
Models (now the second tab), which contradicts the flow. Recommend defaulting to
`AgentTab`.

## Risks
- **Parallel-list desync**: the enum (`21-28`) and the header label slice (`253`)
  are independent literals; if only one is reordered, labels render against the
  wrong content. Mitigation: edit both in the same change and rely on the
  order/default assertions in the tests as the guard.
- **Default-tab intent ambiguity**: reordering the enum silently changes what
  `activeTab: ModelsTab` means only if someone assumed `0`. Here it is a named
  constant, so the risk is the opposite — the default will NOT move unless we
  explicitly set it to `AgentTab`. The proposal must state the intended landing tab.
- **Stale test assertions**: `model_test.go` (`37-38`, `91-92`, `100-101`) and
  `e2e_test.go` (`118-128`) hard-assert the default tab and its tab-key neighbor;
  these WILL fail until updated. They are the safety net that proves the reorder
  and default change landed correctly.
- **Chained-PR coupling**: this is PR 1 of 3. If later PRs (models config flow)
  assume a specific tab index, keep the enum change isolated and documented so the
  chain rebases cleanly.

## Open Questions
1. **Default landing tab** — Confirm the TUI should OPEN on the Agent tab
   (`activeTab: AgentTab`). Recommended: yes, to match "pick the agent first".
2. **Rest-of-tabs order** — After Agent, is the intended order
   `Agent, Models, Judge, Mutation Testing, Playwright` (Models second, matching
   "then configure the models")? The prompt lists exactly this order; confirming
   locks the enum + label layout.
3. **Refactor appetite** — Keep the minimal reorder (Approach 1) for this PR and
   defer the single-descriptor-table refactor (Approach 2) to a follow-up? Given
   the 400-line budget and chained-PR strategy, recommended: yes.

## Ready for Proposal
Yes. The change is well-shaped and low-risk. The orchestrator should present the
executive summary and confirm the three open questions (default = AgentTab,
post-agent order = Models/Judge/Mutation/Playwright, minimal-reorder scope) at the
Human Review Gate before advancing to `propose`.
