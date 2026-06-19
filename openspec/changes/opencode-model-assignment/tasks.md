# Tasks: opencode Model Assignment Bridge

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines (Slice 1) | ~310–380 |
| Estimated changed lines (Slice 2) | ~370–460 |
| 400-line budget risk | Moderate (Slice 1 within budget; Slice 2 near or over) |
| Chained PRs | Yes — two PRs as confirmed |
| Slice 1 goal | Complete standalone bug fix: opencode init writes resolved models into opencode.json |
| Slice 2 goal | Multi-profile + stale-agent cleanup + `archon sync` |

---

## Slice 1 (PR 1) — Standalone Bug Fix

### [x] S1-1 Add `FileBackup` type and `FileBackups` slice to `RollbackManifest`

**File:** `internal/config/rollback.go`

Add `type FileBackup struct { Target, Backup string }` and field
`FileBackups []FileBackup \`json:"file_backups,omitempty"\`` to `RollbackManifest`.
In `Cleanup`, iterate `FileBackups` in reverse: if `Backup != ""`, copy backup
→ target (temp-file + rename for atomicity); if `Backup == ""` and target exists,
remove it (no prior user content). Never add `opencode.json` to `CreatedPaths`.

**Spec coverage:** opencode-overlay § Backup and Rollback for Shared File (all three
scenarios: pre-merge backup recorded; rollback restores; no-prior-file removes).

---

### [x] S1-2 Unit tests for `FileBackups` rollback behaviour

**File:** `internal/config/rollback_test.go`

Four new table-driven cases alongside existing tests:
1. `FileBackup` with prior file → restored after `Cleanup`.
2. `FileBackup` with no prior file (`Backup == ""`) → target removed after `Cleanup`.
3. `FileBackup` + existing `CreatedPaths` → both processed in same `Cleanup` call.
4. `WriteManifest` round-trips `FileBackups` through JSON unmarshal.

**Spec coverage:** opencode-overlay § Backup and Rollback (all three scenarios).

---

### [x] S1-3 Create `internal/opencode/paths.go`

**File:** `internal/opencode/paths.go` (new)

```go
func SettingsPath() string  // ~/.config/opencode/opencode.json
func CachePath() string     // ~/.cache/opencode/models.json
```

Both use `os.UserHomeDir()`; return `""` when home dir is unavailable.

**Spec coverage:** opencode-overlay § Overlay Generation Gate (path resolution
prerequisite); model-resolution § Init Never Fails on Missing Cache.

---

### [x] S1-4 Create `internal/opencode/models.go`

**File:** `internal/opencode/models.go` (new)

Minimal port from gentle-ai: `type Model struct { ID, Name string }`,
`type Provider struct { ID string; Models []Model }`.
`LoadModelsOrEmpty(path string) []Provider` — returns empty slice on missing or
malformed file; never errors.

**Spec coverage:** model-resolution § Cache-Based Resolution; § Init Never Fails on
Missing Cache.

---

### [x] S1-5 Create `internal/opencode/resolve.go` with static map

**File:** `internal/opencode/resolve.go` (new)

`Resolve(name string) (qualified string, ok bool)` in priority order:
1. Contains `/` → pass through, `ok=true`.
2. Cache lookup via `LoadModelsOrEmpty(CachePath())`.
3. `staticModelMap[name]`.
4. `("", false)` — caller appends warning, skips model field.

`staticModelMap` covers all `config.KnownModels` keys:

| Bare name | Qualified ID |
|-----------|--------------|
| `gpt-4` | `openai/gpt-4` |
| `gpt-4o` | `openai/gpt-4o` |
| `gpt-4o-mini` | `openai/gpt-4o-mini` |
| `o3` | `openai/o3` |
| `o3-mini` | `openai/o3-mini` |
| `o4-mini` | `openai/o4-mini` |
| `claude-sonnet-4` | `anthropic/claude-sonnet-4` |
| `claude-haiku-4` | `anthropic/claude-haiku-4` |
| `gemini-2.5-pro` | `google/gemini-2.5-pro` |
| `gemini-2.5-flash` | `google/gemini-2.5-flash` |

