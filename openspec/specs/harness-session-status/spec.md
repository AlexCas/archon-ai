# harness-session-status Specification

## Purpose

Provide a single per-session resume point so an SDD session can continue without
losing context after the agent is closed.

## Requirements

### Requirement: Session status file updated per phase

On EVERY phase transition the orchestrator MUST write/update `SESSION_STATUS.md` at
the repository root, recording the active change, current phase and status, preflight
choices (including the web/Playwright decision), phase history with timestamps, key
artifact paths, open questions, and the next recommended step. It MUST be written at
both the start (`in_progress`) and the end (`completed`) of each phase, even in `auto`
mode. Format follows `session-status-contract`.

#### Scenario: File updated on a phase transition

```gherkin
Scenario: File updated on a phase transition
  Given an active change in the spec phase
  When the orchestrator advances to the design phase
  Then "SESSION_STATUS.md" at the repo root reflects phase "design"
  And it lists the completed phases with timestamps
```

### Requirement: Crash recovery from root

If the agent is closed unexpectedly, `SESSION_STATUS.md` MUST remain at the repo root.
On the next session the orchestrator MUST read it FIRST to restore context before any
other action.

#### Scenario: Resuming after an unexpected close

```gherkin
Scenario: Resuming after an unexpected close
  Given a "SESSION_STATUS.md" left at the repo root from a previous session
  When a new session starts
  Then the orchestrator reads it before evaluating any phase transition
```

### Requirement: Archived with the change

During `sdd-archive` the file MUST be MOVED into the archived change folder and
removed from the root. In engram mode its final contents are stored as an observation
and the root file is deleted.

#### Scenario: Session status archived with the feature

```gherkin
Scenario: Session status archived with the feature
  Given a completed change being archived in openspec mode
  When sdd-archive runs
  Then "SESSION_STATUS.md" is moved into the archived change folder
  And no "SESSION_STATUS.md" remains at the repo root
```
