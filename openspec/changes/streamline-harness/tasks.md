# Tasks: Streamline the Harness

## Review Workload Forecast

| Dimension | Value |
|---|---|
| Estimated changed lines | ~1 600 (removals dominate) |
| 800-line budget risk | Yes — PR 3 is the at-risk slice |
| Chain strategy | Feature Branch Chain |
| Tracker branch | `feature/streamline-harness` |
| Slices | 5 PRs (PR 3 has a prepared 3a/3b contingency split; PR 5 has a 5a/5b split) |

> **Chain strategy: feature-branch-chain.** Tracker branch `feature/streamline-harness`
> accumulates all slices and is the only branch that merges to master. PR1 base = tracker;
> PR2 base = PR1 branch; PR3 base = PR2 branch; PR4 base = PR3 branch; PR5 base = PR3 branch
> (PR4 and PR5 are independent of each other). Apply starts with PR1 only.

---

## Suggested Work Units

| PR | Scope | Key files | Est. lines |
|---|---|---|---|
| PR1 | Remove Impeccable Go plumbing | `config.go`, `config_test.go`, `cmd/archon/config.go`, `main.go`, `init.go`, `impeccable_tab.go` (delete), `model.go`, `model_test.go`, `display.go` + test | ~350 removed |
| PR2 | Remove Graphify Go plumbing + skill/spec | Mirror PR1 files + `graphify_tab.go` (delete), `embed_test.go`, `skills/graphify/` (delete), `openspec/specs/graphify-integration/` (delete) | ~350 removed |
| PR3 | Engram + Impeccable/Graphify out of the skill layer | `persistence-contract.md`, `engram-convention.md` (delete), all 8 phase skills + aux skills, `harness-judge/SKILL.md` | ~400–600 removed |
| PR4 | Preflight prose default | `internal/initcmd/templates.go`, `templates_test.go`, `CLAUDE.md`, `AGENTS.md`, `README.md` | ~200 changed |
| PR5 | Single judge + bugfix track | `.claude/agents/archon-judge.md`, `harness-judge/SKILL.md`, `harness-workflow/SKILL.md`, `sdd-spec/SKILL.md`, `openspec/specs/harness-judge/`, `openspec/specs/harness-bugfix-track/`, `openspec/specs/harness-workflow/` (terminal ordering), `CLAUDE.md` routing hint, `sdd-init` track param | ~300 changed |

---

## PR1 — Remove Impeccable Go Plumbing

**Depends on:** none (greenfield slice from the tracker branch)
**Estimated line-delta:** ~350 lines removed
**Capabilities covered:** partial `[[harness-init]]` — Impeccable config/CLI/init/TUI/status invariants

### Config removal — `internal/config/config.go`

- [x] **[P1-1]** Delete the `Impeccable` struct (`:46-58`).
- [x] **[P1-2]** Delete `ValidImpeccableSeverities` var and the `ValidateImpeccableSeverity` func (`:81-94`).
- [x] **[P1-3]** Delete the `Impeccable Impeccable` field from the `Config` struct (`:111`).
- [x] **[P1-4]** Delete the Impeccable severity normalize + validate block in `Load()` (`:143-151`, 9 lines).
- [x] **[P1-5]** Delete `Impeccable: c.Impeccable` from `Clone()` (`:175`).

### Config test — `internal/config/config_test.go`

- [x] **[P1-6]** Remove the `Impeccable` fields from the `TestConfig_CloneRoundtrip` fixture (`:265-269`) so the roundtrip compiles.
- [x] **[P1-7]** Delete `TestImpeccable_DefaultsAndValidation` (`:458-503`).

### CLI config — `cmd/archon/config.go`

- [x] **[P1-8]** Delete the six `impeccable.*` cases in `setConfigValue` (`:243-267`).
- [x] **[P1-9]** Delete the five `impeccable.*` cases in `getConfigValue` (`:350-359`).
- [x] **[P1-10]** Remove every `impeccable.*` key from the `(supported: …)` error strings at `:324` and `:385`; verify that the trimmed supported list still includes `models.*`, `playwright.*`, `mutation_testing.enabled`, `security.enabled`, `security.profile`.

### CLI main — `cmd/archon/main.go`

