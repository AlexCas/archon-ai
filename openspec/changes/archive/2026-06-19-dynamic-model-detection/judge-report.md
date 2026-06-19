# Judge Phase Report — PR1 (engine)

**Change**: dynamic-model-detection (chained delivery; this report covers PR1 = `internal/models/` engine)
**Verdict**: pass (after one confirmed-fix round)
**Retry**: 1 / 3

## Judgment-Day Result (PR1 diff)
- Judge A: ISSUES FOUND (INFO/SUGGESTION only) → resolved
- Judge B: ISSUES FOUND (INFO/SUGGESTION only) → resolved
- Confirmed (both judges): `parseModels` appended an empty string for a bare `provider/` line (empty model name after prefix strip).
- No CRITICAL, no WARNING (real).

## Gates
- Mutation gate: skipped (`mutation_testing.enabled: false`); Playwright: skipped (not web).

## Independent verification by both judges
- Engine scenarios pass: live catalog when opencode present (live values deliberately disjoint from curated, so the no-leak assertion is real); absent agent hidden (not curated); enumeration failure → curated fallback (both err AND empty-list cases).
- Security review of the repo's FIRST `os/exec` (Judge B): fixed argv, no shell, no input interpolation → no injection; `exec.CommandContext` honors the 2s timeout (kills hung process); `.Output()` discards stderr; non-zero exit → err → curated fallback. All pass.
- `detect()` called once and reused; `defer cancel()` not leaked; filter/fallback logic correct; detector/resolver key names consistent (claude/opencode); `internal/config/model.go` untouched.
- `go build`, `go test ./internal/models/...`, `go vet`, `gofmt -l` all clean.

## Confirmed fix applied (this round)
- `internal/models/opencode.go`: `parseModels` now drops a line that is empty AFTER stripping the `provider/` prefix (bare `provider/` no longer injects an empty model). Doc comment updated (first-slash-only, CRLF-safe, drops empty-after-strip).
- `internal/models/opencode_test.go`: added table cases — bare `provider/` dropped, CRLF trailing-CR stripped, namespaced `a/b/c` keeps `b/c`.
- `design.md`: softened the "composes without edits" claim to "a uniform per-catalog gate (one branch each)" — matches what `ResolveModels` actually does (Judge B INFO 4).
- Re-verified: `go build ./...`, `go vet ./internal/models/...`, `go test ./internal/models/... -count=1` all green (8 parseModels sub-cases + 3 resolver scenarios pass).

## Remaining non-blocking notes (deferred)
- INFO (Judge A): "unparseable output" relies on opencode exiting non-zero on failure; clean-but-bogus stdout could still parse — acceptable for PR1.
- PR2 (TUI) scope: cache-once + free-form scenarios (S "Detection is cached once" / "Free-form unchanged") covered there.

## State Update (PR1)
- Phase: judge (PR1 cycle)
- Status: completed

JUDGMENT: APPROVED

---

# Judge Phase Report — PR2 (TUI), final cycle

**Verdict**: pass
**Retry**: 1 / 3

## Judgment-Day Result (PR2 diff)
- Judge A: APPROVED
- Judge B: APPROVED
- No CRITICAL, no WARNING (real). Findings are INFO/SUGGESTION only.

## Gates
- Mutation gate: skipped; Playwright: skipped (not web).

## Independent verification by both judges
- S "Detection is cached once per Models view": `models.Resolve()` has exactly ONE call site (`model.go` `NewModel`); reload path (`agentInitDoneMsg`) reuses `m.modelsTab.catalog` (no re-detect); `WindowSizeMsg`/tab switches don't reconstruct the tab. Test `TestModelsTab_DetectionCachedOncePerView` proves detect runs once across repeated cycle/type.
- S "Free-form entry and advisory behavior unchanged": `applyToConfig` writes raw `.Value()` (no catalog-membership check); `Validate`/`NormalizeModel` intact. Test `TestModelsTab_FreeFormEntryUnchanged`.
- Detection never during `archon init` (grep: no `internal/models` usage in `initcmd/`/`cmd/`).
- Empty-catalog safe (`append([]string{""}, m.catalog...)` ⇒ len ≥ 1, no div-by-zero).
- `config.StaticModels()` not orphaned (still seeds `KnownModels`). `internal/models/` + `internal/config/model.go` untouched in PR2. All `newModelsTabState` callers updated.
- `go build`, `go test ./...`, `go vet`, `gofmt -l` (PR2 files) clean.

## Fix applied (this round)
- `internal/tui/models_tab.go`: hint text "cycle static models" → "cycle detected models", aligning with the renamed "Available:" label (Judge A cosmetic suggestion). Re-verified green.

## Deferred follow-ups (non-blocking, documented)
- INFO (Judge B): `models.Resolve()` is synchronous in `NewModel`, so `archon tui` startup can block up to the 2s `listTimeout` if opencode is installed-but-slow (bounded; no shellout when opencode absent). A lazy/async resolve (detect when the Models tab is first rendered, with a "detecting…" state) would remove the startup stall — deliberately deferred to a follow-up, not done in this slice.
- SUGGESTION (Judge B): a `NewModel`-level test with an injected resolver would close the last production-wiring test gap; `Resolve()` is a package-level func not parameterized into `NewModel`, so this needs a new seam — deferred.
- SUGGESTION (carried): `LookPathDetector` probes `codex`/`gemini` though only `claude`/`opencode` are consumed — intentional, forward-looking.

## State Update (PR2 / whole change)
- Phase: judge
- Status: completed (PR1 + PR2)

JUDGMENT: APPROVED
