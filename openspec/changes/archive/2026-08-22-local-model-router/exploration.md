# Exploration: local-model-router — deterministic SDD phase dispatch for weak local leaders

## Project Type

**Web testing**: not-web (Go CLI harness — `cmd/archon`, `internal/*`; no browser surface). Playwright stays disabled; no Impeccable recommendation.

## Problem (recap)

The SDD flow depends entirely on the leader MODEL inferring which phase to launch from
natural language. On a weak local leader (`ollama qwen3-orch:latest`, ~5B), implicit
phrasings such as "Trabajemos en esta especificación" fail to launch explore, while the
explicit "…Lanza el agente de exploración" works. The prototype in
`prototype/sdd-router/` already **decided the architecture** empirically (17/18, stable):
a deterministic CODE pre-router handles state-dependent transitions (control words,
start verbs, literal `archon-<phase>` tokens); the MODEL only does fuzzy phase-family
classification; then `harness-workflow` gates legality; then the `archon-<phase>`
subagent runs. This exploration does NOT re-litigate that — it maps the **integration
surface** to formalize it in the harness.

## Current State (how dispatch works today)

- The leader has **no code-assisted routing**. `internal/initcmd/templates.go` bakes the
  orchestrator contract into the generated `CLAUDE.md`/`AGENTS.md`. The only routing
  instruction is `orchestratorRulesClaude` **Rule 2**: "You MUST delegate each phase by
  invoking its `archon-<phase>` subagent…". WHICH phase is left entirely to the model.
- Phase order is canonical in Go: `internal/config/model.go` line 223
  `PhaseOrder = []string{"explore","propose","spec","design","tasks","apply","verify","judge","archive"}`
  and `ValidPhases` (line 209). `ResolvePhaseModels` (line 309) already iterates `PhaseOrder`.
- Phase state lives in `openspec/changes/<name>/state.yaml` (`phase`, `status`). It is
  ALREADY read in Go by `internal/mapgen/scan.go` `readState` (lines 174–189) via the
  `stateYAML{Phase,Status}` struct. This is the reusable state reader.
- The active-change convention is: the change named in `SESSION_STATUS.md` (repo root),
  else the single live folder under `openspec/changes/` that is not `archive/`
  (matches `ROUTER.md` Inputs §2 and the CLAUDE.md SESSION_STATUS contract).
- `cmd/archon/main.go` uses **cobra** with `root.AddCommand(...)` (lines 43–51):
  init, update, rollback, version, status, config, tui, map. Adding a subcommand is a
  one-line `AddCommand` + one `newXxxCmd` constructor (see `newMapCmd`, `newStatusCmd`
  as ~40-line templates).
- Skills are embedded via `skills/embed.go` (`//go:embed */SKILL.md all:_shared`). The
  registry (`skill_count`, `skill_inventory`) is **auto-derived** from the embedded FS on
  `archon init`/`archon update` — `internal/initcmd/update.go` lines 135–136 set
  `SkillCount = len(res.Extracted)` and `SkillInventory = res.Inventory`. There is no
  hardcoded count to bump in Go.

## Integration Surface (the six requested mappings)

### 1. Where the code pre-router should live

**Recommendation: a new `archon route` cobra subcommand** that the leader shells out to.

- Touches: `cmd/archon/main.go` (+1 `AddCommand`, +1 `newRouteCmd` constructor ~40–60
  LOC) and a new `internal/route/route.go` (the deterministic resolver, ~120–180 LOC:
  control-word set, start-verb set, literal `archon-<phase>`/`fase <x>` detection,
  successor lookup, active-change discovery). Reuses `config.PhaseOrder`/`ValidPhases`
  and the `readState` pattern (either export `mapgen.readState` or duplicate the ~10-line
  `stateYAML` struct into `internal/route`).
- Output contract (matches `ROUTER.md` Handoff): one line
  `→ Router: archon-<phase>  (rule: <id>, active-change: <name|none>)` on stdout, or
  `→ Router: ASK …`. The leader reads that line and delegates accordingly. Exit code can
  distinguish resolved vs ASK for scriptability.
- The MODEL classifier stays in the leader/CLAUDE.md prompt (the fuzzy fallback); only
  the deterministic slice becomes code. `archon route` returns a special
  `CLASSIFY` / `ASK` token when no deterministic rule fires, signalling the leader to run
  the in-context classifier for that message.

Options considered:
- **Go helper only (no CLI)**: rejected — the leader is an LLM in another process; it
  cannot call Go functions, it can only shell out. A CLI is the natural boundary.
- **claude hook** (`.claude/settings.json`): a hook could pre-compute routing on every
  user turn, but hooks are provider-specific (breaks opencode/other agents), harder to
  test, and would fire even for non-SDD chatter. A subcommand is provider-neutral and
  unit-testable in Go. Rejected as primary; a hook could later *invoke* `archon route` if
  auto-triggering is wanted, but that is a non-goal for this change.

