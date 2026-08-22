# Design: local-model-router

<!-- proposal: [proposal](proposal.md) | spec: [spec](specs/local-model-router/spec.md) -->

Implements [[local-model-router]].

## Technical Approach

A deterministic Go pre-router (`internal/route` + `archon route` CLI) resolves
state-dependent transitions that a weak local model gets wrong (state arithmetic,
resume-vs-start, multi-match counting). When no code rule fires it emits `CLASSIFY`;
the leader then invokes the fuzzy `skills/sdd-router/SKILL.md` classifier. The router
is READ-ONLY on state; `harness-workflow` stays the sole state writer and legality
gate. Precedence: `explicit-agent > control > implicit > dual-action(D3) > else CLASSIFY`.

## Architecture Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| A1 | Resolver purity | `route.Resolve(Input) Result` — pure function; all IO (state read, discovery) resolved into `Input` by the CLI layer | Table-testable against 18 fixtures with zero filesystem fixtures; ports the prototype's `PATH=code` rows 1:1 |
| A2 | Verb/keyword single source | One `rules.go` with typed sets (`judgeVerbs`, `verifyVerbs`, `conjunctions`) and a `keywordTable` map; the D3 dual-action rule references `judgeVerbs`/`verifyVerbs` **by the same variables** the keyword table's judge/verify rows are built from | Kills the D3 drift risk — judge/verify verb sets cannot diverge from the keyword table because they are the same symbols |
| A3 | Phase order | Import `config.PhaseOrder`; no local list | Spec "PhaseOrder Canonical Source"; successor arithmetic in one place |
| A4 | Active-change discovery | Reuse `mapgen.readState` yaml pattern; SESSION_STATUS parser with fallback chain | Robust to LLM format drift — never hard-fails |
| A5 | CLASSIFY not ASK for fallthrough | Code emits `CLASSIFY`/`path:model`; only D3 and model emit `ASK` | Keeps code/model responsibilities crisp; leader knows when to call the classifier |

## Package API (`internal/route/route.go`)

```go
type Input struct {
    Message      string // raw; Resolve normalizes (lowercase + strip diacritics)
    Phase        string // current phase, "" if none
    Status       string // in_progress|completed, "" if none
    ActiveChange string // resolved name or "none"
}
type Result struct{ Phase, Rule, Path, ActiveChange string }

func Resolve(in Input) Result           // pure: no IO, no LLM
func Normalize(s string) string         // lowercase + diacritic strip
```

`rules.go` holds `judgeVerbs`, `verifyVerbs`, `conjunctions`, `controlWords`,
`implicitVerbs`, `keywordTable map[string][]string` (built from the shared verb
symbols for judge/verify rows). `discover.go` holds `readState` (yaml, per `mapgen`
pattern) and `ActiveChange(root, flag) string` implementing the D2 fallback chain.

## Data Flow

    archon route "<msg>" [--change --phase --status]
        │  CLI resolves Input (flags override discovery + state read)
        ▼
    route.Resolve(Input)  ── pure, first-match ──▶ Result{Phase,Rule,Path,ActiveChange}
        │  stdout: JSON   │  stderr: "→ Router: archon-<phase> (rule:.., active-change:..)"
        ▼
    leader ── CLASSIFY? ──▶ sdd-router model classifier ──▶ phase|ASK
        else ─────────────▶ harness-workflow gate ──▶ archon-<phase> subagent

## Active-Change Fallback Chain (A4)

`--change` flag → parse `Active change:` line in `SESSION_STATUS.md` (regex, tolerant
of markdown bold/spacing) → sole non-archive dir under `openspec/changes/` → `none`.
`none` means explore is the only legal start; control/implicit verbs route to explore.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/route/route.go` | Create | `Input`/`Result`, `Resolve`, `Normalize` |
| `internal/route/rules.go` | Create | verb sets + keyword table (single source) |
| `internal/route/discover.go` | Create | `readState`, `ActiveChange` fallback chain |
| `internal/route/route_test.go` | Create | table tests: 18 fixtures vs `Resolve` |
| `cmd/archon/main.go` | Modify | add `newRouteCmd`; wire into root |
| `skills/sdd-router/SKILL.md` | Create | model classifier contract (provider-neutral) |
| `internal/initcmd/templates.go` | Modify | routing rule into both orchestratorRules blocks |
| `internal/initcmd/templates_test.go` | Modify | assert "archon route" present, ordered before delegate rule |
| golden `CLAUDE.md`/`AGENTS.md` | Modify | regenerate after template edit |

## CLI Contract

`archon route [--json] [--change n] [--phase p] [--status s] "<msg>"`. stdout = JSON
`{phase,rule,path,active_change}` (all required). stderr = human echo
`→ Router: archon-<phase>  (rule: <id>, active-change: <name|none>)`. Exit 0 for any
resolved output incl. `ASK`/`CLASSIFY`; exit 1 only on usage/internal error (corrupt
state, FS failure). `--phase`/`--status` override state read for deterministic tests.

## skill_count Reconciliation (A6)

`sdd-router/SKILL.md` is auto-embedded via `skills/embed.go`. `embeddedSkillCount()`
and `initcmd.Update` recompute `SkillCount` from the embedded FS (dirs carrying a
`SKILL.md`); the `Skills: {{.SkillCount}}` CLAUDE.md line renders from that value.
So `config.yaml` (currently stale 24) and CLAUDE.md (25) both converge to the real
count via `archon update` / re-init — **generated, not hand-edited**. Verify the
post-add count and update `CLAUDE.md`'s prose "Skills: N" and config `skill_count`
through the recompute path in slice B.

## Chained-PR Slices

| Slice | Scope | ~Lines | Budget |
|-------|-------|--------|--------|
| A | `internal/route/*` (route+rules+discover) + `newRouteCmd` + tests | ~330 | under 400; tests dominate |
| B | `sdd-router/SKILL.md` + templates wiring + golden regen + skill_count | ~180 | under 400 |

A is the risk edge: if `route.go` + `rules.go` + `discover.go` + tests exceed 400,
split tests (fixtures) into a follow-up commit within slice A. Flag confirmed at tasks.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | 18 fixtures → `Resolve` code-path rows (#1–14,17,18) | table-driven, per go-testing skill; no FS |
| Unit | discovery precedence, SESSION_STATUS parse, `none` fallback | table-driven with in-memory `fs.FS` |
| Golden | CLAUDE.md/AGENTS.md routing rule + ordering | templates_test |
| Model-path | #15 D3 is CODE-tested (`ambiguous`); #16 ASK + CLASSIFY signal are model-path, NOT unit-tested here | asserted only that code emits `CLASSIFY` |

## Migration / Rollout

No migration. Additive: new package + CLI subcommand + one skill + template lines.
Routing rule inserted **after** the harness-workflow gate rule, before delegate — no
weakening of preflight/vague-request/human-review gates.

## Open Questions

- [ ] Confirm slice-A line count at tasks; pre-plan a test-split commit if over 400.
