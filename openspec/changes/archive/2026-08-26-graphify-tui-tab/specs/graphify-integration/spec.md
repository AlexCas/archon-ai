# Delta for graphify-integration

<!-- [[graphify-integration]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

> Session preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off · Graphify=off.

All scenarios for each requirement are in `graphify-integration.feature` alongside this file.

## ADDED Requirements

### Requirement: R-19 TUI Tab

The harness MUST provide a "Graphify" tab in `archon tui` (`internal/tui/graphify_tab.go`)
that exposes all five `config.Graphify` fields for interactive editing and saving,
mirroring the Impeccable tab structure.

Focus order (bools-first, `graphifyFocusCount = 5`):

| Focus | Field | Control |
|-------|-------|---------|
| 0 | `enabled` | toggle (Enter/Space) |
| 1 | `auto_install` | toggle (Enter/Space) |
| 2 | `semantic` | toggle (Enter/Space) |
| 3 | `version` | textinput |
| 4 | `output_dir` | textinput |

`applyToConfig` MUST coerce a blank `version` input to `DefaultGraphifyVersion`
(`"v0.9.45"`) and a blank `output_dir` input to `DefaultGraphifyOutputDir`
(`".archon/graphify"`). MUST NOT persist `""` for either field.

The tab MUST be wired into `model.go` at all nine canonical sites: `Tab` iota
constant, `Model` field, `NewModel` ctor, two `setWidth` fan-outs
(`WindowSizeMsg` + `agentInitDoneMsg`), key-dispatch switch, `agentInitDoneMsg`
rebuild, `saveConfig` applyToConfig fan-out, `renderTabs` label slice,
`renderTabContent` switch case. No live install probe; no blocking verdict —
parity with the Impeccable tab.

`TestModel_Update_ShiftTabWrapsFromAgent` MUST be updated (Shift+Tab from AgentTab
now wraps to GraphifyTab, the new last tab). `TestModel_renderTabs_Order` MUST be
updated (append `"Graphify"` to the expected label list). `TestGraphifyTabState_ApplyToConfig`
MUST be added to `model_test.go` alongside `TestImpeccableTabState_ApplyToConfig`.

## MODIFIED Requirements

### Requirement: R-05 Preflight Group G

(Previously: the mapping paragraph read only "The `--graphify` flag at init time sets the same value." — no reference to the `archon tui` tab.)

`internal/initcmd/templates.go` MUST add group G after group F in both
`orchestratorRules` blocks and the preflight section. Group G MUST: be its own
arrow-key `AskUserQuestion`; ask in Spanish "¿Activar Graphify para análisis de
grafo de código?"; pre-select "No (recomendado)"; include a mapping paragraph —
"Group G maps to `graphify.enabled` in `.archon/config.yaml`. The `--graphify`
flag at init time or the Graphify tab in `archon tui` set the same value."; and
update the preamble from "A–F"/"six" to "A–G"/"seven".

Root `CLAUDE.md` MUST carry the identical mapping paragraph (hand-edit only;
MUST NOT use `archon init --force` — `templates-go-drift` constraint).