Tiebreak when bare name appears under multiple cache providers: prefer the
static-map provider (deterministic; revisit with auth detection in Slice 2).

**Spec coverage:** model-resolution § Qualified ID Pass-Through; § Cache-Based
Resolution; § Static-Map Fallback; § Init Never Fails on Missing Cache (both
scenarios including unresolvable name → warning, not error).

---

### [x] S1-6 Unit tests for `Resolve`

**File:** `internal/opencode/resolve_test.go` (new)

Four table-driven cases:
1. Qualified ID (`anthropic/claude-sonnet-4`) passes through unchanged.
2. Cache hit: temp `models.json` with matching provider → returns `provider/model`.
3. Cache miss + static map hit: no cache file, bare name in static map → static
   result returned.
4. Static miss: bare name absent from both → `ok=false`, no panic.

**Spec coverage:** model-resolution all four requirements.

---

### [x] S1-7 Create `internal/opencode/jsonmerge.go`

**File:** `internal/opencode/jsonmerge.go` (new)

Port `MergeJSONObjects(base, overlay []byte) ([]byte, error)` from gentle-ai:
recursive object merge; `__replace__` sentinel forces whole-key replacement;
malformed or empty `base` treated as `{}` (no crash). No JSONC stripping needed
for Slice 1.

**Spec coverage:** opencode-overlay § Additive Deep-Merge (both scenarios:
pre-existing content preserved; sentinel forces replacement).

---

### [x] S1-8 Unit tests for `MergeJSONObjects`

**File:** `internal/opencode/jsonmerge_test.go` (new)

Three table-driven cases:
1. Additive merge: overlay adds new keys; existing user keys remain.
2. `__replace__` sentinel: nested map replaced wholesale, not merged.
3. Empty/malformed base: `{}` treated as empty; overlay wins; no error.

**Spec coverage:** opencode-overlay § Additive Deep-Merge (both scenarios) + edge
case from design § Edge Cases (malformed base).

---

### [x] S1-9 Create `internal/opencode/assets/sdd-overlay.json`

**File:** `internal/opencode/assets/sdd-overlay.json` (new)

JSON object with top-level key `"agent"` containing:
- `"archon-orchestrator"`: `"mode": "primary"`, `"prompt": "{file:./AGENTS.md}"`,
  `"permission": { "task": { "__replace__": { "*": "deny", "sdd-init": "allow",
  "sdd-explore": "allow", "sdd-propose": "allow", "sdd-spec": "allow",
  "sdd-design": "allow", "sdd-tasks": "allow", "sdd-apply": "allow",
  "sdd-verify": "allow", "sdd-archive": "allow", "sdd-onboard": "allow" } } }`.
- Ten subagents `sdd-<phase>` (init, explore, propose, spec, design, tasks, apply,
  verify, archive, onboard): `"mode": "subagent"`, `"hidden": true`,
  `"prompt": "{file:~/.config/opencode/skills/sdd-<phase>/SKILL.md}"`,
  `"model": ""` (placeholder; filled by `Inject`), `"tools": {}`.

No JD/review agents (archon has no judgment-day agents in opencode overlay).

**Spec coverage:** opencode-overlay § Orchestrator Agent Definition; § Delegation
Allow-List; § Per-Phase Subagent Definitions.

---

### [x] S1-10 Create `internal/opencode/assets.go`

**File:** `internal/opencode/assets.go` (new)

```go
//go:embed assets/sdd-overlay.json
var overlayJSON []byte
```

Single exported var; no other logic.

**Spec coverage:** prerequisite for `Apply` (S1-11).

---

### [x] S1-11 Create `internal/opencode/inject.go`

**File:** `internal/opencode/inject.go` (new)