- [x] **[P1-11]** Delete `impeccableFlag` var decl (`:87`).
- [x] **[P1-12]** Delete `Impeccable: impeccableFlag` from the `Options{}` literal (`:176`).
- [x] **[P1-13]** Delete the `cmd.Flags().BoolVar(...)` registration for `--impeccable` (`:210`).

### Init — `internal/initcmd/init.go`

- [x] **[P1-14]** Delete `Impeccable bool` from `Options` (`:30-33`).
- [x] **[P1-15]** Delete `opts.Impeccable` arg from the `buildConfig(...)` call (`:91`).
- [x] **[P1-16]** Delete the `impeccable bool` param from `buildConfig` (`:224`) and the `Impeccable: config.Impeccable{...}` block it sets (`:252-254`).

### Init test — `internal/initcmd/init_test.go`

- [x] **[P1-17]** Delete `TestBuildConfig_ImpeccableFlag` (`:621-639`).

### TUI — `internal/tui/impeccable_tab.go`

- [x] **[P1-18]** Delete `internal/tui/impeccable_tab.go` entirely.

### TUI — `internal/tui/model.go`

- [x] **[P1-19]** Delete `ImpeccableTab` from the iota const block (`:29`). `SecurityTab` remains at index 5; `GraphifyTab` shifts from 7→6, `tabCount` drops 8→7. See TODO(PR2) comment in model.go.
- [x] **[P1-20]** Delete the `impeccableTab` field from the model struct (`:52`).
- [x] **[P1-21]** Delete `newImpeccableTabState` initialization in `NewModel` (`:115`).
- [x] **[P1-22]** Delete the `setWidth` call for `impeccableTab` in the `WindowSizeMsg` handler (`:137`).
- [x] **[P1-23]** Delete `case ImpeccableTab:` in the key-routing switch (`:188-193`).
- [x] **[P1-24]** Delete the rebuild and resize calls for `impeccableTab` in `agentInitDoneMsg` (`:225-226`).
- [x] **[P1-25]** Remove `"Impeccable"` from the tab-label slice (`:297`).
- [x] **[P1-26]** Delete `case ImpeccableTab:` in `renderTabContent` (`:328-329`).
- [x] **[P1-27]** Delete the `applyToConfig` call for `impeccableTab` in `saveConfig` (`:370`).

### TUI test — `internal/tui/model_test.go`

- [x] **[P1-28]** Delete `TestImpeccableTabState_ApplyToConfig*` (`:381-440`).
- [x] **[P1-29]** Update any test asserting `tabCount` value (was 8; after both PR1 + PR2 it will be 6 — leave a TODO comment here if only PR1 is applied; the iota shift for `ImpeccableTab` removal at `:29` also requires fixing any hard-coded tab indices that followed it in tests).

### Status display — `internal/status/display.go` + test

- [x] **[P1-30]** Delete the "Impeccable (Design Language)" status block (`:53-65`).
- [x] **[P1-31]** Delete `TestDisplay_Impeccable` (`:137-186`).

### PR1 checklist

- [x] `go build ./...` is green.
- [x] `go vet ./...` is clean.
- [x] `go test ./...` is green (no reference to removed symbols).
- [x] `grep -r 'impeccable\|Impeccable\|ImpeccableTab' internal/ cmd/` returns no hits in Go source (only test-assertion strings that document the expected error message are acceptable).
- [x] Diff is comfortably under 800 lines (638 lines changed: 53 inserted, 585 deleted).
- [ ] PR description references the `streamline-harness` change and `[[harness-init]]`.

---

## PR2 — Remove Graphify Go Plumbing + Skill/Spec

**Depends on:** PR1 merged to the tracker branch
**Estimated line-delta:** ~350 lines removed
**Capabilities covered:** `[[graphify-integration]]` (RETIRED) + remaining `[[harness-init]]` TUI/init invariants (six-tabs, Security-last, Graphify init surface)

### Config removal — `internal/config/config.go`

- [x] **[P2-1]** Delete the `Graphify` struct (`:60-70`).
- [x] **[P2-2]** Delete `DefaultGraphifyVersion` and `DefaultGraphifyOutputDir` consts (`:72-79`).
- [x] **[P2-3]** Delete the `Graphify Graphify` field from the `Config` struct (`:112`).
- [x] **[P2-4]** Delete the Graphify pre-seed lines in `Load()` (`:136-137`).
- [x] **[P2-5]** Delete `Graphify: c.Graphify` from `Clone()` (`:176`).

