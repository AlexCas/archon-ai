# Design — structured-model-resolution (Foundation PR = S0 + S1 combined)

**Change**: `structured-model-resolution`
**Scope**: Foundation PR only (cache reader `internal/opencode` + `ModelRef` type + back-compat
config (Un)Marshal + repointed resolution + CLI/init/consumer compile-fixes).
**Branch**: `feat/structured-models-foundation` (off `master` @ 7360b4d).
**SETTLED inputs (implement exactly, do not re-open)**:
- `specs/model-ref/spec.md` — M1 FullID, M2 fields, M3 Unmarshal, M4 Marshal, M5 Clone, M6 ResolvePhaseModels.
- `specs/opencode-model-cache/spec.md` — C1 DefaultCachePath, C2 LoadModels, C3 LoadModelsOrEmpty, C4 cache-first resolution.

YAML library in use: **`gopkg.in/yaml.v3`** (go.mod:11; `config.go:10`). Custom (un)marshal therefore
uses `*yaml.Node` / `yaml.Marshaler` — NOT `sigs.k8s.io/yaml`.

JSON library for the cache reader: stdlib **`encoding/json`** (matches gentle-ai).

---

## Technical Approach

Two pure additions land first, then the type change ripples through every consumer behind a single
string accessor `FullID()`:

1. **S0** adds `internal/opencode/models.go` — a self-contained, caller-free reader of opencode's
   provider-keyed cache `~/.cache/opencode/models.json`. No archon package imports it yet except
   `internal/models/resolve.go` in step C4.
2. **S1** turns `config.ModelConfig`'s three flat-string fields into `ModelRef` /
   `map[string]ModelRef`, with custom `UnmarshalYAML`/`MarshalYAML` that make an unmigrated
   flat-string `config.yaml` load and re-save **byte-identical**. `FullID()` is the one string
   accessor every downstream consumer (status, TUI, init, CLI, resolution) calls so the codebase
   keeps compiling and rendering unchanged.

Back-compat is the dominant constraint: an empty `Provider` is a first-class advisory-only state;
we NEVER guess a provider for a bare legacy alias.

---

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|---|---|---|---|
| Cache structs | Minimal `Provider{ID,Name,Models}` + `Model{ID,Name,ToolCall,Reasoning}` | Full gentle-ai struct (Cost/Limit/Family/Variants/Env) | Foundation needs only id+name to build FullIDs; extra fields are dead weight this slice. `json` ignores unknown keys, so the 20+ cache fields parse fine. |
| FullID source of truth | `Provider` field + bare `Model`; split on FIRST `/` only when `Model` already slashed | Re-deriving provider from a resolver/alias map | Matches the verified KEY ASYMMETRY (opencode keys bare → join; other-provider keys already slashed → as-is). Never double-prefix. |
| Empty-provider state | Valid everywhere; `FullID()` → bare `Model`; `MarshalYAML` → scalar | Reject / coerce / invent a provider | Byte-identical round-trip for unmigrated configs; "no silent provider invention". |
| (Un)Marshal recursion | `type alias ModelRef` (a.k.a. `modelRefAlias`) inside Unmarshal mapping decode; `MarshalYAML` returns a plain value, never calls itself | Marshal/Unmarshal the same type (infinite recursion) | yaml.v3 re-invokes the custom method on the same type; aliasing strips the methods. |
| Resolution catalog source | Cache FIRST (FullIDs), shell-out fallback when cache absent/empty | Delete `parseModels`/`execLister` | Preserves PR #45; "augment the catalog source, not replace it." |
| `NormalizeModel` fate | Kept for `Validate` (offline advisory); removed from the resolution path | Delete it | Spec M6 mandates retention; `Validate` + its tests still use it. |

---

## 1. S0 — `internal/opencode/models.go` (NEW package)

New file: `internal/opencode/models.go`. Package `opencode`. Imports:
`encoding/json`, `errors`, `fmt`, `os`, `path/filepath`.

### Structs (minimal subset, JSON tags match the real cache)

Verified against `~/.cache/opencode/models.json`: each provider object has
`id`,`name`,`models` (plus `env`,`npm`,`api`,`doc` we ignore); each model object has
`id`,`name`,`tool_call`,`reasoning` (plus family/cost/limit/etc. we ignore). `encoding/json`
silently drops unknown keys, so the minimal structs parse the full cache.

```go
// Model is one model entry within a provider. Only the fields the foundation
// needs are captured; the cache carries many more (family/cost/limit/…) which
// encoding/json silently ignores.
type Model struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    ToolCall  bool   `json:"tool_call"`
    Reasoning bool   `json:"reasoning"`
}

// Provider is one provider in the opencode cache. Models are keyed by the
// cache's model key: BARE under the "opencode" provider (e.g. "deepseek-v4-pro")
// and ALREADY-SLASHED under other providers (e.g. "xai/grok-4").
type Provider struct {
    ID     string           `json:"id"`
    Name   string           `json:"name"`
    Models map[string]Model `json:"models"`
}
```

