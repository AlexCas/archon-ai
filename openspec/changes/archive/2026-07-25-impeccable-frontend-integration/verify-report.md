# Verification Report

Change: [[impeccable-gate]] — Impeccable Frontend-Integration, **PR1 slice (Go wiring)**
Artifacts: [proposal](proposal.md), [spec](specs/impeccable-gate/spec.md),
[feature](specs/impeccable-gate/impeccable-gate.feature), [design](design.md), [tasks](tasks.md)
Persistence mode: OpenSpec (file). Execution mode: interactive.
Branch: `feat/impeccable-pr1-go-wiring` (changes applied, NOT committed).
Scope note: PR1 is Go wiring only for the opt-in `impeccable.enabled` gate; it must be
fully inert at the zero-value default (enabled=false). PR2/PR3 (skills, phase hooks,
judge gate, templates) are out of scope for this slice.

**Verdict: PASS WITH WARNINGS** (one out-of-scope working-tree deletion; no defect in
the impeccable implementation).

---

## Completeness — PR1 tasks (T1.1–T6.1)

All PR1 tasks marked `[x]` in [tasks.md](tasks.md) are implemented and independently
confirmed. PR2 (§7–§12) and PR3 (§13–§15) tasks remain `[ ]` by design — out of the PR1
slice. No PR1 implementation task is unchecked.

| Phase | Tasks | Status |
|-------|-------|--------|
| 1 Config struct + Clone + Load default/validate | 1.1–1.6 | Implemented |
| 2 CLI get/set | 2.1–2.4 | Implemented |
| 3 Init flag threading | 3.1–3.2 | Implemented |
| 4 TUI tab | 4.1–4.6 | Implemented |
| 5 Status block | 5.1 | Implemented |
| 6 Doc update | 6.1 | Implemented |

---

## Build / Vet / Tests (independently executed, this session)

| Command | Result |
|---------|--------|
| `go build ./...` | PASS (exit 0) |
| `go vet ./...` | PASS (exit 0) |
| `go test ./... -count=1` | PASS — 12/12 packages `ok`, 0 failures |

Full suite output (all `ok`):

```
ok  cmd/archon            ok  internal/agent        ok  internal/config
ok  internal/initcmd      ok  internal/mapgen       ok  internal/models
ok  internal/opencode     ok  internal/scaffold     ok  internal/status
ok  internal/tui          ok  internal/version      ok  skills
```

Criterion-specific tests (all PASS):
`TestConfig_CloneRoundtrip`, `TestImpeccable_DefaultsAndValidation`
(absent_block_defaults, invalid_severity_rejected), `TestImpeccableTabState_ApplyToConfig`,
`TestImpeccableTabState_ApplyToConfig_BlankSeverityFallback`,
`TestModel_Update_ShiftTabWrapsFromAgent`, `TestModel_renderTabs_Order`,
`TestDisplay_Impeccable` (disabled/enabled), `TestConfigCmd_UnknownKeyListsImpeccableKeys`,
`TestConfigCmd_ImpeccableSetGet`, `TestConfigCmd_ImpeccableSeverityInvalid`,
`TestBuildConfig_ImpeccableFlag` (on/off).

---

## Acceptance-criteria matrix

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| 1 | Build & tests | **PASS** | build/vet exit 0; `go test ./...` 12/12 ok (above) |
| 2 | Zero-value inertness | **PASS** | see below |
| 3 | Clone completeness | **PASS** | see below (mutation-proven) |
| 4 | Severity validation | **PASS** | see below |
| 5 | CLI keys (get+set, both error strings) | **PASS** | see below |
| 6 | Init flag | **PASS** | see below |
| 7 | TUI tab | **PASS** | see below |
| 8 | Status block | **PASS** | see below |
| 9 | Scope hygiene | **PASS** (impeccable code); WARNING for unrelated deletion | see below |

### 1. Build & tests — PASS
Independently ran `go build ./...`, `go vet ./...`, `go test ./... -count=1`; all exit 0,
12/12 packages `ok`. No reliance on prior claims.

### 2. Zero-value inertness — PASS
- `internal/config/config.go:100-126` `Load()`: pre-seeds ONLY `c.Judge.Enabled = true`
  (line 109). `Impeccable` is NOT pre-seeded — it takes the YAML/zero value. An absent
  `impeccable:` block → `Enabled=false`. Confirmed by
  `TestImpeccable_DefaultsAndValidation/absent_block_defaults`: absent block →
  `Enabled==false` and `Severity=="block-deterministic"` (normalized at read).
- No existing flow branches on `Impeccable` in PR1: no skill consumes the flag yet
  (PR2/PR3), so behavior is unchanged when off. Status prints only `Enabled: false` when
  disabled (`display.go:55-56`); init default is `false` (`init.go:250-252`,
  `TestBuildConfig_ImpeccableFlag/impeccable_off`).
- `Enabled`/`AutoInstall` carry no `omitempty` (`config.go:53-54`), mirroring
  `Playwright.Enabled`/`Security.Enabled`, so a disabled block renders explicitly on disk
  as the precedent requires.

### 3. Clone completeness — PASS (mutation-proven)
- `config.go:147` `Clone()` copies `Impeccable: c.Impeccable` — a full value copy (all
  scalars, no maps/slices), so deep copy is complete.
- `TestConfig_CloneRoundtrip` fixture (`config_test.go:231-237`) sets all FIVE fields
  non-zero: `Enabled:true, AutoInstall:true, Severity:"block-all",
  ProductPath:"PRODUCT.md", DesignPath:"DESIGN.md"`, compared via `reflect.DeepEqual`
  (`config_test.go:257`).