### 2. How the leader invokes it (without weakening HARD GATES)

- Hook point: `internal/initcmd/templates.go` — add a routing instruction to
  `orchestratorRulesClaude` / `orchestratorRulesOpencode` (both harness variants), and
  register a new `sdd-router` skill the leader loads. The instruction: "Before selecting
  a phase to delegate, run `archon route` with the user's message; use its resolved
  phase, or run the in-context classifier when it returns CLASSIFY, or ASK when it returns
  ASK."
- Gate preservation is explicit and load-bearing: the router decides WHICH phase, NEVER
  whether the gates run (`ROUTER.md` Handoff §4). The generated Rules keep their existing
  numbering; the router is inserted as a **new step that feeds** Rule 1 (harness-workflow
  check) and Rule 2 (delegate). Preflight, Vague-Request, and Human-Review gates run
  exactly as today — the router runs AFTER preflight is satisfied and its ASK output can
  itself trigger the vague-request path. Nothing about the gates changes; the router only
  removes the model's burden of *inferring* the target.

### 3. State source of truth (confirmed)

- Truth for `current_phase`/`status`: `openspec/changes/<name>/state.yaml`, owned by
  `harness-workflow` (SKILL.md "State File Format", `phase`/`status` fields). Go already
  parses exactly these two fields in `mapgen/scan.go readState`.
- Truth for the ACTIVE change name: `SESSION_STATUS.md` at repo root (CLAUDE.md
  "Session Status" + harness-workflow Step 3b), with fallback to the sole non-archive
  folder under `openspec/changes/`. The code pre-router must resolve active-change with
  this precedence, then read that folder's `state.yaml`. `STATE = none` when no live
  change exists → `target = explore` (ROUTER control/implicit rules).
- The router READS state; it never writes it. `harness-workflow` remains the only writer
  (atomic temp+rename), preserving the single-writer invariant.

### 4. The skill artifact

- New `skills/sdd-router/SKILL.md` — the canonical, emb-loaded copy of `ROUTER.md`
  reshaped as an LLM-first skill (frontmatter per `skill-creator`: `name`, `description`
  with trigger words, `license`, `metadata.version`; body 180–700 tokens). It documents
  the deterministic rule precedence AND how to call `archon route` + when to fall back to
  the in-context classifier.
- Embedding is automatic: `skills/embed.go` globs `*/SKILL.md`, so a new directory is
  picked up with no code change. `skill_count`/`skill_inventory` are recomputed from the
  FS on `archon update` (`update.go` 135–136) — **no hardcoded number in Go to edit**.
- Registry consistency chores: run the `skill-registry` skill to re-index, and update the
  human-facing counts that ARE hardcoded in prose. Note a pre-existing drift to fix or
  flag: repo `.archon/config.yaml` says `skill_count: 24`, while `CLAUDE.md` and
  `internal/initcmd/templates.go` `{{.SkillCount}}` render 25 today (find shows 26 dirs
  with `SKILL.md`, which includes `_shared`). Adding `sdd-router` shifts these; the change
  must reconcile the count in `.archon/config.yaml` (via `archon update`), the CLAUDE.md
  "Configuration → Skills: N" line, and the CLAUDE.md persona "Skills: 25" line.

### 5. Successor / PhaseOrder reuse (confirmed)

- The code path MUST use `config.PhaseOrder` (model.go:223) — the single canonical order,
  identical to harness-workflow's `PHASE_ORDER` and `ROUTER.md`'s phase order.
- "Next" resolution (control-word `next` rule) is `PhaseOrder[index(current)+1]` — pure
  function of `PhaseOrder`, reusable directly. `ValidPhases` (model.go:209) validates the
  literal `archon-<phase>` token. No new phase-order source is introduced; the router
  imports `internal/config`.

### 6. Risks / open questions / non-goals

**Risks**
- **Review-budget split (400 lines).** Rough estimate: `internal/route/route.go`
  (~150) + resolver tests (~150–250, Go convention favors table tests) + `main.go`
  wiring (~50) + `sdd-router/SKILL.md` (~80) + `templates.go` rule edits + template
  golden-file updates (`templates_test.go`) + count reconciliation. Code+tests alone
  plausibly reach or exceed 400 changed lines. **Flag for `sdd-tasks`: likely a
  2-slice chained PR** — (A) `internal/route` resolver + `archon route` CLI + tests
  (self-contained, verifiable in isolation), (B) leader wiring (templates.go, SKILL.md,
  registry/count reconciliation, golden files). Per the caller's cached strategy
  (`ask-always`), stop and ask before apply if the forecast confirms > 400.
- **Residual #15 dual-action** ("revisa y prueba esto" → should be ASK). The model
  returns `judge`. Per FINDINGS, accept it (harness-workflow + Human Review Gate still
  gate execution) OR add one narrow code rule (judge-verb AND verify-verb joined by
  "y"/"and" → ASK). Recommend: keep the model+gate fallback, document #15 as a known
  edge, avoid a whack-a-mole keyword list. Decision belongs to propose/spec.
