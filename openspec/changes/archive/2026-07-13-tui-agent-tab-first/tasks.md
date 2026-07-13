# Tasks: TUI Opens on the Agent Tab

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~30–40 (additions + deletions) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR (PR 1 of 3 chained series) |
| Delivery strategy | force-chained |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Reorder enum + label + default; update tests | PR 1 | Self-contained; base = master |

---

## Phase 1: Core Reorder — `internal/tui/model.go`

- [x] 1.1 In `model.go:21-28`, reorder the `Tab` iota const block: `AgentTab=0`, `ModelsTab=1`, `JudgeTab=2`, `MutationTab=3`, `PlaywrightTab=4`, `tabCount=5`.
- [x] 1.2 In `model.go:253`, replace the `tabs` string slice with `[]string{"Agent", "Models", "Judge", "Mutation Testing", "Playwright"}` (index-aligned with the new enum).
- [x] 1.3 In `model.go:92`, change `activeTab: ModelsTab,` to `activeTab: AgentTab,`.

## Phase 2: Test Updates — `internal/tui/model_test.go` and `e2e_test.go`

- [x] 2.1 `model_test.go:37-38` — change assertion from `ModelsTab` to `AgentTab` (NewModel default tab).
- [x] 2.2 `model_test.go:91-92` — change assertion from `JudgeTab` to `ModelsTab` (one Tab press from Agent advances to Models).
- [x] 2.3 `model_test.go:100-101` — change assertion from `ModelsTab` to `AgentTab` (Shift+Tab back to default).
- [x] 2.4 `e2e_test.go:126-127` — change assertion from `ModelsTab` to `AgentTab` (full-cycle wrap returns to new default).

## Phase 3: Verification

- [x] 3.1 Run `go build ./...` — must exit 0 with no errors.
- [x] 3.2 Run `go test ./internal/tui/...` — all four updated assertions must pass; no other tests may regress.

## Phase 4: Unplanned Fix (discovered during apply)

- [x] 4.1 `model_test.go` `TestModel_Update_Save` sent `alt+s` (not the bound `ctrl+s`) and only
      passed before by coincidence: the old default tab (`ModelsTab`) routed the unmatched key to
      a textinput, whose `Update` always returns a non-nil blink cmd. With the new default
      (`AgentTab`), the same key falls through `agentTabState.update` and returns `nil`, so the
      test's `cmd == nil` assertion failed — exposing a latent test bug, not a product regression.
      Fixed by sending `tea.KeyMsg{Type: tea.KeyCtrlS}` so the test exercises the actual Save
      binding regardless of active tab.

## Phase 5: Verify-Driven Amendment — Close Gherkin Coverage Gap

`verify` returned PASS WITH WARNINGS: 2 of 5 Gherkin scenarios in
`specs/tui-tabs/tui-tabs.feature` lacked a scenario-specific test (order/wrap
were only correct by source inspection). No production code changed — the
reorder was already implemented and all tests were green. Added exactly the
two missing tests to close the gap to 5/5.

- [x] 5.1 Added `TestModel_renderTabs_Order` in `model_test.go` — asserts the
      rendered tab header emits `Agent`, `Models`, `Judge`, `Mutation Testing`,
      `Playwright` in that order (via `strings.Index` ordering, robust to
      lipgloss/ANSI styling). Covers scenario "Header renders tabs in correct
      order" (@happy).
- [x] 5.2 Added `TestModel_Update_ShiftTabWrapsFromAgent` in `model_test.go` —
      builds a fresh `NewModel` (default `AgentTab`), sends one `Shift+Tab`,
      and asserts the active tab wraps to `PlaywrightTab`. Covers scenario
      "Shift+Tab wraps from Agent to Playwright" (@edge).
- [x] 5.3 Verified `go build ./...`, `go vet ./...`, and
      `go test ./internal/tui/...` all pass with the two new tests green and
      zero regressions. No production code files modified.