- **Mutation proof (executed this session):** temporarily removing the `Impeccable` line
  from `Clone()` made `TestConfig_CloneRoundtrip` FAIL loudly
  (`Impeccable:{Enabled:false...}` got vs `Enabled:true...` want). Restored → PASS. The
  loud contract from exploration.md holds: a dropped field on `archon update` is caught.

### 4. Severity validation — PASS
- Single source of truth: `config.ValidateImpeccableSeverity` (`config.go:66-73`) and
  `ValidImpeccableSeverities` (`config.go:61`).
- `Load()` path: normalizes empty → `"block-deterministic"` (`config.go:118-120`) BEFORE
  validating (`config.go:121-123`), so an absent block is never rejected as `""`. Invalid
  value → error naming the value + three valid options
  (`TestImpeccable_DefaultsAndValidation/invalid_severity_rejected`, fixture
  `severity: foobar`, asserts error contains `"foobar"` and all three
  `ValidImpeccableSeverities`).
- CLI `set` path: `cmd/archon/config.go:204-209` calls the same
  `config.ValidateImpeccableSeverity`; `TestConfigCmd_ImpeccableSeverityInvalid` asserts
  non-zero exit + value + three options.

### 5. CLI keys — PASS
- All 3 keys with values (`enabled`, `auto_install`, `severity`) plus `product_path`,
  `design_path` are in BOTH `setConfigValue` (`config.go:190-215`) and `getConfigValue`
  (`config.go:248-257`). (Note: 5 keys total; the 3 with typed handling are
  enabled/auto_install/severity.)
- Both key-list error strings (`config.go:228` and `config.go:266`) contain all five
  impeccable keys and are **byte-identical** — verified by de-duplicating the two
  `(supported: ...)` substrings: exactly 1 unique string.
  `TestConfigCmd_UnknownKeyListsImpeccableKeys` + `TestConfigCmd_ImpeccableSetGet` pass.

### 6. Init flag — PASS
- `cmd/archon/main.go`: `impeccableFlag bool` (line 84), `Impeccable: impeccableFlag`
  into `initcmd.Options{}` (line 172), `BoolVar(&impeccableFlag, "impeccable", false, ...)`
  (line 205).
- `internal/initcmd/init.go`: `Options.Impeccable bool` (line 31), threaded into
  `buildConfig(...)` (line 89), `buildConfig` signature adds `impeccable bool` (line 222),
  and `Impeccable: config.Impeccable{Enabled: impeccable}` with NO `Severity`
  (lines 250-252) so `Load()` normalizes on next read.
- `TestBuildConfig_ImpeccableFlag` (on/off) passes; the off case asserts `Severity == ""`
  (not pre-seeded), matching design §4.

### 7. TUI tab — PASS
- Enum (`internal/tui/model.go:21-30`): `... SecurityTab, ImpeccableTab, tabCount`.
  `ImpeccableTab` (28) appended after `SecurityTab` (27), before `tabCount` (29). All
  existing indices `AgentTab..SecurityTab` (0..5) are preserved; `tabCount` shifted up 1.
- All 10 lockstep sites present in `model.go`: struct field (50), constructor (112),
  resize setWidth (133), Update routing (183-185), reload (215), post-reload setWidth
  (223), renderTabs label (285), renderTabContent (316-317), saveConfig apply (356).
- `internal/tui/impeccable_tab.go` new: `impeccableTabState` (5 fields + `focused`),
  `impeccableFocusCount = 5`, `newImpeccableTabState`/`update`/`refocus`/`view`
  (title "Impeccable (Design Language) Configuration")/`applyToConfig`/`setWidth`.
- Blank-severity write-back fallback → `"block-deterministic"` (`impeccable_tab.go:162-166`),
  covered by `TestImpeccableTabState_ApplyToConfig_BlankSeverityFallback` (PASS).
- `TestModel_Update_ShiftTabWrapsFromAgent` (wrap target now `ImpeccableTab`) and
  `TestModel_renderTabs_Order` (labels include "Impeccable") both PASS.

### 8. Status block — PASS
- `internal/status/display.go:53-65`: prints an "Impeccable (Design Language)" block. When
  `Enabled: false` → only `Enabled: false` (fields inside the `if Enabled` guard, lines
  56-63). When enabled → Severity always; Product Path / Design Path only when non-empty.
- `TestDisplay_Impeccable` covers both disabled (`Enabled: false` only) and enabled
  (all fields) cases; PASS.

### 9. Scope hygiene — PASS (impeccable code); WARNING (unrelated deletion)
- No PR2/PR3 surfaces touched (verified):
  - `skills/impeccable/SKILL.md`: absent (correct).
  - `internal/initcmd/templates.go`: 0 impeccable references, not in `git diff` — no
    group F.
  - `internal/initcmd/templates_test.go`: not in diff.
  - `CLAUDE.md` / `AGENTS.md`: no impeccable/group-F additions (AGENTS.md only gets the
    one-line `--impeccable` doc per T6.1; README likewise).
  - `skills/harness-judge/SKILL.md` and all other `skills/`: untouched (`git status
    --short skills/` empty).
- `idea-player.md` deletion: the file exists in HEAD (315-line file added in commit
  784e771 "feat: complete ai-orchestration-harness implementation") and is deleted only in
  the working tree — a stray, pre-existing, unrelated deletion. It is NOT part of the
  impeccable change.