### Functions (full signatures)

```go
// DefaultCachePath returns ~/.cache/opencode/models.json for the current user.
// Unlike gentle-ai's version it RETURNS THE ERROR (archon convention) rather than
// swallowing it to "".
func DefaultCachePath() (string, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".cache", "opencode", "models.json"), nil
}

// LoadModels parses the provider-keyed cache into map[providerID]Provider.
// A malformed/partial provider ENTRY is skipped (not fatal); a malformed
// top-level JSON document IS an error. The map key is forced onto p.ID so the
// provider id is authoritative even if the inner "id" is missing/wrong.
func LoadModels(path string) (map[string]Provider, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read models cache %q: %w", path, err)
    }
    var raw map[string]json.RawMessage
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, fmt.Errorf("parse models cache: %w", err)
    }
    providers := make(map[string]Provider, len(raw))
    for id, pjson := range raw {
        var p Provider
        if err := json.Unmarshal(pjson, &p); err != nil {
            continue // skip malformed entry (C2)
        }
        p.ID = id
        providers[id] = p
    }
    return providers, nil
}

// LoadModelsOrEmpty returns an empty map and nil error when the cache file is
// ABSENT; any other read/parse error from LoadModels propagates.
func LoadModelsOrEmpty(path string) (map[string]Provider, error) {
    providers, err := LoadModels(path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return map[string]Provider{}, nil
        }
        return nil, err
    }
    return providers, nil
}
```

Notes:
- `os.ReadFile` returns an error wrapping `os.ErrNotExist` for a missing file; `LoadModels`
  wraps it with `%w`, so `errors.Is(err, os.ErrNotExist)` in `LoadModelsOrEmpty` still matches (C3).
- A document that is valid JSON but not an object (e.g. `[]` / `"x"`) fails the
  `map[string]json.RawMessage` unmarshal → parse error propagates (C3 "Parse error propagates").
- Do NOT port `FixOpenRouterModels`, `DetectAvailableProviders`, `FilterModelsForSDD`,
  `MergeCustomProviders`, `EnrichWithVariants`, variants, auth/settings paths — deferred to S2/S4.
- Note `Reasoning` is captured but unused this slice; it is cheap insurance for the S2 writers and
  documents the cache shape. (If apply prefers strictly-minimal, dropping it is acceptable — keep
  `ToolCall` which S2/S3 will need for tool-capable filtering.)

### testdata fixtures

`internal/opencode/testdata/` already exists (empty). Add:
- `models.json` — well-formed: an `opencode` provider with ≥2 bare-key models (one with
  `tool_call:true`) AND one other provider (e.g. `requesty`) with ≥1 already-slashed key
  (`xai/grok-4`). Used by the happy-path load test and reused by `models/resolve_test.go`.
- `malformed.json` — a valid top-level object with one good provider and one entry whose value is
  the wrong JSON type (e.g. `"opencode": {...valid...}, "broken": 42`) so the broken entry is
  skipped while the good one loads.
- `invalid.json` — not valid JSON (`{` truncated) for the parse-error test.

---

## 2. `config.ModelRef` (M1–M4) — in `internal/config/model.go`

### Struct (yaml tags)

```go
// ModelRef is a structured provider+model assignment. An empty Provider is a
// valid advisory-only state (FullID returns the bare Model); a bare legacy alias
// decodes to this state and re-marshals as the same scalar.
type ModelRef struct {
    Provider string `yaml:"provider,omitempty"`
    Model    string `yaml:"model,omitempty"`
    Effort   string `yaml:"effort,omitempty"`
}
```

### FullID (M1)

```go
// FullID returns the provider-qualified id used by delegations:
//   - Model already contains "/" -> returned as-is (no double-prefix).
//   - non-empty Provider + bare Model -> "<provider>/<model>".
//   - empty Provider -> the bare Model (no leading slash).
func (r ModelRef) FullID() string {
    if strings.Contains(r.Model, "/") {
        return r.Model
    }
    if r.Provider == "" {
        return r.Model
    }
    return r.Provider + "/" + r.Model
}
```

Scenario coverage: `anthropic`+`claude-sonnet-4-6` → `anthropic/claude-sonnet-4-6`;
`opencode`+`deepseek-v4-pro` → `opencode/deepseek-v4-pro`; `openrouter`+`xai/grok-4` →
`xai/grok-4` (slash short-circuit, no double-prefix); `""`+`opus` → `opus`. The slash check is
**first** so an already-slashed Model wins even when Provider is set (matches the M1 edge scenario).

