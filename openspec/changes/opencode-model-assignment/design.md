# Design: opencode Model Assignment Bridge

## Technical Approach

Add a self-contained `internal/opencode/` package that (1) resolves archon's bare
model names into opencode's `provider/model` IDs, (2) injects those IDs into an
embedded agent overlay via a 3-case decision tree, and (3) additively deep-merges the
overlay into the shared `~/.config/opencode/opencode.json`. `archon init` calls one
entry point (`opencode.Apply`) gated on `agentName == "opencode"`, after skill
extraction and template write. A pre-merge backup of `opencode.json` is recorded in
the rollback manifest so it is restored — never deleted — on rollback. This realizes
the `opencode-overlay` and `model-resolution` specs (Slice 1). Slice 2 (multi-profile,
stale cleanup, `archon sync`) builds on the same package without changing Slice 1
call sites.

## Architecture Decisions

| Decision | Options | Choice & Rationale |
|----------|---------|--------------------|
| Where resolution lives | (A) cache-first + static fallback in new `internal/opencode`; (B) static map only in `internal/config`; (C) require qualified input | **A.** Matches gentle-ai and the spec. Cache gives exactly the IDs the user is authenticated for; static map guarantees init never fails when the cache is absent. Keeping it in a new package avoids importing `os`/JSON into `config`. |
| Overlay embed format | (A) one `//go:embed` JSON asset in `internal/opencode/assets/`; (B) build the map in Go; (C) text template | **A.** A literal JSON file is reviewable, diffable, and golden-testable; the JD/review agents are dropped vs gentle-ai (archon has no JD agents). Go-built maps hide the shape from review; templates add escaping noise. |
| Merge approach | (A) port `MergeJSONObjects` deep-merge with `__replace__`; (B) shallow top-level merge; (C) overwrite file | **A.** The spec mandates additive merge preserving user keys, with `__replace__` for the `permission.task` map. Shallow merge would clobber nested user `agent`/`provider` entries; overwrite destroys user config. |
| Backup field on manifest | (A) new generic `FileBackups []FileBackup`; (B) reuse existing single `BackupPath` | **A.** `BackupPath` is hard-coded to AGENTS.md restore logic in `Cleanup`. A separate slice keeps opencode.json restore independent and lets rollback restore both files. |

## Data Flow

    .archon/config.yaml (bare names)
        cfg.Models.Default / cfg.Models.Phases
                  │
                  ▼  opencode.Resolve(name)  ── ~/.cache/opencode/models.json
                  │                              (LoadModelsOrEmpty; static-map fallback)
          provider/model IDs
                  │
                  ▼  Inject(overlay, default, phases, existingAgentKeys)  ── read existing opencode.json
          injected overlay  (3-case decision tree per agent)
                  │
                  ▼  MergeJSONObjects(existing, overlay)  + backup recorded
          ~/.config/opencode/opencode.json   ──→ rollback manifest (FileBackups)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/opencode/paths.go` | Create | `SettingsPath()` → `~/.config/opencode/opencode.json`; `CachePath()` → `~/.cache/opencode/models.json`. Both via `os.UserHomeDir()`; empty string when no home. |
| `internal/opencode/models.go` | Create | Minimal port: `Model`, `Provider` types; `LoadModelsOrEmpty(path)` (tolerates missing/malformed file). No auth/variants/custom-provider code. |
| `internal/opencode/resolve.go` | Create | `Resolve(name) (qualified string, ok bool)` + `staticModelMap`. |
| `internal/opencode/jsonmerge.go` | Create | Port `MergeJSONObjects` + `__replace__` sentinel (drop JSONC comment/trailing-comma stripping unless tests need it; keep malformed-base→empty tolerance). |
| `internal/opencode/inject.go` | Create | `Inject(overlayBytes, defaultModel, phases, existingAgentKeys)`; `readExistingAgentKeys(path)`. Implements the 3-case tree. |
| `internal/opencode/apply.go` | Create | `Apply(opts ApplyOptions) (warnings []string, err error)` — the single init entry point: resolve, inject, backup, merge, write. |
| `internal/opencode/assets/sdd-overlay.json` | Create | Embedded overlay (`//go:embed` in `assets.go`): `archon-orchestrator` primary + 10 hidden subagents. |
| `internal/opencode/assets.go` | Create | `//go:embed assets/sdd-overlay.json` → `var overlayJSON []byte`. |
| `internal/initcmd/init.go` | Modify | After `writeTemplate`, if `agentName == "opencode"` call `opencode.Apply`, record its backup in the manifest, print warnings. |
| `internal/initcmd/templates.go` | Modify | Add `## Phase Models` block to `agentsTemplate` (and optionally `claudeTemplate`). |
| `internal/config/rollback.go` | Modify | Add `FileBackups []FileBackup` (`{Target, Backup string}`); `Cleanup` restores each (copy backup→target) and never deletes the target when a backup exists. |

## Interfaces / Contracts

