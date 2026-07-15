# Verify Report — structured-model-resolution (Foundation PR = S0 + S1)

**Change**: `structured-model-resolution`
**Branch**: `feat/structured-models-foundation`
**Phase**: verify (independent re-verification — read the real code, did not trust the green tests)
**Date**: 2026-06-19

## VERDICT: **PASS**

All four gates exit 0, all 22 spec scenarios are implemented correctly AND covered by a passing
test, no out-of-scope code leaked into the Foundation working-tree diff, and the tricky
correctness paths (FullID slash ordering, no-recursion (un)marshal, cache-first preference,
key asymmetry) behave exactly as specified. Three LOW-severity observations are recorded below;
none block the PR.

---

## 1. Gate output (exact)

| Gate | Result |
|---|---|
| `go build ./...` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `go test ./... -count=1` | exit 0 — all packages `ok` |
| `go clean -testcache && go test ./...` | exit 0 — all packages `ok` (DoD final gate) |
| `gofmt -l` (13 changed .go files) | empty — all formatted |

```
ok  github.com/archon-ai/archon/cmd/archon
ok  github.com/archon-ai/archon/internal/agent
ok  github.com/archon-ai/archon/internal/config
ok  github.com/archon-ai/archon/internal/initcmd
ok  github.com/archon-ai/archon/internal/models
ok  github.com/archon-ai/archon/internal/opencode
ok  github.com/archon-ai/archon/internal/scaffold
ok  github.com/archon-ai/archon/internal/status
ok  github.com/archon-ai/archon/internal/tui
ok  github.com/archon-ai/archon/internal/version
ok  github.com/archon-ai/archon/skills
```

---

## 2. Scenario → code → test traceability (22 scenarios)

### model-ref (M1–M6, 13 scenarios)

| # | Scenario | Code | Test | Status |
|---|---|---|---|---|
| M1.1 | FullID joins provider + bare model | `model.go:24-32` | `TestModelRef_FullID/joins...` | ✅ |
| M1.2 | FullID opencode bare key | `model.go:24-32` | `TestModelRef_FullID/opencode...` | ✅ |
| M1.3 | FullID never double-prefixes slashed id (edge) | `model.go:25-27` slash check FIRST | `TestModelRef_FullID/already-slashed...` | ✅ |
| M1.4 | FullID empty provider → bare, no leading slash (edge) | `model.go:28-30` | `TestModelRef_FullID/empty provider...` (+ explicit no-leading-`/` assert) | ✅ |
| M2.1 | ModelConfig carries structured refs | `model.go:78-82` | `TestModelConfig_StructuredFields` | ✅ |
| M3.1 | Legacy slashed scalar splits | `model.go:39-43` | `TestModelRef_UnmarshalYAML/scalar slashed` | ✅ |
| M3.2 | Legacy bare alias → empty provider (edge) | `model.go:44-46` | `TestModelRef_UnmarshalYAML/scalar bare` | ✅ |
| M3.3 | Mapping node decodes structured | `model.go:48-54` (alias, no recursion) | `TestModelRef_UnmarshalYAML/mapping` | ✅ |
| M4.1 | Flat-string config round-trips byte-identical (happy) | `model.go:60-66` scalar-on-empty | `TestConfig_FlatStringRoundtripByteIdentical` | ✅ (see LOW-1 on scope) |
| M4.2 | Bare alias re-saves as same scalar | `model.go:61-62` | `TestModelRef_MarshalYAML/empty provider...scalar` | ✅ |
| M5.1 | Clone round-trip equality | `config.go:96-100` (`map[string]ModelRef`, value copy) | `TestConfig_CloneRoundtrip` (deep-equal + map-independence mutate) | ✅ |
| M6.1 | Resolution emits FullID when provider present | `model.go:248-260` | `TestResolvePhaseModels/provider-qualified...` | ✅ |
| M6.2 | Resolution emits bare alias when no provider (edge) | `model.go:258` `ref.FullID()` | `TestResolvePhaseModels/empty-provider...` | ✅ |

### opencode-model-cache (C1–C4, 9 scenarios)