### Config test — `internal/config/config_test.go`

- [x] **[P2-6]** Remove the `Graphify` fields from the `TestConfig_CloneRoundtrip` fixture (`:265-269`) and delete any Graphify default/validation test analogue.

### CLI config — `cmd/archon/config.go`

- [x] **[P2-7]** Delete the five `graphify.*` cases in `setConfigValue` (`:269-294`).
- [x] **[P2-8]** Delete the five `graphify.*` cases in `getConfigValue` (`:360-369`).
- [x] **[P2-9]** Remove every `graphify.*` key from the `(supported: …)` error strings at `:324` and `:385`.

### CLI main — `cmd/archon/main.go`

- [x] **[P2-10]** Delete `graphifyFlag` var decl (`:88`).
- [x] **[P2-11]** Delete `Graphify: graphifyFlag` from the `Options{}` literal (`:177`).
- [x] **[P2-12]** Delete the `cmd.Flags().BoolVar(...)` registration for `--graphify` (`:211`).

### Init — `internal/initcmd/init.go`

- [x] **[P2-13]** Delete `Graphify bool` from `Options` (`:30-33`).
- [x] **[P2-14]** Delete `opts.Graphify` arg from the `buildConfig(...)` call (`:91`).
- [x] **[P2-15]** Delete the `graphify bool` param from `buildConfig` (`:224`) and the `Graphify: config.Graphify{...}` block it sets (`:255-258`).

### Init test — `internal/initcmd/init_test.go`

- [x] **[P2-16]** Delete any `TestBuildConfig_GraphifyFlag` test analogue.

### TUI — `internal/tui/graphify_tab.go`

- [x] **[P2-17]** Delete `internal/tui/graphify_tab.go` entirely.

### TUI — `internal/tui/model.go`

- [x] **[P2-18]** Delete `GraphifyTab` from the iota const block (`:30`). After this deletion, `SecurityTab` is at index 5, `tabCount` = 6 — the final post-change state.
- [x] **[P2-19]** Delete the `graphifyTab` field from the model struct (`:53`).
- [x] **[P2-20]** Delete `newGraphifyTabState` initialization in `NewModel` (`:116`).
- [x] **[P2-21]** Delete the `setWidth` call for `graphifyTab` in the `WindowSizeMsg` handler (`:138`).
- [x] **[P2-22]** Delete `case GraphifyTab:` in the key-routing switch (`:194-197`).
- [x] **[P2-23]** Delete the rebuild and resize calls for `graphifyTab` in `agentInitDoneMsg` (`:234-235`).
- [x] **[P2-24]** Remove `"Graphify"` from the tab-label slice (`:297`) — result is six labels ending with `"Security"`.
- [x] **[P2-25]** Delete `case GraphifyTab:` in `renderTabContent` (`:330-331`).
- [x] **[P2-26]** Delete the `applyToConfig` call for `graphifyTab` in `saveConfig` (`:371`).

### TUI test — `internal/tui/model_test.go`

- [x] **[P2-27]** Update the prev-tab-from-Agent test (`:109-122`): change the expected terminal tab from `GraphifyTab` to `SecurityTab`.
- [x] **[P2-28]** Delete any Graphify tab-state test analogue.
- [x] **[P2-29]** Update every test that asserts `tabCount` to expect 6 (was 8 before PR1+PR2; adjust for any partial update left by P1-29).
- [x] **[P2-30]** Update any test that enumerates the tab label list to the six-label set: `Agent, Models, Judge, Mutation Testing, Playwright, Security`.

### Status display — `internal/status/display.go` + test

- [x] **[P2-31]** Delete the Graphify status block.
- [x] **[P2-32]** Delete the Graphify display test.

### Embedded skill — `skills/embed_test.go`

- [x] **[P2-33]** Remove `"graphify"` from the expected embedded-skill list at `:29`.

### Skill directory deletion

