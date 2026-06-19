# Exploration — structured-model-resolution

**Initiative**: Refactor archon's model representation from FLAT STRINGS to a STRUCTURED
provider+model model (like gentle-ai's `ModelAssignment{ProviderID, ModelID, Effort}` +
`FullID()`), read opencode's STRUCTURED cache (`~/.cache/opencode/models.json`,
provider-keyed) instead of stripping the provider, and write per-phase opencode subagents
(+ leader) using `provider/model`. Multi-PR (chained) initiative.

**Branch**: `feat/opencode-phase-subagents` (base `master` @ 7360b4d, post PR #45 overhaul)
**Date**: 2026-06-19
**Supersedes**: `../opencode-phase-subagents` (its per-phase writer = Slice 2 here)

## Verified facts

- **opencode cache exists** at `~/.cache/opencode/models.json` (~2.3 MB), provider-keyed:
  `provider → {id, name, env[], models: {<key> → {id, name, family, tool_call, reasoning, cost, limit}}}`.
- **KEY ASYMMETRY**: under the `opencode` provider, model keys are BARE (`deepseek-v4-flash`,
  `glm-4.7`) → FullID must be `opencode/deepseek-v4-flash`. Under other providers keys are
  ALREADY slashed (`xai/grok-4`) → do NOT double-prefix. Source of truth for FullID =
  `Provider.ID` + bare model key.
- **archon has NO cache reader today.** `internal/models/opencode.go` shells out
  `opencode models opencode-go` and `parseModels` STRIPS the provider.
- **gentle-ai port is real and self-contained** (no cross-package deps): `ModelAssignment`
  + `FullID()` (~17 LOC); `internal/opencode/models.go` cache loader (`DefaultCachePath`,
  `LoadModels`/`LoadModelsOrEmpty` [missing cache → empty, no error], `Provider`/`Model`,
  `FilterModelsForSDD`, `DetectAvailableProviders`).
- KEY INSIGHT: gentle-ai's robustness = it NEVER guesses a provider (captured at TUI pick
  time). The data model, not a resolver, is what works well.

## Inventory of model-string touch points (flat → structured)

- Core type: `internal/config/model.go:9-13` `ModelConfig{Default, Leader string, Phases map[string]string}`.
- `model.go:124-128` `PhaseModel`; `:148-173` `NormalizeModel`; `:179-192` `ResolvePhaseModels`;
  `:82-96` `ValidPhases`/`PhaseOrder` (8 phases, provider-agnostic, survives); `:18-80` curated
  static lists (become fallback/advisory); `:194-205` `Validate`.
- Serialization: `internal/config/config.go:51` `Models ModelConfig`; `:60-76` `Load`;
  `:86-104` `Clone` (hand-rolled, must update — `TestConfig_CloneRoundtrip` guards); `:106-129` `Save`.
- Init/writers: `init.go:23-25` Options flat fields; `:207-239` `buildConfig`; `:94`
  `ResolvePhaseModels` → template; `:101-109` `mergeOpencodeAgent(.., cfg.Models.Leader)`.
  `opencode_mode.go:12-84` single `archon-leader` writer (Slice 2 surface).
  `templates.go:162-171` AGENTS.md "Phase Models" advisory block.
- Resolution/detection: `models/opencode.go:20-52` shell-out + strip; `models/resolve.go:24-50`
  `ResolveModels`/`Resolve` → `[]string`; `models/detect.go`.
- Display/CLI: `status/display.go:53-71`; `cmd/archon/config.go` (`config set/get models.*`,
  leader not settable today); `cmd/archon/main.go:81-90,122-162` init flags + Validate.
- TUI: `tui/models_tab.go` (free-form text inputs, 311 LOC); `tui/model.go:88,93,182,334,349-355`.

## Config YAML: back-compat plan

Today (flat, per `config_test.go:30-33`):
```yaml
models:
  default: claude-sonnet-4
  leader: anthropic/claude-sonnet-4-20250514
  phases: { apply: gpt-4o, verify: claude-haiku-4-5 }
```
Target (structured):
```yaml
models:
  default: { provider: anthropic, model: claude-sonnet-4-6 }
  leader:  { provider: anthropic, model: claude-sonnet-4-6 }
  phases:
    apply:  { provider: opencode, model: deepseek-v4-pro }
    verify: { provider: anthropic, model: claude-haiku-4-5 }
```
Strategy — `ModelRef` type with custom (Un)MarshalYAML:
- `ModelRef{Provider, Model, Effort}` + `FullID()`. `ModelConfig` → `{Default ModelRef; Leader ModelRef; Phases map[string]ModelRef}`.
- `UnmarshalYAML`: ScalarNode (legacy) → split `a/b`→(a,b), bare→("",b); MappingNode → structured.
- `MarshalYAML`: emit SCALAR when `Provider=="" && Effort==""` (byte-compat for unmigrated),
  else mapping. → existing config.yaml loads + re-saves unchanged until user re-picks.