---

## Spec compliance matrix (PR1-scoped Gherkin scenarios)

Only PR1-relevant scenarios are in scope; judge-gate / phase-hook / skill / templates
scenarios belong to PR2/PR3 and are marked DEFERRED (not verifiable in this slice).

| Scenario (feature) | Status | Covering test / evidence |
|--------------------|--------|--------------------------|
| Zero-value config is fully inert | PASS | `TestImpeccable_DefaultsAndValidation/absent_block_defaults`; Load pre-seeds only Judge |
| Clone preserves all Impeccable fields | PASS | `TestConfig_CloneRoundtrip` (mutation-proven) |
| CloneRoundtrip fixture catches missing Clone wiring | PASS | mutation proof: dropping Clone line → test FAIL |
| Invalid severity value rejected at config load | PASS | `TestImpeccable_DefaultsAndValidation/invalid_severity_rejected` |
| severity defaults to block-deterministic | PASS | `absent_block_defaults` asserts `"block-deterministic"` |
| Init with `--impeccable` enables the gate | PASS | `TestBuildConfig_ImpeccableFlag/impeccable_on` |
| Init without `--impeccable` leaves gate disabled | PASS | `TestBuildConfig_ImpeccableFlag/impeccable_off` (Enabled false, Severity "") |
| CLI set impeccable.enabled to true | PASS | `TestConfigCmd_ImpeccableSetGet` |
| CLI rejects invalid severity via set | PASS | `TestConfigCmd_ImpeccableSeverityInvalid` |
| Get unknown impeccable key shows updated key list | PASS | `TestConfigCmd_UnknownKeyListsImpeccableKeys` (both strings identical) |
| Impeccable tab renders current config | PASS | `newImpeccableTabState` seeds all 5 from cfg |
| TUI save persists Impeccable changes | PASS | `TestImpeccableTabState_ApplyToConfig` |
| Tab count and order tests pass with new tab | PASS | `TestModel_Update_ShiftTabWrapsFromAgent`, `TestModel_renderTabs_Order` |
| Status shows Impeccable as disabled | PASS | `TestDisplay_Impeccable/disabled_shows_only_Enabled:_false` |
| Status shows Impeccable as enabled with details | PASS | `TestDisplay_Impeccable/enabled_shows_all_fields` |
| Judge-gate / phase-hook / skill / template scenarios | DEFERRED | PR2/PR3 scope — not in this slice |

---

## Design coherence

Implementation matches design §1–§5 for PR1 exactly: struct field order (§2.1, after
Security), `omitempty` policy (§2.1), Load normalize-then-validate order (§2.2), shared
exported validator (§2.2/§3.1), Clone value copy (§2.4), CloneRoundtrip fixture with all
5 fields (§2.5), CLI cases + byte-identical error strings (§3), flag threading with no
Severity pre-set (§4), TUI insertion last-before-`tabCount` and the 10 lockstep sites
(§5.1–§5.3), and the disabled-vs-enabled status rendering (§1). The design's residual open
question #4 (TUI blank-severity write-back) was resolved in favor of the
`block-deterministic` fallback (§5.2 recommendation), and is tested.

No design deviations found in PR1 scope.

---

## Issues

**CRITICAL:** none.

**WARNING:**
1. `idea-player.md` is deleted in the working tree but is unrelated to the impeccable
   change (it predates this branch, added in 784e771). It should NOT be included in the
   PR1 commit. Recommend `git restore idea-player.md` before committing, or handle the
   deletion in a separate change so PR1 stays scoped to the impeccable wiring.

**SUGGESTION:**
1. `SESSION_STATUS.md` is untracked at the repo root (expected during an active session);
   it will be archived with the change per the session-status contract at archive time.

---

## Final verdict

**PASS WITH WARNINGS.** PR1 (Go wiring) fully satisfies all nine acceptance criteria and
all PR1-scoped Gherkin scenarios with real runtime evidence, including a mutation proof of
the Clone contract. The only warning is an unrelated working-tree deletion
(`idea-player.md`) that must be kept out of the PR1 commit. Ready for the judge phase once
the stray deletion is addressed.

---
---

# Verification Report — PR2 slice (Orchestration prose)

Change: [[impeccable-gate]] — Impeccable Frontend-Integration, **PR2 slice (orchestration prose)**
Artifacts: [proposal](proposal.md), [spec](specs/impeccable-gate/spec.md),
[feature](specs/impeccable-gate/impeccable-gate.feature), [design](design.md), [tasks](tasks.md)
Persistence mode: OpenSpec (file). Execution mode: interactive.
Branch: `feat/impeccable-pr2-orchestration` (changes applied, NOT committed).
Scope note: PR2 is orchestration PROSE only — a new thin skill `skills/impeccable/SKILL.md`
plus flag-gated hooks woven into six existing phase skills. It must be inert unless
`impeccable.enabled: true` (config gate landed in PR1). PR3 (judge gate `harness-judge`,
`templates.go`/`templates_test.go`, repo `CLAUDE.md`) is out of scope for this slice.

**Verdict: PASS WITH WARNINGS** (one unrelated working-tree deletion carried over from
PR1; no defect in the PR2 prose).

---

## Completeness — PR2 tasks (T7.1–T12.2)

All PR2 tasks are marked `[x]` in [tasks.md](tasks.md) and independently confirmed:

