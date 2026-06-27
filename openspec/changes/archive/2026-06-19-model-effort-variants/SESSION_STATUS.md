# SESSION STATUS

## Active Change
`opencode-phase-writers` (Slice 2 of structured-model-resolution) — extend the opencode
writer to emit per-phase `archon-<phase>` subagents (+ leader) using ModelRef.FullID().
Slice 1 (Foundation) SHIPPED in PR #47, merged to master @ 7985c67.

## Current Phase
`judge` — PASS (SHIP, Slice 4 b). Committed + PR #52 opened DIRECT TO MASTER (no stacking).
INITIATIVE COMPLETE (Slices 1-4): #47/#48/#51 merged; #52 open.

## Slice 4 shipped (PR #52)
- Commit 8c7b82e (author AlexCas, no attribution) on feat/structured-models-effort (base master).
- PR #52: https://github.com/AlexCas/archon-ai/pull/52 — reasoning effort → opencode `variant`.
- Option (b): effort from Model.Reasoning flag; no plugin/cache/embed. ~411 LOC (prod ~72).
- Verify+judge (opus): SHIP, 9/9 scenarios, 0 defects; judge probed stale-variant-on-clear → safe.
- Added effort round-trip test as the judge's LOW-note cleanup.

## Initiative status: structured-model-resolution — ALL 4 slices shipped/in-flight
| Slice | PR | state |
|---|---|---|
| 1 Foundation | #47 | merged |
| 2 Writers | #48 | merged |
| 3a Catalog + 3b Picker | #49/#50 (recovered via #51) | merged |
| 4 Effort/variants | #52 | open (direct to master) |

## Follow-ups (out of scope, noted)
- Free-form `q` key blocked by global quit (qwen models untypeable in free-form) — needs parent key routing.
- internal/models/resolve.go (Resolve/cacheModelNames) now unused by TUI — verify CLI/status before removing.
- Stale local branches from prior sessions: fix/opencode-model-assignment, feat/opencode-phase-subagents (not deleted; ask before).
- Archive: openspec changes (structured-model-resolution, opencode-phase-writers, structured-models-tui-picker,
  tui-model-picker, model-effort-variants) can be moved to openspec/changes/archive/ once #52 merges.
Branch: `feat/structured-models-effort` (off master @ 35bd14b). Change: model-effort-variants.
Option (b): effort offered only for Reasoning models (no plugin/cache/embed); write `variant` (omitempty).
Artifacts: openspec/changes/model-effort-variants/{exploration,proposal,design,tasks}.md + specs/model-effort-variants/spec.md.

## --- Slices 1-3 (DONE, in master) record below ---
DONE — Slices 1-3 ALL in master (222ec2a, 7050780, bec859a, a34a2d5 all ancestors of master @ 35bd14b).
PRs #47/#48/#51 merged to master; #49/#50 merged into stacked bases (content recovered via #51).
Stale merged remote branches to clean: feat/structured-models-{foundation,writers,tui-picker,picker-ui}.
Optional: Slice 4 (effort/variants); follow-ups: free-form 'q' key, unused internal/models/resolve.go.

## --- prior recovery note below ---
RECOVERY — stacked-PR mishap fixed. #49(3a)/#50(3b) merged into their stacked BASES, not master,
so master had only S1+S2. Opened PR #51 (feat/structured-models-picker-ui → master) = S3a+S3b.
PR #51: https://github.com/AlexCas/archon-ai/pull/51 — MERGEABLE (state BLOCKED = branch protection/
review required, NOT a conflict). Merge #51 → master gets the full initiative (Slices 1-3).
LESSON: when merging a stack fast, retarget each child PR's base to master before merging, or merge
top-down with retargeting between.

## Slice 3b shipped (PR #50)
- Commit a34a2d5 (author AlexCas, no attribution) on feat/structured-models-picker-ui.
- PR #50: https://github.com/AlexCas/archon-ai/pull/50 — BASE = feat/structured-models-tui-picker (3a). Retarget down the chain as it merges.
- Full models_tab.go rewrite → provider→model picker; model.go cache wiring; test rewrites.
- Verify+judge (opus): SHIP, all scenarios traced, 0 defects. ~880 LOC (prod ~270, rest test) — flagged at PR.
- Cleanups before commit: removed dead sortedProviderIDs + sort import; renamed misleading test.

## Follow-ups (out of scope, noted)
- Parent global `q`→quit intercepts 'q' in free-form → model names with 'q' (qwen2.5) untypeable.
  Pre-existing (old grid had it); needs parent key-routing change. Worth a follow-up PR.
