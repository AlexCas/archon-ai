# Tasks: structured-model-resolution — Foundation PR (S0 + S1)

**Change**: `structured-model-resolution`
**Scope**: Foundation PR only (S0 cache reader + S1 ModelRef type, ModelConfig swap, resolve
repoint, consumer compile-fixes).
**Branch**: `feat/structured-models-foundation` (off `master` @ 7360b4d).
**Design**: `openspec/changes/structured-model-resolution/design.md` — follow it exactly.
**Specs**: `specs/model-ref/spec.md` (M1–M6) · `specs/opencode-model-cache/spec.md` (C1–C4).

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~590 (prod + tests); prod-only ~300–330 |
| 400-line budget | Exceeded (forecast accepted by user per C1) |
| Chained PRs recommended | Optional intra-foundation split: S0 alone (~210 LOC) then S1 stacked; otherwise single Foundation PR |
| Chain strategy | single Foundation PR (agreed) |

---

## Group S0 — `internal/opencode` cache reader package

> Pure addition. No archon package imports it yet (except resolve.go in S1).
> Complete all S0 tasks and pass `go test ./internal/opencode/...` before starting S1.

- [x] **S0-1** Create `internal/opencode/testdata/models.json` — well-formed fixture: an `opencode`
  provider with ≥ 2 bare-key models (at least one with `tool_call: true`) AND one other provider
  (e.g. `requesty`) with at least one already-slashed key (`xai/grok-4`). This fixture is reused
  by `models_test.go` and `internal/models/resolve_test.go`. No Go code yet.
  _(Satisfies: C2 happy-path anchor)_

- [x] **S0-2** Create `internal/opencode/testdata/malformed.json` — valid top-level JSON object with
  one well-formed provider entry (copy the `opencode` provider from S0-1) and one entry whose
  value is the wrong JSON type (e.g. `"broken": 42`). Used by the skip-malformed-entry test.
  _(Satisfies: C2 malformed-skipped scenario)_

- [x] **S0-3** Create `internal/opencode/testdata/invalid.json` — truncated/invalid JSON (e.g. `{`).
  Used by the parse-error propagation test.
  _(Satisfies: C3 parse-error scenario)_

- [x] **S0-4** Create `internal/opencode/models.go` (new package `opencode`) with:
  - `Model` struct (`ID`, `Name`, `ToolCall bool`, `Reasoning bool`; JSON tags matching the cache).
  - `Provider` struct (`ID`, `Name`, `Models map[string]Model`; JSON tags).
  - `DefaultCachePath() (string, error)` — returns `<home>/.cache/opencode/models.json`; returns
    the error from `os.UserHomeDir()` rather than swallowing it.
  - `LoadModels(path string) (map[string]Provider, error)` — reads and parses the provider-keyed
    JSON; skips malformed per-provider entries (continue on unmarshal error); forces `p.ID = id`
    from the map key; returns error on malformed top-level JSON.
  - `LoadModelsOrEmpty(path string) (map[string]Provider, error)` — calls `LoadModels`; returns
    empty map + nil error when `errors.Is(err, os.ErrNotExist)`; propagates all other errors.
  - Imports: `encoding/json`, `errors`, `fmt`, `os`, `path/filepath`.
  _(Satisfies: C1 DefaultCachePath, C2 LoadModels, C3 LoadModelsOrEmpty)_

- [x] **S0-5** Create `internal/opencode/models_test.go` with the following tests (all using testdata
  fixtures from S0-1 – S0-3, never reading `$HOME`):
  - `TestDefaultCachePath` — sets `HOME` via `t.Setenv`; asserts path ends with
    `.cache/opencode/models.json`. _(C1)_
  - `TestLoadModels_WellFormed` — loads `testdata/models.json`; asserts `opencode` provider
    present; bare-key models keyed correctly; `p.ID == map key` for all providers; slashed key
    preserved for the second provider. _(C2 happy)_
  - `TestLoadModels_MalformedEntrySkipped` — loads `testdata/malformed.json`; asserts good
    provider returned, `broken` key absent, nil error returned. _(C2 malformed-skipped)_
  - `TestLoadModelsOrEmpty_Absent` — calls `LoadModelsOrEmpty` with a non-existent path; asserts
    empty map and nil error. _(C3 absent)_
  - `TestLoadModelsOrEmpty_ParseError` — calls `LoadModelsOrEmpty` with `testdata/invalid.json`;
    asserts non-nil error. _(C3 parse-error)_

  Run `go test ./internal/opencode/...` green before continuing.

---