| Task | Artifact | Status |
|------|----------|--------|
| 7.1 | `skills/impeccable/SKILL.md` (NEW thin skill) | DONE |
| 8.1 | `skills/sdd-design/SKILL.md` (read-only hook, Step 2b) | DONE |
| 9.1 | `skills/sdd-apply/SKILL.md` (Step 4c + Rules line) | DONE |
| 10.1 | `skills/sdd-verify/SKILL.md` (advisory NOTE hook) | DONE |
| 11.1 | `skills/sdd-tasks/SKILL.md` (Impeccable pass task) | DONE |
| 12.1 | `skills/sdd-explore/SKILL.md` (group F recommendation, Step 3c) | DONE |
| 12.2 | `skills/sdd-spec/SKILL.md` (`@design` prose note) | DONE |

PR1 tasks (T1.1–T6.1) remain `[x]` (landed prior). PR3 tasks (T13.1–T14.4, T15.x) remain
`[ ]` by design — out of the PR2 slice.

---

## Build / Tests — real runtime evidence

Commands run at repo root on branch `feat/impeccable-pr2-orchestration`:

```
$ go build ./...
BUILD_EXIT=0

$ go test ./... -count=1
ok  github.com/archon-ai/archon/cmd/archon        0.037s
ok  github.com/archon-ai/archon/internal/agent    0.005s
ok  github.com/archon-ai/archon/internal/config   0.009s
ok  github.com/archon-ai/archon/internal/initcmd  0.031s
ok  github.com/archon-ai/archon/internal/mapgen   0.012s
ok  github.com/archon-ai/archon/internal/models   0.003s
ok  github.com/archon-ai/archon/internal/opencode 0.003s
ok  github.com/archon-ai/archon/internal/scaffold 0.008s
ok  github.com/archon-ai/archon/internal/status   0.005s
ok  github.com/archon-ai/archon/internal/tui       1.601s
ok  github.com/archon-ai/archon/internal/version  0.003s
ok  github.com/archon-ai/archon/skills            0.004s
TEST_EXIT=0
```

`skills` package (`embed_test.go`) passes with the new skill present:

```
=== RUN   TestFS_ContainsSkills
--- PASS: TestFS_ContainsSkills (0.00s)
=== RUN   TestFS_SKILLMdAccessible
--- PASS: TestFS_SKILLMdAccessible (0.00s)
PASS
ok  github.com/archon-ai/archon/skills   0.003s
```

**skill_count 24→25 (glob) confirmed empirically.** `skills/embed.go:5` embeds
`*/SKILL.md all:_shared` — a directory glob, so `impeccable/SKILL.md` is auto-embedded
with no code edit. Enumerating glob-matched skill dirs (excluding `_shared`) yields **25**.
`SkillCount` is computed at runtime as `len(extracted)` / `len(res.Extracted)`
(`internal/initcmd/init.go:236,283`, `internal/initcmd/update.go:135`), never a hardcoded
constant, so the count reflects 25 automatically. `embed_test.go` uses a subset allowlist
(not a hard count), so adding a skill cannot break it. The literal `SkillCount` values in
`templates_test.go` (23/15/42/10) are synthetic fixtures unrelated to the real inventory
and PR2 does not touch that file.

---

## Spec / criterion compliance matrix

| # | Criterion | Verdict | Evidence |
|---|-----------|---------|----------|
| 1 | Build/tests green; skills pkg + embed_test pass; count 24→25 | **PASS** | `go build`/`go test` exit 0 above; 25 glob dirs; `embed.go:5` glob; runtime `len(extracted)` |
| 2 | New skill correctness (thin, two-surface, detect, severity, blocked msgs) | **PASS** | `skills/impeccable/SKILL.md` (see below) |
| 3 | Design hook: flag-gated, read-only, no detector, no design.md overwrite | **PASS** | `skills/sdd-design/SKILL.md` Step 2b |
| 4 | Apply hook: agent-behavioral `/impeccable <verb>`, not npx shell-out | **PASS** | `skills/sdd-apply/SKILL.md` Step 4c |
| 5 | Verify hook: advisory NOTE, not CRITICAL | **PASS** | `skills/sdd-verify/SKILL.md` lines 44–50 |
| 6 | tasks/explore/spec hooks | **PASS** | `sdd-tasks`, `sdd-explore` Step 3c, `sdd-spec` `@design` note |
| 7 | Every hook flag-gated on `impeccable.enabled`; inert when off | **PASS** | Each hook opens with the enabled gate + "skip entirely when false" |
| 8 | Scope hygiene: no PR3 files touched | **PASS** | `git diff --name-only HEAD` (see below) |

### Criterion 2 — new skill `skills/impeccable/SKILL.md`

- **Frontmatter valid, matches repo convention.** `name: impeccable`, `description` with a
  `Trigger:` clause, `license: MIT`, `metadata` block (blank line + `version: "1.0"`) —
  parallels reference-style peers (`skill-registry`, `branch-pr`). The extra
  `scope: reference` field is additive and harmless. As a reference (non-phase) skill it
  correctly omits `disable-model-invocation`/`delegate_only`, consistent with peer
  reference skills.
- **Two surfaces clearly distinguished** (lines 24–36): a table separating the `npx CLI`
  (`install`/`update`/`detect`, shelled out) from the `/impeccable <verb>` agent
  slash-commands (`init` + 23 verbs, agent-run), plus an explicit warning: *"Never shell
  out to `npx impeccable init` or `npx impeccable <verb>` — those are agent-run slash
  commands."* (line 34).