- [x] **[P2-34]** Delete `skills/graphify/` directory and all contents.
- [x] **[P2-35]** Delete `skills/impeccable/` directory and all contents (confirm impeccable is not in `embed_test.go` expected list before editing; no list edit needed).

### Live spec retirement (apply-time deletion)

- [x] **[P2-36]** Delete `openspec/specs/graphify-integration/` directory (spec.md + any .feature). This is the physical deletion the `[[graphify-integration]]` delta spec records. Note: the archived copy in `openspec/changes/streamline-harness/specs/graphify-integration/` is the change delta and stays in place.

### PR2 checklist

- [x] `go build ./...` is green.
- [x] `go vet ./...` is clean.
- [x] `go test ./...` is green, including `embed_test.go` (graphify no longer in expected list).
- [x] `grep -r 'graphify\|Graphify\|GraphifyTab' internal/ cmd/ skills/` returns no hits in Go source or skill files (remaining refs in `templates.go`, `sdd-explore/SKILL.md`, `sdd-tasks/SKILL.md`, `chained-pr/SKILL.md` are PR3/PR4 scope — left intentionally).
- [x] TUI iota block ends with `SecurityTab=5, tabCount=6` — verify against `model.go`.
- [x] `openspec/specs/graphify-integration/` no longer exists.
- [x] `skills/graphify/` and `skills/impeccable/` no longer exist.
- [ ] Diff is 1,367 lines (28 ins / 1,339 del) — over 800 due to live spec files (spec.md + .feature ~527 lines). Go+skill+embed changes alone are ~840 lines. Note for orchestrator review.
- [ ] PR description references the `streamline-harness` change, `[[graphify-integration]]` (RETIRED), and `[[harness-init]]`.

---

## PR3 — Engram + Impeccable/Graphify Out of the Skill Layer

**Depends on:** PR1 and PR2 merged to the tracker branch
**Estimated line-delta:** ~400–600 lines removed
**Capabilities covered:** `[[sdd-phase-skills]]` (both: OpenSpec-only persistence + No Impeccable/Graphify hooks)

> **Contingency split:** If the measured diff approaches or crosses 800 lines, split PR3
> into 3a and 3b as described below. Both sub-groups are independently reviewable; 3b
> touches the same files as 3a only for distinct edits (rebase, not logic dependency).

---

### Sub-group 3a — Engram/none Removal (split point if needed)

#### `skills/_shared/persistence-contract.md` rewrite

- [x] **[P3a-1]** Delete the `artifact_store.mode` selection line (`:5`) and the "ASK the user which mode" handshake (`:7`) and the default-resolution line (`:9`).
- [x] **[P3a-2]** Collapse the mode table (`:20-40`) to a single "OpenSpec is the sole store" statement: artifacts are files under `openspec/changes/{change-name}/`, git-tracked, team-shareable.
- [x] **[P3a-3]** Delete the `engram`, `hybrid`, and `none` rows from the mode table, the "engram mode limitation" note (`:29-31`), and the hybrid section (`:44-52`).
- [x] **[P3a-4]** Replace the state-persistence row with the single OpenSpec form: write/read `openspec/changes/{change-name}/state.yaml`.
- [x] **[P3a-5]** Delete the `none`-mode "return only" path (`:69,:74`) and the "default to none" fallback.
- [x] **[P3a-6]** Change the trailing prompt templates (`:103,:123`) that echo `{engram|openspec|hybrid|none}` to `OpenSpec`.

#### `skills/_shared/engram-convention.md` deletion

- [x] **[P3a-7]** Delete `skills/_shared/engram-convention.md` entirely.
- [x] **[P3a-8]** Search for inbound references to `engram-convention.md` in `persistence-contract.md`, all phase skills, and `chained-pr`; remove each link.

#### Per-phase skill engram/none removal (all 8 phase skills + aux)

Apply the three-step pattern per skill: (a) delete the "Artifact store mode (engram|openspec|hybrid|none)" input line, (b) delete every `IF mode is engram/hybrid/none` branch block keeping only the OpenSpec body inline, (c) drop the now-redundant `IF mode is openspec` guard.

