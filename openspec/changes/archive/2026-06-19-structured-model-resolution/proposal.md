# Proposal — structured-model-resolution (Foundation PR)

**Change**: `structured-model-resolution`
**Scope of THIS proposal**: Foundation PR only (Slice 0 + Slice 1 combined).
**Branch**: `feat/opencode-phase-subagents` (Foundation gets its own branch, e.g. `feat/structured-models-foundation`).
**Base**: `master` @ 7360b4d (post PR #45 dynamic-detection overhaul).
**Date**: 2026-06-19
**Authoritative input**: `openspec/changes/structured-model-resolution/exploration.md`.

## Problem statement

Archon represents every model assignment as a **flat string** (`ModelConfig{Default, Leader
string, Phases map[string]string}` in `internal/config/model.go:9-13`). A flat string cannot
carry the **provider** that the model belongs to. Two failures follow from this single design
gap:

1. **opencode gets unqualified models.** opencode delegations need a fully-qualified
   `provider/model` id (e.g. `opencode/deepseek-v4-pro`). Today's resolution path
   (`internal/models/opencode.go`) shells out `opencode models` and `parseModels` **strips the
   provider**, so the value archon carries can never be re-qualified reliably.
2. **No reliable source of provider truth.** archon has **no reader** for opencode's structured
   cache (`~/.cache/opencode/models.json`, provider-keyed). The only data source strips exactly
   the field we need, and there is no place to put it once recovered.

The proven fix (validated against gentle-ai) is a **structured data model**: capture the
provider alongside the model at selection/config time and never guess it later. gentle-ai's
robustness comes from its data model (`ModelAssignment{ProviderID, ModelID, Effort}` +
`FullID()`) plus a structured cache reader — **not** from an alias resolver. This proposal ports
the minimal foundation of that approach into archon.

## Goal / Outcome

Give archon a structured `provider + model` representation and a structured opencode cache
reader, **without** breaking any existing flat-string `config.yaml` and **without** reverting
the dynamic-detection work merged in PR #45. After this PR, every downstream consumer still
reads a plain string (via `FullID()`), so status/init/TUI keep compiling and rendering exactly
as before — while the type can now carry a provider for later slices to exploit.

## Broader initiative & this PR's place in the chain

`structured-model-resolution` is a multi-PR (chained) initiative that refactors archon's model
representation from flat strings to a structured provider-aware model, reads opencode's
structured cache instead of stripping the provider, and writes per-phase opencode subagents
(plus a leader) using fully-qualified ids. The agreed chain:

- **Foundation (THIS PR = S0 + S1 combined)** — cache reader package + `ModelRef` type +
  back-compat config (Un)Marshal + repoint resolution. *Green criterion below.*
- **S2 Writers** *(later)* — `opencode_mode.go` per-phase subagents + leader via `FullID()`;
  templates emit `FullID`. Absorbs the superseded `opencode-phase-subagents` design.
- **S3 TUI picker** *(later)* — rewrite `tui/models_tab.go` into a provider → model two-step
  using the cache catalog.
- **S4 effort/variants** *(later, optional)* — effort/variant third step + opencode variants
  TS plugin.

The ordering hazard motivating S0+S1 first: `ModelRef` (with its `FullID()` string accessor)
must land **before** the writers and TUI, because every consumer calls it to keep rendering a
string. Landing the cache reader (S0) in the same PR de-risks S1's resolution repoint while
staying a pure, caller-free addition.

## In scope (Foundation PR)

1. **NEW package `internal/opencode`** — a structured reader of opencode's provider-keyed cache
   `~/.cache/opencode/models.json`. Port the minimal, self-contained subset from
   `../gentle-ai/internal/opencode/models.go`:
   - `DefaultCachePath()`
   - `Provider` / `Model` structs (the fields archon needs; pricing/limit may ride along)
   - `LoadModels(path)` and `LoadModelsOrEmpty(path)` — **missing cache → empty map, NO error**.
   - Tests (the `testdata` dir already exists per exploration).
   - **Skip** `FilterModelsForSDD`, `DetectAvailableProviders`, variants/`FixOpenRouter`,
     `MergeCustomProviders`, `EnrichWithVariants` — deferred to later slices.

2. **NEW type `config.ModelRef{Provider, Model, Effort}` + `FullID()`** handling the
   opencode-provider **key asymmetry**:
   - bare opencode keys → `opencode/<key>`;
   - already-slashed keys for other providers → **no double-prefix**;
   - empty `Provider` → advisory-only; `FullID()` returns the **bare** `Model`.

3. **Change `config.ModelConfig` fields from flat strings to `ModelRef`** (`Default`, `Leader`)
   and `map[string]ModelRef` (`Phases`), **with back-compat**:
   - custom `UnmarshalYAML` — accept a legacy **bare scalar** (`a/b` → split; bare → empty
     provider) OR a structured **mapping**.
   - custom `MarshalYAML` — emit a **scalar** when `Provider=="" && Effort==""`, else a mapping.
   - Net effect: existing `.archon/config.yaml` files **load and re-save byte-identical** until
     the user re-picks a model. **NEVER guess a provider for a bare legacy alias.**

4. **Update `config.Clone`** for the new types (`TestConfig_CloneRoundtrip` guards the deep copy).

5. **Adapt `ResolvePhaseModels`** so `PhaseModel.Model` = `FullID()` when a provider is present,
   else the bare alias (back-compat). Keep `NormalizeModel` for `Validate` (offline advisory) —
   it is retired only from the resolution path, not deleted.

6. **Repoint `internal/models/resolve.go` `ResolveModels`** to use the new cache reader **WITH
   the existing shell-out (`opencode models`) as fallback**. Do **NOT** delete `parseModels` /
   `execLister` (that would look like reverting PR #45). Keep the `Resolve() []string`
   signature; emit `FullID`s.

7. **Update `cmd/archon/config.go` + `cmd/archon/main.go`** to split a `provider/model` string
   into a `ModelRef` for `config set/get` and the init flags.

## Out of scope (later chained PRs)

- **S2 Writers** — `opencode_mode.go` per-phase subagents + leader via `FullID()`; templates
  emit `FullID`.
- **S3 TUI picker** — `tui/models_tab.go` provider → model two-step rewrite.
- **S4 effort/variants** — third selection step + opencode TS variants plugin.

## Affected files (from exploration)

- **New**: `internal/opencode/models.go` (+ `models_test.go`, `testdata/`).
- `internal/config/model.go` — `ModelRef` + `FullID()`; `ModelConfig` field types;
  (Un)MarshalYAML; adapt `ResolvePhaseModels` (`:179-192`); keep `NormalizeModel` (`:148-173`)
  for `Validate` (`:194-205`).
- `internal/config/config.go` — `Clone` (`:86-104`) for the new types; `Load`/`Save`
  (`:60-129`) unchanged in behavior (rely on the new (Un)Marshal).
- `internal/models/resolve.go` — repoint `ResolveModels` (`:24-50`) to the cache reader with
  shell-out fallback; keep `parseModels`/`execLister` (`internal/models/opencode.go:20-52`).
- `cmd/archon/config.go` — `config set/get models.*` split `provider/model` → `ModelRef`.
- `cmd/archon/main.go` (`:81-90,122-162`) — init flags split + `Validate`.

Consumers that must keep compiling/rendering unchanged via `FullID()`:
`internal/status/display.go:53-71`, `internal/tui/models_tab.go`, `internal/tui/model.go`,
`internal/initcmd/{init.go,opencode_mode.go,templates.go}`.

## Back-compat strategy

| Concern | Mechanism |
| --- | --- |
| Existing flat `config.yaml` loads | `UnmarshalYAML` accepts a bare scalar (`a/b` split; bare → empty provider). |
| Re-save is byte-identical | `MarshalYAML` emits a scalar when `Provider=="" && Effort==""`. |
| No silent provider invention | Bare legacy alias keeps `Provider==""` (advisory-only valid state). |
| `archon update` (Clone → Save) doesn't churn | `Clone` preserves empty-provider entries; Marshal re-emits the original scalar. |
| opencode key asymmetry | `FullID()` builds from `Provider.ID` + bare key; never double-prefixes slashed keys. |
| PR #45 preserved | `ResolveModels` augments the catalog source (cache + shell-out fallback); `parseModels`/`execLister` retained. |

## Risks

- **Back-compat for existing config.yaml (dominant).** Mitigation: dual-accept `Unmarshal` +
  scalar-on-empty `Marshal`; an explicit byte-identical round-trip test on an unmigrated config.
- **Looks like reverting PR #45.** Mitigation: do NOT delete `parseModels`/`execLister`;
  REPOINT `ResolveModels` to the cache **with** shell-out fallback; frame as "augment the
  catalog source," not "replace it."
- **Cache staleness / absence.** Mitigation: `LoadModelsOrEmpty` returns empty-on-missing (no
  error) and the shell-out fallback covers a stale/empty cache.
- **opencode-provider key asymmetry.** Mitigation: build `FullID` from `Provider.ID` + bare
  model key; never double-prefix already-slashed keys.
- **Bare-alias provider ambiguity.** Mitigation: empty `Provider` is valid everywhere; later
  opencode writers (S2) skip/omit rather than write `/claude-sonnet-4`.
- **TUI compile coupling.** Mitigation: land the full `ModelRef` (with the `FullID` string
  accessor) in this single PR so every consumer still has a string to render.

## Green-after-this-slice (success criteria)

- Existing flat-string `config.yaml` **loads and re-saves byte-identical** until the user
  re-picks.
- `go build ./...` and `go test ./...` all pass.
- status / init / TUI still compile and render because every consumer reads a string via
  `FullID()`.
- The merged dynamic-detection feature (PR #45) is **preserved** (augmented, not reverted).

## Size / PR forecast

- Estimated **~400-480 LOC** for the Foundation PR (S0 + S1 combined).
- This **may slightly exceed the D1 budget of 400 changed lines**. Per **C1 (ask-always)** this
  is **flagged for explicit approval**; the user has **accepted** it for this Foundation PR.
- Remaining initiative ships as chained PRs (S2 ~250-350, S3 ~350-400 [maybe 3a/3b], S4 ~200).

## New capabilities to spec

- **`model-ref`** — the structured `ModelRef{Provider, Model, Effort}` type, `FullID()`
  semantics (incl. opencode key asymmetry and empty-provider advisory behavior), and the
  back-compat (Un)MarshalYAML contract on `ModelConfig`.
- **`opencode-model-cache`** — the `internal/opencode` cache reader (`DefaultCachePath`,
  `Provider`/`Model`, `LoadModels`/`LoadModelsOrEmpty` with empty-on-missing) and its use as the
  primary catalog source in `ResolveModels` with the shell-out fallback.
