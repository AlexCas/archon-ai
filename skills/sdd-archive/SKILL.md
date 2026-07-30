---
name: sdd-archive
description: "Archive a completed SDD change by syncing delta specs. Trigger: orchestrator launches archive after implementation and verification."
disable-model-invocation: true
user-invocable: false
license: MIT
metadata:
  
  version: "2.0"
  delegate_only: true
---

> **ORCHESTRATOR GATE**: If you loaded this skill via the `skill()` tool, you are
> the ORCHESTRATOR — STOP. Do NOT execute these instructions inline. Delegate to
> the dedicated `sdd-archive` sub-agent using your platform's delegation primitive
> (e.g., `task(...)`, sub-agent invocation, etc.). This skill is for EXECUTORS
> only.

## Executor Override

If you ARE the `sdd-archive` sub-agent (NOT the orchestrator), the gate above does NOT apply to you. Continue with the phase work below. Do NOT delegate. Do NOT call the Skill tool. You are the executor — execute.


## Language Domain Contract

Generated technical artifacts default to English. Do not inherit the user's conversational language or the active persona's regional voice for SDD artifacts unless the user explicitly requests that artifact language or the project convention requires it.

If Spanish technical artifacts are explicitly requested, use neutral/professional Spanish unless the user explicitly asks for a regional variant.

Public/contextual comments follow the target context language by default. Explicit user language or tone overrides win; Spanish comments default to neutral/professional Spanish unless the user or target context clearly calls for regional tone.

## Purpose

You are a sub-agent responsible for ARCHIVING. You merge delta specs into the main specs (source of truth), then move the change folder to the archive. You complete the SDD cycle.

## What You Receive

From the orchestrator:
- Change name
- Artifact store mode (`engram | openspec | hybrid | none`)
- Structured status from `skills/_shared/sdd-status-contract.md`, including artifact paths, task progress, dependency states, and actionContext
- Any explicit intentional archive override text from the user/orchestrator

## Execution and Persistence Contract

> Follow **Section B** (retrieval) and **Section C** (persistence) from `skills/_shared/sdd-phase-common.md`.

- **engram**: Read `sdd/{change-name}/proposal`, `sdd/{change-name}/spec`, `sdd/{change-name}/design`, `sdd/{change-name}/tasks`, `sdd/{change-name}/verify-report` (all required). Record all observation IDs in the archive report for traceability. Save as `sdd/{change-name}/archive-report`.
- **openspec**: Read and follow `skills/_shared/openspec-convention.md`. Perform merge and archive folder moves.
- **hybrid**: Follow BOTH conventions — persist archive report to Engram (with observation IDs) AND perform filesystem merge + archive folder moves.
- **none**: Return closure summary only. Do not perform archive file operations.

### Task Completion Gate

`sdd-apply` is responsible for marking completed tasks in the persisted tasks artifact. `sdd-archive` is responsible for validating that the persisted artifact reflects the final state before closing the cycle.

Before syncing specs or moving any archive folder, inspect the tasks artifact:

- **engram**: read the full `sdd/{change-name}/tasks` observation.
- **openspec/hybrid**: read `openspec/changes/{change-name}/tasks.md`.

If any implementation task remains unchecked (`- [ ]`):

1. STOP and return `blocked`; do not sync specs, move the change folder, or claim the SDD cycle is complete.
2. Report that `sdd-apply` must be rerun or corrected so it marks completed tasks in the persisted tasks artifact.
3. Only proceed if the orchestrator explicitly instructs you to reconcile stale checkboxes and `apply-progress`/`verify-report` prove every unchecked task is complete. If you do this exceptional repair, record the exact reconciliation reason in the archive report.

The archived audit trail MUST NOT contain stale unchecked tasks for completed work. Internal todo state is not enough; the persisted SDD task artifact is the source of truth for completion visibility.

### Strict-vs-OpenSpec Archive Policy

OpenSpec permits archiving with incomplete artifacts or tasks after a user confirmation. gentle-ai is stricter by default:

- Incomplete implementation tasks block archive unless they are stale checkboxes and apply-progress/verify-report prove completion.
- CRITICAL issues in `verify-report` always block archive. Do not accept an override for CRITICAL verification issues.
- `sdd-archive` does not own normal task completion. `sdd-apply` owns checkbox completion; archive may only perform exceptional mechanical reconciliation with proof from apply-progress and verify-report.
- Missing proposal/spec/design artifacts should be reported. Archive may continue only when the user explicitly chooses an intentional partial archive and the archive report records what was missing.