- [x] **[P3a-9]** `sdd-explore/SKILL.md` — engram/none anchors at `:46-48`, `:53-56`.
- [x] **[P3a-10]** `sdd-propose/SKILL.md` — engram/none anchors at `:41,:47,:49,:83,:89,:184`.
- [x] **[P3a-11]** `sdd-spec/SKILL.md` — engram/none anchors at `:40,:46,:48,:76,:93,:275`.
- [x] **[P3a-12]** `sdd-design/SKILL.md` — engram/none anchors at `:40,:46,:48,:102,:191`.
- [x] **[P3a-13]** `sdd-apply/SKILL.md` — engram/none anchors at `:41,:49,:51,:98,:132,:230`.
- [x] **[P3a-14]** `sdd-verify/SKILL.md` — engram/none mode branches (locate and delete).
- [x] **[P3a-15]** `sdd-tasks/SKILL.md` — engram/none mode branches (locate and delete).
- [x] **[P3a-16]** `sdd-archive/SKILL.md` — engram/none anchors at `:40,:48,:50,:59,:107,:144,:172,:191,:217,:237,:258`.
- [x] **[P3a-17]** `sdd-init/SKILL.md` — engram anchors at `:38,:40,:42,:51,:53,:75-76` (drop the track-less engram handling).
- [x] **[P3a-18]** `sdd-onboard/SKILL.md` — `:37` (`engram|openspec|hybrid|none` → `OpenSpec`).
- [x] **[P3a-19]** `skills/chained-pr/SKILL.md` and `skills/chained-pr/references/chaining-details.md` — remove any engram references.
- [x] **[P3a-20]** `skills/_shared/session-status-contract.md` — remove any engram mode branches.

---

### Sub-group 3b — Impeccable/Graphify Hook Removal from Skills (split point if needed)

- [x] **[P3b-1]** `sdd-explore/SKILL.md` — delete Step 3c Impeccable recommendation block (`:110-118`) and the Graphify code-graph consumption block.
- [x] **[P3b-2]** `sdd-spec/SKILL.md` — delete Impeccable annotation notes (`:112-115,:312-315`). Confirm that `@security` abuse-case hooks are untouched.
- [x] **[P3b-3]** `sdd-design/SKILL.md` — delete Step 2b Impeccable block (`:64-88`).
- [x] **[P3b-4]** `sdd-apply/SKILL.md` — delete Step 4c Impeccable/Graphify block (`:193-208`) and the corresponding Rules line (`:315`).
- [x] **[P3b-5]** `sdd-verify/SKILL.md` — delete Impeccable presence check (`:44-50`).
- [x] **[P3b-6]** `sdd-tasks/SKILL.md` — delete Impeccable pass task (`:253-254,:315-321`) and Graphify Leiden consumption block.
- [x] **[P3b-7]** `skills/harness-judge/SKILL.md` — delete the `impeccable.enabled` Hard Rule (`:23,:30ff`), the Step-1 impeccable config read (`:71-77,:82`), Step 3c entire Impeccable gate block (`:125-165`), the Step-4 result-table impeccable column (`:172-173`), the Output-Contract "### Impeccable Gate" section (`:270-278`), and the Impeccable Error-Handling rows (`:307-310`) and Rules bullet (`:321-323`).

---

### PR3 checklist

- [x] `go build ./...` is green (skill edits are markdown; this verifies no accidental Go touch).
- [x] `go test ./...` is green.
- [x] `grep -rInE 'engram|Engram|hybrid|none.*mode|impeccable|Impeccable|graphify|Graphify|mem_search|mem_get|mem_save|mem_update|topic_key|observation' skills/` returns no hits (excluding the declarative "OpenSpec is the sole persistence mode…" line in `persistence-contract.md`).
- [x] `skills/_shared/engram-convention.md` does not exist.
- [x] `skills/_shared/persistence-contract.md` names OpenSpec as the sole mode.
- [x] `harness-judge/SKILL.md` contains no Impeccable gate, no `impeccable.enabled` reference.
- [x] Security `@security` hooks in `sdd-spec/SKILL.md` are present and unmodified — do a targeted diff to confirm.
- [ ] Diff is 830 lines (3a+3b both done; 30 lines over the 800 soft budget — see note below).
- [ ] PR description references the `streamline-harness` change and `[[sdd-phase-skills]]`.

---

## PR4 — Preflight Prose Default

