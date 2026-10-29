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

**This logic is implemented in code** (`internal/workflow`, covered by its own Go tests) and exposed as `archon workflow status|advance|complete|retry`. Call those commands — do not reimplement the Decision Gates table or the state.yaml write by hand. For claude projects, a `PreToolUse` hook additionally blocks any direct `Edit`/`Write` call on `openspec/changes/*/state.yaml`, so hand-editing it isn't just discouraged, it's refused.

## Phase Sequence

The phase sequence is determined by the change's `track` field in `state.yaml`:

- **`full`** (default; `track` absent or `full`):
  `explore → propose → spec → design → tasks → apply → verify → judge → archive`
- **`bugfix`**: `explore → spec → apply → verify → archive`
  (phases `propose`, `design`, `tasks`, and `judge` are not part of this sequence)
- **`fast`**: `explore → spec → apply → verify → archive`, no judge
  (phases `propose`, `design`, `tasks`, and `judge` are not part of this sequence;
  `fast`'s `spec` phase still produces a full Gherkin spec, unlike `bugfix`'s scoped
  3-section format — see [[harness-fast-track]]. Entered only via `sdd-explore`'s
  sizing suggestion plus a confirmed preflight deviation question, never
  auto-selected. Mechanically, the ONLY way to switch an existing change onto
  `fast` is `archon workflow complete --change {name} --size small --track fast`
  (both flags in the SAME call — `complete` now rejects a second call on an
  already-completed phase, so `--size` cannot be added afterward) at the
  moment `explore` completes — `--track` on `advance` only bootstraps a
  brand-new change and is ignored once `state.yaml` exists; `complete` rejects
  `--track` on every phase except `explore`.)

Each phase has two statuses: `in_progress` | `completed`.

**Terminal ordering (single-PR)**: in the single-PR flow, the sequence after `judge`
passes is: archive commit staged (spec merge, folder move, `archon map`,
`SESSION_STATUS.md` move) → PR opened. Judge gating is unchanged.

**Terminal ordering (Feature Branch Chain)**: in the Feature Branch Chain flow the
archive commit is staged on the **tracker branch** after the integrated judge
passes on the tracker and before the tracker PR merges to `main`. Full rule:
`openspec/specs/harness-workflow/spec.md` "Terminal Phase Ordering (Feature Branch
Chain)" and the `chained-pr` skill. When archive-before-PR is in effect,
Stacked-to-Main is unsupported and converges to Feature Branch Chain at
`sdd-tasks` strategy selection. Full rule:
`openspec/specs/harness-workflow/spec.md` "Stacked-to-Main Archive Convergence" and
the `chained-pr` skill.

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
retry_count: 0
history:
  - {phase: explore, status: completed, ts: "2026-06-10T10:00:00Z"}
  - {phase: propose, status: completed, ts: "2026-06-10T10:05:00Z"}
  - {phase: spec, status: completed, ts: "2026-06-10T10:15:00Z"}
  - {phase: design, status: completed, ts: "2026-06-10T10:30:00Z"}
