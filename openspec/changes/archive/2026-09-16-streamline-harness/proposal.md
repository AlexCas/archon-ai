# Proposal: Streamline the Harness

<!-- Link convention: [[capability]] for capability identity; relative links for
     intra-change navigation. Full rule: skills/_shared/spec-vault.md. -->

## Intent

The harness is over-engineered. Every SDD session pays a 7-group preflight tax; every
phase skill carries dead engram/hybrid/none branches plus Impeccable/Graphify hooks
that are never exercised; the routine judge gate drags a dual-adversarial ceremony; and
there is a single rigid 9-phase track even for one-line bug fixes. This change removes
dead surface and ceremony so the default path is silent and fast, while preserving the
harness's ability to propose a deviation only when a change warrants it. Decisions were
fixed at the explore review gate; this proposal bakes them in.

## Scope

### In Scope
1. **Preflight → prose default.** Replace the 7-group gate with a fixed standard
   default expressed as CLAUDE.md/AGENTS.md prose (no new Go config struct): Ritmo=interactivo,
   Artefactos=OpenSpec, PRs=ask-but-default-chained, Revisión=800. Default path is silent;
   orchestrator surfaces a decision only on deviation (tiny change → suggest single-PR;
   estimate > 800 → flag budget).
2. **Remove Impeccable entirely** (34 touch points: Go config/CLI/TUI/init/display + tests,
   skills, templates, README, CLAUDE.md). Delete `impeccable_tab.go`; fix the `GraphifyTab`/
   tab-iota shift in tests.
3. **Remove Graphify entirely** (Go plumbing + `skills/graphify/` deleted + `embed_test.go`
   list updated + `openspec/specs/graphify-integration/spec.md` removed).
4. **Remove Engram and the `none` mode.** Artifacts are ALWAYS OpenSpec — group B collapses.
   Rewrite `_shared/persistence-contract.md` OpenSpec-only, delete `_shared/engram-convention.md`,
   strip engram/hybrid/none branches from all phase skills.
5. **Single judge in the SDD path.** `harness-judge` performs one focused review (spec
   compliance, design coherence, code quality → pass/fail). Drop the dual/blind ceremony
   from the SDD flow. `judgment-day` STAYS UNCHANGED as standalone opt-in.
6. **Bugfix track.** Add `track: bugfix` to `state.yaml` (back-compat; absent = `full`).
   Bugfix PHASE_ORDER = explore → spec (scoped) → apply → verify → archive. Judge ABSENT.
   Scoped spec is minimal (see Approach for the recommended format).

### Out of Scope
- `judgment-day` internals (untouched; remains invocable via "juzgar"/"dual review").
- Re-adding Impeccable or Graphify as external overlay skills now.
- A `Preflight` Go config struct (explicitly rejected — prose only).
- `archon route` advisory softening (Hotspot 7 — deferred).
- Config-file auto-migration tooling (leftover keys are ignored, see Risks).

## Capabilities

### New Capabilities
- `harness-bugfix-track`: `track` field on `state.yaml` + the 5-phase bugfix sequence and
  its scoped-spec contract. New `openspec/specs/harness-bugfix-track/spec.md`.

### Modified Capabilities
- `harness-workflow`: PHASE_ORDER becomes track-selected; state machine branches on `track`.
- `harness-judge`: single-pass review replaces dual-adversarial delegation in the SDD path.
- `graphify-integration`: removed (spec deleted — capability retired).

## Approach

Mechanical removals (concerns 2–4) precede behavior changes (1, 5, 6). Go plumbing for
Impeccable/Graphify is deleted first (unlocks skill cleanup), then engram/Impeccable/Graphify
branches are stripped from skills, then preflight prose, single judge, and bugfix track land.
`templates.go` is the single source of truth for CLAUDE.md/AGENTS.md — editing it regenerates both.

**Recommended bugfix scoped-spec format** (the one open question): a single-file
`spec.md` in the change folder with three short sections, NO Gherkin suite:
`## Bug` (observed vs expected, 1–3 sentences) · `## Fix Criteria` (a flat checklist of
2–5 verifiable acceptance bullets, e.g. "- [ ] `X` returns Y when Z") · `## Non-Regression`
(1–2 bullets naming existing behavior that must not break). `sdd-spec` in bugfix track is
instructed to emit exactly this shape and skip capability/requirement decomposition.

