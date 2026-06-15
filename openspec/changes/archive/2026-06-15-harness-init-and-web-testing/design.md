# Design: Harness init UX, session status, Gherkin & web testing

## Key Decisions

### Config schema
- Add a `Playwright{Enabled, TestDir, BaseURL}` struct to `internal/config`, parallel
  to `MutationTesting`, marshalled under `playwright:` in `.archon/config.yaml`.
- Curate models into ordered `ClaudeModels` and `OpencodeModels` slices; `KnownModels`
  is derived from them. Validation stays advisory (warn, never reject) so free-form
  models keep working.

### init flow (`internal/initcmd`)
- `detectAgent` no longer requires the folder: an explicit `--agent` wins and the
  folder is created via `ensureAgentDir`. Without a flag and with nothing detected,
  init errors and asks for `--agent`.
- Guard the orchestrator file BEFORE any work: if it exists and `OverwriteTemplate`
  is false, return `ErrTemplateExists` so a declined replace leaves the project
  untouched. The CLI prompts `[y/N]`; `--force` sets `OverwriteTemplate`.

### TUI (`internal/tui`)
- New `PlaywrightTab` mirroring the mutation tab (toggle + test_dir + base_url).
- Models tab: cycle a static catalog with `ctrl+n`/`ctrl+p`; typing remains free-form.
- `archon tui` launches with a default config when none exists, so a project can be
  initialized from the Agent tab. After init, the config is reloaded into the model so
  a later save does not clobber it. An existing orchestrator file triggers an
  in-TUI replace confirmation.

### Skills
- New `_shared/session-status-contract.md`; `harness-workflow` writes `SESSION_STATUS.md`
  per transition; `sdd-archive` moves it into the archive.
- `sdd-spec` emits formal Gherkin `.feature` files. `sdd-explore` determines web/not-web/
  unknown. `sdd-apply` generates Playwright specs from `@web` scenarios. `harness-judge`
  runs the Playwright suite as a post-judgment gate.
- Commit-authorship rule added to orchestrator templates, `work-unit-commits`, `sdd-apply`.

## Alternatives Considered

- **Per-input model dropdown overlay** in the TUI — rejected as heavier UX; the
  `ctrl+n`/`ctrl+p` catalog cycle keeps free-form typing intact without modal state.
- **Auto-detect web only** — rejected in favor of an explicit `playwright.enabled`
  flag with explore-time detection and a preflight question for blank projects, so the
  decision is deterministic and reviewable.

## Non-goals

- Playwright is never enabled for non-web projects.
- The harness does not author product code from the Playwright gate; it only executes.
