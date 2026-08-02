# Tasks: local-model-router

<!-- proposal: [proposal](proposal.md) | spec: [spec](specs/local-model-router/spec.md) | design: [design](design.md) -->

Implements [[local-model-router]].

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~510 total (Slice A ~330, Slice B ~180) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR A (tracker ← slice-A) → PR B (slice-A ← slice-B) |
| Delivery strategy | ask-always |
| Chain strategy | feature-branch-chain |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

> Note: Stacked-to-Main is unsupported under archive-before-PR (openspec mode).
> Converged to Feature Branch Chain per the Archive-before-PR Convergence Gate.
> The orchestrator notified the user at tasks strategy selection.

### Slice A line estimate note

`route.go` (~60) + `rules.go` (~70) + `discover.go` (~60) + `route_test.go` (~110)
+ `main.go` delta (~30) = ~330 lines. This is under 400 on its own, but a test-split
commit is pre-planned: if the combined diff at apply time exceeds 400, split the
table-test fixtures into a follow-up commit within slice A before opening the PR.

### Branch / Merge Order (Feature Branch Chain)

```
main
 └── feat/local-model-router        ← tracker branch (owns archive commit)
      └── feat/lmr-slice-a          ← PR A targets tracker
           └── feat/lmr-slice-b     ← PR B targets slice-A branch
```

Merge order: PR A reviewed + merged into tracker → PR B rebased onto tracker → PR B
reviewed + merged into tracker → integrated judge on tracker passes → archive commit
staged on tracker → tracker PR opens to main.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| Slice A | `internal/route/` package + `archon route` CLI + 18-fixture tests | PR A | base = tracker branch; self-contained; no golden regen needed |
| Slice B | `skills/sdd-router/SKILL.md` + templates wiring + golden regen + skill_count | PR B | base = slice-A branch; depends on slice A merged into tracker |

---

## Slice A — Core Router + CLI + Tests (~330 lines)

### Phase 1: Package Scaffold

- [ ] 1.1 Create `internal/route/` directory; add `route.go` declaring `package route`, `Input` struct (Message, Phase, Status, ActiveChange string), `Result` struct (Phase, Rule, Path, ActiveChange string). Satisfies: design §Package API, spec §Deterministic Code Pre-router.
- [ ] 1.2 Add `Normalize(s string) string` to `route.go`: lowercase via `strings.ToLower`, strip diacritics using `golang.org/x/text/unicode/norm` + `unicode.Mn` category filter. Satisfies: spec §Text Normalization, fixture #1 (accented form).
- [ ] 1.3 Create `internal/route/rules.go` with exported-compatible typed sets: `judgeVerbs`, `verifyVerbs`, `conjunctions`, `controlWords`, `implicitVerbs`; and `keywordTable map[string][]string` keyed by phase name, rows built from the shared verb-set variables for judge/verify entries. Satisfies: design §A2 single-source, spec §Keyword-to-Phase Table Coverage (all 9 phases, ≥2 ES + 2 EN each).
- [ ] 1.4 Create `internal/route/discover.go`: unexported `stateYAML` struct + `readState(fsys fs.FS, p string) (phase, status string)` mirroring the `mapgen.readState` YAML pattern. Satisfies: design §A4, spec §Active-Change Discovery Precedence.
- [ ] 1.5 Add `ActiveChange(root string, flagOverride string) string` to `discover.go` implementing the 4-step fallback chain: flag override → `SESSION_STATUS.md` `Active change:` regex parse → sole non-archive dir under `openspec/changes/` → `"none"`. Satisfies: spec §Active-Change Discovery Precedence, fixtures for `--change` flag and SESSION_STATUS.md scenarios.

### Phase 2: Resolver Logic