## Group S1a — `ModelRef` type (new type + accessors + custom YAML)

> S1a is a pure addition to `internal/config/model.go`. The existing `ModelConfig` fields are
> still strings at the end of S1a; S1b does the swap. This ordering keeps the config package
> compiling green between S1a and S1b.

- [x] **S1a-1** In `internal/config/model.go`, add:
  - `ModelRef` struct (`Provider`, `Model`, `Effort string`; YAML tags with `omitempty`).
  - `FullID() string` method on `ModelRef` — returns `Model` as-is when it contains `/`;
    returns `Provider + "/" + Model` when Provider is non-empty and Model is bare; returns bare
    `Model` (no leading slash) when Provider is empty.
  - `ParseModelRef(s string) ModelRef` function — splits on the FIRST `/`; bare input yields
    `ModelRef{Model: s}` (empty Provider, no guessing).
  - `UnmarshalYAML(node *yaml.Node) error` on `*ModelRef` — scalar path: splits on first `/` or
    stores bare with empty Provider; mapping path: uses `type modelRefAlias ModelRef` alias to
    decode without recursion.
  - `MarshalYAML() (any, error)` on `ModelRef` — returns the bare `Model` string when
    `Provider == "" && Effort == ""`; returns `modelRefAlias(r)` otherwise.
  - Required imports: `strings`, `gopkg.in/yaml.v3` (already imported).
  _(Satisfies: M1 FullID, M3 UnmarshalYAML, M4 MarshalYAML)_

- [x] **S1a-2** In `internal/config/model_test.go`, add:
  - `TestModelRef_FullID` — table-driven: `anthropic`+`claude-sonnet-4-6` → join; `opencode`+
    `deepseek-v4-pro` → `opencode/deepseek-v4-pro`; `openrouter`+`xai/grok-4` → `xai/grok-4`
    (slash short-circuit); `""`+`opus` → `opus` (no leading slash). _(M1 all 4 scenarios)_
  - `TestParseModelRef` — table-driven: `a/b`→`{a,b}`; `a/b/c`→`{a,"b/c"}`; `opus`→`{"",opus}`.
    _(supports §7)_
  - `TestModelRef_UnmarshalYAML` — table-driven scalar `a/b`→`{a,b}`; bare `x`→`{"",x}`; mapping
    form → structured. _(M3 all 3 scenarios)_
  - `TestModelRef_MarshalYAML` — `{"",opus,""}` marshals to scalar node `opus` (not a mapping);
    `{opencode,deepseek-v4-pro,""}` marshals to a mapping. _(M4)_

  Run `go test ./internal/config/...` green before continuing.

---

## Group S1b — `ModelConfig` field swap + Clone + `ResolvePhaseModels`

> Depends on S1a (ModelRef type must exist). Changes the existing field types in `ModelConfig`,
> updates `Clone`, and rewrites `ResolvePhaseModels`. The package will fail to compile until the
> consumer fixups in S1d are also applied — apply S1b and S1d as one contiguous apply session, or
> apply S1b + immediately run the compile-fix steps of S1d before any `go build` check.

- [x] **S1b-1** In `internal/config/model.go`, change `ModelConfig` fields:
  - `Default string` → `Default ModelRef`
  - `Leader string`  → `Leader ModelRef`
  - `Phases map[string]string` → `Phases map[string]ModelRef`
  - YAML tags (`default`, `leader`, `phases`, all with `omitempty`) remain unchanged.
  _(Satisfies: M2 structured fields)_

- [x] **S1b-2** In `internal/config/config.go` `Clone()`, change only the map literal type:
  - `make(map[string]string, len(c.Models.Phases))` → `make(map[string]ModelRef, len(c.Models.Phases))`
  - The struct literal `ModelConfig{Default: c.Models.Default, Leader: c.Models.Leader, ...}`
    and the copy loop `clone.Models.Phases[k] = v` are otherwise unchanged (ModelRef is a value
    type; assignment deep-copies it).
  _(Satisfies: M5 Clone deep-copies structured models)_

- [x] **S1b-3** In `internal/config/model.go`, rewrite `ResolvePhaseModels` to:
  - Iterate `PhaseOrder` (unchanged).
  - Prefer `mc.Phases[p]` when `ref.Model != ""`; fall back to `mc.Default` when not.
  - Omit the phase when `ref.Model == ""` (both phase + default are empty).
  - Emit `PhaseModel{Phase: p, Model: ref.FullID()}`.
  - Remove the `NormalizeModel` calls from this function (NormalizeModel itself is retained for
    `Validate`; do not delete it).
  _(Satisfies: M6 ResolvePhaseModels emits FullID)_