```

`retry_count` persists harness-judge's auto-re-apply attempt count (see `harness-judge`'s retry loop) so it survives a context compaction or a closed-and-reopened session instead of living only in the orchestrating model's own count-as-you-go reasoning. It defaults to 0 and is managed exclusively via `archon workflow retry increment|reset`.

## Hard Rules

- NEVER allow a transition that skips a mandatory phase. The only valid transition from phase N is to phase N+1 within the selected track's sequence.
- ALWAYS read `state.yaml` before evaluating any transition. If the file does not exist, report `blocked` with reason: `state.yaml not found — run sdd-explore first`.
- `judge.enabled` (and harness-judge's computed risk tier) governs the full track only; the bugfix and fast tracks never run judge.
- A phase that does not belong to the selected track is treated as unreachable — return `blocked` with a message naming both the phase and the track (e.g. `propose is not part of the bugfix track`).
- On valid transition: update `state.yaml` with the new phase, set status to `in_progress`, and append a history entry with the current timestamp.
- On phase completion (reported by the sub-agent): update status to `completed` and record the timestamp.
- Idempotent re-entry: if the requested phase matches the current phase and status is `in_progress`, return `allowed` with status `resuming`.
- NEVER modify `state.yaml` for a blocked transition.
- Use atomic write (temp file + rename) when updating `state.yaml`.
- Mutate `state.yaml` ONLY via `archon workflow advance`/`archon workflow complete`/`archon workflow retry` — never hand-edit the YAML with the Edit or Write tool. For claude projects this is not just a convention: a `PreToolUse` hook blocks any direct `Edit`/`Write` call on `openspec/changes/*/state.yaml`.

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

Read `openspec/changes/{change-name}/state.yaml`. Parse `phase`, `status`, and `track` fields. If the file is missing or corrupt, return `blocked`.

- `track` absent → resolve as `full` (back-compat; all existing in-flight changes behave as full track).
- `track: full`, `track: bugfix`, or `track: fast` → honored as-is.
- Any other value → return `blocked` with reason: `unknown track: {v} (supported: full, bugfix, fast)`. MUST NOT silently coerce to `full`.

### Step 2: Evaluate Transition

Run `archon workflow advance --change {name} --to {requested-phase}`. It resolves the
track from Step 1, applies the exact Decision Gates table below, and reports
`allowed`/`blocked` — it is the canonical implementation (`internal/workflow.CheckTransition`),
not a restatement of it:

| Condition | Action |
|---|---|
| `state.yaml` missing | `blocked` — report: `state.yaml not found — run sdd-explore first` |
| Requested phase == current phase, status `in_progress` | `allowed` — status: `resuming` |
| Requested phase == current phase, status `completed` | `blocked` — report: phase already completed |
| Requested phase == next sequential phase | `allowed` — update state to `in_progress` |
| Requested phase is >1 step ahead | `blocked` — report all missing phases in order |
| Requested phase is before current phase | `blocked` — report: backward transitions not allowed |
| `archive` requested with any phase incomplete | `blocked` — report incomplete phases |

Do not re-derive this table by hand from `PHASE_ORDER`/index math — call the command.

### Step 3: Update State (on allowed transition)

`archon workflow advance` already performed the atomic `state.yaml` write (phase,
status `in_progress`, appended history entry) as part of Step 2 — there is no
separate write step to perform here. It also best-effort regenerates
`openspec/map.md` internally; a regen failure is surfaced as a warning on stderr
and never rolls back the `state.yaml` write.

When a phase completes (reported by its sub-agent), run
`archon workflow complete --change {name}` to mark it `completed` and append the
history entry — again, do not hand-edit the YAML.

### Step 3b: Update SESSION_STATUS.md (MANDATORY on every transition)

In the SAME transition, write/update `SESSION_STATUS.md` at the repository ROOT,
following the `session-status-contract` shared module. Update it BOTH when a phase
starts (`in_progress`) and when it completes (`completed`). This file is the
session-level resume point if the agent is closed mid-work.

Get the "Current phase" line and the "Phase History" checklist from
`archon workflow status --change {name}` — it renders both directly from
`state.yaml`, which is the only way those two sections can never drift from the
machine-readable state. Paste that block verbatim; do not hand-type it.

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

- This skill does NOT implement any SDD phase logic, and as of the `archon workflow`
  commands, it does not even implement the transition-legality check or the
  state.yaml write itself — it only decides WHEN to call `archon workflow
  advance`/`complete`/`retry` and what to do with the result. The check and the
  write live in `internal/workflow`, tested independently of this prose.
- This skill does NOT implement any SDD phase logic. It ONLY gates transitions.
- Delegate actual phase work to the corresponding `sdd-*` skill (sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-apply, sdd-verify, sdd-archive).
- The `judge` phase is handled by `harness-judge`, not by an `sdd-*` skill.
- NEVER allow concurrent phase execution — one phase at a time per change.
- Timestamps MUST use ISO 8601 format (UTC).
- On session resume, READ `SESSION_STATUS.md` (repo root) FIRST to restore context before evaluating any transition. See `session-status-contract`.
- `SESSION_STATUS.md` is updated on every transition and is MOVED into the archived
  change folder during `sdd-archive`, within the staged archive commit — before the
  PR is opened in the single-PR flow, or on the **tracker branch** before the
  tracker PR merges to `main` in the Feature Branch Chain flow. See
  `session-status-contract`.