- **detect invocation** documented (lines 50–74): `npx impeccable detect --json .` from
  target root; `--json` REQUIRED; `--no-config`/`--no-inline-ignores` documented as knobs.
  Exit-code/JSON interpretation is explicit (lines 62–74): never rely on exit code for
  pass/fail, parse `--json`, unparseable → advisory (not hard-fail), non-zero + no output
  → `blocked`.
- **Severity mapping table** (lines 77–84): `block-deterministic` (default) / `block-all` /
  `advisory` mapped across deterministic vs LLM-critique buckets — matches spec + design §6.
- **Verbatim blocked messages** present (lines 86–107): node/npx-missing message
  (lines 90–92) and the `auto_install: false` package-missing message (lines 98–102), both
  matching design §6.3 intent.
- **THIN**: lines 18–22 and Rules (lines 120–130) explicitly forbid reimplementing the
  58-rule detector or LLM critique in Go/prose; the only invocation mechanisms are
  `npx impeccable <subcommand>` or agent-run `/impeccable <verb>`. No detector logic is
  inlined.

### Criterion 3 — design hook (`skills/sdd-design/SKILL.md` Step 2b)

Flag-gated (`Only if ... impeccable.enabled: true`). Strictly READ-ONLY reference to
`PRODUCT.md`/`DESIGN.md`. Explicit prohibitions present: *"This hook MUST NOT run `npx
impeccable detect`, MUST NOT run any `/impeccable` slash command, and MUST NOT generate or
overwrite the SDD `design.md` artifact or any Impeccable output file."* Distinguishes the
two paths (`PRODUCT.md`/`DESIGN.md` at target root vs `design.md` in the change dir). Closes
with "skip this step entirely" when disabled.

### Criterion 4 — apply hook (`skills/sdd-apply/SKILL.md` Step 4c)

AGENT-BEHAVIORAL: *"Run the relevant `/impeccable <verb>` ... as **agent slash commands
inside this session** — this is AGENT-BEHAVIORAL, NOT an `npx` shell-out. Do not attempt to
shell out to `/impeccable`; only `install`/`update`/`detect` are real `npx` commands."*
Flag-gated + frontend-scoped; adds a Rules line ("When `impeccable.enabled`, run Impeccable
design verbs on frontend-affecting changes during apply (Step 4c)").

### Criterion 5 — verify hook (`skills/sdd-verify/SKILL.md`)

Advisory only: *"This is a NOTE, not a CRITICAL — missing artifacts are reported as a
recommendation, never a blocker."* Explicitly defers detection to judge (*"Do NOT run `npx
impeccable detect` here; that gate belongs to `harness-judge` (Step 3c)"*) and skips entirely
when disabled. Matches spec scenario "Verify reports missing Impeccable artifacts as
advisory".

### Criterion 6 — tasks / explore / spec hooks

- **sdd-tasks**: Phase-4 task-tree entry + Rules line emit an "Impeccable pass" task only
  when `impeccable.enabled` AND the change touches frontend files; "omit this task entirely"
  when disabled/no-frontend. Task instructs `sdd-apply` to run `/impeccable <verb>`.
- **sdd-explore** Step 3c: on a `web` project-type determination, notes the orchestrator
  SHOULD recommend group F — "recommendation only, never an automatic activation"; does not
  set `impeccable.enabled`; excludes `not-web`/`unknown`. Output template gains a conditional
  group-F recommendation line.
- **sdd-spec**: `@design` is a lightweight PROSE note, NOT a new hard Gherkin tag — explicitly
  to "avoid coupling a new hard tag to the `@web` selector Playwright already owns"; omitted
  when disabled. Matches design §8.1 and spec "left to the design phase → prose".

### Criterion 7 — flag-gating

Every one of the seven hooks opens with an `impeccable.enabled: true` condition and closes
with an explicit "skip entirely / omit / no change to today's behavior when false" clause.
The thin skill itself states (lines 14–16) that no phase should read the file when the flag
is absent/false — the gate is fully inert when off.

### Criterion 8 — scope hygiene

`git diff --name-only HEAD` (tracked) shows only: `idea-player.md` (pre-existing,
unrelated), `tasks.md`, and the six phase skills (`sdd-apply`, `sdd-design`, `sdd-explore`,
`sdd-spec`, `sdd-tasks`, `sdd-verify`). Untracked: `skills/impeccable/` (new) and
`SESSION_STATUS.md`. A targeted grep for the PR3-guarded paths
(`harness-judge/SKILL.md`, `internal/initcmd/templates.go`, `templates_test.go`, repo
`CLAUDE.md`/`AGENTS.md`, `internal/config`, any impeccable `.go`) returned **none** — no
PR3/PR1 Go file was modified. PR2 boundary is clean.

---

## Gherkin scenario mapping (PR2-scoped, prose skills)

PR2 hooks are behavioral prose in phase skills; there is no Go test harness that executes
`sdd-*` skill prose, so these scenarios are verified by source inspection of the skill text
(the only possible evidence for prose hooks) rather than runtime tests:

| Scenario | Covering evidence | Status |
|----------|-------------------|--------|
| Skill is auto-embedded and inventoried (skill_count → 25) | `go test ./skills` PASS + 25 glob dirs + runtime `len(extracted)` | COMPLIANT (runtime) |
| Skill delegates detection to npx, not Go code | `impeccable/SKILL.md` Rules lines 120–130 (no detector reimplementation) | COMPLIANT (inspection) |
| Apply step invokes Impeccable on frontend changes when enabled | `sdd-apply` Step 4c + Rules line | COMPLIANT (inspection) |
| Verify reports missing Impeccable artifacts as advisory | `sdd-verify` NOTE-not-CRITICAL hook | COMPLIANT (inspection) |
| Design references PRODUCT.md/DESIGN.md; unchanged when disabled | `sdd-design` Step 2b | COMPLIANT (inspection) |
| Tasks — add Impeccable pass task | `sdd-tasks` Phase-4 entry + Rules | COMPLIANT (inspection) |
| Explore — recommend group F on web | `sdd-explore` Step 3c | COMPLIANT (inspection) |
| Spec — `@design` prose annotation | `sdd-spec` prose-note hook | COMPLIANT (inspection) |

Note: the judge-gate scenarios (detect execution, blocked-on-missing-node, gate output
contract) are PR3 (`harness-judge` Step 3c) and are intentionally NOT covered here.

---

## Design coherence — PR2

The seven hooks match design §7, §8, §8.1 exactly, including the three design corrections
vs the spec: (1) `init` and the design verbs are `/impeccable` slash commands, not `npx`;
(2) apply is agent-behavioral, not a shell-out; (3) detect is exit-code-tolerant (parse
`--json`). No design deviations found in PR2 scope.

---

## Issues — PR2

**CRITICAL:** none.

**WARNING:**
1. `idea-player.md` is deleted in the working tree but is unrelated to the Impeccable
   change (predates this branch). Same warning as PR1 — it must NOT be included in the PR2
   commit. Recommend `git restore idea-player.md` before committing.

**SUGGESTION:**
1. `skills/impeccable/SKILL.md` frontmatter carries a non-standard `scope: reference` key
   not used by peer skills. It is harmless (additive metadata) but consider dropping it or
   standardizing the convention across reference skills for consistency.
2. `SESSION_STATUS.md` is untracked at the repo root (expected during an active session);
   it will be archived with the change per the session-status contract at archive time.

---

## Final verdict — PR2

**PASS WITH WARNINGS.** PR2 (orchestration prose) fully satisfies all eight verification
criteria: build/tests green with the new skill embedded (count 24→25 confirmed), the thin
skill correctly distinguishes the two Impeccable surfaces and documents detect/severity/
blocked messages without reimplementing the detector, and all seven phase hooks are
flag-gated, correctly typed (design read-only, apply agent-behavioral, verify advisory),
and inert when disabled. Scope is clean — no PR3 or PR1 Go files touched. The only warning
is the unrelated `idea-player.md` deletion, which must be kept out of the PR2 commit. Ready
for the judge phase once the stray deletion is addressed.

---

# Verification Report — PR3 slice (Judge gate + preflight group F + docs sync)

Change: [[impeccable-gate]] — Impeccable Frontend-Integration, **PR3 slice (FINAL: judge
detection gate, preflight group F, templates/CLAUDE.md/AGENTS.md sync)**
Branch: `feat/impeccable-pr3-judge-gate-templates` (applied, NOT committed)
Mode: interactive · Artifact store: OpenSpec · Verified: 2026-07-24

Scope note: PR3 is the terminal slice. It wires the Impeccable detection gate into
`harness-judge` Step 3c, adds preflight group F to the generated templates and the repo
`CLAUDE.md`/`AGENTS.md`, and applies the follow-up count-wording consistency fix (five→six,
A–E→A–F). It also carries the one-sentence rescope of `skills/impeccable/SKILL.md`. Design
authority: design §6 (judge gate), §9 (templates/docs). All PR1 (§1–§6) and PR2 (§7–§12)
tasks were verified in the sections above.

## Completeness — PR3 tasks (T13.1–T14.4)

All PR3 implementation tasks are `[x]` in [tasks.md](tasks.md) and independently confirmed:

| Task | Description | Status |
|---|---|---|
| T13.1 | harness-judge Step 3c 8-step gate flow | done — SKILL.md:125–165 |
| T13.2 | "### Impeccable Gate" output section | done — SKILL.md:270–277 |
| T13.3 | verbatim blocked messages | done — SKILL.md:158–163, 307–308 |
| T13.4 | 5 touchpoints (intro/config/result-table/edge/output) | done — SKILL.md:13,29,71–82,172–178,303–309 |
| T14.1 | templates.go group F + mapping paragraph | done — templates.go:83–92 |
| T14.2 | Impeccable rule in both rule consts | done — templates.go:185,199 |
| T14.3 | templates_test.go assertions | done — templates_test.go:149–150,196–198,242–244 |
| T14.4 | repo CLAUDE.md manual mirror | done — CLAUDE.md:73–78,84,156 |

T15.2 ("Run `go test ./internal/initcmd/...` after PR3 lands") remains `[ ]` — it is the
post-land verification step itself, executed by this report (PASS below). It is a
verification/cleanup task, not a core implementation task, so it is not a blocker.

## Build / Vet / Tests (independently executed, this session)

Full tree, `-count=1`, from the applied-but-uncommitted working tree:

```
$ go build ./...   → exit 0 (no output)
$ go vet ./...     → exit 0 (no output)
$ go test ./... -count=1
ok  cmd/archon 0.038s · internal/agent · internal/config · internal/initcmd 0.031s
ok  internal/mapgen · internal/models · internal/opencode · internal/scaffold
ok  internal/status · internal/tui 1.607s · internal/version · skills
→ all 12 packages PASS, exit 0
```

Targeted (`internal/initcmd`, verbose):

```
--- PASS: TestTemplates_ContainSDDSessionPreflight (AGENTS.md, CLAUDE.md)
--- PASS: TestTemplates_FiveRules            (AGENTS.md, CLAUDE.md)
```

## Per-criterion results

| # | Criterion | Verdict |
|---|---|---|
| 1 | Build / vet / tests green (initcmd especially) | PASS |
| 2 | Judge gate (harness-judge Step 3c) | PASS |
| 3 | templates.go group F (both rule consts, renumbered) | PASS |
| 4 | templates_test.go assertions | PASS |
| 5 | CLAUDE.md + AGENTS.md mirror + count wording consistent | PASS |
| 6 | Scope hygiene (no PR1/PR2 files beyond the one rescope) | PASS |
| 7 | No stray preflight-count inconsistency | PASS |

### Criterion 1 — Build / vet / tests — PASS

All three commands exit 0; all 12 packages pass. `internal/initcmd` (templates_test.go)
passes including the two group-F/rule tests shown above.

### Criterion 2 — Judge gate (harness-judge Step 3c) — PASS

`skills/harness-judge/SKILL.md` implements the gate exactly per design §6:
- Flag-gated on `impeccable.enabled` (Step 3c header line 127 "Only if
  `impeccable.enabled: true` AND judgment-day passed"; hard rule line 29; Step 1 line 82
  "gates Step 3c. Default: `false`").
- Runs `npx impeccable detect --json .` from target-project root (line 141).
- Parses JSON for pass/fail; exit code used ONLY for ran-vs-crash/not-found (line 142–147,
  and Rule line 322: "NEVER relies on exit code alone (exit code is only used to detect
  tool crash / not-found)").
- node/npx missing → `blocked` with verbatim message (line 160–161, 307), never silent
  pass ("Never silent-pass" line 136).
- Deterministic violations → `fail` under default severity (line 149–150).
- LLM critique → advisory unless `block-all` (line 150–153).
- `### Impeccable Gate` output section present (line 270–277) with Status/Severity mode/
  Deterministic count/Advisory count/Details/blocked Reason.