- **Provider-agnosticism.** `.archon/config.yaml agent` may be `claude`, `opencode`,
  etc. `archon route` is a Go binary independent of the agent → provider-neutral by
  construction. But the *invocation instruction* is baked into BOTH `orchestratorRules
  Claude` and `orchestratorRulesOpencode`, and `sdd-router/SKILL.md` must not assume
  Claude-only delegation primitives. Any hook-based auto-trigger would be Claude-only,
  which is why the CLI approach is preferred.
- **Accent/lowercase normalization** must live in code (strip accents, lowercase) so
  `especificación` and `especificacion` match identically — trivial but must be tested.

**Open questions**
- Handoff channel: does `archon route` emit only the echo line (leader parses), or also
  a machine field (JSON) for reliability? Prototype uses the human echo line; a
  `--json` flag may harden parsing for weak leaders.
- Should `archon route` itself discover the active change, or should the leader pass
  `--change <name>`/`--phase`/`--status` explicitly (avoids the router re-implementing
  SESSION_STATUS precedence)? Leaning: router discovers it, single source of truth in code.
- Where exactly the routing step sits relative to the Vague-Request guard (router-first
  vs guard-first) — router ASK output should feed the guard, so guard stays outermost.

**Non-goals**
- Re-deciding the hybrid architecture (settled by the prototype).
- Auto-triggering routing via provider hooks (future, provider-specific).
- Changing harness-workflow's gate logic or the state-write ownership.
- The `keyword-first vs implicit-first` toggle (ROUTER trade-off) — ship implicit-first.

## Affected Areas

- `cmd/archon/main.go` — `+AddCommand(newRouteCmd(...))`, `+newRouteCmd` (~50 LOC).
- `internal/route/route.go` (NEW) — deterministic resolver (~150 LOC) reusing
  `config.PhaseOrder`/`ValidPhases` and the `stateYAML`/`readState` pattern.
- `internal/route/route_test.go` (NEW) — port the 18 `fixtures.md` cases as table tests
  for the deterministic paths; assert PATH=code rows.
- `skills/sdd-router/SKILL.md` (NEW) — LLM-first port of `ROUTER.md`; auto-embedded.
- `internal/initcmd/templates.go` — add the "call `archon route` first" step to
  `orchestratorRulesClaude` + `orchestratorRulesOpencode`; reconcile `Skills: N`.
- `internal/initcmd/templates_test.go` — golden output updates.
- `.archon/config.yaml` + `CLAUDE.md` — reconcile `skill_count` / "Skills: N" (via
  `archon update` + prose edit); run `skill-registry`.
- (READ-ONLY reference, not modified) `internal/mapgen/scan.go readState`,
  `internal/config/model.go PhaseOrder`.

## Approaches

1. **`archon route` CLI subcommand (RECOMMENDED)** — deterministic resolver in
   `internal/route`, leader shells out.
   - Pros: provider-neutral; unit-testable in Go against the 18 fixtures; reuses
     `PhaseOrder`/`readState`; single state-read owner; clean CLI boundary matching the
     LLM/process split.
   - Cons: leader must actually call it (prompt-enforced, not mechanically guaranteed);
     one extra process hop per routed turn.
   - Effort: Medium.

2. **Go helper package, no CLI** — reject: an out-of-process LLM cannot call Go directly.
   - Effort: n/a (not viable as the invocation surface).

3. **claude hook auto-router** — reject as primary: provider-specific, fires on all
   turns, harder to test; viable later as an optional trigger that calls approach 1.
   - Effort: Medium, but wrong layer for a provider-neutral harness now.

## Recommendation

Formalize the prototype as **Approach 1**: a new `archon route` subcommand backed by
`internal/route`, plus a `skills/sdd-router/SKILL.md` and a routing step added to the
generated orchestrator Rules (both harness variants). Reuse `config.PhaseOrder`/
`ValidPhases` and the `mapgen readState` pattern; keep `harness-workflow` as the sole
state writer and legality gate. Expect a likely 2-slice chained PR for the 400-line
budget; leave the #15 dual-action edge to the model+gate fallback (documented), and
reconcile the skill count across `.archon/config.yaml`, `CLAUDE.md`, and templates.

## Risks (summary)

- 400-line budget likely exceeded → plan chained PRs (resolver+CLI, then leader wiring).
- #15 dual-action ambiguity residual → accept via gates or add one narrow code rule.
- Pre-existing skill_count drift (config 24 vs CLAUDE.md 25 vs 26 dirs) must be reconciled.
- Provider-agnosticism: keep the invocation instruction and SKILL.md provider-neutral.

## Ready for Proposal

Yes. The architecture is settled and the integration surface is mapped. Propose should
decide: (a) handoff channel (echo line vs `--json`), (b) active-change discovery in
router vs leader-passed flags, (c) #15 handling, and (d) the chained-PR slicing.