### Action Context Guard

- If structured status reports `actionContext.mode: workspace-planning`, STOP. Do not move workspace changes into repo-local archives or edit linked repos.
- If `allowedEditRoots` is present, archive operations must stay inside those roots.

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Timing

Archive runs AFTER judge passes and BEFORE the PR is opened. Stage the results of
Step 2 (spec merge), Step 3 (folder move), Step 3b (`archon map --backfill` +
`--check`), and Step 3c (`SESSION_STATUS.md` move) into ONE archive commit on the
change branch (Step 3d performs the actual `git add`/`git commit`). The PR is
opened only after this commit is staged. The archive-internal order
(merge → move → map → SESSION_STATUS move) is unchanged; PR-open is EXTERNAL
to this sequence.

### Step 2: Sync Delta Specs to Main Specs

Do not start this step until the **Task Completion Gate** above passes.

**IF mode is `engram`:** Skip filesystem sync — artifacts live in Engram only. The archive report (Step 5) records all observation IDs for traceability.

**IF mode is `none`:** Skip — no artifacts to sync.

**IF mode is `openspec` or `hybrid`:** For each delta spec in `openspec/changes/{change-name}/specs/`:

#### If Main Spec Exists (`openspec/specs/{domain}/spec.md`)

Read the existing main spec and apply the delta:

```
FOR EACH SECTION in delta spec:
├── ADDED Requirements → Append to main spec's Requirements section
├── MODIFIED Requirements → Replace the matching requirement in main spec
├── REMOVED Requirements → Delete the matching requirement from main spec after recording Reason/Migration
└── RENAMED Requirements → Rename the matching requirement while preserving scenarios unless the delta also modifies them
```

**Merge carefully:**
- Match requirements by name (e.g., "### Requirement: Session Expiration")
- Preserve all OTHER requirements that aren't in the delta
- Maintain proper Markdown formatting and heading hierarchy
- For REMOVED requirements, require `(Reason: ...)` and `(Migration: ...)` notes in the delta before deleting from main specs
- For RENAMED requirements, require the old and new requirement names to be explicit

#### If Main Spec Does NOT Exist

The delta spec IS a full spec (not a delta). Copy it directly:

```bash
# Copy new spec to main specs
openspec/changes/{change-name}/specs/{domain}/spec.md
  → openspec/specs/{domain}/spec.md
```

### Step 3: Move to Archive

**IF mode is `engram`:** Skip — there are no `openspec/` directories to move. The archive report in Engram serves as the audit trail.

**IF mode is `none`:** Skip — no filesystem operations.

**IF mode is `openspec` or `hybrid`:** Move the entire change folder to archive with date prefix:

```
openspec/changes/{change-name}/
  → openspec/changes/archive/YYYY-MM-DD-{change-name}/
```

Use today's date in ISO format (e.g., `2026-02-16`).

### Step 3b: Rewrite Vault Links

**IF mode is `openspec` or `hybrid`:** After the folder move, run:

1. `archon map --backfill` — rewrites boundary-crossing relative links inside
   the moved files (e.g. `../../specs/...` gains a `../` level; plain
   `archon map` regenerates `map.md` but does not touch archived-file links,
   so `--backfill` is required here) and regenerates `openspec/map.md`.
   Wikilinks (`[[capability]]`) need no rewrite — they resolve by name, not
   by path.
2. `archon map --check` — verifies no dangling relative links remain.
3. If `--check` exits non-zero, **STOP**: surface the failure to the
   orchestrator and do NOT mark the archive complete. Do not proceed to
   Step 3c or Step 4 until `--check` passes.

**IF mode is `engram` or `none`:** Skip — no `openspec/` filesystem tree exists.

### Step 3c: Archive SESSION_STATUS.md

The session-level resume file lives at the repository ROOT during work. Finalize it
as part of the change's audit trail (see `session-status-contract`):

**IF mode is `openspec` or `hybrid`:** MOVE `SESSION_STATUS.md` from the repo root
into the archived change folder as part of the archive commit staging (before the
PR is opened), then ensure it no longer exists at the root:

```
SESSION_STATUS.md  → openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md
```

**IF mode is `engram`:** Store the final `SESSION_STATUS.md` contents as the
`sdd/{change-name}/session-status` observation, then delete the root file.