- [x] **S1b-4** In `internal/config/model_test.go`, update/add:
  - Rewrite `TestResolvePhaseModels` — build `ModelRef`s instead of strings; assert
    `{opencode,deepseek-v4-pro}` → `opencode/deepseek-v4-pro`; `{"",opus}` → `opus`; default
    fallback; omit-when-empty; PhaseOrder iteration order. Run twice and assert deep-equal
    (determinism). _(M6)_
  - Add `TestModelConfig_StructuredFields` — assign provider-qualified refs to Default, Leader,
    Phases; assert fields preserve Provider+Model. _(M2)_
  - `TestNormalizeModel` and `TestValidate` are UNCHANGED — retain verbatim. _(M6 retention check)_

---

## Group S1c — `config_test.go` back-compat + byte-identity tests

> Depends on S1b (ModelConfig fields must be ModelRef). Adds and updates tests in
> `internal/config/config_test.go`.

- [x] **S1c-1** Update `TestConfig_Load` fixture assertions in `internal/config/config_test.go`:
  - Where existing test asserts `Models.Default == "claude-sonnet-4"`, change to
    `Models.Default == config.ModelRef{Model: "claude-sonnet-4"}`.
  - Where existing test asserts `Models.Phases["apply"] == "gpt-4o"`, change to
    `Models.Phases["apply"] == config.ModelRef{Model: "gpt-4o"}`.
  - Add a mapping-form fixture (`default: {provider: opencode, model: deepseek-v4-pro}`) and
    assert it decodes to `ModelRef{Provider: "opencode", Model: "deepseek-v4-pro"}`.
  _(Satisfies: M2, M3 load scenarios)_

- [x] **S1c-2** Add `TestConfig_FlatStringRoundtripByteIdentical` in `internal/config/config_test.go`:
  - Write the legacy flat-string fixture bytes to a temp file
    (e.g. `default: claude-sonnet-4\nphases:\n  apply: gpt-4o\n`).
  - `Load` the file, then `yaml.Marshal` the `Models` block.
  - Assert the emitted `models:` block bytes are equal to the corresponding block in the input
    (compare ONLY the models block, not whole-file bytes — yaml.v3 re-renders unrelated scalars
    such as `harness_version` and this is pre-existing Save behavior).
  - Assert that an empty-leader config re-saves WITHOUT a `leader:` key.
  - Assert a config with NO `models:` key saves without inventing a `models:` block.
  _(Satisfies: M4 byte-identical round-trip; dominant back-compat guard)_

- [x] **S1c-3** Update `TestConfig_CloneRoundtrip` in `internal/config/config_test.go`:
  - Change the fixture `Default`/`Leader` assignments from plain strings to
    `config.ModelRef{Provider: "anthropic", Model: "claude-sonnet-4-20250514"}` etc.
  - Change the `Phases` map literal to `map[string]config.ModelRef`.
  - Change the mutate-clone assertion at `:223` to
    `clone.Models.Phases["apply"] = config.ModelRef{Model: "MUTATED"}`.
  - Verify the clone equals the original and the original's Phases map is unaffected.
  _(Satisfies: M5 Clone round-trip equality)_

- [x] **S1c-4** Update `TestConfig_Roundtrip` in `internal/config/config_test.go`:
  - Change `Default`/`Leader`/`Phases` literals from strings to `ModelRef`.
  - The `loaded.Models.Leader != original.Models.Leader` comparison works because `ModelRef`
    is a comparable value type.
  _(Satisfies: M2 structured fields round-trip)_

  Run `go test ./internal/config/...` green (all back-compat tests pass) before continuing.

---

## Group S1d — `internal/models/resolve.go` cache-first repoint + tests

> Depends on S0 (opencode package) and S1a (ModelRef + ParseModelRef). The `CacheReader`
> type and injected seam must land before the consumer compile-fixes in S1e.

