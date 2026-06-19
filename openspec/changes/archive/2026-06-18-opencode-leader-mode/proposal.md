# Proposal: Opencode archon-leader mode + configurable leader model (Slice 2)

## Intent

The opencode agent has no archon-driven "leader" primary agent, so users must
manually wire orchestration. Opencode supports custom **primary agents** (under the
top-level `agent` key) that auto-appear in its Tab switcher. This change makes
`archon init` (opencode only) write an `archon-leader` primary agent pointed at the
project `AGENTS.md`, using a user-chosen leader model stored in `.archon/config.yaml`.

## Scope

### In Scope
- New `models.leader` field (full `provider/model-id`) on `ModelConfig`; threaded through `Config.Clone()` + roundtrip fixture + `.archon/config.yaml` serialization.
- `archon init` (opencode only): additively merge an `archon-leader` primary agent into the **project** `opencode.json` (create if absent); `prompt: "{file:./AGENTS.md}"`, `model = models.leader`. Register created/modified paths in rollback manifest. Skip when `models.leader` is empty or agent != opencode.
- TUI: "Leader model" input in the Models tab (opencode only); save/regenerate runs the SAME merge so config and `opencode.json` never drift.

### Out of Scope (deferred, not foreclosed)
- Slice 3: dynamic detection of installed agents/models.
- `archon update` writing the opencode agent (stays skill-only; preserves "update never rewrites orchestrator/user config").
- Leader mode for claude/agents/codex (clean no-op).

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `harness-init`: `archon init` for the opencode agent MUST additively/idempotently merge an `archon-leader` primary agent into the project `opencode.json` using `models.leader`, register it for rollback, and be a no-op for other agents or when `models.leader` is empty.

## Approach

- Config: `Leader string` on `ModelConfig` → `models.leader`, storing the FULL `provider/model-id` verbatim (no prefix stripping). `NormalizeModel`/`Validate` used only for advisory TUI warnings.
- Writer: one shared `mergeOpencodeAgent` (new `internal/initcmd/opencode_mode.go`) — read existing JSON into `map[string]any`, set ONLY `agent.archon-leader`, atomic temp+rename. Called by both init `Run()` (after `writeTemplate`, opencode-gated) and the TUI save/regenerate path.
- TUI: single `textinput` in `models_tab.go`, active only for opencode.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modified | Add `Leader` to `ModelConfig` (`models.leader`). |
| `internal/config/config.go` | Modified | Copy `Leader` in `Config.Clone()`. |
| `internal/config/config_test.go` | Modified | Extend `TestConfig_CloneRoundtrip` fixture. |
| `internal/initcmd/opencode_mode.go` | New | Additive/atomic `mergeOpencodeAgent` JSON merge. |
| `internal/initcmd/init.go` | Modified | Call merge after `writeTemplate` (opencode-gated); thread `models.leader`; register path in `buildRollbackManifest`. |
| `internal/tui/models_tab.go` | Modified | Leader-model input (opencode only). |
| `internal/tui/model.go` | Modified | Run merge in `saveConfig`/`regenerateTemplate` for opencode. |
| `internal/initcmd/opencode_mode_test.go` | New | Merge idempotency / non-clobber / non-opencode no-op. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Merge clobbers unrelated keys in user `opencode.json` | High | Read-modify-write only `agent.archon-leader`; atomic temp+rename; idempotency test. |
| init/TUI write-path drift | Med | Single shared `mergeOpencodeAgent`; both paths call it. |
| New config field dropped on update | Med | Add to `Clone()` + roundtrip fixture (guarded test). |
| Non-opencode regression | Med | Gate on agent; explicit no-op test. |
| **Exceeds 400-line budget (D1)** | Med | Spans config + writer + TUI + tests; likely 350-450 lines. Heads-up: may need a chained PR (config+writer first, TUI second) — re-forecast at tasks. |

## Rollback Plan

- Revert `model.go`/`config.go` (drop `Leader`) and test edits.
- Delete `internal/initcmd/opencode_mode.go` + its test; revert init.go/model.go/models_tab.go.
- Per-project: `archon`'s `.archon/rollback.json` lists the written `opencode.json` path so init rollback removes/restores it. No config migration to undo.

## Dependencies

- Opencode external schema (config location, `agent` shape, `{file:...}` prompt, `provider/model-id`) — CONFIRMED in exploration addendum (2026-06-18).

## Success Criteria

- [ ] `models.leader` survives `Clone()`/roundtrip and serializes under `models:`.
- [ ] `archon init --agent opencode` with a leader model writes/merges `agent.archon-leader` into project `opencode.json` (creating it if absent) without clobbering existing keys; re-running is idempotent.
- [ ] Non-opencode agents and empty `models.leader` write nothing.
- [ ] Written path appears in the rollback manifest.
- [ ] TUI Leader-model save produces a byte-identical merge to init (no drift).
- [ ] `go test ./...` passes.
