# Design: Pin the judge phase model via an archon-judge subagent

## Technical Approach

Approach (a), locked. Add `judge` to `config.PhaseOrder` and `config.ValidPhases`
so the existing claude codegen path emits `.claude/agents/archon-judge.md`, a
config key (`models.phases.judge`), a CLAUDE.md "Phase Models" row, a TUI row, and
`config get/set` support — all from one resolver. The ONE generator change is a
judge special-case body in `renderClaudeAgent` (no `sdd-judge` skill exists; judge
must delegate to `judgment-day`). `harness-judge` Step 2 changes from inline
`judgment-day` to delegating to the `archon-judge` subagent. The workflow state
machine (textual `harness-workflow/SKILL.md`) already lists judge and is NOT
touched. Maps to specs `claude-phase-subagents` and `harness-judge`.

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Where judge goes in `PhaseOrder` | Append AFTER `verify`, BEFORE `archive`: `{...,"tasks","apply","verify","judge","archive"}` | Matches the human phase order (verify → judge → archive). `PhaseOrder` drives ONLY codegen row order (CLAUDE.md, TUI, agent-file iteration), NOT the state machine; this gives the natural reading order in CLAUDE.md. |
| Where judge defaults to verify | In `cmd/archon/main.go` when building `modelFlags`, BEFORE `buildConfig` | `buildConfig` drops empty phase values, so an unset judge would fall to `Default`, not verify. Computing `judge = modelJudgeFlag or modelVerifyFlag` at the flag layer makes "mirror verify" deterministic and keeps `buildConfig`/`ResolvePhaseModels` generic. |
| Judge body source | Special-case literal wrapper in `renderClaudeAgent` | No `sdd-judge` skill; generic body would emit a dangling `skills/sdd-judge/SKILL.md` ref and a "do not delegate" line that contradicts judge's job. |
| New config struct field? | No | Judge lives in the existing `Models.Phases` map. `Clone()` already deep-copies the whole map, so judge round-trips with zero struct change. |