| # | Scenario | Code | Test | Status |
|---|---|---|---|---|
| C1.1 | Default cache path | `opencode/models.go:33-39` | `TestDefaultCachePath` | ✅ |
| C2.1 | Well-formed cache yields providers/models (happy) | `models.go:45-64` | `TestLoadModels_WellFormed` (incl. `p.ID==key`, slashed key preserved, `tool_call`) | ✅ |
| C2.2 | Malformed provider entry skipped (edge) | `models.go:56-59` `continue` | `TestLoadModels_MalformedEntrySkipped` | ✅ |
| C3.1 | Absent cache → empty map, nil error (edge) | `models.go:68-77` `errors.Is(ErrNotExist)` | `TestLoadModelsOrEmpty_Absent` | ✅ |
| C3.2 | Parse error propagates (error) | `models.go:51-53` + `:72-74` | `TestLoadModelsOrEmpty_ParseError` | ✅ |
| C4.1 | Resolution prefers cache; lister NOT invoked (happy) | `resolve.go:62-64` | `TestResolveModels_PrefersCache` (asserts `!lister.called`) | ✅ |
| C4.2 | Falls back to shell-out when cache empty (edge) | `resolve.go:65-72` | `TestResolveModels_FallbackWhenCacheEmpty` | ✅ |
| C4.3 | PR #45 detection path stays reachable (edge) | `parseModels`/`execLister` retained (`opencode.go:18,36`) | `TestParseModels` + `TestResolveModels_LiveEnumerationFallsBackToCurated` | ✅ |
| C4.* | Cache names provider-qualified FullID form | `resolve.go:31-45` `cacheModelNames` | `TestCacheModelNames_Mapping` (opencode/<key>, slashed as-is, sorted) | ✅ |

**Every one of the 22 scenarios is implemented and covered by a passing test. No unimplemented,
mis-implemented, or untested scenario found.**

---

## 3. Adversarial correctness review (the tricky bits)

I wrote a throwaway in-package probe (since removed) to exercise paths the committed tests do not,
and reasoned about each. Findings:

- **`FullID()` order of checks** — slash short-circuit is FIRST (`model.go:25`), then empty-provider,
  then join. Verified empirically:
  - `{opencode, "xai/grok-4"}` → `"xai/grok-4"` (slash wins over provider; no `opencode/xai/grok-4`). ✅
  - `{"", "opus"}` → `"opus"`, no leading slash. ✅
  - `{"opencode", ""}` → `"opencode/"` (provider + empty model). This only arises from a malformed
    input like `set models.default opencode/` and is consistent with the documented join rule; not a
    bug. (LOW-2)
- **`UnmarshalYAML`** — scalar splits on FIRST `/` (`a/b/c` → `{a, b/c}`), bare → empty provider
  (never guessed), mapping path uses `type modelRefAlias ModelRef` so it does not infinitely recurse.
  Confirmed compiles + decodes + the full suite passes (no stack overflow). ✅
- **`MarshalYAML`** — scalar ONLY when `Provider=="" && Effort==""`. Verified the non-scalar cases
  emit a mapping:
  - `{opencode, ""}` → `provider: opencode` mapping. ✅
  - `{"", opus, high}` (effort-only) → `model: opus\neffort: high` mapping. ✅
  - A mixed `ModelConfig` with a provider-qualified `Leader` marshals the leader as a nested mapping
    while `default`/`phases.apply` stay bare scalars, and re-parses back to the same structured refs.
    ✅ (byte-identity for the legacy scalars within the block holds; see LOW-1.)
  - Alias prevents marshaler re-entry (returns a plain string or `modelRefAlias`, never `ModelRef`). ✅
- **`ResolvePhaseModels`** — emits `ref.FullID()`; selection is `!ok || ref.Model == ""` → fall back
  to Default, omit when both empty; `PhaseOrder` iteration + determinism preserved; `NormalizeModel`
  is NOT called in resolution but IS still used by `Validate` (`model.go:270`). Matches M6 exactly. ✅
- **`LoadModelsOrEmpty`** — absent → `{}` + nil (via `errors.Is(err, os.ErrNotExist)` through the
  `%w`-wrapped read error); invalid JSON → error propagates; malformed per-provider entry skipped not
  fatal. All three confirmed by fixtures `models.json` / `malformed.json` / `invalid.json`. ✅
- **`ResolveModels` cache-first** — cache present → FullID names, lister untouched (test asserts
  `!lister.called`); cache absent/empty → shell-out then curated fallback. CacheReader seam injected;
  all callers pass a reader (`Resolve()` → `defaultCacheReader`; every test passes a stub). ✅
- **opencode key asymmetry** — `cacheModelNames` wraps each key in
  `config.ModelRef{Provider: provID, Model: key}` then calls `FullID()`: bare opencode key
  → `opencode/<key>`; already-slashed key under another provider → as-is (FullID slash short-circuit).
  Traced through and confirmed by `TestCacheModelNames_Mapping`. ✅

---

## 4. Regression check (unmigrated flat-string configs)

No behavior change for existing flat-string configs:
- **status display** (`display.go:55-68`) reads `.FullID()`; for an empty-provider ref `FullID()` ==
  the bare model == the old flat string. Same rendered output.
- **config get / list** (`config.go:119-211`) same — `FullID()` of a bare ref is the bare string.
  `TestConfigCmd_Get/GetPhase/List/ListEmpty/SetRoundtrip` pass unchanged.
