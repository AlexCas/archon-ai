# Exploration — model-effort-variants (Slice 4, OPTIONAL)

**Goal**: let users pick an effort/reasoning level for models that support it, and have opencode
use it via the per-agent `variant` field in opencode.json. Branch `feat/structured-models-effort`
(== master; Slices 1-3 merged). Date 2026-06-19.

## Decisive finding
The FULL gentle-ai pipeline needs a runtime opencode PLUGIN (`model-variants.ts`) that writes a
variants cache, installed via an embedded-asset subsystem. **archon has NO embed/asset/plugin-install
machinery at all** (no `go:embed`, no assets dir, no `.ts`). Option (a) = build that whole subsystem
+ a first-run-empty data dependency. Disproportionate for an optional slice.

## Already in master (foundation paid for)
- `config.ModelRef.Effort` exists (unused). `MarshalYAML` already emits a MAPPING `{provider,model,effort}`
  when Effort set, scalar otherwise — round-trip tested. NO config changes needed.

## The gap (common to all options that DO something)
- `internal/opencode.Model` has NO `Variants` field (only ID,Name,ToolCall,Reasoning).
- `config.ResolvePhaseModels` flattens ModelRef→FullID and DROPS Effort.
- `internal/initcmd/opencode_mode.go` writes only `model`, no `variant`; structs lack a Variant field.
- TUI picker has no effort step.

## Options
| Opt | What | Size | Verdict |
|---|---|---|---|
| (a) Full port | Model.Variants + variants cache + NEW embed/plugin-install subsystem + TS plugin + effort step + variant write | very large (~600+ LOC + TS + embed infra), fragile first-run | NOT worth it |
| (b) Reasoning-derived (no plugin) | offer fixed low/medium/high ONLY for `Model.Reasoning==true` (gentle-ai itself trusts the Reasoning flag for this); effort picker sub-mode; write `variant`. No plugin/cache/embed. | medium ~150-250 LOC, all Go | RECOMMENDED |
| (c) Free-form effort | user types effort in the picker; write `variant`. No detection. | small ~80-150 LOC | viable minimal |
| (d) Skip | nothing; Effort stays unused | 0 | defensible |

## Recommended: (b). Concrete edits
- `config/model.go`: add `Effort` to `PhaseModel`; set it in `ResolvePhaseModels` from `ref.Effort`.
- `initcmd/opencode_mode.go`: add `Variant string \`json:"variant,omitempty"\`` to archonLeaderAgent +
  archonPhaseAgent; populate leader from `models.Leader.Effort`, phase from `pm.Effort`. (archon owns the
  archon-* keys and rewrites them whole each run, so omitempty is idempotency-safe — verify with a test.)
- `tui/models_tab.go`: add `effortSelect` sub-mode + `effortCursor`; after model select, if the picked
  `opencode.Model.Reasoning`, enter effortSelect (options: "default"→"" , low, medium, high) → set
  `row.ref.Effort`; else apply as today. Render the effort list; wire into update()/view().
- Tests: ResolvePhaseModels carries Effort; opencode_mode golden (variant present when set, absent when
  not; idempotent re-run); tui effort sub-mode only for reasoning models, default→empty.

## Determinism/back-compat
Marshaling deterministic (scalar effortless / mapping with effort); existing configs untouched.
opencode.json keys sorted; adding `variant` stays deterministic; archon rewrites whole archon-* entries
each run so omitempty leaves no stale variant. Confirm with idempotency test.

## gentle-ai refs (read-only)
`../gentle-ai/internal/opencode/models.go` (Variants/EffortLevels/LoadVariants/EnrichWithVariants),
`internal/assets/opencode/plugins/model-variants.ts`, `internal/tui/screens/model_picker.go` ModeEffortSelect,
`internal/components/sdd/inject.go:2189` (variant write).
