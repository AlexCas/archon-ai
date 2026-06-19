# Exploration: opencode-model-assignment

Bug: "El modo Archon Leader en opencode no reconoce ningún modelo que se le pase como
configuración." After `archon init` for the opencode agent, the per-phase / default
models written to `.archon/config.yaml` have no effect — opencode keeps using its
default model for every agent/phase.

## Current State

### How archon-ai handles models today (confirmed by code reading)

1. **`archon init` writes models ONLY into `.archon/config.yaml`.**
   - `cmd/archon/main.go:94-122` collects `--model` and `--model-<phase>` flags and
     passes them as `Options.ModelDefault` / `Options.ModelPhases`.
   - `internal/initcmd/init.go:60` calls `buildConfig`, which at `init.go:166-169`
     stores them under `Models.Default` and `Models.Phases` (`internal/config/model.go`
     `ModelConfig`).
   - The config is serialized to `.archon/config.yaml` (`cfg.Save()` at `init.go:62`).

2. **Nothing ever reads `cfg.Models` to influence the agent runtime.** A repo-wide
   search for consumers of `cfg.Models` returns only:
   - `internal/status/display.go:34-47` — prints the models for `archon status`.
   - `cmd/archon/config.go:116-172` — `archon config get/set` for models.
   - `internal/tui/models_tab.go` — TUI editor that reads/writes the same YAML.
   - `internal/config/config.go:61-65` — deep-copy in `Clone()`.
   None of these emit anything opencode reads. `.archon/config.yaml` is a private
   archon file; opencode does not read it.

3. **archon-ai never generates or merges an `opencode.json`.** The only JSON file the
   repo touches is `.archon/rollback.json` (`internal/config/rollback.go`). There is
   no `opencode` package, no JSON deep-merge, and no `encoding/json` import anywhere in
   `internal/` outside rollback. So no `agent.<name>.model` wiring ever reaches opencode.

4. **The rendered `AGENTS.md` template has no "Phase Models" guidance.**
   `internal/initcmd/templates.go` — `orchestratorSections` + `agentsTemplate`
   (lines 117-134) contain phase order, preflight, gates, and rules, but **no** Phase
   Models advisory block telling the orchestrator to request a model per delegated
   phase. (For comparison, the committed root `CLAUDE.md` in this repo DOES have a
   "## Phase Models" section, but the generated opencode `AGENTS.md` does not.)

### Why opencode ignores the config (root cause confirmed)

opencode selects per-agent models from its own settings file
`~/.config/opencode/opencode.json` under `agent.<name>.model` (provider-qualified,
e.g. `anthropic/claude-sonnet-4-...`), plus a top-level `model` as the root default.
archon-ai writes none of this. Therefore:

- There is no `agent.<phase>` sub-agent map → opencode runs the orchestrator as a
  single primary agent using its built-in default model, and any delegated phase
  inherits that same default. The `.archon/config.yaml` model values are inert.

Additional structural note: `internal/initcmd/init.go:49` extracts skills to
`~/.config/opencode/skills/` and symlinks them into `.opencode/skills/`. So opencode
is genuinely the target runtime, but archon stops at skills + an `AGENTS.md` file and
never produces the `opencode.json` agent/model overlay that would make "Archon Leader"
delegate to per-phase sub-agents with assigned models.

**Conclusion: all three orchestrator hypotheses are confirmed, verbatim.** The bug is
a missing integration layer, not a defect in existing code.

## How gentle-ai solves this (the reference to follow)

Data flow: `.archon`-style model choices → provider-qualified IDs → injected into an
embedded JSON overlay → deep-merged into `~/.config/opencode/opencode.json`.

