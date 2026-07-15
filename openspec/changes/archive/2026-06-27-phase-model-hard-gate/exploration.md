# Exploration: phase-model-hard-gate

**Change**: `phase-model-hard-gate`
**Date**: 2026-06-27

---

### Project Type

**Web testing**: not-web

This is a Go CLI tool with a Bubbletea TUI and no browser-facing surface.

---

### Current State

Archon generates two kinds of orchestrator files:

1. **`AGENTS.md`** (opencode) / **`CLAUDE.md`** (claude) — a single markdown prompt
   injected into the leader model. Both use the same template
   (`internal/initcmd/templates.go`, `orchestratorTrailer`, lines 163–171). When
   `PhaseModels` is non-empty, the template emits:

   ```
   ## Phase Models
   Advisory: when delegating an SDD phase, request the model below for that phase by
   passing `model: <id>` to the Agent/Task delegation tool. This is a preference, not a
   hard gate; if the platform cannot honor per-delegation model selection, proceed with
   the default model.
   ```

   Both `agentsTemplate` and `claudeTemplate` are identical strings (same
   `orchestratorSections + orchestratorTrailer`, lines 178 and 180). The "advisory,
   not a hard gate" language is injected into **both** harnesses from the same source.

2. **opencode**: `internal/initcmd/opencode_mode.go` — `mergeOpencodeAgent` /
   `MergeOpencodeAgent` reads `opencode.json`, writes `archon-leader` (mode `primary`)
   and one `archon-<phase>` entry per `config.ResolvePhaseModels` result (mode
   `subagent`, `hidden: true`, `model: pm.Model`, `variant: pm.Effort`). This is a
   **real hard gate** — opencode uses the agent key when invoked by name, so the model
   is bound at definition time, not at call-time. Writer is called from:
   - `internal/initcmd/init.go:101-109` (gated `agentName == "opencode"`)
   - `internal/tui/model.go:343-346` (gated `cfg.Agent == "opencode"`)

3. **claude**: There is **no** `.claude/agents/archon-<phase>.md` writer anywhere in
   `internal/initcmd/` or `internal/tui/`. The grep for `".claude/agents"` in `.go`
   files returns zero production hits. Only `SESSION_STATUS.md` mentions the path in
   prose. So for claude projects, no per-phase subagent files are ever generated.
   The leader (`CLAUDE.md`) runs all phases on its inherited model.

