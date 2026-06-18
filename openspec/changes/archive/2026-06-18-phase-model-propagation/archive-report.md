# Archive Report

**Change**: phase-model-propagation
**Archived**: 2026-06-18T16:55:00Z
**Archived to**: `openspec/changes/archive/2026-06-18-phase-model-propagation/`
**Artifact store**: openspec

## Outcome

Propagates the configured per-phase models from `.archon/config.yaml` (`models.phases`
/ `models.default`) into the generated orchestrator file. Init and the TUI
"regenerate template" path now emit an advisory, deterministic phase→model block that
instructs the orchestrator to request the configured model when delegating each SDD
phase. Model values are normalized to identifiers the delegation tool accepts; unknown
values are surfaced to the user.

All work is landed on `master`:
- PR #36 — feat/phase-model-propagation (config core + template rendering + plumbing)
- PR #35 — prerequisite archive of the prior change (archon-update)

## Verification & Judgment

- **Verify**: PASS WITH WARNINGS — no CRITICAL issues.
- **Judge**: PASS (APPROVED, 2 rounds, 0 re-apply cycles). Mutation gate skipped
  (disabled); Playwright gate skipped (non-web Go CLI).

## Specs Synced to Source of Truth

| Domain | Action | Details |
|--------|--------|---------|
| harness-init | Updated | 5 ADDED requirements + 7 Gherkin scenarios appended (no requirements duplicated from the prior archon-update archive) |

ADDED requirements appended to `openspec/specs/harness-init/spec.md`:
- Rendered phase→model block
- Normalization to real model IDs
- Phase model resolution and fallback
- Deterministic phase ordering
- Unknown model values surfaced to the user

Scenarios appended to `openspec/specs/harness-init/harness-init.feature`:
- Init renders a phase model for a configured phase (@happy)
- TUI regeneration produces the same block as init (@happy)
- Display name is normalized to an accepted identifier (@happy)
- Phase falls back to the default model (@happy)
- Phase omitted when no model resolves (@edge)
- Multiple configured phases render in canonical order (@edge)
- Garbage model value is surfaced (@error)

Updated source-of-truth files:
- `openspec/specs/harness-init/spec.md` (appended)
- `openspec/specs/harness-init/harness-init.feature` (appended)

## Archive Contents

- exploration.md
- proposal.md
- specs/ (harness-init — spec.md + harness-init.feature)
- design.md
- tasks.md (15/15 implementation tasks complete + 4 verification tasks)
- verify-report.md
- judge-report.md
- state.yaml (phase: archive, status: completed)
- archive-report.md
- SESSION_STATUS.md (moved from repo root)

## Tasks

All implementation tasks checked complete across Phases 1–3 (config core, template
rendering, plumbing) plus Phase 4 verification tasks. No stale unchecked tasks.

## Notes / Follow-ups

- The phase→model block is ADVISORY. It instructs the orchestrator LLM to request the
  configured model per phase; it does not hard-enforce model selection. Confirm the
  platform's per-delegation model selection mechanism honors the block as expected.
- Verify completed with non-CRITICAL warnings only; no follow-up blocking items.
