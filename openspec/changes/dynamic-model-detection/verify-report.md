# Verify Report — dynamic-model-detection (PR1 / engine cycle)

**Scope:** PR1 (engine) ONLY — Phases 1–4 of `tasks.md`. PR2 (TUI, Phases 5–6) is
chained delivery and intentionally NOT implemented yet. Its unchecked tasks and
the two TUI-scoped spec scenarios are EXPECTED and marked **deferred to PR2**,
not blocking findings. No archive in this cycle.

**Branch:** `feat/dynamic-model-detection` (off `master`).
**Artifact store:** openspec.
**Verdict:** PASS (PR1).

---

## 1. Task Completeness (PR1: Phases 1–4)

| Task | Description | File | Status |
|------|-------------|------|--------|
| 1.1 | `CLIDetector` type + `LookPathDetector` iterating `opencode, claude, codex, gemini` via `exec.LookPath` | `internal/models/detect.go` | DONE |
| 2.1 | `OpencodeLister` interface (`List(ctx)`) + `execLister` running `exec.CommandContext(ctx,"opencode","models","opencode-go")` | `internal/models/opencode.go` | DONE |
| 2.2 | `parseModels([]byte) []string` — split, trim, skip blanks, strip `provider/` | `internal/models/opencode.go` | DONE |
| 3.1 | `ResolveModels(detect, lister)` — claude→curated; opencode→live `List` w/ 2s timeout, fallback to curated on err/empty | `internal/models/resolve.go` | DONE |
| 3.2 | `Resolve()` convenience = `ResolveModels(LookPathDetector, execLister{})` | `internal/models/resolve.go` | DONE |
| 4.1 | "Installed opencode shows the live catalog" test | `internal/models/resolve_test.go` | DONE |
| 4.2 | "Only installed agents' models are offered" test | `internal/models/resolve_test.go` | DONE |
| 4.3 | "Live enumeration error falls back silently" test (err + empty) | `internal/models/resolve_test.go` | DONE |
| 4.4 | `parseModels` table test | `internal/models/opencode_test.go` | DONE |
| 4.5 | Verify PR1: build/test/vet (no subprocess) | this report | DONE |

**PR2 (deferred — NOT in scope, expected unchecked):** Phase 5 (5.1–5.4 TUI wiring)
and Phase 6 (6.1–6.4 TUI tests + verify) remain `[ ]`. Deferred to PR2 per chained
delivery plan; not a finding.

All 10 PR1 tasks implemented and `[x]` in `tasks.md`.

---

## 2. Runtime Evidence (real execution)

### `go build ./...`
```
exit=0
```
Whole module builds, including the new `internal/models/` package.

### `go test -v ./internal/models/... -count=1`
```
=== RUN   TestParseModels
=== RUN   TestParseModels/strips_provider_prefix
=== RUN   TestParseModels/skips_blank_and_whitespace-only_lines
=== RUN   TestParseModels/keeps_no-slash_lines_as-is
=== RUN   TestParseModels/trims_surrounding_whitespace
=== RUN   TestParseModels/empty_input_yields_nil
--- PASS: TestParseModels (0.00s)
=== RUN   TestResolveModels_InstalledOpencodeShowsLiveCatalog
--- PASS: TestResolveModels_InstalledOpencodeShowsLiveCatalog (0.00s)
=== RUN   TestResolveModels_OnlyInstalledAgentsOffered
--- PASS: TestResolveModels_OnlyInstalledAgentsOffered (0.00s)
=== RUN   TestResolveModels_LiveEnumerationFallsBackToCurated
=== RUN   TestResolveModels_LiveEnumerationFallsBackToCurated/error_from_lister
=== RUN   TestResolveModels_LiveEnumerationFallsBackToCurated/empty_result
--- PASS: TestResolveModels_LiveEnumerationFallsBackToCurated (0.00s)
PASS
ok  	github.com/archon-ai/archon/internal/models	0.002s
exit=0
```
All tests PASS. The three engine scenario tests plus the `parseModels` table test
(5 sub-cases) all green.

### `go vet ./internal/models/...`
```
exit=0
```
Clean.

### `gofmt -l internal/models/`
```
(no output)
exit=0
```
All files formatted.