**Depends on:** PR3 merged to the tracker branch
**Estimated line-delta:** ~200 lines changed
**Capabilities covered:** `[[harness-workflow]]` "Preflight is a Standard Default Expressed as Prose"; `[[harness-init]]` "Artifact store is always OpenSpec"

### Template — `internal/initcmd/templates.go`

- [x] **[P4-1]** Delete lines 56-88 of `orchestratorSections` (the A–G arrow-key question list). Replace with a concise "SDD Session Preflight" prose block stating the fixed default: Ritmo=interactivo · Artefactos=OpenSpec · PRs=ask-but-default-chained · Revisión=800, applied silently.
- [x] **[P4-2]** Keep lines 90-91 (group E — Playwright; reword to drop the "group E" framing).
- [x] **[P4-3]** Delete lines 93-97 (group F mapping, `impeccable.enabled`).
- [x] **[P4-4]** Delete lines 99-103 (group G mapping, `graphify.enabled`).
- [x] **[P4-5]** Replace lines 105-109 (hard-gate rules referencing "seven per-group questions" and "§openspec§, §engram§, or §both§") with the prose deviation contract: default applied without asking; surface a single scoped question only on (a) estimate > 800 lines, (b) trivially small change, or (c) explicit user request.
- [x] **[P4-6]** In `orchestratorRulesClaude` (`:187-199`) and `orchestratorRulesOpencode` (`:203-215`): delete Rule 8 (`impeccable.enabled`) and Rule 9 (`graphify.enabled`); keep Rule 7 (playwright.enabled); renumber trailing rules after the two deletions (former Rule 10 → Rule 8, former Rule 11 → Rule 9).

### Template test — `internal/initcmd/templates_test.go`

- [x] **[P4-7]** Remove Group F/G assertions (`:149-151`) and the Group B engram/"Ambos" assertions.
- [x] **[P4-8]** Remove Rule 8 / Rule 9 assertions (`:205`).
- [x] **[P4-9]** Add assertion: rendered output contains the prose default text (e.g. "Ritmo=interactivo" or "SDD Session Preflight").
- [x] **[P4-10]** Add assertion: rendered output does NOT contain "Impeccable", "Graphify", "Engram", or "Ambos".

### Rendered orchestrator files — `CLAUDE.md` and `AGENTS.md` (repo root)

- [x] **[P4-11]** Edit `CLAUDE.md` directly to mirror the template edits: replace the 7-group block with prose preflight, drop groups B-engram/F/G, drop rules 8–9.
- [x] **[P4-12]** Edit `AGENTS.md` with the same prose preflight changes.
- [x] **[P4-13]** Confirm that the committed CLAUDE.md and AGENTS.md match what `RenderClaudeMD`/`RenderAgentsMD` would produce from the updated template (run `archon update` or invoke the render functions in a test if available).

### README — `README.md`

- [x] **[P4-14]** Strip the `--impeccable` flag description (`:103`).
- [x] **[P4-15]** Strip the Impeccable and Graphify TUI-tab descriptions (`:162`).
- [x] **[P4-16]** Strip the `impeccable:` and `graphify:` config snippets (`:190-192,:260-262`).

### PR4 checklist

- [x] `go build ./...` is green.
- [x] `go vet ./...` is clean.
- [x] `go test ./...` is green, including `templates_test.go` with the new prose assertions.
- [x] `grep -i 'impeccable\|graphify\|engram\|ambos' CLAUDE.md AGENTS.md README.md internal/initcmd/templates.go` returns no hits.
- [x] `CLAUDE.md` and `AGENTS.md` contain the prose "SDD Session Preflight" block and the silent-default + deviation rule.
- [x] Playwright group E is still present in the template output.
- [x] Diff is comfortably under 800 lines (377 lines: 112 ins / 265 del).
- [ ] PR description references the `streamline-harness` change, `[[harness-workflow]]`, and `[[harness-init]]`.

---

## PR5 — Single Judge + Bugfix Track

**Depends on:** PR3 merged to the tracker branch (independent of PR4)
**Estimated line-delta:** ~300 lines changed
**Capabilities covered:** `[[harness-judge]]`, `[[harness-workflow]]` (Phase State Machine track-aware + Single-Judge terminal ordering), `[[harness-bugfix-track]]`

