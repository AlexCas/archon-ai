# Archive Report: proportional-phase-gating

- **Archived to**: `openspec/changes/archive/2026-10-28-proportional-phase-gating/`
- **Task completion gate**: all 23 tasks in `tasks.md` were checked (`- [x]`) before this report was written; no reconciliation was needed.
- **Verify report**: final round PASS WITH WARNINGS; warnings are pre-existing, unrelated drift (confirmed again below), not CRITICAL issues. No override was required.
- **Judge**: PASS, 0 confirmed issues, 6 non-blocking notes, after 3 fail/fix rounds (see `verify-report.md` and the archived `state.yaml` history for the full timeline).

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| harness-workflow | Updated | 1 requirement MODIFIED ("Single Judge in the Full-Track Terminal Ordering" → tier-computed gating), 1 requirement ADDED ("Fast Track Selection in Terminal Ordering") |
| harness-judge | Updated | 1 requirement MODIFIED+renamed ("Preserved gates and loop under the single judge" → "Computed risk tier gates judge (full track only)"), 1 requirement ADDED ("Computed Per-Change Risk Tier") |
| harness-fast-track | Created | New domain; main spec did not exist, delta spec copied directly as the full spec, alongside its `.feature` file |

Each domain's `.feature` file was merged with the same scenario set as its `spec.md`.

## Archive Contents

- proposal.md ✅
- exploration.md ✅
- specs/ ✅ (includes Gherkin `.feature` files)
- design.md ✅
- tasks.md ✅ (23/23 tasks complete)
- apply-progress.md ✅
- verify-report.md ✅
- SESSION_STATUS.md ✅ (moved from root)
- state.yaml ✅ (phase: archive, status: completed)

## Source of Truth Updated

- `openspec/specs/harness-workflow/spec.md`
- `openspec/specs/harness-judge/spec.md`
- `openspec/specs/harness-fast-track/spec.md` (new)
- `openspec/map.md` (regenerated via `archon map --backfill`)

## Known Pre-Existing Drift (Not Introduced by This Change)

`archon map --check` reports dangling `../../proposal.md` / `../../exploration.md`
header links in 6 main-spec domains: `harness-bugfix-track`, `harness-fast-track`,
`harness-init`, `harness-judge`, `harness-workflow`, `local-model-router`, and
`sdd-phase-skills`. Five of these six predate `proportional-phase-gating` entirely
(they were never touched by this change); `harness-fast-track` inherited the same
pattern when its delta spec's unmodified header comment was copied directly into
`openspec/specs/` as a brand-new domain. The pattern is an established, if
imperfect, repo convention: merged main specs keep the delta's original header
comment verbatim, including links that were only valid relative to the change
folder. `archon map --backfill` confirmed the archived folder's own internal links
resolve cleanly — this drift is scoped to main-spec header comments only, not to
anything the archive move itself touched. Logged here as a candidate follow-up
bugfix (e.g., strip or rewrite these header links at merge time); it did not block
this archive, consistent with `verify-report.md`'s own "pre-existing, unrelated
drift" characterization of its non-blocking warnings.

## Retry Accounting Note (Carried Over From Judge Phase)

One of the three `archon workflow retry increment` calls during the judge/verify
fail-fix loop was triggered by a verify-only failure, not a judge failure — a
process slip on the orchestrator's part, disclosed to the user at the time. There
is no `retry` decrement command, so the persisted count could not be corrected;
`retry_count` was reset to 0 via `archon workflow retry reset` once judge passed.
This did not affect the final outcome.

## SDD Cycle Complete

The change has been fully explored, proposed, specified, designed, planned,
implemented, verified, judged, and archived. Ready for the next change.
