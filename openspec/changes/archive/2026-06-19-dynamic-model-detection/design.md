# Design — dynamic-model-detection (Slice 3)

## Technical Approach

Hybrid Approach D, built behind the existing `StaticModels()` seam. A new pure
engine (`internal/models/`) composes the offered catalog from injected
dependencies: a PATH **CLI detector** and an **opencode lister**. The real
implementations use `os/exec` (the repo's first), bounded by `context` timeout;
unit tests inject fakes — no real subprocess. `StaticModels()` stays the curated
source-of-truth (fallback + `KnownModels` seed). A new `ResolveModels(detector,
lister)` produces the dynamic, ordered list the TUI consumes, cached once at
Models-view open. `NormalizeModel`, `Validate`, and free-form acceptance are
untouched.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Engine home | New `internal/models/` package | Keeps `os/exec` out of `config` (pure); resolver depends on `config` catalogs without a cycle |
| Detector seam | `type CLIDetector func() map[string]bool` (real impl wraps `exec.LookPath`) | Mirrors `Detect(fsys fs.FS)` injection style; trivial to fake |
| Lister seam | `type OpencodeLister interface { List(ctx) ([]string, error) }` | Real impl runs the subprocess; fake returns canned/err |
| Catalog source per agent | Iterate curated catalogs that exist; gate by detected CLI | Catalog-agnostic — composes once Slice 1's Gemini/OpenAI land |
| opencode models | Live `opencode models opencode-go` when present; curated `OpencodeModels` ONLY on enumeration failure | Fixes drift; absent opencode ⇒ hidden (not curated) |
| Claude / remote-only | Stay curated | Nothing locally enumerable |
| `KnownModels` | Stays curated-derived (from `StaticModels()`) — unchanged | `Validate` is advisory, free-form always accepted; no reason to couple it to env-dependent detection |
| Caching | Resolved once in `newModelsTabState`, stored on tab state | Spec: once per view, never per keystroke, never at init |
| Failure UX | Silent fallback/hide; never error or block | Spec hard rule |

## Data Flow

```
TUI Models-view open
  └─ newModelsTabState(cfg, resolved)        // resolved computed once at open
        resolved = models.ResolveModels(detector, lister)
           detector() -> {"opencode":true, "claude":true, ...}   // exec.LookPath
           for each curated catalog whose CLI is detected:
             - opencode + present -> lister.List(ctx)            // live
                 on err/timeout/empty -> config.OpencodeModels   // curated fallback
             - claude/other present -> curated list
             - CLI absent -> SKIP (hidden)
        -> ordered []string (Claude curated first, then live/curated opencode, …)
  └─ cycleStaticModel / "Available:" hint read m.catalog (cached), not StaticModels()
```

## Interfaces / Contracts

```go
// internal/models/detect.go  (PR1)
type CLIDetector func() map[string]bool

func LookPathDetector() map[string]bool { // real impl
    out := map[string]bool{}
    for _, c := range []string{"opencode", "claude", "codex", "gemini"} {
        if _, err := exec.LookPath(c); err == nil { out[c] = true }
    }
    return out
}

// internal/models/opencode.go  (PR1)
type OpencodeLister interface {
    List(ctx context.Context) ([]string, error)
}

type execLister struct{} // real impl
func (execLister) List(ctx context.Context) ([]string, error) {
    cmd := exec.CommandContext(ctx, "opencode", "models", "opencode-go")
    out, err := cmd.Output()
    if err != nil { return nil, err }
    return parseModels(out), nil
}

// parseModels strips the "provider/" prefix from each non-empty line.
func parseModels(b []byte) []string {
    var ms []string
    for _, ln := range strings.Split(string(b), "\n") {
        ln = strings.TrimSpace(ln)
        if ln == "" { continue }
        if i := strings.IndexByte(ln, '/'); i >= 0 { ln = ln[i+1:] }
        ms = append(ms, ln)
    }
    return ms
}

// internal/models/resolve.go  (PR1)
func ResolveModels(detect CLIDetector, lister OpencodeLister) []string {
    present := detect()
    var out []string
    if present["claude"] { out = append(out, config.ClaudeModels...) }
    if present["opencode"] {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()
        if live, err := lister.List(ctx); err == nil && len(live) > 0 {
            out = append(out, live...)
        } else {
            out = append(out, config.OpencodeModels...) // installed-but-failed fallback
        }
    }
    // future curated catalogs (Gemini/OpenAI) gate the same way when they exist
    return out
}

// Default convenience used by the TUI:
func Resolve() []string { return ResolveModels(LookPathDetector, execLister{}) }
```

## File Changes

| File | PR | Change |
|------|----|--------|
| `internal/models/detect.go` | PR1 | `CLIDetector` type + `LookPathDetector` (exec.LookPath) |
| `internal/models/opencode.go` | PR1 | `OpencodeLister` iface, `execLister` (CommandContext+timeout), `parseModels` |
| `internal/models/resolve.go` | PR1 | `ResolveModels(detector, lister)`, `Resolve()` convenience |
| `internal/models/*_test.go` | PR1 | Engine unit tests with injected fakes (no subprocess) |
| `internal/config/model.go` | — | Unchanged (curated lists, `StaticModels`, `KnownModels`, `Validate`) |
| `internal/tui/models_tab.go` | PR2 | `modelsTabState.catalog []string`; `newModelsTabState(cfg, catalog)`; `cycleStaticModel`/hint read `m.catalog`; rename hint "Static:" → "Available:" |
| `internal/tui/model.go` | PR2 | Compute `models.Resolve()` at open; pass into `newModelsTabState` (both call sites: line 88 & 175) |
| `internal/tui/models_tab_test.go` | PR2 | Cycling/hint/cache tests using an injected catalog slice |

## Testing Strategy

No real subprocess in unit tests; inject `CLIDetector` func + fake `OpencodeLister`.

| Spec scenario | Test |
|---------------|------|
| Installed opencode → live catalog | detector{opencode:true}, lister returns live list → output contains live, not curated |
| Only installed agents offered | detector{opencode:false} → no opencode models in output; present agents remain |
| Cached once per view | `newModelsTabState` resolves once; cycling/typing reads `m.catalog`; resolver fn not re-invoked (call counter on fake) |
| Live enumeration error → curated fallback | detector{opencode:true}, lister returns err/timeout/empty → output == `config.OpencodeModels` |
| Free-form / Validate unchanged | existing `config` tests unchanged; `Validate` semantics asserted intact |

`parseModels` gets a table test (prefixed lines, blank lines, no-slash lines).

## Migration / Rollout

No data/config migration. Chained PRs stacked to main: **PR1** = pure injectable
engine + tests (no UI). **PR2** = TUI wiring (cache at open, hint rename) + tests,
stacked on PR1. Rollback = revert; curated `StaticModels()` seam remains intact.

## Open Questions

- **Base branch**: deferred to apply (Slice 1 Gemini/OpenAI in unmerged PR #40).
  `ResolveModels` composes future catalogs via a uniform per-catalog gate — adding
  Gemini/OpenAI is one small gated branch each (not a zero-edit drop-in; there is
  no generic catalog registry yet, which would be premature here).
- **Timeout value**: 2s proposed; confirm at apply if opencode cold-start is slow.
