# Proposal: Harness init UX, session status, Gherkin & web testing

## Why

Seven gaps were identified in the Archon harness and CLI/TUI:

1. No session-level resume point — context is lost if the agent closes mid-session.
2. Test/use cases were not authored in formal Gherkin, only loose Given/When/Then bullets.
3. No init switch to enable Playwright web-test generation/execution (parallel to mutation testing).
4. The harness did not generate or run Playwright tests for web projects from the use cases.
5. `archon init` (CLI and TUI) depended on the agent folder pre-existing and could not bootstrap a blank project.
6. An existing `CLAUDE.md`/`AGENTS.md` was overwritten silently, with no chance to decline.
7. Model selection offered only free-text input — no curated static list of Claude / Opencode Go models.

A cross-cutting hard rule was also missing: commits made through the harness must carry only the user's authorship.

## What Changes

- **harness-init**: init creates the agent folder when missing; guards an existing orchestrator file with a replace prompt (and aborts if declined); adds a `--playwright` flag; offers a curated static model catalog (Claude + Opencode Go) plus free-form entry in the TUI; adds a Playwright TUI tab.
- **harness-session-status**: a `SESSION_STATUS.md` file at the repo root, updated on every phase transition, read first on resume, and archived with the change.
- **harness-testing**: `sdd-spec` emits formal Gherkin `.feature` files; web projects are detected in `sdd-explore`; Playwright specs are generated from Gherkin during apply and executed as a judge-phase gate after verify; controlled by `playwright.enabled`.
- **harness-commits**: commits authored solely by the user — no `Co-Authored-By` or tool attribution.

## Capabilities

### New Capabilities
- `harness-session-status` — per-session resume file lifecycle.
- `harness-testing` — Gherkin-formal specs + web detection + Playwright generation/execution.
- `harness-commits` — commit authorship policy.

### Modified Capabilities
- `harness-init` — folder creation, existing-file guard, Playwright flag, static model selection.

## Impact

- **Go**: `internal/config` (Playwright struct, curated models), `internal/initcmd` (folder creation, template guard, playwright flag), `internal/tui` (Playwright tab, model catalog, init-without-config), `cmd/archon` (flags, config keys, prompt), `internal/status`.
- **Skills**: new `_shared/session-status-contract.md`; updates to `harness-workflow`, `harness-judge`, `sdd-explore`, `sdd-spec`, `sdd-verify`, `sdd-tasks`, `sdd-apply`, `sdd-archive`, `work-unit-commits`, `_shared/openspec-convention.md`; orchestrator templates.
- **Risk**: Low for non-web projects (Playwright opt-in, default off). Behavior change: existing orchestrator files are no longer overwritten silently.
