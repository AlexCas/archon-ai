# Verify + Judge Report — model-effort-variants (Slice 4, option b)

- Repo: /home/alexcasdev/Projects/archon-ai
- Branch: feat/structured-models-effort (== master + uncommitted change)
- Method: read the real code, ran the gates, wrote disposable adversarial probes (removed after running). No production code modified.

## VERDICT: SHIP

All four gates pass. All 9 spec scenarios trace to code and at least one test. The three highest-risk
behaviors (effortless byte-identity, stale-variant clearing, non-reasoning re-pick clearing stale
effort) were independently confirmed with disposable probes, not just by trusting the shipped tests.

---

## PART 1 — VERIFY

### Gates (exact output)

| Gate | Result |
|------|--------|
| `go build ./...` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `gofmt -l` on the 6 changed files | no output (all formatted) |
| `go clean -testcache && go test ./... -count=1` | all packages `ok` |

Test run:
```
ok  github.com/archon-ai/archon/cmd/archon            0.045s
ok  github.com/archon-ai/archon/internal/agent        0.008s
ok  github.com/archon-ai/archon/internal/config       0.011s
ok  github.com/archon-ai/archon/internal/initcmd      0.025s
ok  github.com/archon-ai/archon/internal/models       0.009s
ok  github.com/archon-ai/archon/internal/opencode     0.009s
ok  github.com/archon-ai/archon/internal/scaffold     0.014s
ok  github.com/archon-ai/archon/internal/status       0.009s
ok  github.com/archon-ai/archon/internal/tui          1.185s
ok  github.com/archon-ai/archon/internal/version      0.010s
ok  github.com/archon-ai/archon/skills                0.011s
```

### Scenario traceability (4 reqs / 9 scenarios)

| Scenario | Code | Test |
|----------|------|------|
| Reasoning model offers effort selection | models_tab.go:222-230 (Reasoning→effortSelect), updateEffortSelect Enter sets ref.Effort | TestModelsTab_ReasoningModelEntersEffortSelect, TestModelsTab_EffortSelectHighSetsEffort |
| Non-reasoning model skips the effort step | models_tab.go:227 `else { m.mode = rowNav }` | TestModelsTab_NonReasoningModelSkipsEffortSelect |
| Default effort option maps to empty | models_tab.go:250-251 `if opt=="default" { opt="" }` | TestModelsTab_EffortSelectDefaultMapsToEmpty |
| ResolvePhaseModels carries effort | model.go:261 `Effort: ref.Effort` (same resolved `ref` as Model) | TestResolvePhaseModels/"phase ref effort carried…", "default fallback ref effort…", "empty effort stays empty" |
| variant present when effort set | opencode_mode.go:90 (leader), :99 (phase) | TestMergeOpencodeAgent_LeaderVariantPresent, _PhaseVariantPresent |
| variant omitted when effort empty | `Variant string json:"variant,omitempty"` opencode_mode.go:22,37 | TestMergeOpencodeAgent_LeaderVariantAbsentWhenEmpty, _PhaseVariantAbsentWhenEmpty |
| Re-run is byte-identical | full-key overwrite + sorted MarshalIndent + atomic rename | TestMergeOpencodeAgent_IdempotentWithMixedEffort (+ probe TestAdv_IdempotentEffortSet) |
| Effort persists in config round-trip | model.go MarshalYAML mapping (:62-68) / UnmarshalYAML mapping (:48-54) | TestModelRef_MarshalYAML/"effort-bearing ref marshals as mapping" (+ probe round-trip below) |
| (Effort-less idempotency, implicit) | omitempty + field order unchanged | existing effortless tests unchanged + probe TestAdv_EffortlessNoVariantOrder |

**Gap (LOW, not blocking):** The "Effort persists in config" scenario is covered at the
Marshal and Unmarshal sides separately (MarshalYAML→mapping with effort key tested;
UnmarshalYAML mapping decode tested) but there is **no single end-to-end Marshal→Unmarshal
round-trip test asserting `Effort` survives**, and the UnmarshalYAML mapping test case does not
include an `effort:` key nor assert the decoded `Effort`. The contract states this is foundation
work that must "not regress" — it does not regress. I confirmed the true round-trip empirically
(probe below): `{provider,model,effort}` marshals as a mapping and reloads with `Effort` intact;
effortless refs stay scalar. Recommend a one-line round-trip test as cleanup, but not a ship blocker.

---

## PART 2 — JUDGE (adversarial)