1. **Overlay assets** — `internal/assets/opencode/sdd-overlay-single.json` (and
   `-multi.json`). The overlay defines under `"agent"`:
   - one `"mode": "primary"` orchestrator (`gentle-orchestrator`) with
     `"prompt": "{file:./AGENTS.md}"` and a `permission.task` map (using a
     `"__replace__"` sentinel) that allows delegation ONLY to the named SDD/JD
     sub-agents and denies `"*"`;
   - one `"mode": "subagent", "hidden": true` entry per SDD phase
     (`sdd-init`, `sdd-explore`, … `sdd-archive`, `sdd-onboard`) and per JD agent,
     each with an Executor-Override `prompt` ("You are an SDD executor for the X
     phase … Read your skill file at ~/.config/opencode/skills/<phase>/SKILL.md")
     and a `tools` map.

2. **Provider resolution** — `internal/opencode/models.go`. opencode needs
   `provider/model-id`, but archon stores bare names (`gpt-4o`, `claude-sonnet-4`).
   gentle-ai resolves this from opencode's OWN model catalog cache:
   - `DefaultCachePath()` → `~/.cache/opencode/models.json`; `LoadModels` /
     `LoadModelsOrEmpty` parse it into `map[providerID]Provider{Models: map[id]Model}`.
   - `DetectAvailableProviders` filters to providers the user can actually use
     (OAuth in `~/.local/share/opencode/auth.json`, env vars like `ANTHROPIC_API_KEY`,
     or the built-in `opencode` provider) AND that have a tool-call-capable model.
   - `MergeCustomProviders` folds in custom providers declared in the user's
     `opencode.json`. `FilterModelsForSDD` keeps only tool-call models.
   - The TUI then offers (provider, model, optional reasoning `variant`) and persists
     a `model.ModelAssignment{ProviderID, ModelID, Effort}` whose `FullID()` returns
     `provider/model-id` (`internal/model/model_assignment.go`).

3. **Injection + merge** — `internal/components/sdd/inject.go`.
   - `injectModelAssignments` (line 2189) walks the overlay's `"agent"` map and, for
     EACH agent, applies this decision tree:
     1. **Explicit TUI assignment wins** → set `"model": assignment.FullID()` and
        `"variant": Effort` (or `""`).
     2. **Agent already a key in the user's existing `opencode.json`** → skip; let the
        deep merge preserve whatever the user has (possibly no model at all).
     3. **Otherwise, fall back to the root model** (`readOpenCodeRootModel`, the
        top-level `"model"` in `opencode.json`) → inject it so the sub-agent does not
        silently inherit the orchestrator model. JD agents are excluded from this
        fallback to keep judge diversity.
   - The injected overlay is then deep-merged into `opencode.json` via
     `internal/components/filemerge/json_merge.go` `MergeJSONObjects` (recursive
     object merge; `"__replace__"` sentinel forces whole-key replacement so the
     delegation `permission.task` map is replaced, not merged).

Key insight on the bare-name problem: gentle-ai does NOT keep its own model-name map.
It reads opencode's own `~/.cache/opencode/models.json` to enumerate real
`provider/model` IDs the user is authenticated for, and stores the fully-qualified ID.

## The gap in archon-ai (what is missing)

To make the configured models reach opencode, archon-ai needs the entire bridge that
gentle-ai has and archon lacks:

1. **An embedded opencode overlay asset** defining the primary `archon-orchestrator`
   (or similar) agent + one `mode: subagent` per SDD phase with Executor-Override
   prompts pointing at the extracted skills. (archon currently ships skills only.)
2. **A JSON deep-merge helper** (`MergeJSONObjects` equivalent) — none exists today.
3. **Provider/model resolution**: turn bare names in `.archon/config.yaml`
   (`claude-sonnet-4`, `gpt-4o`) into `provider/model-id`. Today archon has only the
   bare-name `KnownModels` set in `internal/config/model.go`.
4. **An injection step** that writes `agent.<phase>.model` (+ optional `variant`) into
   the overlay, with a decision tree (explicit phase assignment > existing user agent
   preserved > default model fallback).
5. **A write/merge into `~/.config/opencode/opencode.json`** wired into `archon init`
   (and ideally a re-runnable `archon sync` path so editing models in the TUI / via
   `archon config set` re-applies them).
6. **"Phase Models" guidance in the generated `AGENTS.md`** so the orchestrator agent
   knows to delegate per phase (defense-in-depth, matching the root `CLAUDE.md`).

## The bare-name → provider/model-id problem

`.archon/config.yaml` stores bare names (`gpt-4o`, `claude-sonnet-4`); opencode needs
`provider/model-id`. Three resolution strategies:

| Strategy | How | Pros | Cons |
|----------|-----|------|------|
| A. Read opencode cache | Parse `~/.cache/opencode/models.json`, match bare name → `provider/id` | Exactly the IDs the user is authenticated for; future-proof; matches gentle-ai | Cache may be absent until opencode runs once; need a fallback |
| B. Static map in archon | Extend `KnownModels` into `name → provider/id` | No external file dependency; deterministic tests | Goes stale; only covers hardcoded names; wrong if user uses a different provider |
| C. Require qualified input | Have users store `provider/model` directly in `.archon/config.yaml` | Zero resolution code; unambiguous | Breaks existing bare-name config and TUI/CLI UX; not backward compatible |

Recommended: **A with a B-style fallback.** Resolve against
`~/.cache/opencode/models.json` when present (reusing gentle-ai's `LoadModelsOrEmpty`
approach in a small `internal/opencode` package). When the cache is missing or the
name is unambiguous, fall back to a small static provider map for the handful of names
already in `KnownModels`. Also accept an already-qualified `provider/model` value
verbatim (pass-through) so power users are never blocked.

## Affected Areas

- `internal/initcmd/init.go` — `Run` must add an opencode overlay-injection +
  `opencode.json` merge step (gated on `agentName == "opencode"`).
- `internal/initcmd/templates.go` — add a "Phase Models" section to `agentsTemplate`
  (and optionally clarify the orchestrator/sub-agent delegation expectation).
- **NEW** `internal/opencode/` (package) — `opencode.json` path helpers, models-cache
  loader, bare-name → `provider/model` resolver. Port the minimal slice of
  gentle-ai's `internal/opencode/models.go`.
- **NEW** overlay asset, e.g. `internal/initcmd/assets/opencode/sdd-overlay.json`
  (embedded via `//go:embed`) — primary orchestrator + per-phase subagents.
- **NEW** JSON deep-merge helper (port `internal/components/filemerge/json_merge.go`
  with the `__replace__` sentinel) — likely `internal/filemerge/` or under
  `internal/opencode/`.
- **NEW** injection function (port `injectModelAssignments`'s decision tree, simplified
  to SDD phases only; archon has no JD agents).
- `cmd/archon/main.go` — optionally add an `archon sync` subcommand to re-apply the
  overlay after TUI/CLI model edits; also surface a warning when the opencode cache is
  absent.
- `internal/config/rollback.go` / `buildRollbackManifest` — the merged
  `opencode.json` is a SHARED global file (user may have prior content). Rollback must
  NOT delete it; it should restore a backup of the pre-merge state. This needs a backup
  step in `init.go` similar to the AGENTS.md backup path that already exists.
- `internal/tui/models_tab.go` — currently edits bare names; after the fix it should
  ideally drive provider/model selection (longer-term; not required for the minimal fix
  if bare-name resolution is in place).

## Approaches

1. **Full gentle-ai port (overlay + cache-driven resolution + merge + sync + TUI
   provider picker).**
   - Pros: behavior-complete; matches the proven reference; handles auth-aware
     provider detection and reasoning variants.
   - Cons: large for archon's tiny codebase; pulls in models cache, auth detection,
     custom-provider merge, variants plugin — most unused by archon today.
   - Effort: High.

2. **Minimal bridge (RECOMMENDED).** Embed one `sdd-overlay.json` (primary
   `archon` orchestrator + per-phase subagents), add a small JSON deep-merge, a
   bare-name→`provider/model` resolver that prefers `~/.cache/opencode/models.json`
   and falls back to a static map / qualified pass-through, an injection step with the
   3-case decision tree, and merge into `~/.config/opencode/opencode.json` during
   `archon init` (with a backup for rollback). Add the "Phase Models" section to
   `AGENTS.md`. Provide an `archon sync` to re-apply after edits.
   - Pros: fixes the bug end-to-end; scoped to opencode; reuses archon's existing
     embed + symlink patterns; testable with golden JSON.
   - Cons: introduces JSON-merge and a new `internal/opencode` package; must handle the
     shared-file backup/rollback carefully.
   - Effort: Medium.

3. **Docs-only / AGENTS.md guidance, no opencode.json.**
   - Pros: tiny.
   - Cons: does NOT fix the bug — opencode still won't read per-agent models without an
     `agent.<name>.model` map. Rejected.

## Recommendation

Approach 2 (minimal bridge). Concretely:

1. Add `internal/opencode/` with: `opencode.json` + models-cache path helpers; a
   `Resolve(bareOrQualifiedName) -> "provider/model"` function that (a) passes through
   values already containing `/`, (b) matches against `~/.cache/opencode/models.json`
   when present, (c) falls back to a small static `name → provider/model` map covering
   the existing `KnownModels`.
2. Add an embedded `sdd-overlay.json` with a `mode: primary` orchestrator
   (`prompt: {file:./AGENTS.md}`, `permission.task` allow-list of sdd phases) and one
   `mode: subagent, hidden: true` entry per phase with an Executor-Override prompt
   pointing to `~/.config/opencode/skills/<phase>/SKILL.md`.
3. Add a JSON deep-merge with `__replace__` support (port from filemerge).
4. Add `injectModels(overlay, defaultModel, phases, existingAgentKeys)` implementing
   the explicit > existing > default decision tree, mapping `.archon` phase names to
   the `sdd-<phase>` agent keys and writing `model` (+ `variant` if archon later adds
   effort).
5. In `init.go`, when `agentName == "opencode"`: back up the current
   `~/.config/opencode/opencode.json`, inject + deep-merge the overlay, write it, and
   record the backup in the rollback manifest (do NOT add the file itself to
   `CreatedPaths` for deletion).
6. Add the "## Phase Models" section to `agentsTemplate`.
7. Add `archon sync` (or extend init) so model edits via TUI / `archon config set`
   re-apply to `opencode.json`.

This resolves the bare-name mapping (step 1) explicitly while keeping the surface small.

## Risks

- **Shared global file**: `~/.config/opencode/opencode.json` may already contain the
  user's own agents/providers/settings. The merge MUST be additive and backup-driven;
  rollback must restore the backup, never delete the file. This is the highest-risk
  area.
- **Stale model cache**: `~/.cache/opencode/models.json` may be absent on a fresh
  machine (opencode hasn't run). Resolution must degrade gracefully (static fallback /
  pass-through) and warn rather than fail init.
- **Wrong provider guess**: a bare name like `claude-sonnet-4` could exist under more
  than one provider; without auth-aware detection the static fallback may pick the
  wrong provider. Mitigate by preferring the cache + authenticated providers, and by
  letting users store a qualified `provider/model` to override.
- **Overlay/skill name drift**: subagent prompt paths point at
  `~/.config/opencode/skills/<phase>/SKILL.md`; these must stay in lockstep with the
  extracted skill folder names, or delegation breaks silently.
- **Backward compatibility**: existing `.archon/config.yaml` files store bare names; the
  resolver must keep accepting them so re-running init/sync on an existing project does
  not error.
- **Variants/effort**: archon's `ModelConfig` has no effort field today; the overlay
  should omit `variant` rather than emit `""` unless/until archon adds effort, to avoid
  unintended overrides.

## Ready for Proposal

Yes. Root cause is confirmed (missing opencode.json overlay/injection layer; bare-name
vs provider/model mismatch). The orchestrator should tell the user we will add a minimal
opencode-model bridge: an embedded SDD agent overlay, a bare-name→provider/model
resolver backed by opencode's own model cache, a JSON deep-merge into
`~/.config/opencode/opencode.json` during `archon init` (with backup-based rollback),
"Phase Models" guidance in `AGENTS.md`, and a re-apply (`archon sync`) path. Open
question to confirm with the user before propose: (a) is single-overlay sufficient or do
they want gentle-ai's multi-profile concept, and (b) acceptable fallback behavior when
the opencode model cache is missing.
