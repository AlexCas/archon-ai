# Archive Report

**Change**: archon-update
**Archived**: 2026-06-18T01:45:19Z
**Archived to**: `openspec/changes/archive/2026-06-18-archon-update/`
**Artifact store**: openspec

## Outcome

Adds a safe, version-aware `archon update` command that refreshes installed skills
from the embedded set without rewriting the orchestrator template or resetting user
config, plus a truthful skill-inventory at init time (real `SKILL.md` frontmatter
versions instead of a hardcoded value).

All work is landed on `master`:
- PR #30 — foundation (feat/archon-update-foundation)
- PR #31 / #33 — command (feat/archon-update-command)
- PR #32 — tracker (feature/archon-update → master)

## Verification & Judgment

- **Verify**: PASS WITH WARNINGS — no CRITICAL issues. Warnings W1/W2 concern
  command-layer output assertions verified by inspection/shared-path inference;
  suggestions S1 (`judge` config field does not yet exist) and S2 (copy-mode
  `--prune` doc note) recorded as follow-ups.
- **Judge**: PASS (retry 1/3). Dual review surfaced 5 confirmed issues, all fixed
  in re-apply and confirmed resolved on re-judge. Mutation gate skipped (disabled);
  Playwright gate skipped (non-web CLI).

## Specs Synced to Source of Truth

| Domain | Action | Details |
|--------|--------|---------|
| harness-update | Created | New main spec + Gherkin feature (no prior main spec existed; full delta copied) |
| harness-init | Updated | 1 ADDED requirement ("Truthful skill inventory versions") appended to spec + 2 scenarios appended to feature |

Updated source-of-truth files:
- `openspec/specs/harness-update/spec.md` (new)
- `openspec/specs/harness-update/harness-update.feature` (new)
- `openspec/specs/harness-init/spec.md` (appended)
- `openspec/specs/harness-init/harness-init.feature` (appended)

## Archive Contents

- exploration.md
- proposal.md
- specs/ (harness-update, harness-init — spec.md + .feature each)
- design.md
- tasks.md (21/21 tasks complete)
- verify-report.md
- judge-report.md
- state.yaml (phase: archive, status: completed)
- archive-report.md
- SESSION_STATUS.md (moved from repo root)

## Tasks

21/21 implementation tasks checked complete. No stale unchecked tasks.

## Notes / Follow-ups

- S1: spec language references a `judge` field in `.archon/config.yaml` that does
  not yet exist in `config.Config`; preservation is currently vacuous. Reconcile
  the spec or add the field in a future change.
- W1/W2: candidate command-layer output assertions for a follow-up test pass.
