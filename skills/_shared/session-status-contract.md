# Session Status Contract

Shared module. Defines `SESSION_STATUS.md`: a single, human-readable, per-session
snapshot that lets any agent resume an SDD session without losing context — even
if the previous agent was closed mid-phase.

## Why this exists

`state.yaml` (per change) and Engram observations track phase state, but they are
machine-oriented and scoped per change. `SESSION_STATUS.md` is the fast,
at-a-glance resume point for the WHOLE session, kept at the repository ROOT while
work is in progress.

## Location & lifecycle

- **Active session**: `SESSION_STATUS.md` lives at the repository root.
- **Written/updated**: on EVERY phase transition — at the start of a phase
  (status `in_progress`) and again when the phase completes (status `completed`).
- **Crash recovery**: if the agent is closed unexpectedly, the file remains at the
  root. The next session MUST read it FIRST, before any other action, to restore
  context.
- **Archive**: during `sdd-archive`, MOVE the file into the archived change folder
  (`openspec/changes/archive/YYYY-MM-DD-{change-name}/SESSION_STATUS.md`) alongside
  the feature artifacts, then delete it from the root. In Engram-only mode, store
  its final contents as an observation and remove the root file.
- **One file per session**: if a new change starts in the same session, the file
  is updated to reflect the active change. The header always names the current
  change.

## Format

Write Markdown with this exact top-level structure:

```markdown
# Session Status

- **Session started**: <ISO 8601 UTC>
- **Last updated**: <ISO 8601 UTC>
- **Active change**: <change-name>
- **Current phase**: <phase> (<in_progress | completed>)
- **Next recommended**: <phase or action>

## Preflight
- Execution mode: <interactive | auto>
- Artifact store: <openspec | engram | both>
- Chained PR strategy: <ask-always | single-pr-default | force-chained | auto-forecast>
- Review budget: <N> lines
- Web project (Playwright): <yes | no | unknown>

## Phase History
- [x] explore — completed <ts>
- [x] propose — completed <ts>
- [ ] spec — in_progress <ts>
- [ ] design
- [ ] tasks
- [ ] apply
- [ ] verify
- [ ] judge
- [ ] archive

## Artifacts
- proposal: openspec/changes/<name>/proposal.md
- specs: openspec/changes/<name>/specs/<domain>/spec.md
- features (Gherkin): openspec/changes/<name>/specs/<domain>/*.feature
- design / tasks / verify-report: <paths as they are produced>

## Open Questions / Blockers
- <bullet list, or "none">

## Resume Hint
<One or two sentences: what the next agent should do first to continue.>
```

## Rules

- Use ISO 8601 UTC timestamps everywhere.
- Update via atomic write (temp file + rename).
- Keep it concise — it is a resume aid, not a transcript. Link to artifacts by
  path rather than copying their content.
- NEVER commit `SESSION_STATUS.md` from the root as a permanent file; it is a
  working-session artifact that gets archived (moved) at the end. It MAY be left
  uncommitted at the root between sessions for crash recovery.
- The orchestrator owns this file. Sub-agents report their results; the
  orchestrator records them here on each transition.
