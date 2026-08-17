# Proposal: Graphify integration — Slice A (advisory code-graph gate)

> Preflight (cached): mode=interactive · store=openspec · PR=ask-always ·
> budget=800 · Playwright=off · Impeccable=off. Scope: **Slice A, Approach 2**
> (single PR, TUI tab deferred). Built on [exploration](exploration.md).

## Intent

Give `sdd-explore` and `sdd-tasks` an optional, deterministic *code graph* of the
repo (tree-sitter AST via [Graphify](https://github.com/Graphify-Labs/graphify);
no embeddings) so comprehension and slice boundaries beat blind grep. Opt-in,
default-off, **advisory / never-blocking** — mirrors the `Impeccable` precedent.

## Scope

### In Scope
- `graphify.*` config block in `internal/config/config.go` (follows the scalar
  `Playwright`/`Security` shape — no severity, no `Load()` validation).
- `--graphify` init flag (`cmd/archon/main.go`) + `archon config` get/set + both
  "supported keys" strings (`cmd/archon/config.go`), plus `archon status` block.
- Preflight **group G** (Spanish, own arrow-key, recommended "No") in both
  `orchestratorRules` blocks and the preflight section (`internal/initcmd/templates.go`).
- New thin skill `skills/graphify/SKILL.md` — decides WHEN/HOW to call the tool,
  reimplements nothing.
- Read-only advisory consumption in exactly two phases: `sdd-explore`, `sdd-tasks`.
- Skill auto-indexed by `skill-registry` regeneration; regenerated root
  `CLAUDE.md`/`AGENTS.md` (skill count 25→26, dynamic — no Go constant).
- Tracked excerpt: `graph-report.excerpt.md` (see Approach).

### Out of Scope (deferred)
- **TUI tab** (`internal/tui/graphify_tab.go` + `model.go` wiring + test, ~320
  lines) — parked trivial follow-up.
- Slice B (graph diff in verify, edge evidence in judge); Slice C (bridge
  `internal/mapgen` spec graph ↔ code graph in archive).

## Capabilities

### New Capabilities
- `graphify-integration`: opt-in code-graph gate — config block, `--graphify`
  flag, preflight group G, the `graphify` skill, advisory explore/tasks
  consumption, tracked report excerpt.

### Modified Capabilities
- `harness-init`: init gains the `--graphify` flag and `graphify` config block.
- `sdd-phase-skills`: explore + tasks gain a gated advisory graph step.

## Approach

**Config fields** (`graphify:` block): `enabled` (bool, false — the gate);
`auto_install` (bool, false — may run `uv`/`graphify install`, never silently);
`version` (string, `"v8"` — pin the fast-moving upstream against API drift);
`output_dir` (string, `.archon/graphify` — gitignored, keeps `graph.json`/`.html`
untracked); `semantic` (bool, false — one switch for all LLM features; off = pure
local AST).

**Surface separation** (the Impeccable failure mode): harness may shell only
`graphify extract|query|path|explain`. `/graphify` is an agent slash-command —
NEVER shelled. `python -m graphify.serve` (MCP) is a future option, not a Slice A
dependency (headless/cron auth caveat).

**Tracked excerpt**: `sdd-explore` writes a ≤40-line (~2 KB) excerpt of
`GRAPH_REPORT.md` to `openspec/changes/<change>/graph-report.excerpt.md` for
review traceability; refreshed whenever the graph is re-extracted for the change.
Full `graph.json`/`graph.html`/`GRAPH_REPORT.md` stay untracked in `.archon/graphify/`
(`.gitignore:10` covers `.archon/`).

**Staleness**: skill-side only — advisory mtime-vs-HEAD note, no config knob.

**Degradation** (invariant): missing binary / no Python|uv / failed extraction /
unparseable output / stale graph → emit one-line note, fall back to grep/read,
**continue**. Never fail a phase, never block a gate, never return `blocked`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/config.go` (+test) | Modified | `Graphify` struct, `Config` field, `Clone()` copy |
| `cmd/archon/config.go` (+test) | Modified | get/set cases + both key strings |
| `cmd/archon/main.go`, `internal/initcmd/init.go` | Modified | `--graphify` flag → `Options` → `buildConfig` |
| `internal/initcmd/templates.go` (+test) | Modified | preflight group G + mapping paragraph |
| `internal/status/display.go` (+test) | Modified | Graphify status block |
| `CLAUDE.md`, `AGENTS.md` (root) | Modified | regenerated: group G + count 25→26 |
| `skills/graphify/SKILL.md` | New | thin orchestration skill |
| `skills/sdd-explore`, `skills/sdd-tasks` `SKILL.md` | Modified | gated advisory step |
| `openspec/changes/graphify-integration/graph-report.excerpt.md` | New | tracked report excerpt |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Estimate creeps past 800 budget | Low | ~516 lines w/o TUI tab; flag loudly if it grows |
| Young upstream (v8) schema drift | Med | `version` pin + advisory parse-fail fallback |
| Second runtime (Python/uv) friction | Med | default-off + no-install-at-init + advisory fallback |
| Shelling `/graphify` (agent verb) | Low | skill forbids it explicitly (Impeccable precedent) |
| Merge collision w/ `local-model-router` (config.go, templates.go) | Med | mechanical; second-merger resolves |
| Code-graph vs spec-graph (`mapgen`) naming conflation | Low | keep vocabulary distinct; enables Slice C |

## Rollback Plan

Single PR, additive and default-off. Revert the merge commit — no data migration,
no persisted state (outputs live in gitignored `.archon/graphify/`). Users on
`graphify.enabled: false` (the default) see zero behavior change either way.

## Dependencies

- Optional runtime: Python + `uv`/`pipx` and the `graphify` CLI, only when
  `enabled: true`. Node (for `npx impeccable`) already required — this is a second
  optional runtime, contained by default-off.

## Success Criteria

- [ ] `graphify.*` block round-trips through config `Load`/`Clone`, `archon config`, `archon status`.
- [ ] `--graphify` sets the flag at init; preflight renders group G in both orchestrator files.
- [ ] With `enabled: false`, explore/tasks behavior is byte-identical to today.
- [ ] With `enabled: true` but binary absent, both phases note-and-continue (no failure).
- [ ] `graph-report.excerpt.md` is written/tracked; `graph.json`/`graph.html` stay untracked.