### Field ORDER / effortless byte-identity — PASS
`Variant` is declared AFTER `Model` in both structs (opencode_mode.go:22, :37) with `omitempty`.
Mode/Hidden/Model/Description/Prompt order is unchanged. Probe `TestAdv_EffortlessNoVariantOrder`
confirmed effortless output contains NO `variant` key and preserves the exact pre-slice key order
(leader: mode→description→model→prompt; phase: mode→hidden→model→description→prompt). Existing
effortless fixed-field tests pass unchanged (initcmd package `ok`).

### Idempotency + stale-variant clearing (KEY RISK) — PASS
Traced `mergeOpencodeAgent`: it reads the existing doc, then does a FULL overwrite of each entry
via `agents[key] = struct{...}` (opencode_mode.go:86, :95) — not a field-level merge. So a re-serialize
without `Variant` (omitempty) leaves no stale `variant`. Confirmed with probes:
- `TestAdv_ClearEffortRemovesVariant`: write with `Effort:"high"`, re-run same model with empty
  Effort → `variant` key gone.
- `TestAdv_ForeignVariantKeyOverwritten`: seeded an opencode.json where `archon-spec` already had
  `"variant":"STALE"` → after merge the STALE key is gone AND the unrelated `user-thing` agent and
  other top-level keys survive.
- `TestAdv_IdempotentEffortSet`: re-run with effort SET is byte-identical.

### State machine — PASS
- `effortSelect` cursor bounds: clamped `[0, len(effortOptions)-1]` (models_tab.go:241-248). Probe
  `TestAdv_EffortCursorBounds` confirmed Up clamps at 0 and repeated Down clamps at 3.
- Entering effortSelect: model already set in `updateModelSelect` (ref built BEFORE branching,
  :223) so effortSelect only ADDS `ref.Effort`.
- Esc→modelSelect preserves the set model (TestModelsTab_EscFromEffortSelectGoesBackToModelSelect).
- Re-entering modelSelect after Esc and re-picking rebuilds `row.ref` fresh via `refFromCacheKey`
  (empty Effort) → no stale carry; a reasoning re-pick re-enters effortSelect with `effortCursor=0`.
- Non-reasoning pick CLEARS a previously-set effort: `refFromCacheKey` returns a fresh ref with
  empty Effort and the non-reasoning branch skips effortSelect. Confirmed by probe
  `TestAdv_NonReasoningRepickClearsEffort`. **No path leaves stale effort on a row.**

### ResolvePhaseModels ref trace — PASS
`ref` is the single variable: `ref = mc.Phases[p]`, reassigned to `mc.Default` on fallback
(model.go:254-257). Both `ref.FullID()` and `ref.Effort` read the SAME `ref` on the same line
(:261). No model-from-phase / effort-from-default mismatch is possible.

### Leader effort-only edge — PASS
Leader is written only when `leaderFull != ""` (opencode_mode.go:85). An effort-only leader
(Model empty, Effort set) → `FullID()==""` → leader not written. No weird partial entry.

### TUI safety — PASS
No stderr writes in the new code. `effortOptions` is a fixed package-level slice
`{"default","low","medium","high"}` → deterministic render order. Non-key msgs in effortSelect are
ignored (dispatch at models_tab.go:124-133 only forwards non-key msgs to the textinput while in
freeForm; effortSelect needs none). Dispatch switch covers `effortSelect` (:144-145).

### Scope — PASS
- `MarshalYAML`/`UnmarshalYAML` untouched by this diff (foundation).
- No plugin / variants-cache / embedded-asset / `Variants` field introduced.
- Save path changed only via the two struct fields + two populate lines.
- Free-form path unchanged: `ParseModelRef` never sets Effort (model.go:73-78) → free-form rows
  keep empty Effort.

### Go correctness — PASS
No unused vars; `go vet` clean. Cursor indexing is guarded (modelSelect already guards empty list
by falling back to freeForm at :199-203; effortSelect operates on a fixed non-empty slice).

---

## LOW notes (non-blocking)
1. Add an end-to-end config round-trip test asserting `Effort` survives Marshal→Unmarshal, and add
   an `effort:` field + `Effort` assertion to the existing UnmarshalYAML mapping test case
   (model_test.go:175-180). Behavior verified correct empirically; this is test-coverage hygiene.
2. (Informational) `effortSelect` deliberately offers no per-model validation of which efforts a
   given reasoning model actually supports — by design (option b derives availability only from the
   `Reasoning` flag). Matches the spec.

## Probes used (all disposable, removed after running)
- config: effort Marshal→Unmarshal round-trip; full ModelConfig round-trip (leader/phase mapping,
  default scalar) — both PASS.
- initcmd: clear-effort-removes-variant; idempotent-with-effort-set; foreign-STALE-variant-overwritten;
  effortless-no-variant-and-order — all PASS.
- tui: non-reasoning-repick-clears-effort; effort-cursor-bounds — both PASS.