- Result-table row added mirroring Playwright (line 172–178: 4th "impeccable gate" column,
  `fail or blocked` degrades overall exactly like Playwright).
- Error-handling rows mirror Playwright (line 307–309: node/npx-missing, package-missing,
  config-absent).
- All three severity values handled — `block-deterministic` (default), `block-all`,
  `advisory` — enumerated at line 148–153 and Rule 322. (grep count: block-deterministic ×4,
  block-all ×3, advisory ×6.)

### Criterion 3 — templates.go group F — PASS

`internal/initcmd/templates.go`:
- Group F preflight bullet added after group E (line 83–85): "F. Impeccable (Diseño de
  interfaz)" with recommended No / opt-in Sí.
- Group F mapping paragraph added after the group E mapping (line 90–94): "Group F maps to
  `impeccable.enabled` … The `--impeccable` flag … or the Impeccable tab in `archon tui`".
- New rule added in BOTH consts, byte-identical: `orchestratorRulesClaude` line 185 and
  `orchestratorRulesOpencode` line 199 — "7. When impeccable.enabled, run Impeccable
  subcommands during apply and the detection gate after judge passes".
- Downstream rules renumbered consistently in both: 7→8 (judge fail re-apply), 8→9 (commit
  authorship). No duplicate or missing numbers; the `--impeccable` flag / config struct /
  TUI tab the paragraph references all exist (config.go:52,90; internal/tui/impeccable_tab.go).