- [ ] 2.1 Implement `Resolve(in Input) Result` in `route.go` applying four rules in strict top-to-bottom first-match order: `explicit-agent` → `control` → `implicit` → `keyword`; implicit is above keyword. Satisfies: spec §Deterministic Code Pre-router, spec §Implicit-above-keyword Precedence.
- [ ] 2.2 Implement `explicit-agent` rule in `Resolve`: scan normalized message for `archon-<phase>` token OR phase name in navigation position (e.g. after "al", "el", "lanza el agente de"); emit `Result{Phase: matched, Rule: "explicit-agent", Path: "code"}`. Covers fixtures #5, #17, #18.
- [ ] 2.3 Implement `control` rule: if message contains a control word (`siguiente|continuemos|adelante|sigamos|continúa`), check `Input.Status`/`Input.Phase`; `in_progress` → `resume` rule; `completed` → `next` rule using `config.PhaseOrder` for successor; `ActiveChange == "none"` → `next-nochange` rule targeting `explore`. Satisfies: spec §PhaseOrder Canonical Source, fixtures #6–#9.
- [ ] 2.4 Implement `implicit` rule: if message contains an implicit-start verb (`trabajemos|empecemos|comencemos|hagamos|armemos|pongamonos|arranquemos`) and `ActiveChange != "none"` → `implicit-resume`; else → `implicit-start` targeting `explore`. Satisfies: spec §Implicit-above-keyword Precedence, fixtures #1–#4.
- [ ] 2.5 Implement D3 `ambiguous` rule (checked BEFORE keyword fallback but AFTER explicit/control/implicit): if normalized message contains a judge-verb AND a verify-verb with a conjunction present → `Result{Phase:"ASK", Rule:"ambiguous", Path:"code"}`. Satisfies: spec §Narrow Dual-Action Ambiguity Rule, fixture #15.
- [ ] 2.6 Implement `keyword` rule: scan `keywordTable`; exactly one phase matches → emit that phase with `Rule: "keyword"`; zero or multi-match → fall through to CLASSIFY. Satisfies: spec §Keyword-to-Phase Table Coverage, fixtures #10–#14, keyword-outline scenarios.
- [ ] 2.7 Implement CLASSIFY fallthrough: `Result{Phase:"CLASSIFY", Rule:"classify", Path:"model", ActiveChange: in.ActiveChange}`. Satisfies: spec §Model Classifier Fallback, fixture #16.

### Phase 3: CLI Subcommand

- [ ] 3.1 Add `newRouteCmd(stdout, stderr io.Writer) *cobra.Command` in `cmd/archon/main.go`. Flags: `--json` (no-op, JSON is always emitted), `--change string`, `--phase string`, `--status string`; positional arg: message string. Satisfies: spec §Machine-Readable JSON Output Contract, design §CLI Contract.
- [ ] 3.2 In `newRouteCmd` body: call `ActiveChange(root, flagChange)`, build `Input` (applying `--phase`/`--status` flag overrides), call `Resolve`, marshal `Result` to JSON, write to stdout; write human echo to stderr: `→ Router: archon-<phase>  (rule: <id>, active-change: <name|none>)`. Exit 0 for all resolved outputs including `ASK`/`CLASSIFY`; exit 1 only on marshal or FS error. Satisfies: spec exit-code scenarios, design §Data Flow.
- [ ] 3.3 Wire `newRouteCmd` into `newRootCmd` via `root.AddCommand(newRouteCmd(stdout, stderr))`. Satisfies: design §File Changes (cmd/archon/main.go modify).

### Phase 4: Table Tests

- [ ] 4.1 Create `internal/route/route_test.go` with a table-driven `TestResolve` covering all 18 fixtures via `route.Resolve` directly (no FS, no subprocess). Each row: name, Input fields (Message, Phase, Status, ActiveChange), want Result (Phase, Rule, Path). Use `t.Run(tt.name, ...)`. Satisfies: design §Testing Strategy (unit, no FS), spec fixtures #1–#15, #17, #18; go-testing skill §table-driven pattern.
  - Fixture #16 asserting `CLASSIFY` output is included as a code-path row.
  - Fixture #15 (`ambiguous`) confirms D3 code-path; model-side behavior (#16b) is excluded from unit tests (model-path).
- [ ] 4.2 Add `TestNormalize` table test in `route_test.go`: accented → stripped form, uppercase → lowercase, combined. Satisfies: spec §Text Normalization, feature file normalization background.
- [ ] 4.3 Add `TestActiveChange` table test in a separate `discover_test.go` using `testing/fstest.MapFS` for in-memory fixtures: flag-override wins, SESSION_STATUS.md parsed, sole-folder fallback, `none` fallback. Satisfies: spec §Active-Change Discovery Precedence, feature file discovery scenarios.
- [ ] 4.4 Run `go test ./internal/route/...` and `go build ./cmd/archon/...`; confirm all tests pass and binary builds clean. Pre-plan test-split commit: if total slice-A diff exceeds 400 lines, isolate fixture rows 9–18 into a separate commit before opening PR A.

