# Session Status

## Active Change
`streamline-harness` — reduce harness over-engineering across preflight, judge flow, config surface, and add a lightweight bug-fix SDD path.

## Current Phase
ARCHIVE — IN PROGRESS on tracker `feature/streamline-harness`. Integrated judge PASS (all 6 concerns, cross-slice consistent, green). Release decided: v0.13.0 — "Streamlined harness". Next: archive commit → #114 → master → tag v0.13.0 + gh release.

## ARCHIVE reminder (corrected)
Live specs in this repo are delta-framed BY CONVENTION (harness-judge, sdd-init, etc. all "# Delta for X"). PR5's live-spec edits were REVERTED so ARCHIVE performs the spec placement/merge (its Rule-9 job). Archive must place ALL 7 change-folder deltas into openspec/specs/: graphify-integration (removal), harness-init, harness-judge, harness-workflow, harness-bugfix-track (new), sdd-phase-skills — plus folder move + archon map + SESSION_STATUS move. (graphify-integration LIVE spec was already deleted in PR2.)

## Notes
- PR2 also removed orphaned `skills/impeccable/SKILL.md` (PR1's Go-only scope left it behind) — document in PR2 body.
- Build env: use GOROOT=/home/linuxbrew/.linuxbrew/opt/go/libexec (go1.26.5 binary vs 1.26.4 stdlib cache mismatch).

## Branch Topology (Feature Branch Chain)
- Tracker: `feature/streamline-harness` (base master) — draft PR #114, planning artifacts (a119681)
- PR1: `feat/streamline-remove-impeccable` (base tracker) — commit ee0201e, PR #115 OPEN, judge APPROVED ✅
- PR2 branches from PR1 (immediate parent) once we proceed.

## PR Chain Progress
- [x] PR1 Remove Impeccable — PR #115 (apply/verify/judge all green; 638-line diff)
- [x] PR2 Remove Graphify — PR #116 (base PR1; judge APPROVED; 18 files, 72+/1411−; size:exception approved)
- [x] PR3 Engram + Imp/Graph out of skills — PR #117 (base PR2; judge FAIL→retry1→APPROVED; 22 files, ~908 lines; size:exception approved)
- [x] PR4 Preflight prose default — PR #118 (base PR3; judge APPROVED; 7 files, 140+/289−; Rules reconciled surgically per user decision)
- [x] PR5 Single judge + bugfix track — PR #119 (base PR4; judge PASS; behavior-only 12 files, 159+/78−=237; live-spec placement deferred to archive)

## PR Stack (all OPEN, awaiting review/merge)
- #114 tracker (draft) ← #115 PR1 ← #116 PR2 ← #117 PR3 ← #118 PR4 ← #119 PR5
- Merge order: PR1→…→PR5 down into the tracker, then integrated judge on tracker, then archive on tracker, then #114 → master.

## Deferred follow-ups (post-chain cleanup PR)
- CLAUDE.md Session-Status "archive" bullet diverges from generic template line (archon update would overwrite).
- AGENTS.md one extra blank line before ## Rules vs template concatenation.
- PR1 loader-lenient regression subtest (leftover impeccable: block loads clean).

## Non-blocking follow-ups (noted)
- PR1: add a `Load()` regression subtest feeding a leftover `impeccable:` block asserting err==nil (guards against an accidental KnownFields(true) regression).

## Preflight (session decisions)
- **A. Ritmo**: Interactivo
- **B. Artefactos**: OpenSpec (engram out of scope)
- **C. PRs**: Preguntar, default encadenados (Feature Branch Chain)
- **D. Revisión**: 800 líneas
- **E. Playwright**: No
- **F. Impeccable**: No
- **G. Graphify**: No

## Scope (user intent)
1. Preflight becomes a fixed standard default (interactive · openspec · ask-but-chained-default · 800 lines); harness may still propose a different choice when warranted. Remove engram from scope.
2. Remove Impeccable and Graphify from the harness entirely (config, TUI, skills, templates, docs). Re-add externally only if needed later.
3. Simplify judge flow: a single judge is enough in the SDD path. Keep the `judgment-day` skill as an opt-in flow outside SDD.
4. New lightweight SDD flow for small bugs: explore → scoped spec → apply → manual verification (skip the full pipeline).
5. Investigate and propose ideas to make the harness flow more agile.

## Completed Phases
- explore — 2026-09-16 (exploration.md created)
- propose — 2026-09-16 (proposal.md created)
- spec — 2026-09-16 (6 capability specs created under specs/)
- design — 2026-09-16 (design.md created)
- tasks — 2026-09-16 (tasks.md created; 5 PRs, ~65 tasks)

## Key Artifacts
- `openspec/changes/streamline-harness/tasks.md` (NEW — 5 chained PRs, ~65 tasks, contingency splits 3a/3b and 5a/5b)
- `openspec/changes/streamline-harness/exploration.md`
- `openspec/changes/streamline-harness/proposal.md`
- `openspec/changes/streamline-harness/specs/harness-bugfix-track/spec.md` + `.feature` (NEW capability)
- `openspec/changes/streamline-harness/specs/harness-workflow/spec.md` (MODIFIED — track-aware, single-judge, preflight prose)
- `openspec/changes/streamline-harness/specs/harness-judge/spec.md` (MODIFIED — single focused judge; judgment-day untouched)
- `openspec/changes/streamline-harness/specs/graphify-integration/spec.md` (RETIRED)
- `openspec/changes/streamline-harness/specs/harness-init/spec.md` (MODIFIED — no Impeccable/Graphify surface; 6 TUI tabs; OpenSpec-only)
- `openspec/changes/streamline-harness/specs/sdd-phase-skills/spec.md` (MODIFIED — OpenSpec-only persistence; no Impeccable/Graphify hooks)

## Bugfix scoped-spec format (decided)
Single spec.md, no Gherkin: `## Bug` (observed vs expected) · `## Fix Criteria` (2–5 verifiable bullets) · `## Non-Regression` (1–2 bullets).

## Slicing plan (5 chained PRs, order 1→2→3→4→5)
- PR1 Impeccable Go · PR2 Graphify Go+skill/spec · PR3 Engram+Imp/Graph out of skills (dep 1,2) · PR4 Preflight prose (dep 3) · PR5 Single judge + bugfix track (dep 3). PR4/PR5 mutually independent.

## Resolved Decisions (review gate, 2026-09-16)
1. Preflight defaults live as PROSE in CLAUDE.md — no new Go config struct.
2. Remove `none` artifact mode too — artifacts always OpenSpec (group B collapses to one path).
3. bugfix track: judge is ABSENT from the bugfix PHASE_ORDER (not opt-in).

## Open Questions (recommend in propose)
- Exact format of the bug-fix "scoped spec" (minimal acceptance criteria, no full Gherkin suite).

## Recommended slicing (5 chained PRs, <800 lines each)
- PR1 Remove Impeccable (Go) · PR2 Remove Graphify (Go) · PR3 Remove Engram + Impeccable/Graphify from skills (dep: PR1,PR2) · PR4 Preflight simplification · PR5 Single judge + bug-fix track

## Next Step
Human Review Gate on tasks → on approval, run sdd-apply starting with PR1 (Remove Impeccable Go plumbing).