### UnmarshalYAML (M3)

```go
func (r *ModelRef) UnmarshalYAML(node *yaml.Node) error {
    if node.Kind == yaml.ScalarNode {
        s := node.Value
        if i := strings.Index(s, "/"); i >= 0 { // split on the FIRST "/"
            r.Provider = s[:i]
            r.Model = s[i+1:]
            return nil
        }
        r.Provider = ""    // bare alias -> empty provider, NEVER guessed
        r.Model = s
        return nil
    }
    type modelRefAlias ModelRef          // strips the custom method -> no recursion
    var tmp modelRefAlias
    if err := node.Decode(&tmp); err != nil {
        return err
    }
    *r = ModelRef(tmp)
    return nil
}
```

Scenarios: `"anthropic/claude-sonnet-4-20250514"` → `{anthropic, claude-sonnet-4-20250514}`;
`"opus"` → `{"", opus}`; mapping `{provider: opencode, model: deepseek-v4-pro}` → structured.
Split on the FIRST `/` mirrors `FullID`'s slash semantics (a slashed model under a real provider
round-trips: scalar `a/b/c` → `{a, b/c}` → `FullID` `a/b/c`).

### MarshalYAML (M4)

```go
func (r ModelRef) MarshalYAML() (any, error) {
    if r.Provider == "" && r.Effort == "" {
        return r.Model, nil    // SCALAR — byte-identical for unmigrated configs
    }
    type modelRefAlias ModelRef          // mapping, no recursion
    return modelRefAlias(r), nil
}
```

Scenarios: `{"", opus, ""}` → scalar `opus` (not a mapping); a ref with a provider → mapping with
`provider`/`model` (and `effort` only if set, via `omitempty`). Returning a value (not a
`*yaml.Node`) is sufficient for yaml.v3; the alias prevents the marshaler re-entering itself.

**Byte-identical guarantee**: the unmigrated fixture (`default: claude-sonnet-4`,
`phases: {apply: gpt-4o}`) decodes every value to `{Provider:"", Effort:""}`, so every value
re-marshals as the original scalar. `omitempty` on `ModelRef`/map fields preserves field presence
(see edge case in §9).

---

## 3. `config.ModelConfig` change (M2) — `internal/config/model.go:9-13`

**Before:**
```go
type ModelConfig struct {
    Default string            `yaml:"default,omitempty"`
    Leader  string            `yaml:"leader,omitempty"`
    Phases  map[string]string `yaml:"phases,omitempty"`
}
```
**After:**
```go
type ModelConfig struct {
    Default ModelRef            `yaml:"default,omitempty"`
    Leader  ModelRef            `yaml:"leader,omitempty"`
    Phases  map[string]ModelRef `yaml:"phases,omitempty"`
}
```
yaml tags stay `default`/`leader`/`phases`. (See §9 for the `omitempty`-on-struct caveat — an empty
`ModelRef{}` is NOT omitted by yaml.v3, but the existing fixtures always set Default, and Leader was
already frequently empty/omitted as a plain string; behavior for a fully-empty models block is
covered by the round-trip test.)

---

## 4. `config.Clone` (M5) — `internal/config/config.go:96-101`

**Before:**
```go
Models: ModelConfig{Default: c.Models.Default, Leader: c.Models.Leader, Phases: make(map[string]string, len(c.Models.Phases))},
SkillInventory:  make([]SkillInventory, len(c.SkillInventory)),
}
for k, v := range c.Models.Phases {
    clone.Models.Phases[k] = v
}
```
**After:**
```go
Models: ModelConfig{Default: c.Models.Default, Leader: c.Models.Leader, Phases: make(map[string]ModelRef, len(c.Models.Phases))},
SkillInventory:  make([]SkillInventory, len(c.SkillInventory)),
}
for k, v := range c.Models.Phases {
    clone.Models.Phases[k] = v
}
```
`ModelRef` is a value type with no reference fields, so assignment deep-copies it; only the map
literal type changes (`map[string]string` → `map[string]ModelRef`). The copy loop is otherwise
unchanged. `TestConfig_CloneRoundtrip` guards this.

---

## 5. `ResolvePhaseModels` (M6) — `internal/config/model.go:179-192`