> **Contingency split:** If the measured diff approaches or crosses 800 lines, split PR5
> at the natural seam into 5a (single judge) and 5b (bugfix track). They share no source
> files except the already-authored spec deltas. Apply 5a first, then 5b on top.

---

### Sub-group 5a — Single Judge (split point if needed)

#### `.claude/agents/archon-judge.md` rewrite

- [ ] **[P5a-1]** Replace the body of `archon-judge.md` with a single-review brief: read all files modified by the change plus the change's spec/design; evaluate spec compliance, design coherence, and code quality; return one verdict (`pass`/`fail`) with issues. Preserve frontmatter `model: claude-opus-4-8`. Do NOT instruct the agent to run `judgment-day`, launch a second/blind judge, or apply fixes.

#### `skills/harness-judge/SKILL.md`

- [ ] **[P5a-2]** Rewrite the `description` frontmatter field and the Purpose line (`:13`): "dual adversarial review" → "single focused review"; drop "judgment-day" from the trigger prose.
- [ ] **[P5a-3]** In Hard Rules (`:24`): change "delegate the dual review to archon-judge" → "delegate a single focused review to archon-judge"; keep "do NOT run inline on the orchestrator's model".
- [ ] **[P5a-4]** Rename Step 2 (`:84-91`) from "Delegate Dual Review" to "Delegate Single Review"; remove "dual adversarial", "two blind judges", "synthesis", and the "archon-judge invokes judgment-day internally" sentence; rewrite to: delegate one focused review, capture the returned `pass`/`fail` + issues.
- [ ] **[P5a-5]** Replace every remaining "judgment-day" token in the skill body (outside the step that was already rewritten) with "the single judge / archon-judge review" to keep pass/fail wording consistent.
- [ ] **[P5a-6]** Verify that the following are NOT changed: Step 0 (`judge.enabled` gate), Step 3 mutation gate, Step 3b Playwright gate, Step 4 evaluate, Step 6 3-retry re-apply loop, the Structured Feedback output format.

#### Live spec — `openspec/specs/harness-judge/spec.md`

- [ ] **[P5a-7]** Apply the change's authored `specs/harness-judge/spec.md` delta to the live spec: replace the dual-review requirement with the "Single Focused Judge in the SDD Path" MODIFIED requirement and append the "judgment-day Remains Standalone and Untouched" ADDED requirement.

#### `harness-workflow` terminal ordering note

- [ ] **[P5a-8]** Confirm that `skills/harness-workflow/SKILL.md` already reflects (or add the single sentence): in the full track, exactly one judge runs after verify and before archive; no second judge phase is inserted. (No further harness-workflow SKILL change is required for this sub-group; track-awareness is 5b.)

---

### Sub-group 5b — Bugfix Track (split point if needed)

#### `skills/harness-workflow/SKILL.md` — track-aware PHASE_ORDER

- [ ] **[P5b-1]** In Step 1 (`:86`): after parsing `phase`/`status`, also parse `track`. Missing → resolve as `full`. Unknown value → return `blocked` with reason `unknown track: {v} (supported: full, bugfix)`.
- [ ] **[P5b-2]** Replace the single hard-coded `PHASE_ORDER` array (`:93`) with a track-selected definition:
  - `bugfix`: `[explore, spec, apply, verify, archive]`
  - `full` (default): `[explore, propose, spec, design, tasks, apply, verify, judge, archive]`
- [ ] **[P5b-3]** Confirm all downstream `.index()` transition logic (`:94-98`) is unchanged — it operates on whichever array `track` selected.
- [ ] **[P5b-4]** In phase-skipping-prevention (`:62`): add note that the "N → N+1 only" rule applies within the selected array; a phase not in the selected track is treated as unreachable → `blocked` with a message naming both the phase and the track.
- [ ] **[P5b-5]** Add one Hard-Rule sentence: "`judge.enabled` governs the full track only; the bugfix track never runs judge."

#### `sdd-spec/SKILL.md` — bugfix detection + 3-section scoped spec

