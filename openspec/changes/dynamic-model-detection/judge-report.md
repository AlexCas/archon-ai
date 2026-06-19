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
