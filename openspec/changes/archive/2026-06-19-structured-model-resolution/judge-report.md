# Judge Report — structured-model-resolution (Foundation S0+S1)

**Date**: 2026-06-19
**Method**: Blind dual adversarial review (Judge A + Judge B, independent, opus).
**Initial verdict**: DO-NOT-SHIP (both judges). Re-apply retry 1 in progress.

## Consensus / confirmed issues

### Issue 1 — Slashed legacy scalar breaks byte-stability (A:HIGH / B:MED) — FIX
`internal/config/model.go` `MarshalYAML` emits a scalar ONLY when `Provider=="" && Effort==""`.
A legacy provider-qualified scalar (`anthropic/claude-sonnet-4-...`, the documented `--leader`
form) unmarshals to `{Provider,Model}` and re-marshals as a multi-line MAPPING → violates the
"existing config.yaml re-saves byte-identical" contract. Both judges reproduced it empirically.
The byte-identity test only covered bare aliases, so it missed this.
**FIX**: emit the scalar `FullID()` whenever `Effort==""` (drop the `Provider==""` condition);
mapping only when `Effort!=""`. Add a round-trip test fixture with a slashed scalar
(`anthropic/claude-...`) asserting scalar-form preservation.

### Issue 2 — Cache catalog explosion (B:HIGH) — FIX
`internal/models/resolve.go` `cacheModelNames` flattens EVERY provider (real cache = 145
providers / 5278 models) into the returned `[]string`, whereas the retained shell-out fallback
returns only `opencode models opencode-go` (bare) names. The TUI hint renders all 5278 joined;
the picker cycles all. Spec sanctioned the FullID *format* (C4) but not dropping the opencode-go
scoping → unintended divergence between cache and fallback paths.
**FIX**: scope `cacheModelNames` to the `opencode` provider only (match the fallback scope),
emitting `opencode/<key>` FullIDs. Add/adjust a test asserting only opencode-provider models
are returned.

## Decision (user, 2026-06-19) — not a fix

### Issue 3 — NormalizeModel dropped from resolution changes Claude template advisory (B:MED)
Per spec M6, `NormalizeModel` is not used in resolution, so the Claude template "Phase Models"
advisory now shows `claude-opus-4-8` instead of the `opus` family alias. Spec-sanctioned.
**USER DECISION: ACCEPT id verbatim** — no extra work. The advisory will carry the full id;
it is valid and consistent with the new resolution. CLAUDE.md visible output changes accordingly.

## LOW notes (non-blocking, no fix this slice)
- A-L1 / B-L1: effort-only ModelRef invisible (FullID==""), and an empty-provider+slashed-model
  is not structurally round-trip-stable — but neither state is reachable via ParseModelRef /
  cacheModelNames. Latent only.
- A-L2: corrupt cache silently degrades to shell-out with no warning. Minor; consider a stderr
  note in a follow-up.
- Model.Reasoning captured but unused (intentional, per design).

## Items both judges verified CLEAN
- `go build` / `go vet` / `go test ./...` green; PR #45 `parseModels`/`execLister` retained &
  reachable; CacheReader seam callers updated (no nil-panic); FullID core edge cases; UnmarshalYAML
  no-recursion + never-guesses-provider; Clone deep-copies Phases; ResolvePhaseModels honors
  PhaseOrder; LoadModels malformed-skip; DefaultCachePath HOME-unset errors; no out-of-scope leak.

## Re-judge (retry 1)

**Verdict: SHIP.** Both DO-NOT-SHIP issues from round 1 are resolved; no regression introduced.

### Gate output (exact)
- `go build ./...` → clean (exit 0)
- `go vet ./...` → clean (exit 0)
- `gofmt -l` on the 5 changed files → no output (all formatted)
- `go clean -testcache && go test ./... -count=1` → all packages `ok`:
  cmd/archon, internal/agent, internal/config, internal/initcmd, internal/models,
  internal/opencode, internal/scaffold, internal/status, internal/tui, internal/version, skills.

### Issue 1 — legacy slashed scalar round-trip (RESOLVED)
`ModelRef.MarshalYAML` now emits a scalar `FullID()` whenever `Effort == ""`; only an
effort-bearing ref marshals as a mapping. Confirmed:
- `TestConfig_SlashedScalarRoundtripByteIdentical` PASS — `anthropic/claude-sonnet-4-20250514`
  round-trips byte-identical (scalar→scalar), no `provider:` mapping churn.