## Data Flow

    --model-judge (or fallback --model-verify) ─┐
    --model-verify ─────────────────────────────┤→ modelFlags["judge"]
                                                 │
    modelFlags → buildConfig → Models.Phases.judge
                                  │
                                  ▼
              ResolvePhaseModels (PhaseOrder incl. judge)
                  │                         │
                  ▼                         ▼
         writeClaudeAgents          phaseModelsClaude template
         renderClaudeAgent(judge)   → CLAUDE.md "Phase Models" judge row
         → archon-judge.md (wrapper, bare model id)

    harness-judge Step 2 ──delegate──▶ archon-judge subagent ──runs──▶ judgment-day

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/config/model.go` | Modify | Add `"judge": true` to `ValidPhases` (line 153 map); append `"judge"` after `"verify"` in `PhaseOrder` (line 167). Update the `PhaseOrder` comment (no longer "excludes judge"). `ResolvePhaseModels` needs NO code change — it iterates `PhaseOrder` and already chains `Phases[p]`→`Default`→omit. |
| `cmd/archon/main.go` | Modify | Add `modelJudgeFlag` var (line ~88) + `--model-judge` flag (line ~199). In the `modelFlags` map (line 122), set `"judge"` to `modelJudgeFlag` if non-empty, else `modelVerifyFlag` (the mirror-verify default). Add it to the `config.Validate` warning loop coverage (it is already iterated since it lives in `modelFlags`). |
| `internal/initcmd/claude_mode.go` | Modify | In `renderClaudeAgent`, branch on `pm.Phase == "judge"`: emit the judge wrapper body instead of the generic `skills/sdd-<phase>/SKILL.md` body. Frontmatter (name/description/model) stays identical in shape; the 8 existing phases stay byte-identical. |
| `internal/initcmd/templates.go` | None (verify only) | `phaseModelsClaude` ranges `.PhaseModels`; judge appears once automatically via `PhaseOrder`. No edit. |
| `cmd/archon/config.go` | Modify | Update the two "valid phases" error strings (lines 181, 209) to include `judge`. The `ValidPhases[phase]` guard itself now passes for judge via the model.go change. |
| `internal/tui/models_tab.go` | None | Row built by iterating `config.PhaseOrder` (line 104); judge row appears automatically. The `rows` capacity hint `1+len(PhaseOrder)+1` self-adjusts. No edit. |
| `internal/config/config.go` | None | `Clone()` deep-copies `Models.Phases` wholesale; judge needs no new clone line. |
| `skills/harness-judge/SKILL.md` | Modify | Step 2: delegate dual review to the `archon-judge` subagent (not inline `judgment-day`). Update the "ALWAYS invoke judgment-day inline" hard rule to "delegate to archon-judge". Preserve Step 0 gate, Steps 3/3b/4/5/6, retry cap, feedback + output contracts. |
| `skills/judgment-day/SKILL.md` | None | Runs unchanged inside the archon-judge subagent. The subagent IS the executor; judgment-day's own delegation of its two blind judges is internal and unaffected. |
| `.claude/agents/archon-judge.md` | New (generated) | Frontmatter `model: claude-opus-4-8` (bare) + judgment-day wrapper body (below). Regenerated at every init. |
| `.archon/config.yaml`, `CLAUDE.md` | Regenerated | `models.phases.judge` key (alphabetical slot, between explore/propose) + CLAUDE.md judge row (PhaseOrder slot, after verify). |

## Interfaces / Contracts

**`renderClaudeAgent` judge branch** — body the generator emits (after the shared
`---`/name/description/`model:`/`---`/blank-line frontmatter):

```text
You are the Archon SDD judge executor. There is no sdd-judge skill: your job is the
dual adversarial review. Run the `judgment-day` skill against the current change
(all files modified by the change), then report its verdict (APPROVED or ESCALATED,
with confirmed/suspect issues) back to `harness-judge`. Do NOT apply fixes or
re-verify yourself — harness-judge owns the re-apply loop and the gates.
```

This deliberately avoids the generic "Follow `skills/sdd-judge/SKILL.md`. Do NOT
delegate" lines (judge MUST delegate to judgment-day).

**Default-to-verify resolution order** (deterministic):
1. `--model-judge <v>` set → `modelFlags["judge"] = v`.
2. else `--model-verify <v>` set → `modelFlags["judge"] = v`.
3. else both empty → judge absent from `Phases`; `ResolvePhaseModels` falls to
   `Default`; omitted only if `Default` is also empty (same as every phase).

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit (config) | `ResolvePhaseModels` includes judge; judge `Phases` entry wins; judge→`Default` fallback; judge omitted when nothing resolves; `len == len(PhaseOrder)` cases updated to 9 | Extend `internal/config/model_test.go` `TestResolvePhaseModels` (the 8-vs-9 count asserts at lines 313, 386 must move to 9). |
| Unit (codegen) | archon-judge.md IS written when judge resolves; body invokes `judgment-day`, reports to `harness-judge`, does NOT contain `skills/sdd-judge/SKILL.md`; frontmatter model is bare; 8 existing bodies still reference their `sdd-<phase>` skill | INVERT `TestWriteClaudeAgents_WritesOneFilePerResolvablePhase` lines 78–82 (judge MUST now exist). Add `TestWriteClaudeAgents_JudgeBodyIsWrapper`. Extend `TestWriteClaudeAgents_BodyPointsAtPhaseSkill` to confirm judge is exempt. |
| Unit (idempotency) | Re-run byte-identical incl. archon-judge.md | `TestWriteClaudeAgents_ReRunByteIdenticalAndPreservesUserFiles` loop already iterates `PhaseOrder`; picks up judge for free once added. |
| Unit (config round-trip) | judge survives Clone + marshal/reload | `config_test.go` `TestConfig_CloneRoundtrip` covers the map generically; add a judge key to its fixture. |
| Integration (CLI) | `config set/get models.phases.judge` no "unknown phase"; CLAUDE.md "Phase Models" lists judge after verify | `cmd/archon` config tests + templates_test golden (regen golden to add judge row in PhaseOrder position). |

## Migration / Rollout

No data migration. Existing `.archon/config.yaml` files WITHOUT a `judge` key:
`ResolvePhaseModels` falls back to `Default` for judge, so on the next `archon init`
or TUI save the judge key/agent/row appear with no user action and no breakage.
Reading an old config never errors (judge is just an absent map key). Rollback:
revert the four code edits + the harness-judge Step 2 edit; next init stops emitting
archon-judge.md and the judge key; delete any stale `.claude/agents/archon-judge.md`.

## Open Questions

- [ ] None blocking. (Confirmed: `Clone()` needs no edit; `models_tab.go` needs no
  edit; `templates.go` needs no edit; `ResolvePhaseModels` needs no logic change.)