- [x] **S1d-1** In `internal/models/resolve.go`, add:
  - `CacheReader` type: `type CacheReader func() (map[string]opencode.Provider, error)`.
  - `defaultCacheReader()` private function — calls `opencode.DefaultCachePath()` then
    `opencode.LoadModelsOrEmpty(path)`.
  - `cacheModelNames(cache CacheReader) []string` private function — calls `cache()`; returns
    nil when err != nil or map is empty; wraps each model key in
    `config.ModelRef{Provider: provID, Model: key}` and calls `FullID()`; returns sorted output.
  - Update `ResolveModels` signature:
    `func ResolveModels(detect CLIDetector, lister OpencodeLister, cache CacheReader) []string`
  - In the `opencode` branch: try `cacheModelNames(cache)` first; fall back to `lister.List`
    (shell-out, PR #45 path) when cache names is nil/empty; fall back to `config.OpencodeModels`
    (curated) when shell-out also fails/empty.
  - Update `Resolve()` to pass `defaultCacheReader`:
    `return ResolveModels(LookPathDetector, execLister{}, defaultCacheReader)`.
  - Add imports: `sort`, `github.com/archon-ai/archon/internal/opencode`.
  - Retain `parseModels`, `execLister`, and existing curated-fallback logic unchanged (PR #45
    preserved).
  _(Satisfies: C4 cache-first resolution, C4 fallback, C4 PR #45 preserved)_

- [x] **S1d-2** Update `internal/models/resolve_test.go` — all existing `TestResolveModels_*` tests:
  - Add the third `cache CacheReader` argument: pass `func() (map[string]opencode.Provider, error) { return nil, nil }`
    (empty-cache stub) so the shell-out / curated behavior is exactly preserved.
  _(Existing test behavior unchanged; compile fix only)_

- [x] **S1d-3** Add new tests in `internal/models/resolve_test.go`:
  - `TestResolveModels_PrefersCache` — inject a `CacheReader` that returns an `opencode`
    provider with bare keys + a sentinel lister that records whether it was called; assert
    returned names are `opencode/<key>` FullID form AND the lister was NOT invoked. _(C4 prefers)_
  - `TestResolveModels_FallbackWhenCacheEmpty` — inject a `CacheReader` returning empty map +
    a lister returning known live names; assert live names used (shell-out fallback). _(C4 fallback)_
  - `TestCacheModelNames_Mapping` (inline in resolve_test or separate) — bare opencode key →
    `opencode/<key>`; slashed key from another provider → as-is; output sorted. _(C4 FullID form)_
  - `TestParseModels` in `internal/models/opencode_test.go` — UNCHANGED (retain verbatim; guards
    that `parseModels`/`execLister` are still present and PR #45 path is reachable). _(C4 PR #45)_

  Run `go test ./internal/models/...` green before continuing.

---

## Group S1e — Consumer compile-fixes (CLI, init, status, TUI)

> Depends on S1b (ModelConfig fields are ModelRef). These files fail to compile once S1b lands.
> Apply all five sub-tasks as one apply session; `go build ./...` must be green after S1e-5.

- [x] **S1e-1** Update `cmd/archon/config.go`:
  - In `setConfigValue` (`:155`): `cfg.Models.Default = value` →
    `cfg.Models.Default = config.ParseModelRef(value)`.
  - In `setConfigValue` (`:184`): `make(map[string]string)` →
    `make(map[string]config.ModelRef)`.
  - In `setConfigValue` (`:186`): `cfg.Models.Phases[phase] = value` →
    `cfg.Models.Phases[phase] = config.ParseModelRef(value)`.
  - In `getConfigValue` (`:196`): `return cfg.Models.Default, nil` →
    `return cfg.Models.Default.FullID(), nil`.
  - In `getConfigValue` (`:211`): `return cfg.Models.Phases[phase], nil` →
    `return cfg.Models.Phases[phase].FullID(), nil`.
  - In `newConfigListCmd` (`:119–135`): replace every direct use of `cfg.Models.Default` and
    `cfg.Models.Phases[phase]` as strings with `.FullID()` calls; replace the `== ""`/`!= ""`
    comparisons with `.FullID() == ""`/`.FullID() != ""`.
  _(Supports §7a)_

- [x] **S1e-2** Update `internal/initcmd/init.go` `buildConfig`:
  - `var phases map[string]string` → `var phases map[string]config.ModelRef`.
  - `make(map[string]string)` → `make(map[string]config.ModelRef)`.
  - `phases[k] = v` → `phases[k] = config.ParseModelRef(v)`.
  - `Default: modelDefault` → `Default: config.ParseModelRef(modelDefault)`.
  - `Leader: modelLeader` → `Leader: config.ParseModelRef(modelLeader)`.
  - `Options` struct and `main.go` flag wiring remain UNCHANGED (string params; split happens
    in `buildConfig` only).
  - `init.go:102` `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader)` →
    `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader.FullID())`.
  _(Supports §7b)_

- [x] **S1e-3** Update `internal/status/display.go` (`:53–71`):
  - `:55` `cfg.Models.Default == ""` → `cfg.Models.Default.FullID() == ""`
  - `:58` `cfg.Models.Default != ""` → `cfg.Models.Default.FullID() != ""`
  - `:59` `..., cfg.Models.Default` → `..., cfg.Models.Default.FullID()`
  - `:68` `..., cfg.Models.Phases[phase]` → `..., cfg.Models.Phases[phase].FullID()`
  _(Supports §7c)_

- [x] **S1e-4** Update `internal/tui/models_tab.go`:
  - `:50` `newModelInput("Default model", cfg.Models.Default)` →
    `newModelInput("Default model", cfg.Models.Default.FullID())`
  - `:57` `value = cfg.Models.Phases[phase]` → `value = cfg.Models.Phases[phase].FullID()`
  - `:69` `newModelInput("Leader model", cfg.Models.Leader)` →
    `newModelInput("Leader model", cfg.Models.Leader.FullID())`
  - `:283` `cfg.Models.Default = m.inputs[...].Value()` →
    `cfg.Models.Default = config.ParseModelRef(m.inputs[...].Value())`
  - `:285–286` `make(map[string]string)` → `make(map[string]config.ModelRef)`
  - `:293` `cfg.Models.Phases[phase] = value` →
    `cfg.Models.Phases[phase] = config.ParseModelRef(value)`
  - `:302` `cfg.Models.Leader = m.inputs[idx].Value()` →
    `cfg.Models.Leader = config.ParseModelRef(m.inputs[idx].Value())`
  - Add `"github.com/archon-ai/archon/internal/config"` import if not already present.
  _(Supports §7c)_

- [x] **S1e-5** Update `internal/tui/model.go`:
  - `:334` `MergeOpencodeAgent(m.projectDir, cfg.Models.Leader)` →
    `MergeOpencodeAgent(m.projectDir, cfg.Models.Leader.FullID())`
  _(Supports §7c)_

---

## Group S1f — Optional: `config_test.go` provider-qualified CLI test

> Low-risk, optional. Add only if the test is missing and would prevent catching a set/get
> regression for provider-qualified models.

- [x] **S1f-1** (OPTIONAL) In `cmd/archon/config_test.go`, add
  `TestConfigCmd_SetProviderQualified`: call `config set models.default opencode/deepseek-v4-pro`
  then `config get models.default`; assert the returned value is `opencode/deepseek-v4-pro`.
  _(Confirms ParseModelRef + FullID seam end-to-end; existing bare-model tests are unchanged)_

---

## Group S1g — Full verification gates

> Run after ALL S0 + S1a–S1e tasks are applied. All commands must exit 0.

- [x] **S1g-1** `go build ./...` — zero errors. Full compile check confirms all consume-fixup
  seams are wired. _(M2, M5, M6 compile-fix gate)_

- [x] **S1g-2** `go vet ./...` — zero warnings. _(code quality gate)_

- [x] **S1g-3** `go test ./internal/opencode/...` — five cache-reader tests green. _(C1–C3)_

- [x] **S1g-4** `go test ./internal/config/...` — all ModelRef, ModelConfig, round-trip, Clone,
  and ResolvePhaseModels tests green. _(M1–M6)_

- [x] **S1g-5** `go test ./internal/models/...` — all resolve tests green, including the three
  new cache-first/fallback/PR-#45 tests; `TestParseModels` still passes. _(C4)_

- [x] **S1g-6** `go test ./cmd/archon/...` `go test ./internal/initcmd/...`
  `go test ./internal/status/...` `go test ./internal/tui/...` — all consumer tests green.
  _(compile-fix correctness; M1 FullID rendering)_

- [x] **S1g-7** `go test ./...` — full suite green as final gate.

---

## Definition of Done

These criteria must ALL be met before the Foundation PR is opened:

- [x] Existing flat-string `config.yaml` (e.g. `default: claude-sonnet-4`, `phases.apply: gpt-4o`)
  loads and re-saves with its `models:` block byte-identical (`TestConfig_FlatStringRoundtripByteIdentical`
  passes).
- [x] `go build ./...` and `go vet ./...` both exit 0.
- [x] `go test ./...` exits 0 (all 22+ spec scenarios covered by passing tests).
- [x] status / init / TUI still compile and render correctly — every consumer reads a string via
  `FullID()`, every writer converts with `ParseModelRef()`.
- [x] PR #45 shell-out detection path is preserved and still reachable (`TestParseModels` +
  `TestResolveModels_FallbackWhenCacheEmpty` pass; `parseModels`/`execLister` not deleted).
- [x] `go test ./...` is run a final time on a clean build (`go clean -testcache && go test ./...`).
