---
name: sdd-init
description: "Trigger: sdd init, iniciar sdd, openspec init. Initialize SDD context, testing capabilities, registry, and persistence."
disable-model-invocation: true
user-invocable: false
license: MIT
metadata:
  
  version: "3.0"
  delegate_only: true
---

> **ORCHESTRATOR GATE**: If you loaded this skill via the `skill()` tool, you are
> the ORCHESTRATOR — STOP. Do NOT execute these instructions inline. Delegate to
> the dedicated `sdd-init` sub-agent using your platform's delegation primitive
> (e.g., `task(...)`, sub-agent invocation, etc.). This skill is for EXECUTORS
> only.

## Executor Override

If you ARE the `sdd-init` sub-agent (NOT the orchestrator), the gate above does NOT apply to you. Continue with the phase work below. Do NOT delegate. Do NOT call the Skill tool. You are the executor — execute.

## Language Domain Contract

Generated technical artifacts default to English. Do not inherit the user's conversational language or the active persona's regional voice for SDD artifacts unless the user explicitly requests that artifact language or the project convention requires it.

If Spanish technical artifacts are explicitly requested, use neutral/professional Spanish unless the user explicitly asks for a regional variant.

Public/contextual comments follow the target context language by default. Explicit user language or tone overrides win; Spanish comments default to neutral/professional Spanish unless the user or target context clearly calls for regional tone.

## Activation Contract

Run this phase when the orchestrator/user asks to initialize SDD in a project. You are the phase executor: do the work yourself, do not delegate, and do not behave like the orchestrator.

## Hard Rules

- Detect the real stack, conventions, architecture, testing tools; never guess.
- Follow `../_shared/openspec-convention.md` and write file artifacts.
- Always persist testing capabilities in `openspec/config.yaml` `testing:`.
- Always build `.atl/skill-registry.md`.
- If `openspec/` already exists, report what exists and ask before updating it.
- `openspec/map.md` is seeded by the Go init step (`createOpenSpecDir` in `internal/initcmd/init.go`), not hand-created by this skill — rely on it existing after `archon init` runs.

## Decision Gates

| Input | Action |
|---|---|
| strict TDD marker/config found | Use that value. |
| no marker/config but test runner exists | Default `strict_tdd: true`. |
| no test runner | Set `strict_tdd: false` and explain unavailable. |

## Track Parameter

`sdd-init` accepts an optional `track` parameter (default `full`). When the user's request signals a bug fix ("Es un bug", "fix a bug", "bug report"), the orchestrator passes `track: bugfix`. Write the resolved track value into `state.yaml` at change creation:

```yaml
track: bugfix   # or omit for full (default)
phase: explore
status: in_progress
```

Supported values: `full` | `bugfix`. Reject any other value with an error naming both supported values. No CLI flag is required — natural-language routing is the primary trigger; the `track` parameter is passed by the orchestrator when it initializes the change.

## Execution Steps

1. Inspect project files (`package.json`, `go.mod`, `pyproject.toml`, CI, lint/test config) and summarize stack/conventions.
2. Detect test runner, test layers, coverage, linter, type checker, and formatter.
3. Resolve Strict TDD from agent marker, `openspec/config.yaml`, detected runner fallback, or no-runner fallback.
4. Resolve `track` from the orchestrator-supplied parameter (default `full`). Write the resolved value into the new `state.yaml`.
5. Initialize persistence for the resolved mode.
6. Build `.atl/skill-registry.md` using the skill-registry scan rules.
7. Persist testing capabilities and project context.
8. Return the structured initialization envelope.

## Output Contract

Return `status`, `executive_summary`, `artifacts`, `next_recommended`, and `risks`. Include project, stack, persistence mode, Strict TDD status, testing capability table, artifact paths, registry path, and next `/sdd-explore` or `/sdd-new` step.

## References

- [references/init-details.md](references/init-details.md) — detection checklist, config skeleton, and output templates.
- `../_shared/openspec-convention.md` — openspec layout and rules.