- internal/models/resolve.go (Resolve/cacheModelNames) now UNUSED by the TUI — possible cleanup
  (kept intact this slice; may still be used by CLI/status — verify before removing).
- Slice 4 (OPTIONAL): effort/variants picker step + opencode variants TS plugin (needs Model.Variants
  field added). Only if effort/reasoning control is wanted.

## --- Slice 3b chain status below ---
`feat/structured-models-picker-ui` STACKED on 3a (@ bec859a, PR #49). Change folder: openspec/changes/tui-model-picker.
Artifacts: proposal/design/tasks.md + specs/tui-model-picker/spec.md (8 reqs, 14 scenarios).
Size forecast ~440 LOC (prod ~230 + test ~210) — at/over D1, flag at PR (C1).

## Slice 3b settled decisions (user 2026-06-19)
- Hosting: IN-TAB per-row SUB-MODE (Enter expands the focused row into provider list → model list,
  inline). Reuse agent_tab's hand-rolled cursor pattern (NOT bubbles/list).
- Leader row: ALSO uses the picker (with free-form fallback), for consistency.
- Free-form: ALWAYS available (a key toggles free-form entry on any row) — escape hatch for models
  not in the cache.
- Legacy bare-alias preservation: REQUIREMENT — an untouched legacy ModelRef{Provider:""} (e.g. "opus")
  is kept verbatim; never force it through the picker / never corrupt it.
- Corrupt-cache warning: inline in the tab view (TUI-safe), keys off LoadModelsOrEmpty err; absent→silent.
- Picker consumes 3a helpers (DetectAvailableProviders, FilterModelsForSDD). Effort/variants OUT.

## --- Slice 3a (shipped, PR #49) record below ---

## Slice 3a shipped (PR #49)
- Commit bec859a (author AlexCas, no attribution) on feat/structured-models-tui-picker.
- PR #49: https://github.com/AlexCas/archon-ai/pull/49 — BASE = feat/structured-models-writers (Slice 2).
  RETARGET to master after #48 merges (GitHub usually auto-retargets stacked PRs on base merge).
- Added FilterModelsForSDD, DetectAvailableProviders (exported), hasToolCallModel (priv) to internal/opencode.
- Verify+judge (opus): SHIP, all scenarios traced, 0 defects, 2 LOW notes.
- ~116 LOC, under D1.

## Remaining work
- Slice 3b (own cycle) — TUI two-step picker + inline corrupt-cache warning + legacy bare-alias
  preservation + free-form fallback + models_tab rewrite. Consumes 3a helpers. ~300-400 LOC, design
  decisions to settle: picker hosting (in-tab sub-mode), legacy preservation contract, free-form
  trigger, leader-row picker-vs-free-form, warning copy/placement. Effort/variants OUT (no Variants in Model).
- Slice 4 (optional) — effort/variants + opencode variants TS plugin (needs Model.Variants).
Merge order: #48 (Slice 2) → #49 (Slice 3a) → future 3b → optional 4. Archive openspec change after full ship.
Branch: `feat/structured-models-tui-picker` STACKED on Slice 2 (@ 7050780, PR #48 open). Retarget to master after #48 merges.

## Slice 3 split (user 2026-06-19): 3a NOW, 3b later
- 3a = data layer: port `hasToolCallModel`, `FilterModelsForSDD`, `DetectAvailableProviders`
  (SIMPLIFIED: tool_call + always-opencode; NO Env/auth port, no Provider struct change) into
  `internal/opencode` + tests. Corrupt-vs-absent seam already exists in LoadModelsOrEmpty (tested).
  Capability: `opencode-provider-catalog`. Additive, no UI change. ~60-100 LOC.
- 3b (later, own cycle) = TUI two-step picker + inline corrupt-cache warning + legacy bare-alias
  preservation + free-form fallback + models_tab test rewrite.

## Slice 3 scope additions (from user, 2026-06-19)
- Voseo fix: NO-OP — templates.go:118 already "¿Quieres ajustar...?" (note was stale, dropped).
- Cache LOW note (A-L2): FOLD INTO Slice 3 — surface a TUI-safe in-UI warning when the opencode
  cache is present but unreadable (NOT a stderr write; Resolve() is TUI-only, stderr corrupts Bubbletea).

## --- Slice 2 (shipped, PR #48) record below ---

## Slice 2 shipped (PR #48)
- Commit 7050780 (author AlexCas, no tool attribution) on feat/structured-models-writers.
- PR #48: https://github.com/AlexCas/archon-ai/pull/48 (base master). ~213 LOC, under D1.
- opencode_mode.go now writes archon-<phase> subagents via ResolvePhaseModels FullIDs + leader.
- Verify+judge combined (opus): SHIP, 10/10 scenarios traced, 0 defects, 2 LOW notes.
- Excluded: CLAUDE.md (transient), SESSION_STATUS.md, opencode-phase-subagents/ (superseded), pre-existing untracked openspec dirs.

## Remaining chained PRs (future sessions)
- Slice 3 — TUI picker: rewrite tui/models_tab provider→model two-step (~350-400 LOC, maybe 3a/3b).
- Slice 4 (optional) — effort/variants (ModelRef.Effort) + opencode `variant` field + variants TS plugin.
Archive the structured-model-resolution openspec change only after the full initiative ships.
Branch: `feat/structured-models-writers` (off master @ 7985c67).
Artifacts: openspec/changes/opencode-phase-writers/{exploration,proposal,design,tasks}.md +
specs/opencode-phase-writers/spec.md. Capability: opencode-phase-writers.

## Slice 2 settled decisions (carried from superseded opencode-phase-subagents, adapted to ModelRef)
- Keys `archon-<phase>` for the 8 config.PhaseOrder phases (judge excluded).
- Per-phase entry: mode:"subagent", hidden:true, model, description, prompt.
  description = "Archon SDD <phase> phase"; prompt = "{file:./AGENTS.md}".
- Model = resolved ModelRef chain Phases[phase] → Default → Leader, then .FullID() (NOT verbatim string).
- All 8 phases always emitted with fallback model when merge runs.
- No-op when leader+default+all phases empty.
- EXTEND mergeOpencodeAgent (likely change signature to take config.ModelConfig).
- Templates "Phase Models" advisory already emits FullID (foundation); confirm in design.

## --- prior slice 1 record below (shipped) ---

## Pivot history (this session)
1. Closed PR #46 (old change `opencode-model-assignment`, superseded by master's overhaul PR #45).
2. Started fresh change `opencode-phase-subagents` (VERBATIM approach). Ran explore→propose→spec→design.
3. User questioned vs gentle-ai. Re-explored gentle-ai: its robustness = provider captured
   at selection time + structured opencode cache, NOT an alias resolver. archon uses flat strings.
4. User chose Option C (full gentle-ai port) → new broader SDD cycle + chained PRs.
   `opencode-phase-subagents` is now SUPERSEDED (see its SUPERSEDED.md); its per-phase
   writer becomes Slice 2 here.

## Current Phase
`judge` — PASS (SHIP after retry 1). Foundation PR committed + opened.
Branch: `feat/structured-models-foundation` (off master @ 7360b4d).

## Foundation PR shipped (Slice 1)
- Commit 222ec2a (author AlexCas, NO tool attribution), conventional.
- PR #47: https://github.com/AlexCas/archon-ai/pull/47 (base master). Size flagged in body (~948 LOC, mostly mechanical test fixtures; prod ~237).
- Included: code + `openspec/changes/structured-model-resolution/` artifacts.
- Excluded: CLAUDE.md (pre-existing transient), SESSION_STATUS.md (transient),
  opencode-phase-subagents/ (superseded), other pre-existing untracked openspec dirs.

## Next chained PRs (future sessions)
- Slice 2 — Writers: opencode_mode.go per-phase subagents + leader via FullID(); templates emit FullID. (absorbs superseded opencode-phase-subagents design) ~250-350 LOC.
- Slice 3 — TUI picker: rewrite models_tab provider→model two-step (maybe 3a/3b). ~350-400 LOC.
- Slice 4 (optional) — effort/variants + opencode variants TS plugin. ~200 LOC.
Each chains off the prior. Archive the openspec change only after the full initiative ships.

## Follow-up LOW notes (from judge, non-blocking)
- Corrupt opencode cache silently degrades to shell-out with no stderr warning (consider a note).
- Effort-only ModelRef is invisible (FullID==""); not consumed until Slice 4.

## Judge round 1 (DO-NOT-SHIP) + fixes applied (retry 1)
- Issue 1 (A:HIGH/B:MED) slashed scalar churned scalar→mapping. FIXED: MarshalYAML emits scalar
  FullID() when Effort=="" (refines spec M4: scalar for ANY no-effort ref; mapping only w/ effort).
  No churn risk (no mapping-form configs exist yet). New test TestConfig_SlashedScalarRoundtripByteIdentical
  + updated TestModelRef_MarshalYAML.
- Issue 2 (B:HIGH) cacheModelNames returned all 145 providers/5278 models. FIXED: scoped to the
  "opencode" provider (const opencodeProviderID), matches shell-out scope. New test
  TestCacheModelNames_NoOpencodeProvider + updated TestCacheModelNames_Mapping.
- Issue 3 (B:MED) NormalizeModel dropped → Claude template shows verbatim id. USER DECISION: accept.
- All 11 pkgs green (cache cleared), gofmt clean, vet clean. Report: judge-report.md.

## Apply Result (Foundation S0+S1)
- All 30 tasks done. NEW pkg `internal/opencode/` (models.go + tests + 3 testdata fixtures).
- MODIFIED: config/model.go (ModelRef + (Un)Marshal + ParseModelRef + ResolvePhaseModels→FullID),
  config/config.go (Clone), models/resolve.go (CacheReader cache-first + shell-out fallback),
  cmd/archon/config.go+main wiring, initcmd/init.go, status/display.go, tui/models_tab.go+model.go,
  + 10 test files (mechanical ModelConfig→ModelRef fixture migration).
- Independent verification: go build ./... OK; go vet OK; go test ./... ALL 11 pkgs green.
- SIZE: prod ~237 LOC (160 mod + 77 new pkg); test ~711 LOC churn; TOTAL code ~948 ins / 173 del.
  EXCEEDS D1 400 — flag at PR (C1). Prod near forecast; overage is mechanical test fixtures.
- parseModels/execLister retained (PR #45 preserved). yaml lib = gopkg.in/yaml.v3.
- Byte-identity: TestConfig_FlatStringRoundtripByteIdentical passes (models block byte-identical).
(design DONE, approved by user incl. 3 open Qs accepted:
byte-identity scoped to models block; ResolveModels gains CacheReader param; keep Model.Reasoning.)
spec DONE/approved. propose DONE/approved.
Tasks artifact: `openspec/changes/structured-model-resolution/tasks.md`
(7 groups, 30 tasks: S0×5, S1a×2, S1b×4, S1c×4, S1d×4, S1e×5, S1f×1 optional, S1g×7 gates.
Ordering: S0 → S1a → S1b+S1d (parallel after S1a) → S1c (after S1b) → S1e (after S1b) → S1f optional → S1g gates.)
Design artifact: `openspec/changes/structured-model-resolution/design.md`
(apply-ready: S0 cache reader file/funcs/structs; ModelRef + (Un)MarshalYAML + ParseModelRef;
ModelConfig field swap; Clone; ResolvePhaseModels→FullID; resolve.go cache-first+shell-out
fallback w/ injected CacheReader; CLI/init/status/TUI compile-fixes via FullID/ParseModelRef;
22-scenario test plan; ~590 LOC prod+tests, prod ≈300-330).
Prototype-verified: legacy flat models block round-trips byte-identical; empty Leader omitted;
no-models config emits no models block. NormalizeModel/Validate retained; parseModels/execLister
retained (PR #45 preserved).
Capabilities: `model-ref`, `opencode-model-cache`. Scoped to the FOUNDATION PR = S0+S1 combined.
First chained PR = cache reader (`internal/opencode`) + `ModelRef` + back-compat
(Un)MarshalYAML + repoint `ResolveModels`. ~400-480 LOC, may exceed D1 → flag at PR (C1).
S2 (writers), S3 (TUI picker), S4 (variants) = subsequent chained PRs, future cycles.
Branch: `feat/opencode-phase-subagents` (Foundation will get its own branch e.g. feat/structured-models-foundation).
Base: `master` @ 7360b4d.

## Spec Artifacts Written (2026-06-19)
- `openspec/changes/structured-model-resolution/specs/model-ref/spec.md` (full spec; 6 requirements, 13 scenarios)
- `openspec/changes/structured-model-resolution/specs/model-ref/model-ref.feature`
- `openspec/changes/structured-model-resolution/specs/opencode-model-cache/spec.md` (full spec; 4 requirements, 9 scenarios)
- `openspec/changes/structured-model-resolution/specs/opencode-model-cache/opencode-model-cache.feature`
Total: 10 requirements, 22 scenarios. Both are NEW full specs (no prior main spec for either capability).

## Explore key findings (this change)
- opencode cache `~/.cache/opencode/models.json` is provider-keyed. KEY ASYMMETRY: `opencode`
  provider keys are BARE (→ FullID `opencode/<key>`); other providers keys already slashed
  (no double-prefix). archon has NO cache reader today.
- Back-compat: `ModelRef{Provider,Model,Effort}` + custom (Un)MarshalYAML — accept both bare
  scalar (legacy) and mapping; marshal scalar-on-empty so unmigrated config.yaml is byte-stable.
- NEVER guess provider for a bare alias; empty Provider = advisory-only valid state.
- Slicing validated: S0 cache reader (opt) → S1 foundation type+migration → S2 writers →
  S3 TUI picker → S4 variants. Ordering: ModelRef must precede writers+TUI (compile coupling).
- Do NOT delete parseModels/execLister (would look like reverting PR #45); REPOINT
  ResolveModels to cache + shell-out fallback.

## Chosen approach (user, 2026-06-19) — Option C
- STRUCTURED model data model: provider + model (+ effort/variant later), like gentle-ai's
  `ModelAssignment{ProviderID, ModelID, Effort}` + `FullID()`.
- Read opencode's STRUCTURED cache `~/.cache/opencode/models.json` (provider-keyed) instead
  of `opencode models opencode-go` + strip-provider (current `internal/models/opencode.go`).
- Per-phase opencode subagents written with `provider/model` (FullID), leader too.
- TUI: pick provider → then model (today free-form text in models_tab).
- Implement as CHAINED PRs:
  - Slice 1 — Foundation: structured type + structured cache reader + back-compat config migration (~250-350 LOC)
  - Slice 2 — Writers: opencode_mode.go per-phase subagents + leader via FullID; templates (~200 LOC)
  - Slice 3 — TUI picker: provider→model (~200 LOC)
  - Slice 4 (optional) — variants/effort + opencode TS plugin (~150 LOC)

## gentle-ai reference (mapped this session, ../gentle-ai)
- `internal/model/model_assignment.go` — `ModelAssignment{ProviderID, ModelID, Effort}`, `FullID()=ProviderID+"/"+ModelID`.
- `internal/opencode/models.go` — catalog loader from `~/.cache/opencode/models.json` (provider-keyed);
  `LoadModels`/`LoadModelsOrEmpty` (missing cache → empty, no error), `Provider`/`Model`,
  `FilterModelsForSDD`, `DetectAvailableProviders`, `MergeCustomProviders`, `EnrichWithVariants`.
- `internal/components/sdd/inject.go:2189` `injectModelAssignments` — 3-case tree:
  explicit assignment→FullID()+variant; existing agent→preserve; else rootModelID (top-level
  `model` in opencode.json) verbatim, JD excluded.
- variants: separate cache `~/.gentle-ai/cache/model-variants.json` produced by an opencode
  plugin `internal/assets/opencode/plugins/model-variants.ts` (calls `client.provider.list()`).
- KEY: gentle-ai never guesses provider — captured at pick time. Robustness = data model, not resolver.

## archon current state (master overhaul, to refactor)
- `internal/config/model.go` — `ModelConfig{Default, Leader string, Phases map[string]string}` (FLAT strings);
  `NormalizeModel` (alias→family/curated), `ResolvePhaseModels`, `PhaseOrder` (8 phases, judge excluded).
- `internal/models/opencode.go` — shells out `opencode models opencode-go`, `parseModels` STRIPS provider.
- `internal/initcmd/opencode_mode.go` — `mergeOpencodeAgent` writes single `archon-leader` (model verbatim).
- Consumers of model strings: `internal/status/display.go`, `internal/tui/models_tab.go`, `internal/initcmd/templates.go`.

## Preflight Choices (session, unchanged)
- A1 Interactive
- B1 OpenSpec
- C1 ask-always (review budget gate)  → forecast >400 LOC → CHAINED PRs agreed
- D1 400 changed lines per PR
- E1 Playwright disabled (Go CLI project)

## Completed Phases (this change)
- (none yet — re-explore in progress)

## Open Questions (for explore to settle)
- Config YAML schema for structured models: shape + back-compat for existing flat-string config.yaml.
- Migration: how to resolve an existing flat string (e.g. "opus") to provider+model at load (catalog lookup? leave as-is until re-picked?).
- How much of `NormalizeModel`/`ResolvePhaseModels` survives vs is retired for the opencode path.
- Exact slice boundaries + LOC per slice; which slice the TUI picker lands in.
- Whether Slice 1 can ship without breaking the just-merged TUI/status consumers.

## Next Recommended Step
Human Review Gate on the TASKS artifact (`tasks.md`): show group breakdown, total task count,
ordering risks, and Definition of Done. Ask "¿Quieres ajustar algo en esta fase antes de continuar?".
On approval, advance to `apply` phase (create branch `feat/structured-models-foundation` off
`master` @ 7360b4d and apply tasks in order).

## Side note still pending (separate fix, carried over)
`internal/initcmd/templates.go` Human Review Gate prompt uses voseo ("¿Querés ajustar...?"),
violates Leader Persona no-voseo. Out of scope here.