### Criterion 4 — templates_test.go — PASS

- `TestTemplates_ContainSDDSessionPreflight` now asserts `"F. Impeccable (Diseño de
  interfaz)"` and `"Group F maps to \`impeccable.enabled\`"` (lines 149–150), run against
  both AGENTS.md and CLAUDE.md renderings.
- `TestTemplates_FiveRules` shared-rules slice includes the new rule 7 and the renumbered
  8/9 (lines 191–198), and the negative guard was updated from "no rule 9" to "no rule 10 /
  exactly 9 rules" (lines 242–244).
- Both tests PASS (verbose output above).

### Criterion 5 — Repo CLAUDE.md + AGENTS.md mirror — PASS

- Group F question bullet in repo `CLAUDE.md` is byte-identical to the rendered templates.go
  string after the `§`→backtick substitution (automated `diff` returned IDENTICAL).
- Group F mapping paragraph byte-identical across templates.go (rendered) and CLAUDE.md.
- The new rule line collapses to a single unique string across templates.go (both consts),
  CLAUDE.md, and AGENTS.md (`grep -h … | sort -u` → one line).
- `AGENTS.md` correctly keeps its pre-existing legacy `F1/F2` answer-code question format
  (by design) while mirroring the mapping paragraph (AGENTS.md:66–71) and the rule
  (AGENTS.md:146–148). Verified only the paragraph + rule are mirrored there, not the
  arrow-key question shape.
- Count-wording consistency confirmed: `grep -E "all five|five per-group|group A–E|A–E|
  five choices|five questions"` across `CLAUDE.md`, `AGENTS.md`, `templates.go` → **none
  found**. The "A–F", "six per-group", "all six" wording is present in both templates.go
  (lines 58,98,100) and CLAUDE.md (lines 42,82,84).

### Criterion 6 — Scope hygiene — PASS

`git diff --name-only` against the PR2 tip (HEAD = ab2daac) shows exactly the intended set:
AGENTS.md, CLAUDE.md, internal/initcmd/templates.go, internal/initcmd/templates_test.go,
skills/harness-judge/SKILL.md, skills/impeccable/SKILL.md, plus tasks.md (progress
tracking) and the pre-existing idea-player.md deletion.
- No PR1 Go files touched: `grep -E "internal/(config|tui|status)/|cmd/archon/"` on the
  changed set → NONE.
