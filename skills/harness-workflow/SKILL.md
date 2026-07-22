---
name: harness-workflow
description: "Trigger: phase transition, state check, workflow gate, next phase. Enforce SDD phase state machine and block invalid transitions."
license: MIT
metadata:
  
  version: "1.0"
  scope: orchestrator-gate
---

## Purpose

Gate every SDD phase transition. Read `openspec/changes/{name}/state.yaml`, enforce the linear phase sequence, and block any attempt to skip mandatory phases.

## Phase Sequence

```
explore → propose → spec → design → tasks → apply → verify → judge → archive
```

Each phase has two statuses: `in_progress` | `completed`.

## Activation Contract

Load when the orchestrator needs to:
1. Check whether a phase transition is valid
2. Report the current workflow state for a change
3. Advance a change to the next phase

The orchestrator MUST invoke this skill BEFORE delegating to any `sdd-*` sub-agent.

## State File Format

Location: `openspec/changes/{change-name}/state.yaml`

```yaml
phase: design
status: completed
history:
  - {phase: explore, status: completed, ts: "2026-06-10T10:00:00Z"}
  - {phase: propose, status: completed, ts: "2026-06-10T10:05:00Z"}
  - {phase: spec, status: completed, ts: "2026-06-10T10:15:00Z"}
  - {phase: design, status: completed, ts: "2026-06-10T10:30:00Z"}
```

## Hard Rules

- NEVER allow a transition that skips a mandatory phase. The only valid transition from phase N is to phase N+1.
- ALWAYS read `state.yaml` before evaluating any transition. If the file does not exist, report `blocked` with reason: `state.yaml not found — run sdd-explore first`.
- On valid transition: update `state.yaml` with the new phase, set status to `in_progress`, and append a history entry with the current timestamp.
- On phase completion (reported by the sub-agent): update status to `completed` and record the timestamp.
- Idempotent re-entry: if the requested phase matches the current phase and status is `in_progress`, return `allowed` with status `resuming`.
- NEVER modify `state.yaml` for a blocked transition.
- Use atomic write (temp file + rename) when updating `state.yaml`.

## Decision Gates

| Condition | Action |
|---|---|
| `state.yaml` missing | `blocked` — report: `state.yaml not found — run sdd-explore first` |
| Requested phase == current phase, status `in_progress` | `allowed` — status: `resuming` |
| Requested phase == current phase, status `completed` | `blocked` — report: phase already completed |
| Requested phase == next sequential phase | `allowed` — update state to `in_progress` |
| Requested phase is >1 step ahead | `blocked` — report all missing phases in order |
| Requested phase is before current phase | `blocked` — report: backward transitions not allowed |
| `archive` requested with any phase incomplete | `blocked` — report incomplete phases |

## Execution Steps

### Step 1: Read State

Read `openspec/changes/{change-name}/state.yaml`. Parse `phase` and `status` fields. If the file is missing or corrupt, return `blocked`.

### Step 2: Evaluate Transition

Compare the requested phase against the current phase using the linear sequence:

```
PHASE_ORDER = [explore, propose, spec, design, tasks, apply, verify, judge, archive]
current_index = PHASE_ORDER.index(current_phase)
requested_index = PHASE_ORDER.index(requested_phase)
```

Valid transitions:
- `requested_index == current_index` AND `status == in_progress` → `allowed` (resuming)
- `requested_index == current_index + 1` AND `status == completed` → `allowed` (advancing)
- Everything else → `blocked`

### Step 3: Update State (on allowed transition)

Write updated `state.yaml` using atomic write (temp file + rename):
- Set `phase` to the requested phase
- Set `status` to `in_progress` (unless resuming, keep `in_progress`)
- Append history entry: `{phase: requested, status: in_progress, ts: <now>}`

Execution order is fixed: `state.yaml` MUST be written (temp+rename complete)
BEFORE the next sub-step runs. Immediately after, shell out to `archon map` to
regenerate `openspec/map.md` so it reflects the new phase/status. A regen
failure (non-zero exit) MUST be surfaced to the orchestrator as a WARNING —
it MUST NOT roll back the `state.yaml` write or block the recorded transition;
state is LLM-owned, the map regen is best-effort.

### Step 3b: Update SESSION_STATUS.md (MANDATORY on every transition)

In the SAME transition, write/update `SESSION_STATUS.md` at the repository ROOT,
following the `session-status-contract` shared module. Update it BOTH when a phase
starts (`in_progress`) and when it completes (`completed`). This file is the
session-level resume point if the agent is closed mid-work.

- Record: active change, current phase + status, preflight choices (including the
  web/Playwright decision), phase history with timestamps, key artifact paths,
  open questions, and the next recommended step.
- Use atomic write (temp file + rename).
- Do NOT skip this step — even in `auto` mode.

### Step 4: Report

Return structured response:

```markdown
## Workflow State

**Change**: {change-name}
**Current phase**: {phase}
**Status**: {status}
**Transition**: {allowed | blocked}
**Next recommended**: {next phase if status is completed, or current phase if in_progress}
```

If blocked, append:

```markdown
**Reason**: {why blocked}
**Missing phases**: {comma-separated list if skipping detected}
```

## Error Handling

| Condition | Behavior |
|---|---|
| Corrupt `state.yaml` (invalid YAML) | `blocked` — report: `corrupt state.yaml — delete and re-run sdd-explore` |
| Unknown phase name in request | `blocked` — report: `unknown phase: {name}` |
| Filesystem write failure | `blocked` — report: `failed to update state.yaml: {error}` |
| Change directory missing | `blocked` — report: `change directory not found: openspec/changes/{name}/` |

## Rules

- This skill does NOT implement any SDD phase logic. It ONLY gates transitions.
- Delegate actual phase work to the corresponding `sdd-*` skill (sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-apply, sdd-verify, sdd-archive).
- The `judge` phase is handled by `harness-judge`, not by an `sdd-*` skill.
- NEVER allow concurrent phase execution — one phase at a time per change.
- Timestamps MUST use ISO 8601 format (UTC).
- On session resume, READ `SESSION_STATUS.md` (repo root) FIRST to restore context before evaluating any transition. See `session-status-contract`.
- `SESSION_STATUS.md` is updated on every transition and is MOVED into the archived change folder during `sdd-archive`.