- **AGENTS.md phase-models block** — `ResolvePhaseModels` returns `ref.FullID()`; for legacy bare
  refs that is the bare alias, identical to before. `templates_test.go` fixtures (now ModelRef) assert
  the same rendered strings.
- **opencode leader merge** — `init.go:102` and `model.go:334` pass `cfg.Models.Leader.FullID()`;
  `MergeOpencodeAgent` still takes a string. Bare leader → unchanged string.
- **byte-identity** — `TestConfig_FlatStringRoundtripByteIdentical` confirms the `models:` block
  scalars survive a load→marshal, an empty leader emits no `leader:` key, and a config with no
  `models:` key does not gain one.

---

## 5. Out-of-scope leak check

The relevant Foundation diff is the **working-tree** change set (the Foundation work is currently
uncommitted on this branch; HEAD is the PR #45 merge `7360b4d`). `git diff --stat` (working tree)
covers exactly:

- Production: `internal/opencode/models.go` (new) + testdata, `internal/config/model.go`,
  `internal/config/config.go`, `internal/models/resolve.go`, and the five consumer compile-fixes
  (`cmd/archon/config.go`, `internal/initcmd/init.go`, `internal/status/display.go`,
  `internal/tui/models_tab.go`, `internal/tui/model.go`).
- Tests: the matching `_test.go` files plus mechanical `string → config.ModelRef` / `.FullID()`
  adaptations in `templates_test.go`, `update_test.go`, `display_test.go`, `model_test.go`,
  `e2e_test.go`, `models_tab_test.go`. Inspected — every changed line in those fixup files is a
  pure type adaptation; no behavioral assertion changed.

Confirmed ABSENT from the Foundation working-tree diff: no `opencode_mode.go` per-phase writer
changes, no TUI picker rewrite, no variants/EnrichWithVariants, no `FixOpenRouterModels` /
`DetectAvailableProviders` / `MergeCustomProviders` ports. Scope matches S0+S1 exactly.

(Note: `git diff master...HEAD` shows many more files — `opencode_mode.go`, `update.go`,
`playwright_tab.go`, etc. — but those are PRE-EXISTING on this branch from prior merged PRs, not
Foundation work. The Foundation delta is the uncommitted working tree, ~742 changed lines, in line
with the ~590 forecast plus the unavoidable mechanical test fixups.)

---

## 6. Byte-identity scope sanity check

The test marshals `cfg.Models` (the block), not the whole file, and compares against the
yaml.v3-re-emitted form. This is the faithful, design-§9-documented scoping: yaml.v3 re-renders
unrelated scalars (drops quotes on `harness_version`) and normalizes indentation, so whole-file
byte-equality is neither expected nor asserted. The check correctly proves what matters: the
`models:` block's scalar VALUES (`claude-sonnet-4`, `gpt-4o`) survive a round-trip unchanged, no
phantom `leader:` appears, and no phantom `models:` block is invented for a config without one.

---

## LOW-severity notes (record, do not block)

- **LOW-1 — "ByteIdentical" test name vs. what it asserts.** `TestConfig_FlatStringRoundtripByteIdentical`
  does NOT compare against the literal original on-disk bytes (the fixture uses 2-space indent;
  the assertion `wantModels` uses yaml.v3's 4-space `phases.apply` indent). It verifies scalar-form
  preservation across a re-marshal, which is the real (and correct) guarantee per design §9. Consider
  renaming to `...ScalarFormPreserved` in a later slice, or add a one-line comment that indentation is
  intentionally yaml.v3-normalized. No functional impact.
- **LOW-2 — Empty-provider + slashed model is not round-trip-stable in its structured decomposition.**
  `ModelRef{Provider:"", Model:"xai/grok-4"}` marshals to scalar `"xai/grok-4"`, which unmarshals to
  `{Provider:"xai", Model:"grok-4"}`. `FullID()` is preserved both ways (`xai/grok-4`), so no
  user-visible/contract regression, and this state is not producible via `ParseModelRef`
  (`set ... xai/grok-4` yields `{xai, grok-4}` directly). Pure internal-shape asymmetry; matches the
  documented split-on-first-`/` rule. Worth a comment near `MarshalYAML`/`UnmarshalYAML` only.
- **LOW-3 — `Model.Reasoning` captured but unused this slice.** As the design flagged (intentional,
  cheap insurance for S2/S3). No action needed.

---

## Definition of Done — confirmed

- [x] Legacy flat-string `config.yaml` loads and re-saves with its `models:` block scalar-preserved.
- [x] `go build ./...` and `go vet ./...` exit 0.
- [x] `go test ./...` exits 0 (all 22 scenarios covered by passing tests).
- [x] status / init / TUI compile and render unchanged via `FullID()` / `ParseModelRef()`.
- [x] PR #45 shell-out path preserved and reachable (`parseModels`/`execLister` present; tests pass).
- [x] `go clean -testcache && go test ./...` green on a clean build.