4. **Config and resolution**: `internal/config/model.go` holds `ModelConfig{Default,
   Leader ModelRef, Phases map[string]ModelRef}` (lines 80-84), `PhaseOrder` (8 phases,
   `judge` excluded, line 167), `ResolvePhaseModels` (lines 251-264), and `PhaseModel`
   struct (lines 196-200). `ModelRef.FullID()` emits `provider/model` when provider is
   set, or bare model otherwise. These are production-ready from the Foundation PR (PR
   #47, `structured-model-resolution`).

---

### Diagnosis Accuracy Check

All four points in the orchestrator's diagnosis are confirmed, with minor clarifications:

| Point | Verdict | Notes |
|-------|---------|-------|
| 1. Shared template emits "advisory, not a hard gate" for BOTH harnesses | **Confirmed** — `templates.go:163-171` is shared between `agentsTemplate` and `claudeTemplate`. | No per-harness divergence yet. |
| 2. opencode already has a real hard gate via `mergeOpencodeAgent` | **Confirmed** — `opencode_mode.go:63-121` writes `archon-<phase>` subagents with fixed model; called from `init.go:101` and TUI `model.go:343`. The advisory text in `AGENTS.md` is therefore incorrect/misleading for opencode too, since the REAL binding is in `opencode.json`. | |
| 3. No `.claude/agents` writer exists in any production Go file | **Confirmed** — zero hits when grepping `internal/` for `.claude/agents`. | The `.claude/` directory at repo root contains only a `skills/` subdirectory, not an `agents/` subdirectory. |
| 4. Config types are production-ready | **Confirmed** — `ModelRef`, `ModelConfig`, `PhaseOrder`, `ResolvePhaseModels`, `PhaseModel` all exist in `internal/config/model.go`. | `PhaseOrder` excludes `judge` (8 phases), matching the opencode writer. |

One inaccuracy to note: the diagnosis says `model:` frontmatter in Claude Code subagents
accepts full model IDs like `claude-opus-4-8`. This is accurate per current Claude Code
docs, but the platform precedence chain (`CLAUDE_CODE_SUBAGENT_MODEL` env > per-call
param > frontmatter > inherited) should NOT be mentioned in generated `CLAUDE.md`
content — it is implementation noise that confuses the leader. The exploration notes it
here for design awareness only.

---

### Affected Areas

- `internal/initcmd/templates.go` — shared template writer for both `CLAUDE.md` and
  `AGENTS.md`; the `orchestratorTrailer` "Phase Models" block (lines 162–171) needs to
  be split per-agent so each harness reflects its actual binding mechanism.

- `internal/initcmd/opencode_mode.go` — the proven pattern to mirror for the claude
  side. Exports `MergeOpencodeAgent` as the integration seam. The new claude writer
  should follow the same signature shape: `WriteClaude<Phase>Agents(projectDir string,
  models config.ModelConfig) (written []string, err error)`.

- `internal/initcmd/init.go` — init callsite (lines 99-109) gates opencode writer on
  `agentName == "opencode"`. A parallel claude gate (`agentName == "claude"`) is where
  the new claude writer hooks in. Rollback manifest registration follows the same pattern
  (line 105-108).

- `internal/tui/model.go` — TUI save path (lines 321-354) gates opencode merge on
  `cfg.Agent == "opencode"` (line 343). The same save function needs a parallel
  `cfg.Agent == "claude"` gate calling the new writer. No new exported seam is needed
  beyond the production function.

- `internal/config/model.go` — `PhaseOrder` (line 167), `ResolvePhaseModels` (lines
  251-264), `PhaseModel` struct (lines 196-200). Already correct and used by the
  opencode writer; the claude writer reuses these as-is.

- `internal/config/config.go` — no changes needed. `Config` struct already holds
  `Models ModelConfig`.

---

### Approaches

1. **Dedicated per-phase agent files — mirror opencode pattern**
   Write `.claude/agents/archon-<phase>.md` files (one per SDD phase) with frontmatter
   `model: <full-id>` plus a minimal prompt body. Reuse `config.ResolvePhaseModels` to
   enumerate phases; use `ModelRef.FullID()` for the model string. Add a new
   `WriteClaude<Phase>Agents` (or `writeClaudeAgents`) function in
   `internal/initcmd/` (e.g. `claude_mode.go`) mirroring `opencode_mode.go`. Export it
   so the TUI save path calls it. Update `init.go` and `model.go` save path to call it.
   Also rewrite the `## Phase Models` template block for `CLAUDE.md` to refer to
   `archon-<phase>` subagents as the actual binding, not an advisory param.
   - **Pros**: real hard gate matching opencode's pattern; agent files are version-controlled;
     consistent architecture across both harnesses; no leader cognitive load about model
     params.
   - **Cons**: generates up to 8 new files per init; the `.claude/agents/` directory must
     be created; requires rollback registration for each file; test surface grows.
   - **Effort**: Medium

2. **Rely on orchestrator's per-invocation `model` param only**
   Keep no per-phase agent files. Rewrite the template block in `CLAUDE.md` to instruct
   the leader to pass `model: <id>` when spawning each phase via the Task tool. This
   is purely a docs fix — no Go code changes beyond the template.
   - **Pros**: zero new files; simplest implementation.
   - **Cons**: NOT a hard gate — it relies on the leader correctly following instructions
     every single delegation call. Leader drift or hallucination breaks it silently.
     Inconsistent with opencode which has a real gate.
   - **Effort**: Low

3. **Both: per-phase agent files + corrected orchestrator docs**
   Generate `.claude/agents/archon-<phase>.md` (Approach A) AND update `CLAUDE.md` to
   name those agents as the binding mechanism (replacing the advisory block). The leader
   delegates to named agents (`archon-explore`, `archon-spec`, etc.); Claude Code routes
   to the correct file and honors its `model:` frontmatter. This is the full symmetry with
   opencode.
   - **Pros**: full hard gate; agent files enforce model even if leader instruction is
     imperfect; docs accurately describe reality; parity with opencode.
   - **Cons**: same as Approach A — more files, slightly larger rollback manifest.
   - **Effort**: Medium (same as A since docs fix is trivial)

---

### Recommendation

**Approach C (Both)** — generate `.claude/agents/archon-<phase>.md` per-phase files
AND update the `CLAUDE.md` `## Phase Models` block to reference named `archon-<phase>`
agents as the actual binding.

Rationale:
- opencode already ships a real hard gate; claude should have parity.
- Approach B (docs-only) is structurally fragile — it can silently break without any
  observable failure.
- The opencode pattern is already proven and tested; the claude writer is a straightforward
  parallel. The effort delta between B and C is small (one new file, one new function, one
  parallel gate in `init.go` and `model.go`).
- The template split (separate `claudeTemplate` from `agentsTemplate` for the Phase Models
  block) is a prerequisite either way — so Approach C costs only marginally more than B.

**Scope for the claude writer** (file: `internal/initcmd/claude_mode.go`):
- Function `writeClaudeAgents(projectDir string, models config.ModelConfig) ([]string, error)`
  iterating `config.ResolvePhaseModels(models)`, writing each `.claude/agents/archon-<phase>.md`
  with YAML frontmatter (`model: <FullID>`) and a minimal stub body. Returns written paths for
  rollback.
- Export `WriteClaudeAgents` as the TUI seam.
- Gate in `init.go` on `agentName == "claude"` (parallel to opencode gate at line 101).
- Gate in `tui/model.go:saveConfig` on `cfg.Agent == "claude"` (parallel to opencode gate at
  line 343).

**Template fix** (`internal/initcmd/templates.go`):
- Split `claudeTemplate` to use a per-claude `orchestratorTrailerClaude` constant that replaces
  the "advisory" Phase Models block with a "hard gate via named subagents" description listing
  `archon-<phase>` as the agent to invoke.
- Keep `agentsTemplate` (opencode) unchanged in structure but optionally update its Phase Models
  block to reflect that the real gate is in `opencode.json`, not a call-time param.

---

### Risks

- **Phase Models block divergence**: `claudeTemplate` and `agentsTemplate` currently share
  identical content. Splitting them risks them drifting independently and becoming
  inconsistent in the common sections. Mitigation: keep `orchestratorSections` shared and
  only split the `orchestratorTrailer` portion.

- **Rollback completeness**: Up to 8 new `.claude/agents/archon-<phase>.md` files need to be
  registered in the rollback manifest. Missing any leaves orphan files after `archon undo`.
  Mitigation: collect all written paths from `writeClaudeAgents` and append to
  `rollback.CreatedPaths` before `WriteManifest` (same pattern as opencode at `init.go:105-108`).

- **TUI idempotency**: The TUI save path calls `MergeOpencodeAgent` which is idempotent
  (atomic rename, sorted keys). The claude writer must also be idempotent — re-running
  overwrites files in place without creating duplicates. Mitigation: write unconditionally
  (same as opencode does with `Rename`).

- **Empty phases**: If `ResolvePhaseModels` returns no phases (no models configured),
  the writer should be a no-op rather than writing empty or minimal files. Mitigation:
  same guard as `opencode_mode.go:63-68` — skip when `len(phases) == 0`.

- **Relation to structured-model-resolution Slice 2/3/4**: This change overlaps in scope
  with the `opencode-phase-subagents` change (now SUPERSEDED) and with S2 ("Writers") of
  `structured-model-resolution`. The MEMORY.md note says S1 shipped (PR #47), S2/S3/S4
  pending. This change IS effectively S2 (Writers) generalized to include a claude writer
  alongside the already-shipped opencode writer. Proceeding here should be treated as
  completing S2 of the broader initiative.

---

### Ready for Proposal

**Yes** — the problem is well-bounded, the approach is clear, and the pattern to mirror
(`mergeOpencodeAgent` in `opencode_mode.go`) already exists and is tested. The orchestrator
should present the recommendation (Approach C) to the user, note that this is effectively
S2 of `structured-model-resolution` extended to include the claude harness, and confirm
scope before moving to propose.