**IF mode is `none`:** Delete the root file (no persisted audit trail).

If `SESSION_STATUS.md` is absent at the root (e.g., already archived), note it and continue.

### Step 3d: Stage and Commit Archive Changes

**IF mode is `openspec` or `hybrid`:** Stage and commit all archive changes onto the change branch as ONE commit:

1. `git add` the merged main spec(s) (Step 2), the moved change folder (Step 3),
   the regenerated `openspec/map.md` (Step 3b), and the moved `SESSION_STATUS.md`
   (Step 3c).
2. Create ONE commit with subject `chore(archive): archive {change-name}`
   (conventional commit format). Keep the body focused on what was archived
   (specs synced, folder moved, map regenerated) — not on the tooling that
   produced it.
3. **Commit Attribution HARD RULE**: the commit is authored SOLELY by the
   user's git account. Do NOT add `Co-Authored-By` trailers, "Generated with"
   lines, or any agent/assistant/tool attribution.
4. This commit satisfies the branch-PR precondition: the archive commit MUST
   be staged on the change branch BEFORE the PR is opened (single-PR flow).

**IF mode is `engram` or `none`:** Skip — there is no branch commit in these
modes; the archive report (Step 5) is the audit trail.

### Step 4: Verify Archive

**IF mode is `openspec` or `hybrid`:** Confirm:
- [ ] Main specs updated correctly
- [ ] Change folder moved to archive
- [ ] `archon map --check` passed after the move (Step 3b)
- [ ] Archive contains all artifacts (proposal, specs, design, tasks, SESSION_STATUS.md)
- [ ] `SESSION_STATUS.md` no longer exists at the repo root
- [ ] Archived `tasks.md` has no unchecked implementation tasks, unless the orchestrator explicitly approved archive-time stale-checkbox reconciliation backed by apply-progress/verify-report proof
- [ ] Active changes directory no longer has this change
- [ ] Archive commit created on the change branch (Step 3d) with subject `chore(archive): archive {change-name}`, authored solely by the user's git account
- [ ] PR has NOT been opened before this archive commit is staged (single-PR flow)

**IF mode is `engram`:** Confirm all artifact observation IDs are recorded in the archive report and the tasks observation has no unchecked implementation tasks unless the orchestrator explicitly approved archive-time stale-checkbox reconciliation backed by apply-progress/verify-report proof.

**IF mode is `none`:** Skip verification — no persisted artifacts.

### Step 5: Persist Archive Report

**This step is MANDATORY — do NOT skip it.**

Follow **Section C** from `skills/_shared/sdd-phase-common.md`.
- artifact: `archive-report`
- topic_key: `sdd/{change-name}/archive-report`
- type: `architecture`

### Step 6: Return Summary

Return to the orchestrator:

```markdown
## Change Archived

**Change**: {change-name}
**Archived to**: `openspec/changes/archive/{YYYY-MM-DD}-{change-name}/` (openspec/hybrid) | Engram archive report (engram) | inline (none)

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| {domain} | Created/Updated | {N added, M modified, K removed requirements} |

### Archive Contents
- proposal.md ✅
- specs/ ✅ (includes Gherkin `.feature` files)
- design.md ✅
- tasks.md ✅ ({N}/{N} tasks complete)
- SESSION_STATUS.md ✅ (moved from root)

### Source of Truth Updated
The following specs now reflect the new behavior:
- `openspec/specs/{domain}/spec.md`

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
```

## Rules

- NEVER archive a change that has CRITICAL issues in its verification report
- If the user explicitly approves a non-critical partial archive or stale-checkbox reconciliation, record the exact reason in the archive report and mark the archive as intentional-with-warnings
- NEVER archive completed work while `tasks.md` / the tasks observation still shows stale unchecked implementation tasks
- ALWAYS sync delta specs BEFORE moving to archive
- When merging into existing specs, PRESERVE requirements not mentioned in the delta
- Use ISO date format (YYYY-MM-DD) for archive folder prefix
- If the merge would be destructive (removing large sections), WARN the orchestrator and ask for confirmation
- The archive is an AUDIT TRAIL — never delete or modify archived changes
- If `openspec/changes/archive/` doesn't exist, create it
- Archive is pre-PR in the single-PR flow — never open the PR before the archive commit is staged
- Apply any `rules.archive` from `openspec/config.yaml`
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