**Preflight deviation rule (prose):** default choices are applied silently. The orchestrator
asks a scoped one-line question ONLY when (a) estimate exceeds the 800-line budget, or (b) the
change is trivially small (suggest single-PR), or (c) the user explicitly asks to adjust.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/config.go`, `cmd/archon/{config,main}.go`, `internal/initcmd/init.go` | Removed | Impeccable + Graphify structs, CLI cases/flags, buildConfig params |
| `internal/tui/{impeccable_tab,graphify_tab}.go`, `model.go` | Removed/Modified | Delete tab files; drop iota entries; fix tab-index shift |
| `internal/status/display.go` + all `*_test.go` | Removed/Modified | Impeccable/Graphify blocks + tests |
| `internal/initcmd/templates.go` (+ `templates_test.go`) | Modified | Collapse preflight to prose default; drop groups B(engram)/F/G + rules 8–9 |
| `CLAUDE.md`, `AGENTS.md`, `README.md` | Modified | Preflight prose, single-judge rule, bugfix note; strip removed features |
| `skills/_shared/persistence-contract.md` | Modified | Rewrite OpenSpec-only |
| `skills/_shared/engram-convention.md`, `skills/{graphify,impeccable}/` | Removed | Deleted |
| All 8 phase skills + `harness-judge`, `harness-workflow`, `sdd-init`, `sdd-onboard`, `chained-pr` | Modified | Strip engram/none + Impeccable/Graphify; single-judge; track support |
| `openspec/specs/{harness-workflow,harness-judge}/spec.md` | Modified | Track selection; single-judge model |
| `openspec/specs/graphify-integration/spec.md` | Removed | Retired |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Tab-iota shift breaks TUI index tests | High | Update `model_test.go` prev/next-tab expectations; SecurityTab becomes last |
| `Clone()`/`CloneRoundtrip` misses a removed field | Low | Roundtrip test fails loudly; remove Impeccable/Graphify from fixture |
| Existing `.archon/config.yaml` has `impeccable.*`/`graphify.*` blocks | Med | Lenient `yaml.Unmarshal` (no `KnownFields`) — leftover keys are silently ignored, no error |
| `archon config set/get impeccable.*` after removal | Med | Now an unknown-key error (by design); document in README migration note |
| Skill edits leave dangling engram/Impeccable references | Med | Grep sweep per PR; `embed_test.go` guards the embedded skill list |
| Bugfix track drops judge — reduced safety net | Med | Verify + Human Review Gate (after scoped spec) retained; judge re-enterable via full track |

## Rollback Plan

Each concern is a self-contained PR; revert the offending PR(s). Removed Go config keys are
additive-free deletions — reverting restores the struct/flag/tab verbatim. No data migration:
`state.yaml` without `track` still reads as `full`. Existing config files never required the
removed keys, so no rollback of user data is needed.

## Dependencies

- PR 3 (skill cleanup) depends on PRs 1–2 (Go removal) landing first.
- PRs 4 and 5 are independent of each other; recommended order 1 → 2 → 3 → 4 → 5.

## Proposed PR Slicing (ask-but-default-chained, 800-line budget)

Validated the exploration's 5-PR plan; refined slice 5 to keep single-judge and bugfix-track
each independently reviewable. All slices are under 800 lines.

| PR | Scope | Est. | Depends on |
|----|-------|------|-----------|
| 1 | Remove Impeccable Go plumbing (config/CLI/main/init/TUI/display + tests) | ~350 removed | — |
| 2 | Remove Graphify Go plumbing (mirror PR 1) + delete `skills/graphify/` + `embed_test.go` + retire `graphify-integration` spec | ~350 removed | — |
| 3 | Remove Engram + Impeccable/Graphify from skills; rewrite `persistence-contract.md` OpenSpec-only; delete `engram-convention.md` | ~400–600 removed | 1, 2 |
| 4 | Preflight prose default in `templates.go`/`CLAUDE.md`/`AGENTS.md`/`README.md` (drop groups B-engram/F/G, rules 8–9; add silent-default + deviation rule) | ~200 changed | 3 |
| 5 | Single judge (`harness-judge` + `harness-judge` spec) **and** bugfix track (`state.yaml` `track`, `harness-workflow` skill+spec, `harness-bugfix-track` spec, CLAUDE.md routing hint) | ~300 changed | 3 |

## Success Criteria

- [ ] No `impeccable`/`graphify`/`engram` references remain in Go code, skills, templates, or specs (grep-clean); builds and all tests pass.
- [ ] `archon config set impeccable.enabled true` returns an unknown-key error; an existing config file carrying those blocks loads without error.
- [ ] TUI shows 6 tabs (SecurityTab last); index tests pass.
- [ ] A fresh session applies the standard preflight silently and only asks on a genuine deviation.
- [ ] The SDD judge phase runs one focused review; `judgment-day` still triggers only on explicit request.
- [ ] A change with `track: bugfix` traverses explore → spec (scoped, 3-section format) → apply → verify → archive with judge absent; a change without `track` behaves as `full`.