- NEVER guess a provider for a bare alias (empty Provider is a valid advisory-only state).
- Update `Clone`; `config set/get` split `provider/model` strings.

## NormalizeModel / ResolvePhaseModels fate

- `ResolvePhaseModels` SURVIVES; returns `PhaseModel.Model` = FullID when provider present,
  else bare alias (back-compat). Powers AGENTS.md advisory block (should emit FullID so
  opencode delegations get `model: opencode/deepseek-v4-pro`).
- `NormalizeModel` retired from resolution path, KEPT for `Validate` (offline advisory).

## Slicing plan (chained PRs, each green)

Ordering hazard: `ModelRef` type must precede writers + TUI (they call `.FullID()`).

- **Slice 0 (optional, ~120 LOC)**: cache reader only. New `internal/opencode` pkg
  (`DefaultCachePath`, `Provider`/`Model`, `LoadModels`/`LoadModelsOrEmpty`) + tests. Pure
  addition, no callers. De-risks Slice 1. (testdata dir already exists.)
- **Slice 1 Foundation (~300-380 LOC)**: `ModelRef` + (Un)MarshalYAML + change `ModelConfig`
  fields; adapt `ResolvePhaseModels` (FullID-or-alias); update `Clone`; repoint
  `models/resolve.go` at the cache reader WITH shell-out fallback (keep `Resolve()[]string`,
  emit FullIDs); `config set/get`/`main.go` split. Green: flat YAML loads/saves byte-identical.
- **Slice 2 Writers (~250-350 LOC)**: `opencode_mode.go` leader via `Leader.FullID()` + loop
  `PhaseOrder` writing `archon-<phase>` subagents; update call sites `init.go`/`tui/model.go`;
  templates emit FullID. (Absorbs the superseded opencode-phase-subagents design.)
- **Slice 3 TUI picker (~350-400 LOC, maybe split 3a/3b)**: rewrite `models_tab.go` to
  provider→model two-step using the cache catalog. Green: writers+type already structured.
- **Slice 4 (optional, ~200 LOC)**: effort/variants third step + gentle-ai `EnrichWithVariants`
  + opencode variants TS plugin.

## Risks

- **Back-compat for existing config.yaml** (dominant) → dual-accept Unmarshal + scalar-on-empty
  Marshal; verify `archon update` (Clone→Save) doesn't churn unmigrated entries.
- **Looks-like-reverting-PR #45** → do NOT delete `parseModels`/`execLister`; REPOINT
  `ResolveModels` to cache with shell-out fallback. Frame as "augment catalog source."
- **Cache staleness/absence** → `LoadModelsOrEmpty` empty-on-missing + shell-out fallback;
  TUI degrades to free-form.
- **opencode-provider key asymmetry** → build FullID from `Provider.ID` + bare key; never
  double-prefix already-slashed keys.
- **Bare-alias provider ambiguity** → empty Provider valid everywhere; opencode writer must
  skip / omit model key rather than write `/claude-sonnet-4`.
- **TUI compile coupling** → Slice 1 must land the full `ModelRef` (with FullID string
  accessors) in one PR so every consumer still has a string to render.

## Key file:line references
- `internal/config/model.go:9-13,82-96,124-128,148-205`; `config.go:51,60-129`
- `internal/initcmd/{init.go:23-25,94,101-109,207-239; opencode_mode.go:12-84; templates.go:162-188}`
- `internal/models/{opencode.go:20-52, resolve.go:24-50, detect.go}`
- `internal/status/display.go:53-71`; `cmd/archon/{config.go, main.go:81-162}`
- `internal/tui/{models_tab.go, model.go:88,93,182,334,349-355}`
- gentle-ai: `../gentle-ai/internal/model/model_assignment.go:7-16`;
  `../gentle-ai/internal/opencode/models.go:13-157,271-320`