### No real subprocess in tests
```
grep -nE "exec\.|Resolve\(\)|execLister" internal/models/*_test.go
grep_exit=1   # no matches
```
Confirmed: test files contain no `exec.*`, no `Resolve()`, and no `execLister`
usage. Tests inject a `fakeLister` and inline detector funcs only — no real
`opencode` binary is invoked.

---

## 3. Spec Compliance Matrix

Source: `specs/harness-init/harness-init.feature`.

| Scenario (tag) | Mapped test | Result |
|----------------|-------------|--------|
| Installed opencode shows the live catalog (@happy) | `TestResolveModels_InstalledOpencodeShowsLiveCatalog` — detector{opencode,claude}=true, fake live list; asserts live present, curated `OpencodeModels` absent, claude curated present | PASS |
| Only installed agents' models are offered (@happy) | `TestResolveModels_OnlyInstalledAgentsOffered` — detector{opencode:false}; asserts no live & no curated opencode, claude remains | PASS |
| Live enumeration error falls back silently (@error) | `TestResolveModels_LiveEnumerationFallsBackToCurated` — opencode present, lister err + empty cases; asserts output == `config.OpencodeModels` | PASS |
| (engine correctness for live source) | `TestParseModels` table — prefix strip, blanks, no-slash, trim, empty→nil | PASS |
| Detection is cached once per Models view (@edge) | — | **DEFERRED to PR2** (TUI cache at view open; Phase 6.1) |
| Free-form entry and advisory behavior unchanged (@happy) | — | **DEFERRED to PR2** (TUI input / `NormalizeModel`/`Validate`; Phase 6.3) |

The two deferred scenarios are TUI-scoped (`internal/tui/`), correctly out of PR1's
pure-engine scope. Not failures.

---

## 4. Design Coherence (vs `design.md`)

| Design requirement | Implementation | Match |
|--------------------|----------------|-------|
| Pure `internal/models/` package, `os/exec` out of `config` | New package; only `detect.go`/`opencode.go` import `os/exec` | YES |
| Injectable `CLIDetector` func | `type CLIDetector func() map[string]bool` + `LookPathDetector` | YES |
| Injectable `OpencodeLister` interface | `type OpencodeLister interface { List(ctx) ([]string, error) }` + `execLister` | YES |
| `ResolveModels` filter: absent CLI → skip (hidden, not curated) | claude/opencode gated on `present[...]`; absent → not appended | YES |
| opencode: live-with-curated-fallback (err/empty → `config.OpencodeModels`) | `if live,err:=List(); err==nil && len(live)>0 { use live } else { curated }` | YES |
| `exec.CommandContext` with 2s timeout | `context.WithTimeout(..., listTimeout)` where `listTimeout = 2*time.Second` | YES |
| `parseModels` strips `provider/` prefix | `IndexByte(ln,'/')` → `ln[i+1:]`; trims; skips blanks; keeps no-slash | YES |
| `config/model.go` UNCHANGED | `git status` shows no change to `internal/config/model.go` | YES |
| No real subprocess in tests | grep confirms no `exec.`/`Resolve()`/`execLister` in `*_test.go` | YES |
| `Resolve()` = `ResolveModels(LookPathDetector, execLister{})` | Exact | YES |

No deviations from design.

---

## 5. Issues

**CRITICAL:** None.

**WARNING:** None.

**SUGGESTION:**
- (S1) PR1 work is currently uncommitted/untracked in the working tree (`git status`
  shows `?? internal/models/`); the active branch `feat/dynamic-model-detection`
  points at the same commit as `master`. Functionally correct for verification, but
  the engine code and tests should be committed on the feature branch before PR1 is
  opened. Non-blocking for the verify gate.
- (S2) `LookPathDetector` probes `codex` and `gemini` too, but `ResolveModels`
  only consumes `claude`/`opencode` on this branch. This is intentional
  (catalog-agnostic, composes with Slice 1) and documented; no action needed.

---

## Final Verdict

**PASS (PR1 / engine).** All Phase 1–4 tasks implemented and checked. `go build`,
`go test`, `go vet`, and `gofmt` all clean with real execution evidence. The three
in-scope engine spec scenarios plus the `parseModels` table test map to passing
tests. Implementation matches `design.md` with zero deviations; `config/model.go`
is untouched; tests inject fakes and never shell out. PR2 (TUI, Phases 5–6) and the
two TUI-scoped scenarios are correctly deferred to the next chained PR — not
blocking findings. Only non-blocking suggestions noted (commit the work on the
feature branch before opening PR1).