- `TestConfig_FlatStringRoundtripByteIdentical` PASS — bare alias `opus` still a scalar (no regression).
- `TestModelRef_MarshalYAML` / `_UnmarshalYAML` / `TestModelRef_FullID` / `TestParseModelRef` all PASS.

**Reasoning on the M4 deviation (scalar for any no-effort ref; mapping only with effort).**
Sound. The system-wide observable identity of a ModelRef is `FullID()` — verified that NO
production consumer reads `.Provider`/`.Model` directly (grep across internal/ + cmd/: zero hits;
status/display, cmd/archon/config, initcmd, tui, ResolvePhaseModels all go through `FullID()`).
A round-trip probe over edge refs shows `FullID()` is preserved in EVERY case
(legacy slashed, bare alias, multi-slash model, embedded slash, empty-model `anthropic/`, fully
empty). Mapping→scalar convergence is intended and harmless:
- `{provider:openrouter, model:xai/grok-4}` (no effort) marshals to `xai/grok-4` and unmarshals
  back to `{provider:xai, model:grok-4}`. Struct fields churn, but `FullID()` is `xai/grok-4`
  both before and after — delegation behaves identically and the struct churn is unobservable.
- This input is reachable via `ParseModelRef("openrouter/xai/grok-4")`, but a fixed-point probe
  confirms it normalizes to `xai/grok-4` on the FIRST save and stays there on every subsequent
  save (load→save→load→save is byte-stable). So there is a one-time semantic-preserving
  normalization, never ongoing churn. The `openrouter` prefix was already dropped by `FullID()`
  before any delegation regardless of marshal behavior.
- Edge `{Provider:"anthropic", Model:"", Effort:""}` → `FullID()`=`anthropic/` → marshals to
  scalar `anthropic/` → unmarshals back identically (fixed point). Only reachable by a user
  typing a trailing slash; not harmful, not churning.

This is the same latent point round-1 logged as A-L1/B-L1 ("empty-provider+slashed-model not
structurally round-trip-stable — latent only"). Re-confirmed LATENT/LOW, non-blocking: invisible
to all consumers and byte-stable across re-saves.

### Issue 2 — cacheModelNames over-broad scope (RESOLVED)
`cacheModelNames` now scopes to `const opencodeProviderID = "opencode"`, returns nil when that
provider is absent, emits `opencode/<key>` FullIDs sorted. Confirmed:
- `TestCacheModelNames_Mapping` PASS — non-opencode provider (`requesty`/`xai/grok-4`) excluded,
  exactly 2 opencode models returned, sorted.
- `TestCacheModelNames_NoOpencodeProvider` PASS — nil when no opencode entry.
- `TestResolveModels_PrefersCache` PASS — cache preferred; shell-out lister NOT invoked.
- `TestResolveModels_FallbackWhenCacheEmpty` PASS — empty cache → shell-out lister invoked
  (PR #45 `parseModels`/`execLister` path reachable). Curated fallback tests still PASS.
- Cache-vs-fallback scope now matches (both opencode-go scoped).
- Sole `Resolve()` caller is `internal/tui/model.go:88` (`catalog := models.Resolve()`); the
  catalog is consumed as opaque pick/display strings, so FullID-form names are fine — no breakage.

### Round-1 CLEAN items re-checked (still clean)
- UnmarshalYAML no-recursion: `type modelRefAlias ModelRef` strips the method (model.go:48). ✔
- Clone deep-copy: ModelRef is an all-string value type; Default/Leader copy by value, Phases map
  rebuilt entry-by-entry (config.go:96–101). `TestConfig_CloneRoundtrip` PASS. ✔
- ResolvePhaseModels emits `ref.FullID()` in PhaseOrder, omit-when-empty (model.go:250–263). ✔
- LoadModels malformed-skip: `continue` on per-provider unmarshal error (opencode/models.go:58). ✔

### Scope
No out-of-scope changes from the fixes: Issue 1 touched only the MarshalYAML condition and Issue 2
only the cache scoping; consumers go through the unchanged `FullID()`/`ParseModelRef` seam. The
broader working-tree changes are the pre-existing structured-models foundation, not the fixes.

### Residual LOW (non-blocking)
- A-L1/B-L1 (re-confirmed): structured `{provider, slashed-model}` and `{provider, ""}` are not
  STRUCT-identical across round-trip, but ARE FullID-stable and save-fixed-point. Invisible to all
  consumers. Latent only.
- A-L2: corrupt cache silently degrades to shell-out with no stderr note. Consider a follow-up.