```go
func Inject(overlay []byte, defaultModel string, phases map[string]string,
    existingAgentKeys map[string]bool) ([]byte, error)
func readExistingAgentKeys(settingsPath string) (map[string]bool, error)
```

`Inject` iterates every `agent.sdd-<phase>` key in the overlay JSON. For each:
1. If `phases[phase]` resolves to a non-empty qualified ID → set that model.
2. Else if `existingAgentKeys[key]` is true → omit the agent entry from the overlay
   (the merge will preserve the user's existing entry).
3. Else → set the resolved `defaultModel` (prevents silent inheritance of the
   orchestrator's model).

`readExistingAgentKeys` reads `opencode.json` at the given path, extracts top-level
`agent` object keys; returns empty map + nil error when file is absent.

**Spec coverage:** opencode-overlay § Model Injection Decision Tree (all three
scenarios: explicit wins; existing user agent preserved; default fallback).

---

### [x] S1-12 Unit tests for `Inject`

**File:** `internal/opencode/inject_test.go` (new)

Three table-driven cases:
1. Explicit phase assignment wins: `phases["apply"] = "openai/gpt-4o"` →
   `agent.sdd-apply.model` equals `openai/gpt-4o`.
2. Existing user agent preserved: `existingAgentKeys["sdd-spec"] = true` and no
   explicit phase → `sdd-spec` absent from inject output (merge keeps user entry).
3. Default fallback: no explicit phase, no existing key →
   `agent.sdd-<phase>.model` equals resolved default.

**Spec coverage:** opencode-overlay § Model Injection Decision Tree (all three
scenarios).

---

### [x] S1-13 Create `internal/opencode/apply.go`

**File:** `internal/opencode/apply.go` (new)

```go
type ApplyOptions struct {
    SettingsPath string
    CachePath    string
    DefaultModel string
    Phases       map[string]string
}
func Apply(opts ApplyOptions) (backupPath string, warnings []string, err error)
```

Steps:
1. If `SettingsPath == ""` → append warning, return `("", warnings, nil)` (no home
   dir edge case).
2. Resolve `DefaultModel` via `Resolve`; on `ok=false` append warning.
3. Resolve each entry in `Phases`; on `ok=false` append warning, skip that phase.
4. Read existing `opencode.json` (absent = `{}`); call `readExistingAgentKeys`.
5. `Inject(overlayJSON, resolvedDefault, resolvedPhases, existingKeys)`.
6. Backup: if prior file existed, copy to `<settingsPath>.backup.<timestamp>` via
   temp-file + rename; set `backupPath`.
7. `MergeJSONObjects(existing, injected)`.
8. Write merged result via temp-file + rename (atomic write).
9. Return `backupPath, warnings, nil`.

`opencode.json` is NEVER added to `CreatedPaths`.

**Spec coverage:** opencode-overlay § Overlay Generation Gate; § Backup and Rollback
for Shared File; model-resolution § Init Never Fails on Missing Cache; design §
Edge Cases (no home dir, missing cache, malformed base, no prior file).

---

### [x] S1-14 Golden tests for `Apply`

**File:** `internal/opencode/apply_test.go` (new)

Two golden-file cases using `t.TempDir()`:
1. Empty `opencode.json` (or no file) → merged result matches golden fixture;
   `backupPath` empty when no prior file; no warnings for known model names.
2. Pre-populated `opencode.json` with user agents/providers → merged result
   preserves all user keys; `backupPath` non-empty; rollback restores original
   (copy backup → target, assert content matches original).

Golden fixtures in `internal/opencode/testdata/`.

**Spec coverage:** opencode-overlay § Additive Deep-Merge (pre-existing content
preserved); § Backup and Rollback (backup recorded; rollback restores; no deletion).

---

### [x] S1-15 Invariant test: overlay agent keys match extracted `sdd-*` folders

**File:** `internal/opencode/overlay_invariant_test.go` (new)

Read `assets/sdd-overlay.json`; collect `agent` keys minus `archon-orchestrator`.
Read `skills.FS` (from `skills/embed.go`); collect directory names matching
`sdd-*`. Assert the two sets are equal. Fails the build if a new sdd-* skill is
added without a matching overlay entry (or vice versa).

**Spec coverage:** opencode-overlay § Per-Phase Subagent Definitions (subagent keys
match extracted skill folders scenario).

---

### [x] S1-16 Add `## Phase Models` section to `agentsTemplate` (and `claudeTemplate`)

**File:** `internal/initcmd/templates.go`

Insert the following section in both `agentsTemplate` and `claudeTemplate` after the
`## Configuration` block (before `## State Management`):

```
## Phase Models
Per-phase models are wired via the opencode.json agent definitions under
`~/.config/opencode/opencode.json`. Each SDD phase agent (`sdd-<phase>`) carries
its resolved `provider/model` ID so opencode uses the correct model per phase
rather than the orchestrator default.
```

The confirmed decision specifies this section added to BOTH templates.

**Spec coverage:** opencode-overlay § Phase Models Documentation in AGENTS.md.

---

### [x] S1-17 Unit test: `agentsTemplate` and `claudeTemplate` contain Phase Models section

**File:** `internal/initcmd/templates_test.go`

Add two assertions (one per render function) that the rendered output contains the
substring `"## Phase Models"`. Run alongside existing template tests.

**Spec coverage:** opencode-overlay § Phase Models Documentation in AGENTS.md
(AGENTS.md includes Phase Models section scenario).

---

### [x] S1-18 Wire `opencode.Apply` into `archon init` (opencode-gated)

**File:** `internal/initcmd/init.go`

After `writeTemplate`, if `agentName == "opencode"`:
1. Call `opencode.Apply(ApplyOptions{ SettingsPath: opencode.SettingsPath(),
   CachePath: opencode.CachePath(), DefaultModel: opts.ModelDefault,
   Phases: opts.ModelPhases })`.
2. If `backupPath != ""`, append `config.FileBackup{Target: opencode.SettingsPath(),
   Backup: backupPath}` to `rollback.FileBackups` and call `rollback.WriteManifest()`
   again.
3. Print any warnings to stderr.
4. On non-nil error: return wrapped error (init fails cleanly).

Non-opencode agents: no overlay step (gate enforced by the `if` check).

**Spec coverage:** opencode-overlay § Overlay Generation Gate (both scenarios:
opencode triggers; non-opencode skipped); design § Migration/Rollout
(opencode.json not in CreatedPaths).

---

### [x] S1-19 Slice 1 verification task (build + test suite + smoke check)

**Files:** all Slice 1 files above

1. `go build ./...` — must compile with no errors.
2. `go test ./internal/opencode/... ./internal/config/... ./internal/initcmd/...` —
   all tests green; golden files generated on first run and committed.
3. Manual smoke: `archon init` in a temp opencode project with
   `--model claude-sonnet-4 --model-apply gpt-4o`; assert:
   - `~/.config/opencode/opencode.json` contains `agent.archon-orchestrator` with
     `mode: primary`.
   - `agent.sdd-apply.model` equals `openai/gpt-4o`.
   - `agent.sdd-explore.model` equals `anthropic/claude-sonnet-4` (default fallback).
   - Pre-existing user keys (if any) survive the merge.
   - Rollback manifest `file_backups` has one entry for `opencode.json` (when prior
     file existed).

**Spec coverage:** end-to-end confirmation of all Slice 1 requirements.

---

## Slice 2 (PR 2) — Multi-Profile + Stale Cleanup + `archon sync`

> Depends only on Slice 1 public functions. PR 2 base = PR 1 branch.

### S2-1 Add named SDD profile support to `internal/opencode/`

**Files:** `internal/opencode/profiles.go` (new),
`internal/opencode/assets/sdd-overlay-<profile>.json` (new per profile)

Define `Profile` struct with name, per-phase model overrides, and agent-key prefix.
`LoadProfile(name string)` reads the matching embedded overlay asset.
Default profile (`"default"`) maps to the existing `sdd-overlay.json`.

**Spec coverage:** opencode-overlay § Multi-Profile Overlays [Slice 2] (named profile
generates its own agents scenario).

---

### S2-2 Stale archon-managed agent cleanup on re-apply

**File:** `internal/opencode/cleanup.go` (new)

`RemoveStaleAgents(settingsPath string, managedKeys []string) error` reads existing
`opencode.json`, removes entries from the `agent` object whose keys are in
`managedKeys` but NOT in the new profile's agent set, preserves all other keys,
and writes atomically.

**Spec coverage:** opencode-overlay § Stale Agent Cleanup [Slice 2] (both scenarios:
stale archon agent removed; user-authored agents preserved).

---

### S2-3 Profile selection in `archon init` and `Apply`

**File:** `internal/opencode/apply.go`, `internal/initcmd/init.go`

Add `Profile string` field to `ApplyOptions`. When non-empty, `Apply` loads the
named profile overlay. Wire `--profile` flag through `archon init` → `Options` →
`Apply`.

**Spec coverage:** opencode-overlay § Multi-Profile Overlays [Slice 2].

---

### S2-4 Add `archon sync` subcommand

**File:** `cmd/archon/main.go`

New cobra subcommand `sync` that:
1. Loads `.archon/config.yaml` from the current project dir.
2. Calls `opencode.Apply` with current resolved models (same logic as init step).
3. Runs stale-agent cleanup for replaced profiles.
4. Prints warnings; exits non-zero on hard error.

**Spec coverage:** opencode-overlay § Re-apply via archon sync [Slice 2] (edited
models re-applied scenario).

---

### S2-5 Tests for multi-profile, stale cleanup, and `archon sync`

**Files:** `internal/opencode/profiles_test.go`,
`internal/opencode/cleanup_test.go`, `cmd/archon/e2e_test.go`

- Profile test: loading a named profile produces agent keys prefixed for that
  profile.
- Cleanup test: stale managed key removed; user key preserved.
- `archon sync` e2e: edit `models.phases.apply` in config → run `sync` → assert
  `opencode.json` has updated model; user keys still present.

**Spec coverage:** opencode-overlay § Multi-Profile Overlays; § Stale Agent Cleanup;
§ Re-apply via archon sync [Slice 2].

---

## Spec Requirement Coverage Matrix

| Spec | Requirement | Task(s) |
|------|-------------|---------|
| opencode-overlay | Overlay Generation Gate | S1-18 |
| opencode-overlay | Orchestrator Agent Definition | S1-9, S1-14 |
| opencode-overlay | Delegation Allow-List | S1-9, S1-14 |
| opencode-overlay | Per-Phase Subagent Definitions | S1-9, S1-15 |
| opencode-overlay | Model Injection Decision Tree | S1-11, S1-12 |
| opencode-overlay | Additive Deep-Merge | S1-7, S1-8, S1-14 |
| opencode-overlay | Backup and Rollback for Shared File | S1-1, S1-2, S1-13, S1-14 |
| opencode-overlay | Phase Models Documentation | S1-16, S1-17 |
| opencode-overlay | Multi-Profile Overlays [Slice 2] | S2-1, S2-3 |
| opencode-overlay | Stale Agent Cleanup [Slice 2] | S2-2, S2-5 |
| opencode-overlay | Re-apply via archon sync [Slice 2] | S2-4, S2-5 |
| model-resolution | Qualified ID Pass-Through | S1-5, S1-6 |
| model-resolution | Cache-Based Resolution | S1-4, S1-5, S1-6 |
| model-resolution | Static-Map Fallback | S1-5, S1-6 |
| model-resolution | Init Never Fails on Missing Cache | S1-5, S1-6, S1-13 |
