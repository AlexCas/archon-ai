# Proposal: Pin the judge phase model via an archon-judge subagent

## Intent

Judge is the only SDD phase whose model is not a hard gate. The dual review
(`judgment-day`) runs inline under `harness-judge`, so it inherits the
orchestrator's model instead of a pinned one. We empirically validated that the
other 8 phases honor their frontmatter `model:`. Judge should too — its review
must run under a frontmatter-pinned model that mirrors verify (`claude-opus-4-8`).

Approach **(a)** is chosen and NOT re-opened: create an `archon-judge` subagent
whose frontmatter `model:` is the binding gate; `harness-judge` delegates the
dual review to it instead of running `judgment-day` inline.

## Scope

### In Scope
- Emit `.claude/agents/archon-judge.md` at init/re-init, model pinned to opus.
- Wire the `models.phases.judge` config key, `--model-judge` flag, `ValidPhases`.
- `harness-judge`: delegate dual review to `archon-judge` instead of inline.
- `archon-judge.md`: thin wrapper that invokes `judgment-day`.
- CLAUDE.md "Phase Models": add `judge: claude-opus-4-8`.

### Out of Scope
- Phase ORDER semantics — judge stays a separate phase (verify → judge → archive);
  no change to the `harness-workflow` state machine.
- Mutation-testing and Playwright gates — untouched.
- Making judge skill-gated differently — it stays gated by `harness-judge`.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
None at the spec level — this is harness codegen + skill wiring. (spec phase may
record a delta for the model-resolution capability if one exists; no requirement
text changes.)

## Approach

The decisive design point is the **two decoupled phase orders**:
- `config.PhaseOrder` (`internal/config/model.go:167`) drives ONLY codegen
  (`writeClaudeAgents`, the CLAUDE.md "Phase Models" rows, the TUI rows).
- The workflow state machine's sequence is a TEXTUAL list inside
  `skills/harness-workflow/SKILL.md:79` and ALREADY contains `judge`.

So adding `judge` to `config.PhaseOrder`/`ValidPhases` does NOT disturb the
workflow sequence — it only switches on the codegen wiring. **Recommended
resolution: add judge to `config.PhaseOrder` and `ValidPhases`.** This gives the
config key, CLAUDE.md row, `config set/get`, and TUI row for free and keeps a
single resolver path.

One catch forces a small generator change: `renderClaudeAgent`
(`internal/initcmd/claude_mode.go:61`) hardcodes a generic body —
`Follow skills/sdd-<phase>/SKILL.md` + `Do NOT delegate; execute yourself`. There
is no `sdd-judge` skill, and judge MUST delegate. So judge needs a special-cased
body. `archon-judge.md` is a thin wrapper: pin model in frontmatter, then
instruct the executor to run `judgment-day` for dual review and report the verdict
to `harness-judge` (no inline review, no fix logic).

Model source: there is no Go default-seeding table; per-phase models come from
init flags → `models.phases`. Add a `--model-judge` flag defaulting to verify's
resolved value so judge resolves to opus even when unset (or, if the flag is
empty, `ResolvePhaseModels` falls back to `Default`). harness-judge Step 2 changes
from "invoke judgment-day inline" to "delegate to the archon-judge subagent".

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modified | Add `judge` to `PhaseOrder` + `ValidPhases` |
| `internal/initcmd/claude_mode.go` | Modified | Special-case judge body in `renderClaudeAgent` (wrapper, not `sdd-judge`) |
| `cmd/archon/main.go` | Modified | Add `--model-judge` flag → `modelFlags["judge"]` |
| `cmd/archon/config.go` | Modified | Update two "valid phases" error strings |
| `.claude/agents/archon-judge.md` | New (generated) | Frontmatter `model: claude-opus-4-8` + judgment-day wrapper body |
| `.archon/config.yaml` | Modified (generated) | New `models.phases.judge` entry |
| `skills/harness-judge/SKILL.md` | Modified | Step 2 delegates to archon-judge, not inline |
| `CLAUDE.md` | Modified (generated) | "Phase Models" gains `judge` row |
| `internal/tui/models_tab.go` | Auto | Judge model row appears (iterates PhaseOrder) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Generic `renderClaudeAgent` body emits a non-existent `sdd-judge` ref | High if unguarded | Special-case judge body — explicit in scope |
| Re-init overwrites a user-tuned judge model | Low | Same atomic, idempotent write path as the other 8; honors `models.phases.judge` |
| Adding judge to PhaseOrder changes `ResolvePhaseModels` output for existing repos | Low | Resolver already handles missing phases (falls back to Default / omits); covered by existing round-trip tests + a new judge case |
| "Phase Models" intro implies a subagent for every phase | None now | True for judge under approach (a) — it IS a subagent |
| Workflow sequence accidentally altered | None | `harness-workflow` PHASE_ORDER is textual and already includes judge; untouched |

## Rollback Plan

Revert the `config.PhaseOrder`/`ValidPhases` additions, the `renderClaudeAgent`
judge case, the `--model-judge` flag, and the `harness-judge` Step 2 edit. On the
next `archon init`, `archon-judge.md` and the `models.phases.judge` key stop being
emitted; harness-judge returns to inline `judgment-day`. Delete any stale
`.claude/agents/archon-judge.md`. No data migration.

## Dependencies

None external.

## Success Criteria

- [ ] `archon init` (fresh and re-init) emits `.claude/agents/archon-judge.md` with `model: claude-opus-4-8`, byte-identical on re-run.
- [ ] `.archon/config.yaml` gains `models.phases.judge`; `archon config get/set models.phases.judge` works; TUI shows a judge row.
- [ ] `harness-judge` delegates dual review to `archon-judge`; no inline judgment-day.
- [ ] CLAUDE.md "Phase Models" lists `judge: claude-opus-4-8`.
- [ ] Workflow sequence and the mutation/Playwright gates are unchanged.

## Estimated Size

Roughly 90–150 changed lines: ~6 lines Go (PhaseOrder/ValidPhases), ~12 lines
generator special-case, ~3 lines flag + map entry, ~2 error strings, ~15 lines
across the two skills, plus regenerated `archon-judge.md`/CLAUDE.md/config and a
new resolver test case. **Fits a single PR under the 400-line budget; no
chained-PR split needed** (preflight C is ask-always — flag only if spec/design
expand the surface materially).