---

## Slice B — SKILL.md + Templates Wiring + Reconciliation (~180 lines)

> Depends on slice A merged into the tracker branch. Base branch for PR B = `feat/lmr-slice-a`.

### Phase 5: Model Classifier Skill

- [ ] 5.1 Create `skills/sdd-router/SKILL.md`: frontmatter (`name: sdd-router`, description, `user-invocable: false`), invocation contract (leader calls ONLY when `archon route` emits `CLASSIFY`), full 9-phase keyword table (Spanish + English, ≥2 per language per phase), output contract (one echo line `→ Router: archon-<phase>` or `ASK`; MUST NOT start executing a phase in the same turn), provider-neutral language (no Claude-specific primitives). Satisfies: spec §sdd-router Skill — Model Classifier Contract, feature file scenario "Model classifier emits one echo line".
- [ ] 5.2 Verify `skills/embed.go` auto-embeds `sdd-router/SKILL.md` at build time (no manual registration needed). Satisfies: design §skill_count Reconciliation (A6).

### Phase 6: Leader Wiring

- [ ] 6.1 Modify `orchestratorRulesClaude` in `internal/initcmd/templates.go`: insert a new rule before the existing rule 2 (delegate): "Before delegating a phase, run `archon route '<message>'` and use its resolved phase; invoke the model classifier (`skills/sdd-router`) when output is `CLASSIFY`; surface ASK to the user." Renumber subsequent rules. The inserted rule MUST appear after rule 1 (harness-workflow gate) and before the delegate rule. Satisfies: spec §Leader Wiring (orchestratorRules), feature file scenario "archon route instruction", design migration note (gates remain intact).
- [ ] 6.2 Apply the identical routing rule insertion to `orchestratorRulesOpencode` in `templates.go`. Same ordering constraint. Satisfies: spec §Leader Wiring, design §File Changes (both variants).
- [ ] 6.3 Verify that preflight, vague-request, and human-review gate rules are not moved, weakened, or reordered by the insertion. Satisfies: spec §Leader Wiring ("existing rules… remain intact and downstream from it").

### Phase 7: Golden Updates + skill_count

- [ ] 7.1 Modify `internal/initcmd/templates_test.go`: add assertion that rendered `CLAUDE.md` contains the string `"archon route"` AND that the routing rule line appears before the harness-workflow delegation line. Same check for `AGENTS.md`. Satisfies: feature file scenario "archon route instruction", spec §Leader Wiring golden scenario.
- [ ] 7.2 Run `go test ./internal/initcmd/...` to capture any rendering failures after template edits. Fix failures; confirm no existing golden assertions regressed. Satisfies: design §Testing Strategy (golden layer).
- [ ] 7.3 Run `archon update` (or `go run ./cmd/archon update`) in the repo root to trigger the `embeddedSkillCount()` recompute path; confirm the new count includes `sdd-router`. Then update `config.yaml` `skill_count` and the `Skills: N` prose line in `CLAUDE.md` via the recompute value — do NOT hand-edit the count. Satisfies: design §skill_count Reconciliation (A6), proposal risk §skill_count drift.

### Phase 8: Verification

- [ ] 8.1 Run `go test ./internal/route/... ./cmd/archon/... ./internal/initcmd/...` and confirm all pass. Satisfies: design §Testing Strategy (all layers).
- [ ] 8.2 Run `go build ./cmd/archon/...` and invoke `archon route --phase none --status "" "Trabajemos en esta especificacion"` locally; confirm JSON output `{"phase":"explore","rule":"implicit-start","path":"code","active_change":"none"}`. Validates fixtures #1/#3/#4 end-to-end.
- [ ] 8.3 Invoke `archon route --phase verify --status in_progress "Revisa y prueba esto"` and confirm `{"phase":"ASK","rule":"ambiguous",...}` with exit 0. Validates fixture #15 (D3) end-to-end.
- [ ] 8.4 Invoke `archon route --phase spec --status in_progress "Que opinas del clima?"` and confirm `{"phase":"CLASSIFY","rule":"classify","path":"model",...}` with exit 0. Validates fixture #16 fallthrough.
- [ ] 8.5 Confirm `skill_count` is consistent across `config.yaml`, generated `CLAUDE.md`, and the embedded FS count from `embeddedSkillCount()`. Satisfies: proposal §Success Criteria (skill_count consistent).