- [ ] **[P5b-6]** At entry, read `track` from `state.yaml`. Branch on `track: bugfix`:
  - Write a single `openspec/changes/{change-name}/spec.md` with EXACTLY three sections: `## Bug` (observed vs expected, 1–3 sentences), `## Fix Criteria` (flat checklist 2–5 bullets, `- [ ] X returns Y when Z`), `## Non-Regression` (1–2 bullets naming existing behavior that must not break).
  - Explicitly bypass the "Gherkin Feature Files (MANDATORY)" block (`:95-122`) and the Rules that mandate a `{domain}.feature` per domain (`:303-304`).
  - Emit no `.feature` file.
- [ ] **[P5b-7]** Confirm `track: full` (or absent) path is unchanged — existing full spec flow runs without modification.

#### `sdd-init/SKILL.md` — track parameter

- [ ] **[P5b-8]** Add support for a `track` parameter (default `full`); write the resolved track value into `state.yaml` at change creation. Document that natural-language routing ("Es un bug") is the primary trigger; no CLI flag is required.

#### CLAUDE.md / AGENTS.md — routing hint

- [ ] **[P5b-9]** In the Vague Request Guard region of `CLAUDE.md` (and `AGENTS.md` / `templates.go`), add a brief note: when the user's request is a bug report (not a feature), the orchestrator asks "¿Es un bug o una nueva funcionalidad?" and, on "bug", initializes the change with `track: bugfix` via `sdd-init`.

#### Live specs — new and updated

- [ ] **[P5b-10]** Create `openspec/specs/harness-bugfix-track/spec.md` (and copy `harness-bugfix-track.feature`) from the change's authored `specs/harness-bugfix-track/` delta into the live spec directory. Create the directory if it does not exist.
- [ ] **[P5b-11]** Apply the `specs/harness-workflow/spec.md` delta to the live spec at `openspec/specs/harness-workflow/spec.md`: apply the MODIFIED "Phase State Machine" requirement (track-aware), the MODIFIED "Phase Skipping Prevention" (track-aware), and the ADDED "Single Judge in the Full-Track Terminal Ordering" requirement; add the ADDED "Preflight is a Standard Default Expressed as Prose" requirement (if not already applied by PR4 — coordinate to avoid double-apply).

---

### PR5 checklist

- [ ] `go build ./...` is green.
- [ ] `go vet ./...` is clean.
- [ ] `go test ./...` is green.
- [ ] `archon-judge.md` contains no reference to `judgment-day` invocation.
- [ ] `harness-judge/SKILL.md` contains no "dual adversarial", "two blind judges", or "judgment-day" invocation (judgment-day's own skill file is byte-for-byte unchanged).
- [ ] `harness-workflow/SKILL.md` PHASE_ORDER is track-selected; `bugfix` sequence is `[explore, spec, apply, verify, archive]`.
- [ ] `sdd-spec/SKILL.md` bugfix branch emits no `.feature` file and uses the 3-section format.
- [ ] `openspec/specs/harness-bugfix-track/` exists with `spec.md` and the `.feature` file.
- [ ] `openspec/specs/harness-judge/spec.md` reflects single-judge requirement and judgment-day standalone requirement.
- [ ] CLAUDE.md / AGENTS.md contain the bug-vs-feature routing hint.
- [ ] Diff is under 800 lines if kept whole; if split, each of 5a and 5b is under 800 lines individually.
- [ ] PR description references the `streamline-harness` change, `[[harness-judge]]`, `[[harness-workflow]]`, and `[[harness-bugfix-track]]`.

---

## Cross-PR: Grep Sweep

After each PR apply, run the following grep sweeps to confirm the relevant surface is clean. Include these in the PR checklist.

| PR | Sweep command |
|---|---|
| PR1 | `grep -r 'impeccable\|Impeccable\|ImpeccableTab' internal/ cmd/` |
| PR2 | `grep -r 'graphify\|Graphify\|GraphifyTab' internal/ cmd/ skills/ openspec/specs/` |
| PR3 | `grep -r 'engram\|hybrid\|none.*mode\|Impeccable\|Graphify' skills/` |
| PR4 | `grep -i 'impeccable\|graphify\|engram\|ambos' CLAUDE.md AGENTS.md README.md internal/initcmd/templates.go` |
| PR5 | `grep -r 'judgment-day' .claude/agents/archon-judge.md skills/harness-judge/` |