**Before:**
```go
func ResolvePhaseModels(mc ModelConfig) []PhaseModel {
    var out []PhaseModel
    for _, p := range PhaseOrder {
        id, ok := NormalizeModel(mc.Phases[p])
        if !ok {
            id, ok = NormalizeModel(mc.Default)
        }
        if !ok {
            continue
        }
        out = append(out, PhaseModel{Phase: p, Model: id})
    }
    return out
}
```
**After:**
```go
// ResolvePhaseModels returns phase->model pairs in canonical PhaseOrder. For each
// phase it prefers the explicit Phases entry, falls back to Default, and omits the
// phase when neither yields a model. The emitted Model is ref.FullID():
// "<provider>/<model>" when a provider is present, else the bare alias.
func ResolvePhaseModels(mc ModelConfig) []PhaseModel {
    var out []PhaseModel
    for _, p := range PhaseOrder {
        ref, ok := mc.Phases[p]
        if !ok || ref.Model == "" {
            ref = mc.Default
        }
        if ref.Model == "" {
            continue
        }
        out = append(out, PhaseModel{Phase: p, Model: ref.FullID()})
    }
    return out
}
```
- Resolution NO LONGER calls `NormalizeModel` (M6: removed from resolution path). The selection
  criterion is now `ref.Model != ""` (a ref with a model is a real assignment; alias normalization
  no longer gates it).
- `PhaseModel{Phase,Model string}` shape unchanged.
- Scenarios: phase ref `{opencode, deepseek-v4-pro}` → `opencode/deepseek-v4-pro`; phase ref
  `{"", opus}` → `opus`; unset phase falls back to Default; empty Default + empty phase → omitted;
  deterministic via `PhaseOrder` iteration (unchanged).
- **Behavioral note for the existing `model_test.go` TestResolvePhaseModels**: those cases used
  display strings like `"Sonnet 4.6"` and expected the normalized alias `"sonnet"`. After M6 the
  function returns `ref.FullID()` (the raw `Model`, no normalization). Those existing cases MUST be
  rewritten to construct `ModelRef`s and expect FullID output (see Test plan §8).
- `NormalizeModel` (`:148-173`), `providerFamilies`, the curated lists, and `Validate`
  (`:194-205`) are UNCHANGED and still consume strings; `Validate` is still called from the CLI/init
  with the user-supplied `provider/model` or bare string before it is split into a `ModelRef`.

---

## 6. `internal/models/resolve.go` repoint (C4)

