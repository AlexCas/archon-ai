# Proposal: Dynamic Model Detection (Slice 3)

## Intent

Static catalogs in `internal/config/model.go` drift: curated `OpencodeModels` lists `glm-5`/`kimi-k2.5`/`qwen3.7-plus`, but `opencode models opencode-go` returns `glm-5.2`/`kimi-k2.7-code`/`minimax-m3`. The TUI also shows models for agent CLIs the user has not installed. Make the catalog reflect what is actually available; keep free-form entry working.

## Scope

### In Scope
- PATH-based agent-CLI detector in `internal/agent/` (`exec.LookPath`), injectable.
- Opencode model lister that shells out to `opencode models opencode-go` (context timeout, parses `provider/model` lines), injectable, with curated fallback on any error.
- Resolver feeding the `StaticModels()` consumption point: composes detected agents' catalogs (opencode live when present), filters out non-installed agents, catalog-agnostic over whatever curated lists exist.
- TUI Models tab consumes the resolved list (cycling + renamed "Available:" hint); detection cached once at tab open.

### Out of Scope
- Enumerating Claude/Gemini/OpenAI models (remote; stay curated). *(deferred)*
- Auth-based provider filtering (full opencode catalog for v1). *(deferred)*
- Detection during `archon init` (TUI-open only). *(deferred)*

## Capabilities

> Contract for sdd-spec. `harness-init` governs init + the static catalog/TUI behavior (Slices 1–2 modified it).

### New Capabilities
- None.

### Modified Capabilities
- `harness-init`: the "Static model selection with free-form fallback" requirement changes — the offered catalog becomes DYNAMIC (filtered by detected agent CLIs; opencode models enumerated live with curated fallback) instead of a fixed static list. Free-form entry and advisory `Validate`/`NormalizeModel` behavior are unchanged.

## Approach

Hybrid live enumeration (exploration Approach D):
1. `DetectCLIs()` → which agent CLIs are on PATH (`exec.LookPath`, repo's first `os/exec`, narrow/injectable).
2. If `opencode` present, list live via `opencode models opencode-go` (timeout-bounded, never `--refresh`); on ANY error fall back to curated `OpencodeModels`.
3. Claude stays curated. Filter out catalogs whose CLI is absent.
4. Resolver returns the same ordered shape `StaticModels()` produces, seeding the TUI and `KnownModels` unchanged.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/agent/detect.go` (or new file) | New | Injectable PATH CLI detector |
| `internal/config/model.go` (or new `internal/models/`) | New/Modified | Injectable opencode lister + resolver behind `StaticModels()` seam |
| `internal/tui/models_tab.go` | Modified | Consume resolved list; cache at open; "Available:" hint |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| First `os/exec` surface | Med | Context timeout, injectable, never fail the TUI |
| `opencode models` output contract drifts | Med | Parse defensively; curated fallback |
| Test determinism (env-dependent) | High | Inject detector/lister fake; no real subprocess in unit tests |
| Cross-slice: Slice 1 `Gemini`/`OpenAI` catalogs unmerged (PR #40) | Med | Catalog-agnostic iteration; base-branch decision deferred to apply |
| Review budget (detector+lister+resolver+TUI+tests) near 400 lines | Med | See budget read below; chained split if exceeded |

## Rollback Plan

Revert the slice commit(s). The `StaticModels()` seam and curated lists remain intact, so the TUI falls back to the pre-existing static behavior with no data migration or config change.

## Dependencies

- `opencode` CLI on PATH for live enumeration (optional; curated fallback otherwise).
- Composes with Slice 1 (`GeminiModels`/`OpenAIModels`, PR #40) once merged — must NOT assume those exist.

## Success Criteria

- [ ] Models tab hides catalogs for agent CLIs not on PATH; free-form entry still accepted.
- [ ] When `opencode` is installed, the live catalog (not the stale curated one) is shown.
- [ ] Offline / opencode-absent / subprocess-error degrades silently to curated lists.
- [ ] Detection runs once at TUI open (cached), not per keystroke or at `init`.
- [ ] `NormalizeModel`, `Validate`, `KnownModels` derivation unchanged; unit tests use injected fakes (no real subprocess).
