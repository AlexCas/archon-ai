# Design: Multi-provider per-phase models (Slice 1)

## Technical Approach

Generalize the Claude-only `NormalizeModel` (`internal/config/model.go:97-113`) into a
provider-aware matcher that recognizes curated Gemini/OpenAI/Opencode catalog ids in
addition to Claude family tokens. The existing whole-token tokenizer, the
`opus→sonnet→haiku` priority, advisory `Validate`, and both render paths are PRESERVED.
The only logic change is inside `model.go`; `templates.go`, `init.go`, `tui/model.go`
are confirmed provider-agnostic and unchanged. Implements the `harness-init` MODIFIED
"Normalization to real model IDs" and ADDED "Cross-provider precedence" requirements.

## Architecture Decisions

### Decision: One ordered provider table consumed by the existing tokenizer

**Choice**: A package-level `providerFamilies []providerFamily`, walked in fixed order
Claude → Gemini → OpenAI → Opencode. Claude rows match a family token and emit the
short alias; non-Claude rows match a curated catalog id and emit that id as-is.

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Single ordered table (exploration A1) | Minimal departure; one structure to review; deterministic by order | **Chosen** |
| Per-provider dispatch funcs (A2) | More surface; duplicated tokenizer boilerplate | Rejected (refactor target if quirks appear) |
| Closure matcher list (A3) | Flexibility we don't need; easy to make non-deterministic | Rejected |

**Rationale**: smallest reviewable step that keeps the proven octopus-safe semantics in
one place. Table order encodes the precedence requirement directly.

### Decision: Canonical output split (Claude alias vs. catalog id)

**Choice**: Claude → short alias (`opus`/`sonnet`/`haiku`); Gemini/OpenAI/Opencode →
matched catalog id verbatim.

**Rationale**: Claude has a stable family-alias system (avoids dated-ID churn in
CLAUDE.md); the other providers have no comparable alias — their accepted id IS the
catalog string. Confirmed decision; not re-litigated.

### Decision: Matching semantics per provider

- **Claude**: token equals a family alias (`opus`/`sonnet`/`haiku`) — unchanged.
- **Non-Claude**: the lowercased+trimmed input equals a curated catalog id (exact,
  case-insensitive). Catalog ids carry version digits, so whole-token family matching
  does not apply; an exact-id check keeps `gpt-4` distinct from `gpt-4o` and avoids
  spurious collisions.

## Data Flow

    value ─► NormalizeModel
              │ lower+trim; tokenize (existing FieldsFunc, octopus-safe)
              ▼
        walk providerFamilies in fixed order
          Claude row:    any token == alias?      → alias
          non-Claude row: input == catalog id?     → id (as-is)
          first match wins ─────────────────────► (id, true)
          no match ─────────────────────────────► ("", false)

`ResolvePhaseModels`, `Validate`, `StaticModels`, `KnownModels` call through unchanged
and inherit multi-provider support.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modify | Add `GeminiModels`, `OpenAIModels` vars; `providerFamilies` table; rewrite `NormalizeModel` body to walk it; fold new catalogs into `StaticModels()`. |
| `internal/config/model_test.go` | Modify | Invert the three non-Claude `ok=false` rows; add per-provider + precedence tables. |
| `internal/initcmd/templates_test.go` | Modify | Add render assertion: a non-Claude default emits `## Phase Models`. |
| `internal/initcmd/templates.go`, `init.go`, `internal/tui/model.go`, `models_tab.go` | None | Provider-agnostic; confirmed no edits. |

## Interfaces / Contracts

Signature is unchanged — `func NormalizeModel(s string) (id string, ok bool)`. New
internal structure (non-obvious part):

```go
type providerFamily struct {
    families []string // Claude family tokens (whole-token match)
    catalog  []string // non-Claude catalog ids (exact match); canonical = the id
}

// Walked in fixed precedence: Claude, Gemini, OpenAI, Opencode.
var providerFamilies = []providerFamily{
    {families: claudeFamilies},            // emit alias
    {catalog: GeminiModels},               // emit id
    {catalog: OpenAIModels},
    {catalog: OpencodeModels},
}
```

`NormalizeModel`: for a Claude row, return the family token if any whole token equals
it (existing inner loop); for a catalog row, return the id if the trimmed/lowered input
equals it. First row that matches wins.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|--------------|----------|
| Unit (config) | Claude rows unchanged | Keep all existing `ok=true` Claude rows + `octopus`/`Opues 4.8` (scenarios: display-name, whole-token guard, typo). |
| Unit (config) | Non-Claude inversion | Rewrite `glm-5`/`kimi-k2.5`/`gpt-4`(→pick a real id, e.g. `gpt-4o`) rows to `ok=true`, id as-is. |
| Unit (config) | Per-provider tables | One curated id per Gemini/OpenAI/Opencode → itself (scenarios: Gemini/OpenAI/Opencode catalog id). |
| Unit (config) | Precedence/collision | A value matching Claude + a later catalog → Claude alias (scenario: colliding value). |
| Render (initcmd) | Non-Claude block emits | `ResolvePhaseModels(ModelConfig{Default: <gemini id>})` → `## Phase Models` non-empty (scenarios: non-Claude default, byte-identical across paths via existing matching test). |
| Unit (config) | `Validate` advisory | Typo still warns; curated non-Claude ids no longer warn (scenario: typo omitted, not rejected). |

## Migration / Rollout

No migration required. Generated CLAUDE.md/AGENTS.md changes only on the next
`archon init` / TUI regenerate. No config schema or on-disk data changes.

## Open Questions

- [ ] **CONFIRM curated catalog contents** (external facts that drift). Proposed:
  `GeminiModels = ["gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash"]`;
  `OpenAIModels = ["gpt-4o", "gpt-4o-mini", "gpt-4.1", "o3", "o4-mini"]`.
  `OpencodeModels` reused as-is. The `model_test.go` `gpt-4` row should adopt a confirmed
  OpenAI id (e.g. `gpt-4o`).
- [ ] Size: ~1 function rewrite + 2 small vars + test-table edits ≈ 120–170 lines —
  fits the 400-line budget (D1); single PR, no chaining.