Cache becomes the PRIMARY opencode catalog source; shell-out is the fallback. `parseModels` and
`execLister` are RETAINED (PR #45). `Resolve() []string` signature unchanged; cache-sourced names
are provider-qualified FullIDs.

Add a small injected reader seam so the cache path is unit-testable without touching `$HOME`
(mirrors the existing `CLIDetector`/`OpencodeLister` injection style):

```go
// CacheReader loads the opencode provider cache. The real implementation reads
// the default cache file; tests inject a fake returning canned providers.
type CacheReader func() (map[string]opencode.Provider, error)

// defaultCacheReader reads ~/.cache/opencode/models.json, empty-on-missing.
func defaultCacheReader() (map[string]opencode.Provider, error) {
    path, err := opencode.DefaultCachePath()
    if err != nil {
        return nil, err
    }
    return opencode.LoadModelsOrEmpty(path)
}
```

**Before** (`:24-44`): `ResolveModels(detect CLIDetector, lister OpencodeLister) []string` — opencode
branch tries `lister.List`, else curated.

**After**: extend the signature with the reader and prefer the cache:
```go
func ResolveModels(detect CLIDetector, lister OpencodeLister, cache CacheReader) []string {
    present := detect()
    var out []string

    if present["claude"] {
        out = append(out, config.ClaudeModels...)
    }

    if present["opencode"] {
        if names := cacheModelNames(cache); len(names) > 0 {
            out = append(out, names...)            // cache FIRST, FullID form
        } else {
            ctx, cancel := context.WithTimeout(context.Background(), listTimeout)
            defer cancel()
            if live, err := lister.List(ctx); err == nil && len(live) > 0 {
                out = append(out, live...)          // PR #45 shell-out fallback
            } else {
                out = append(out, config.OpencodeModels...) // curated fallback
            }
        }
    }
    return out
}

// cacheModelNames flattens the opencode cache into FullID strings. Returns nil
// when the cache is absent/empty/errored (caller falls back to the shell-out).
func cacheModelNames(cache CacheReader) []string {
    providers, err := cache()
    if err != nil || len(providers) == 0 {
        return nil
    }
    var names []string
    for provID, prov := range providers {
        for key := range prov.Models {
            ref := config.ModelRef{Provider: provID, Model: key}
            names = append(names, ref.FullID())
        }
    }
    sort.Strings(names) // determinism: map iteration is unordered
    return names
}
```
`Resolve()` wires the real reader:
```go
func Resolve() []string {
    return ResolveModels(LookPathDetector, execLister{}, defaultCacheReader)
}
```

Mapping detail: each provider's bare/slashed model key is wrapped in a `ModelRef` and emitted via
`FullID()` — opencode keys become `opencode/<key>`, slashed keys stay as-is. The previous code
appended unsorted curated/live slices in source order; cache output is **sorted** for determinism
(map iteration order is random). `defaultCacheReader` uses `LoadModelsOrEmpty` so a missing cache is
empty (not an error) → `cacheModelNames` returns nil → shell-out fallback runs. The shell-out and
curated-fallback branches are byte-for-byte the prior behavior, so PR #45 is preserved and remains
reachable (C4 "PR #45 detection path stays reachable").

New import in `resolve.go`: `sort` and `github.com/archon-ai/archon/internal/opencode`.

---

## 7. CLI / init seams

A shared helper converts the user's `provider/model` (or bare) string into a `ModelRef`. Place it in
`internal/config/model.go` next to `ModelRef` so both `cmd/archon` and `initcmd` reuse it:

```go
// ParseModelRef splits a user-supplied "provider/model" string into a ModelRef.
// A bare value (no "/") yields an empty Provider (advisory-only). Splitting on the
// FIRST "/" mirrors FullID/UnmarshalYAML, so ParseModelRef and UnmarshalYAML agree.
func ParseModelRef(s string) ModelRef {
    if i := strings.Index(s, "/"); i >= 0 {
        return ModelRef{Provider: s[:i], Model: s[i+1:]}
    }
    return ModelRef{Model: s}
}
```

### 7a. `cmd/archon/config.go`

`setConfigValue` (`:155, :186`) — assign a parsed ref:
- Before: `cfg.Models.Default = value`
- After: `cfg.Models.Default = config.ParseModelRef(value)`
- Before: `cfg.Models.Phases[phase] = value` (with `make(map[string]string)` at `:184`)
- After: `cfg.Models.Phases[phase] = config.ParseModelRef(value)` (and `make(map[string]config.ModelRef)`)

`getConfigValue` (`:196, :211`) — render FullID:
- Before: `return cfg.Models.Default, nil`
- After: `return cfg.Models.Default.FullID(), nil`
- Before: `return cfg.Models.Phases[phase], nil`
- After: `return cfg.Models.Phases[phase].FullID(), nil`

`newConfigListCmd` (`:119-136`) — comparisons and formatting move to FullID:
- `:119` `cfg.Models.Default == ""` → `cfg.Models.Default.FullID() == ""`
- `:124` `cfg.Models.Default != ""` → `cfg.Models.Default.FullID() != ""`
- `:125` `... %s ..., cfg.Models.Default` → `..., cfg.Models.Default.FullID()`
- `:135` `... %s ..., cfg.Models.Phases[phase]` → `..., cfg.Models.Phases[phase].FullID()`

`Validate(value)` at set time (`:51`) is UNCHANGED — it still validates the raw user string before
the split. Existing `config_test.go` fixtures (bare `claude-sonnet-4`, `gpt-4o`) round-trip:
`get` returns `FullID()` = the bare model (no provider) = the same string. So `TestConfigCmd_Get`,
`GetPhase`, `SetRoundtrip` (`o3`), `List`, `ListEmpty` pass unchanged.

### 7b. `cmd/archon/main.go` + `internal/initcmd/init.go`

The init flags carry `provider/model` (or bare) strings. The split happens when populating the
config. Two equivalent placements — choose **(B)** (split in `buildConfig`) to keep `Options` a thin
string DTO and the `main.go` flag wiring untouched:

- `internal/initcmd/init.go` `Options` (`:23-25`) — UNCHANGED (`ModelDefault string`,
  `ModelLeader string`, `ModelPhases map[string]string`). `main.go:151-160` flag wiring UNCHANGED.
- `internal/initcmd/init.go` `buildConfig` (`:207-239`) — split when assembling the ref:

**Before** (`:208-236`):
```go
var phases map[string]string
for k, v := range modelPhases {
    if v != "" {
        if phases == nil {
            phases = make(map[string]string)
        }
        phases[k] = v
    }
}
...
Models: config.ModelConfig{
    Default: modelDefault,
    Leader:  modelLeader,
    Phases:  phases,
},
```
**After:**
```go
var phases map[string]config.ModelRef
for k, v := range modelPhases {
    if v != "" {
        if phases == nil {
            phases = make(map[string]config.ModelRef)
        }
        phases[k] = config.ParseModelRef(v)
    }
}
...
Models: config.ModelConfig{
    Default: config.ParseModelRef(modelDefault),
    Leader:  config.ParseModelRef(modelLeader),
    Phases:  phases,
},
```
`buildConfig`'s Go signature is unchanged (still takes the three string params). `main.go`'s
`Validate` advisory calls (`:133-149`) are UNCHANGED — they validate the raw flag strings; the
leader's `!strings.Contains(modelLeaderFlag, "/")` guard still works on the raw string.

### 7c. Consumer compile-fixes (status, TUI) — MANDATORY in this PR

These read/write `Models.{Default,Leader,Phases}` as strings and will NOT compile once the fields
are `ModelRef`. The spec frames them as "keep compiling/rendering unchanged via `FullID()`"; the
minimal edits below achieve that. They are NOT the S2 writer rewrite — just the type adaptation.

**`internal/status/display.go:55-68`** — render FullID on read:
- `:55` `cfg.Models.Default == ""` → `cfg.Models.Default.FullID() == ""`
- `:58` `cfg.Models.Default != ""` → `cfg.Models.Default.FullID() != ""`
- `:59` `..., cfg.Models.Default` → `..., cfg.Models.Default.FullID()`
- `:68` `..., cfg.Models.Phases[phase]` → `..., cfg.Models.Phases[phase].FullID()`

**`internal/tui/models_tab.go`** — FullID on load, ParseModelRef on save:
- `:50` `newModelInput("Default model", cfg.Models.Default)` → `..., cfg.Models.Default.FullID())`
- `:57` `value = cfg.Models.Phases[phase]` → `value = cfg.Models.Phases[phase].FullID()`
- `:69` `newModelInput("Leader model", cfg.Models.Leader)` → `..., cfg.Models.Leader.FullID())`
- `:283` `cfg.Models.Default = m.inputs[...].Value()` → `cfg.Models.Default = config.ParseModelRef(...Value())`
- `:285-286` `make(map[string]string)` → `make(map[string]config.ModelRef)`
- `:293` `cfg.Models.Phases[phase] = value` → `cfg.Models.Phases[phase] = config.ParseModelRef(value)`
- `:302` `cfg.Models.Leader = m.inputs[idx].Value()` → `cfg.Models.Leader = config.ParseModelRef(...)`
- Check whether `config` is already imported in `models_tab.go`; add the import if not.

**`internal/tui/model.go`** — `:334` `MergeOpencodeAgent(m.projectDir, cfg.Models.Leader)` →
`..., cfg.Models.Leader.FullID())` (MergeOpencodeAgent still takes a string this slice;
`:354` `ResolvePhaseModels(cfg.Models)` is unchanged — same `ModelConfig` arg).

`internal/initcmd/init.go:102` `mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader)` →
`..., cfg.Models.Leader.FullID())` (signature stays `string`; S2 will change the writer itself).

---

## 8. Test plan (22 scenarios → tests)

### `internal/opencode/models_test.go` (NEW) — covers C1, C2, C3
| Test | Scenario(s) | Assert |
|---|---|---|
| `TestDefaultCachePath` | C1 default cache path | returns `<home>/.cache/opencode/models.json`; set `HOME`/`os.UserHomeDir` via `t.Setenv("HOME", ...)` |
| `TestLoadModels_WellFormed` | C2 well-formed | `testdata/models.json` → `opencode` provider present, models keyed by bare key; other provider has slashed key; `p.ID == map key` |
| `TestLoadModels_MalformedEntrySkipped` | C2 malformed skipped | `testdata/malformed.json` → good provider returned, broken entry absent, nil error |
| `TestLoadModelsOrEmpty_Absent` | C3 absent | nonexistent path → empty map, nil error |
| `TestLoadModelsOrEmpty_ParseError` | C3 parse error | `testdata/invalid.json` → non-nil error |

### `internal/config/model_test.go` (MODIFY/ADD) — covers M1, M6
| Test | Scenario(s) | Assert |
|---|---|---|
| `TestModelRef_FullID` (new, table) | M1 ×4 | join; opencode bare; already-slashed no double-prefix; empty provider → bare, no leading `/` |
| `TestResolvePhaseModels` (REWRITE existing) | M6 ×2 + order/determinism | build `ModelRef`s; phase `{opencode,deepseek-v4-pro}` → `opencode/deepseek-v4-pro`; phase `{"",opus}` → `opus`; default fallback; omit-when-empty; PhaseOrder order; twice deep-equal |
| `TestNormalizeModel`, `TestValidate` | (unchanged) | retained verbatim — M6 keeps `NormalizeModel`/`Validate` |
| `TestParseModelRef` (new, table) | (supports 7) | `a/b`→{a,b}; `a/b/c`→{a,`b/c`}; `opus`→{"",opus} |

### `internal/config/config_test.go` (MODIFY) — covers M2, M3, M4, M5
| Test | Scenario(s) | Assert |
|---|---|---|
| `TestConfig_Load` (fixture update) | M2, M3 | existing fixture (`default: claude-sonnet-4`, `phases.apply: gpt-4o`) now asserts `Models.Default == ModelRef{Model:"claude-sonnet-4"}`, etc.; ADD a mapping-form fixture (`default: {provider: opencode, model: deepseek-v4-pro}`) → structured ref |
| `TestModelConfig_StructuredFields` (new) | M2 | assign provider-qualified refs to Default/Leader/Phases; fields preserve provider+model |
| `TestModelRef_UnmarshalYAML` (new, table) | M3 ×3 | scalar `a/b`→{a,b}; bare `x`→{"",x}; mapping→structured |
| `TestModelRef_MarshalYAML` (new, table) | M4 (bare alias) | `{"",opus,""}` marshals to scalar `opus` (assert the emitted node is scalar, not mapping) |
| `TestConfig_FlatStringRoundtripByteIdentical` (NEW, explicit) | M4 happy | write the LEGACY flat fixture bytes; `Load` then `yaml.Marshal`; assert the emitted **`models:` block** equals the input's models block (VERIFIED behavior: empty `Leader` is omitted, `default`/`phases` re-emit as the same bare scalars). Do NOT assert whole-file equality — yaml.v3 re-renders unrelated scalars (e.g. drops quotes on `harness_version`). Dominant back-compat guard. |
| `TestConfig_CloneRoundtrip` (UPDATE fixture) | M5 | change fixture `Default/Leader` from strings to `ModelRef{Provider:..,Model:..}` and `Phases` to `map[string]ModelRef`; the mutate-clone assertion at `:223` becomes `clone.Models.Phases["apply"] = config.ModelRef{Model:"MUTATED"}` |
| `TestConfig_Roundtrip` (UPDATE fixture) | M2 | `Default/Leader/Phases` literals become `ModelRef`; the `loaded.Models.Leader != original.Models.Leader` compare works on the value type (`ModelRef` is comparable) |

### `internal/models/resolve_test.go` (MODIFY) + `opencode_test.go` — covers C4
| Test | Scenario(s) | Assert |
|---|---|---|
| All existing `TestResolveModels_*` | (signature) | add a third arg: a `CacheReader` returning empty (`func() (map[string]opencode.Provider, error){return nil,nil}`) so the shell-out/curated behavior is exactly preserved |
| `TestResolveModels_PrefersCache` (new) | C4 prefers cache | cache reader returns an `opencode` provider with bare keys + a sentinel lister; assert FullID names (`opencode/<key>`) appear AND the lister is NOT invoked (use a lister that records calls / returns sentinel that must be absent) |
| `TestResolveModels_FallbackWhenCacheEmpty` (new) | C4 fallback + PR #45 reachable | cache reader returns empty map; lister returns live names → live names used (shell-out path) |
| `cacheModelNames` mapping (new, may live in resolve_test) | C4 FullID form | opencode bare key → `opencode/<key>`; slashed key → as-is; output sorted |
| `TestParseModels` (`opencode_test.go`) | (unchanged) | retained verbatim — `parseModels`/`execLister` not deleted (PR #45 preserved) |

### `cmd/archon/config_test.go` (no new file needed; verify pass) — supports 7
All existing tests use bare models (`claude-sonnet-4`, `gpt-4o`, `o3`); `FullID()` of an
empty-provider ref equals the bare string, so `Get`/`GetPhase`/`SetRoundtrip`/`List`/`ListEmpty`
pass unchanged. OPTIONAL add `TestConfigCmd_SetProviderQualified`: `set models.default
opencode/deepseek-v4-pro` then `get` → `opencode/deepseek-v4-pro` (proves the split+FullID seam).

### Existing TUI/status tests
Re-run; any that assert on `cfg.Models.*` strings get the `.FullID()` / `ParseModelRef` treatment in
their fixtures (same mechanical change as production). No new behavior to test beyond compile + the
unchanged rendered string for empty-provider refs.

---

## 9. Determinism / edge cases

- **Model containing a slash** (`FullID`): the `strings.Contains(r.Model, "/")` short-circuit
  returns it as-is even when `Provider` is non-empty — no double-prefix (M1 edge). `ParseModelRef`
  and `UnmarshalYAML` both split on the FIRST `/`, so `a/b/c` → `{a, b/c}` → `FullID` `a/b/c`
  (lossless round-trip).
- **Empty-provider advisory flow**: `{"",opus}` flows `ResolvePhaseModels` → `PhaseModel.Model =
  "opus"` (bare), and status/CLI render `FullID()` = `"opus"`. No leading slash anywhere. No change
  to the rendered value vs. today's flat string.
- **Cache absent vs present**: `defaultCacheReader` → `LoadModelsOrEmpty` → empty map on missing →
  `cacheModelNames` nil → shell-out fallback. Present + non-empty → cache FullIDs, lister untouched.
- **yaml.v3 recursion**: both `UnmarshalYAML` (mapping branch) and `MarshalYAML` use a local
  `type modelRefAlias ModelRef` to drop the custom methods and avoid infinite re-entry. `MarshalYAML`
  returns a plain value (string or alias struct), never a `ModelRef`.
- **`omitempty` with the custom `MarshalYAML` (VERIFIED via prototype)**: because `MarshalYAML`
  returns the bare `Model` STRING for an empty-provider ref, an empty `ModelRef{}` marshals to the
  empty string `""`, which yaml.v3 treats as an empty scalar and OMITS under the `omitempty` map
  field. Confirmed empirically: a legacy block `{default: claude-sonnet-4, phases: {apply: gpt-4o}}`
  with an unset `Leader` re-marshals WITHOUT a `leader:` key (Leader omitted) and the `default`/
  `phases.apply` values come back as the same bare scalars. A fully-unset config (no `models:` key)
  re-marshals with NO `models:` block (`Config.Models` `omitempty` at `config.go:51` covers the
  zero `ModelConfig`). So the empty-struct-`omitempty` worry does NOT materialize with the
  string-returning `MarshalYAML`.
- **Scope of "byte-identical"**: the guarantee is over the `models:` block specifically. yaml.v3
  may re-render UNRELATED scalars (e.g. it strips the quotes on `harness_version: "1.0.0"` → `1.0.0`)
  — that is pre-existing Save behavior unrelated to this change. The byte-identical test MUST compare
  the `models:` block (or normalize the rest), NOT assume the whole file is unchanged. **Flag for
  apply**: assert on the models block; also assert the "minimal config" (no `models:`) Saves without
  inventing one (`TestConfig_Load` already has this fixture).
- **Map iteration order** in `cacheModelNames`: sorted with `sort.Strings` for byte-stable output.

---

## 10. Size estimate (LOC)

| File | Action | ~LOC (incl. tests where noted) |
|---|---|---|
| `internal/opencode/models.go` | new | ~70 |
| `internal/opencode/models_test.go` | new | ~90 |
| `internal/opencode/testdata/{models,malformed,invalid}.json` | new | ~50 |
| `internal/config/model.go` (ModelRef + FullID + (Un)Marshal + ParseModelRef + ModelConfig + ResolvePhaseModels) | modify | ~70 |
| `internal/config/model_test.go` | modify | ~60 |
| `internal/config/config.go` (Clone) | modify | ~3 |
| `internal/config/config_test.go` (fixtures + byte-identical test) | modify | ~55 |
| `internal/models/resolve.go` (CacheReader + cacheModelNames + repoint) | modify | ~40 |
| `internal/models/resolve_test.go` | modify | ~50 |
| `cmd/archon/config.go` (set/get/list FullID + split) | modify | ~12 |
| `cmd/archon/config_test.go` (optional provider-qualified) | modify | ~15 |
| `internal/initcmd/init.go` (buildConfig split + leader FullID) | modify | ~10 |
| `internal/status/display.go` (FullID renders) | modify | ~4 |
| `internal/tui/models_tab.go` (FullID/ParseModelRef) | modify | ~10 |
| `internal/tui/model.go` (leader FullID) | modify | ~2 |
| **Total (prod + tests)** | | **~590**; **prod-only ≈ 300–330** |

The change exceeds D1 (400 changed lines). This was forecast (~400–480) and ACCEPTED by the user for
the Foundation PR under C1 (ask-always). **Flag again at PR creation.** If review pressure is high,
the only clean intra-foundation split is S0 (cache reader + tests, ~210 LOC, pure addition, no
callers) as a first chained PR, then S1 stacked on it — but the agreed plan keeps them combined.

---

## Open questions (for the Human Review Gate)

- [ ] **Byte-identity scope (RESOLVED by prototype, confirm acceptable)** — the string-returning
  `MarshalYAML` makes empty refs omit cleanly and re-emit legacy scalars verbatim WITHIN the
  `models:` block. yaml.v3 still re-renders unrelated scalars (drops quotes on `harness_version`),
  so the round-trip test asserts on the `models:` block, not the whole file. Confirm this scoping is
  acceptable (it matches pre-existing Save behavior).
- [ ] **`ResolveModels` signature change** — adding the `CacheReader` param changes a public-ish
  function signature; all in-repo callers are `Resolve()` and tests, so it is safe. Confirm no
  external consumer depends on the 2-arg form.
- [ ] **Drop `Model.Reasoning`?** — captured but unused this slice. Keep (cheap, documents shape) or
  trim to strictly-minimal. `ToolCall` must stay (S2/S3 need it). Recommend KEEP.