```go
// internal/opencode
func SettingsPath() string                    // ~/.config/opencode/opencode.json
func CachePath() string                        // ~/.cache/opencode/models.json
func Resolve(name string) (qualified string, ok bool)
func MergeJSONObjects(base, overlay []byte) ([]byte, error)
func Inject(overlay []byte, defaultModel string, phases map[string]string,
    existingAgentKeys map[string]bool) ([]byte, error)

type ApplyOptions struct{ SettingsPath, CachePath, DefaultModel string; Phases map[string]string }
func Apply(opts ApplyOptions) (backupPath string, warnings []string, err error)
```

`Resolve` order: (1) value already contains `/` → pass through, `ok=true`; (2) cache
lookup over `LoadModelsOrEmpty` (match bare name to a `provider`+model id); (3)
`staticModelMap[name]`; (4) else `("", false)` — caller appends a warning and skips
that agent's `model`. The static map covers every `config.KnownModels` key:
`gpt-4`/`gpt-4o`/`gpt-4o-mini`→`openai/…`, `claude-sonnet-4`/`claude-haiku-4`→
`anthropic/…`, `gemini-2.5-pro`/`gemini-2.5-flash`→`google/…`, `o3`/`o3-mini`/
`o4-mini`→`openai/…`.

**Phase→agent key mapping:** `.archon` phase names (`explore`,`apply`,…) map to overlay
keys `sdd-<phase>`. `models.default` resolves to the fallback injected into every
subagent lacking an explicit phase assignment. `variant` is never written (archon has
no effort field — avoids stale-variant override).

**Overlay skeleton** (`assets/sdd-overlay.json`): `archon-orchestrator` with
`"mode":"primary"`, `"prompt":"{file:./AGENTS.md}"`, and `permission.task` =
`{"__replace__":{"*":"deny","sdd-init":"allow",…,"sdd-onboard":"allow"}}` (all ten
phases). Ten subagents `sdd-<phase>` each with `"mode":"subagent"`, `"hidden":true`,
Executor-Override `prompt` → `~/.config/opencode/skills/sdd-<phase>/SKILL.md`, and a
`tools` map. No JD/review agents.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `Resolve`: qualified pass-through, cache hit, cache-miss→static, static-miss→`ok=false` | Table tests; temp `models.json` for cache cases |
| Unit | `MergeJSONObjects`: additive (preserve user keys), `__replace__` whole-key replace, no prior file (empty/malformed base) | Table tests on byte literals |
| Unit | `Inject` 3-case tree: explicit phase wins; existing user agent preserved (untouched); default fallback injected | Assert `agent.sdd-*.model` per case |
| Golden | Full `Apply` against an empty + a pre-populated `opencode.json`; assert merged shape, preserved user keys, restored on rollback | `go-testing` golden file |
| Invariant | Overlay subagent keys == extracted `skills/sdd-*` folder names | Read `skills.FS` dirs, compare to overlay `agent` keys (minus orchestrator) |
| Unit | `agentsTemplate` renders a "Phase Models" section | Assert substring in `RenderAgentsMD` output |

## Edge Cases & Error Handling

- **Missing cache** → `LoadModelsOrEmpty` returns empty; static fallback runs; init never fails.
- **Unresolved name** → `Resolve` returns `ok=false`; `Apply` appends a warning, skips that agent's `model`; init succeeds.
- **Malformed existing opencode.json** → merge treats base as empty (matches gentle-ai); user is warned via backup retention; no crash.
- **No prior opencode.json** → backup path empty; rollback may remove the created file (no user content to lose).
- **No home dir** → `SettingsPath`/`CachePath` return `""`; `Apply` returns a warning and skips overlay (init still succeeds).
- **Concurrency** → write via temp-file + `os.Rename` (matches `Config.Save`/template write) for atomicity.

## Migration / Rollout

No data migration. Re-running `archon init` on an existing project re-backs-up and
re-merges idempotently (additive merge + per-run backup). `opencode.json` MUST NOT be
added to `CreatedPaths`.

## Slice Boundary

**Slice 1 (PR 1) — complete, testable bug fix:** `paths.go`, `models.go`, `resolve.go`,
`jsonmerge.go`, `inject.go`, `apply.go`, `assets/sdd-overlay.json` + `assets.go`;
`init.go` wiring; `templates.go` Phase Models; `rollback.go` `FileBackups`; all unit +
golden + invariant tests. Single default overlay, no `archon sync`.

**Slice 2 (PR 2):** named multi-profile overlays + per-profile agent keys; stale
archon-managed-agent cleanup on re-apply; `archon sync` subcommand in `cmd/archon`
re-applying after TUI/CLI model edits. Depends only on Slice 1 public funcs.

## Open Questions

- [ ] Cache match semantics for an ambiguous bare name present under multiple providers — pick first authenticated, or prefer the static-map provider? (Default: prefer static-map provider for determinism; revisit in Slice 2 with auth detection.)
- [ ] Should `claudeTemplate` also gain the Phase Models section, or only `agentsTemplate`? (Spec requires only `agentsTemplate`.)
