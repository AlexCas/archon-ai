# Persistence Contract (shared across all SDD skills)

## Storage Model

Artifacts are always stored as files under `openspec/changes/{change-name}/`, git-tracked, and team-shareable. OpenSpec is the sole persistence mode — there is no engram, hybrid, or none path.

## OpenSpec Artifact Layout

Artifacts live at deterministic paths inside the change folder:

| Artifact | Path |
|----------|------|
| Exploration | `openspec/changes/{change-name}/exploration.md` |
| Proposal | `openspec/changes/{change-name}/proposal.md` |
| Delta specs | `openspec/changes/{change-name}/specs/{domain}/spec.md` |
| Gherkin features | `openspec/changes/{change-name}/specs/{domain}/{domain}.feature` |
| Design | `openspec/changes/{change-name}/design.md` |
| Tasks | `openspec/changes/{change-name}/tasks.md` |
| Apply progress | `openspec/changes/{change-name}/apply-progress.md` |
| Verify report | `openspec/changes/{change-name}/verify-report.md` |
| Archive report | `openspec/changes/{change-name}/archive-report.md` |

Follow `skills/_shared/openspec-convention.md` for the full layout and rules.

## State Persistence

Write and read DAG state via:

```
openspec/changes/{change-name}/state.yaml
```

## Sub-Agent Context Rules

Sub-agents launch with a fresh context and NO access to the orchestrator's instructions.

Who reads, who writes:
- SDD phase (with dependencies): sub-agent reads artifact files directly; sub-agent writes its artifact.
- SDD phase (no dependencies, e.g. explore): sub-agent writes its artifact.

Sub-agents must read the concrete `artifactPaths` or `contextFiles` supplied by the orchestrator's structured status rather than assuming fixed filenames.

## Orchestrator Prompt Instructions for Sub-Agents

SDD (with dependencies):
```
Read these artifacts before starting:
  openspec/changes/{change-name}/proposal.md
  openspec/changes/{change-name}/spec.md (or specs/{domain}/spec.md)
  openspec/changes/{change-name}/design.md
  openspec/changes/{change-name}/tasks.md

PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, write your artifact to the path defined in
openspec-convention.md for this phase. If you return without writing the
artifact file, the next phase CANNOT find it and the pipeline BREAKS.
```

SDD (no dependencies):
```
PERSISTENCE (MANDATORY — do NOT skip):
After completing your work, write your artifact to the path defined in
openspec-convention.md for this phase. If you return without writing the
artifact file, the next phase CANNOT find it and the pipeline BREAKS.
```

## Sub-Agent Response Ordering

When a sub-agent persists artifacts (via file writes), the persistence write MUST happen BEFORE the final text response. The sub-agent's absolute last output must be text, never a tool call.

## Skill Registry

The orchestrator pre-resolves skill paths from the skill registry and injects them as `## Skills to load before work` in the launch prompt. Sub-agents read those exact `SKILL.md` files before task-specific work.

To generate/update: run the `skill-registry` skill, or run `sdd-init`.

Sub-agent skill loading: check for a `## Skills to load before work` block in your prompt — if present, read those exact files. If not present, check for `SKILL: Load` instructions as a fallback. If neither exists, proceed without — this is not an error.

## Detail Level

The orchestrator may pass `detail_level`: `concise | standard | deep`. This controls output verbosity but does NOT affect what gets persisted — always persist the full artifact.