- The only PR2 phase-hook skill touched is `skills/impeccable/SKILL.md`, and its diff is
  the single intended sentence: `"The judge gate is the ONLY place that executes
  Impeccable:"` → `"The judge gate (\`harness-judge\` Step 3c) is the place that executes
  Impeccable's detection:"`. No other phase-hook skill (sdd-design/apply/verify/tasks/
  explore/spec) was modified in PR3.

### Criterion 7 — No stray count inconsistency — PASS

Adding group F introduced no internal contradiction: all "five→six" and "A–E→A–F" call
sites were updated together (question intro, "ask the six per-group questions", "all six
choices"). No remaining "five"/"A–E" reference exists in any of the three synced docs
(grep evidence in Criterion 5). templates_test.go's rule-count guard is also consistent
(exactly 9 rules).

## Gherkin scenario mapping (PR3-scoped)

Source: `specs/impeccable-gate/impeccable-gate.feature`.

| Scenario | Covering evidence | Status |
|---|---|---|
| Judge gate passes when no violations (:129) | harness-judge Step 3c line 149–150 (`block-deterministic`, 0 violations → pass) + output section :270 | COVERED (prose) |
| Judge gate blocks on deterministic violations (:136) | Step 3c line 149; result table :174 fail column | COVERED (prose) |
| Judge gate skipped when disabled (:144) | flag gate line 127,134 "no invocation, no section, no column" | COVERED (prose) |
| Judge gate blocked when Node/npx absent (:151) | Step 3c line 136,160; error row :307; "Never silent-pass" :136 | COVERED (prose) |
| npx not found + auto_install false (:187) | Step 3c line 138–140; error row :308 verbatim message | COVERED (prose) |
| auto_install true triggers install (:195) | Step 3c line 138 "run `npx impeccable install` once, then continue" | COVERED (prose) |
| Generated CLAUDE.md includes group F (:205) | TestTemplates_ContainSDDSessionPreflight (F. Impeccable + mapping) — PASS at runtime | COVERED (test) |
| Generated CLAUDE.md includes Impeccable rule (:211) | TestTemplates_FiveRules asserts rule 7 in both variants — PASS | COVERED (test) |
| Template tests assert group F + rule (:217) | `go test ./internal/initcmd/...` — PASS | COVERED (test) |
| Skill delegates detection to npx (:231) | impeccable/SKILL.md Rules :122–127 (never reimplement; npx only) | COVERED (prose) |

The judge-gate/design/apply/auto_install scenarios are prose-behavioral (they describe
runtime harness behavior that executes only when `impeccable.enabled: true` with the
external `npx impeccable` tool present); their contract is verified by source inspection of
the skill instructions, consistent with the design's decision to keep Impeccable a thin
orchestration layer with no Go-side detector. The three template scenarios have real
passing Go tests.

## Design coherence — PR3

Matches design §6 (judge gate: 8-step flow, severity mapping, verbatim blocked messages,
result-table + output-contract mirroring Playwright) and §9 (group F question, mapping
paragraph, rule in both consts, test assertions, repo doc mirror). The impeccable/SKILL.md
rescope aligns the skill's earlier "ONLY place" claim with the now-authoritative
harness-judge Step 3c, removing a cross-file contradiction. No design deviations found.

## Issues — PR3

**CRITICAL:** none.

**WARNING:**
1. `idea-player.md` remains deleted in the working tree (carried from PR1/PR2). Unrelated to
   Impeccable (predates the branch; original deletion in commit 784e771). It must NOT be
   included in the PR3 commit — recommend `git restore idea-player.md` before committing.

**SUGGESTION:**
1. Repo `CLAUDE.md:158` still reads `- Skills: 24 (embedded via archon init)`, but the
   `impeccable` skill (added in PR2) makes the true embedded count 25 — the PR2 report
   confirmed the generated `{{.SkillCount}}` renders 25. This static line is OUT OF PR3
   SCOPE (PR3 T14.4 only mirrors the group F + rule additions, not the skill count), and it
   is templated dynamically at `archon init` time, so generated projects are correct. Still,
   consider bumping the repo CLAUDE.md line to 25 in a follow-up for repo self-consistency.
2. `SESSION_STATUS.md` is untracked at the repo root (expected during an active session); it
   will be archived with the change per the session-status contract at archive time.

---

## Final verdict — PR3

**PASS WITH WARNINGS.** PR3 (the final slice) satisfies all seven verification criteria:
build/vet/tests are green across all 12 packages with the templates tests passing; the
harness-judge Step 3c gate is fully specified (flag-gated, JSON-parsed, exit-code-only-for-
crash, verbatim blocked messages, all three severities, output section and result/error
rows mirroring Playwright); templates.go adds group F plus the new rule in both consts with
consistent renumbering; templates_test.go asserts all of it and passes; the repo CLAUDE.md
and AGENTS.md mirror the rendered strings byte-for-byte (with AGENTS.md keeping its legacy
question format by design); the five→six / A–E→A–F count wording is now consistent with no
stray "five"/"A–E" references anywhere; and scope is clean — no PR1 Go files and only the
single intended one-sentence rescope in the one PR2 skill. The sole warning is the unrelated
`idea-player.md` deletion, which must be kept out of the PR3 commit. Ready for the judge
phase once the stray deletion is addressed.
