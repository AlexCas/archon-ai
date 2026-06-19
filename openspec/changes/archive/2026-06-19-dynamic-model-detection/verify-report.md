# Verify Report — dynamic-model-detection (FINAL — full change, PR1 + PR2)

**Scope:** Whole change — Phases 1–6 of `tasks.md`. PR1 (engine, Phases 1–4) was
verified + judged PASS and delivered as PR #43. This FINAL cycle covers PR2 (TUI
wiring, Phases 5–6) on branch `feat/dynamic-model-detection-tui` (stacked on the
PR1 branch `feat/dynamic-model-detection`) and confirms the ENTIRE change is
complete: all tasks `[x]`, all 5 spec scenarios mapped to passing tests. Ready to
archive next.

**Branch:** `feat/dynamic-model-detection-tui` (stacked on `feat/dynamic-model-detection`).
**Artifact store:** openspec.
**Verdict:** PASS (full change — PR1 + PR2).

---

## 1. Task Completeness (all phases [x])

| Task | Description | File | PR | Status |
|------|-------------|------|----|--------|
| 1.1 | `CLIDetector` type + `LookPathDetector` (opencode, claude, codex, gemini) | `internal/models/detect.go` | PR1 | DONE |
| 2.1 | `OpencodeLister` interface + `execLister` (`opencode models opencode-go`) | `internal/models/opencode.go` | PR1 | DONE |
| 2.2 | `parseModels` — split, trim, skip blanks, strip `provider/` | `internal/models/opencode.go` | PR1 | DONE |
| 3.1 | `ResolveModels(detect, lister)` — claude→curated; opencode→live w/ 2s timeout, fallback on err/empty | `internal/models/resolve.go` | PR1 | DONE |
| 3.2 | `Resolve()` = `ResolveModels(LookPathDetector, execLister{})` | `internal/models/resolve.go` | PR1 | DONE |
| 4.1 | "Installed opencode shows the live catalog" test | `internal/models/resolve_test.go` | PR1 | DONE |
| 4.2 | "Only installed agents' models are offered" test | `internal/models/resolve_test.go` | PR1 | DONE |
| 4.3 | "Live enumeration error falls back silently" test | `internal/models/resolve_test.go` | PR1 | DONE |
| 4.4 | `parseModels` table test | `internal/models/opencode_test.go` | PR1 | DONE |
| 4.5 | Verify PR1 (build/test/vet, no subprocess) | (PR1 report) | PR1 | DONE |
| 5.1 | `catalog []string` field + `newModelsTabState(cfg, catalog)` stores it | `internal/tui/models_tab.go` | PR2 | DONE |
| 5.2 | `cycleStaticModel` builds cycle list from `m.catalog` (empty-lead) | `internal/tui/models_tab.go` | PR2 | DONE |
| 5.3 | `view` hint reads `m.catalog`; label "Static:"→"Available:" | `internal/tui/models_tab.go` | PR2 | DONE |
| 5.4 | `models.Resolve()` once at open; passed to BOTH call sites | `internal/tui/model.go` | PR2 | DONE |
| 6.1 | "Detection is cached once per Models view" test | `internal/tui/models_tab_test.go` | PR2 | DONE |
| 6.2 | Cycling + "Available:" hint render from injected catalog | `internal/tui/models_tab_test.go` | PR2 | DONE |
| 6.3 | "Free-form entry and advisory behavior unchanged" test | `internal/tui/models_tab_test.go` | PR2 | DONE |
| 6.4 | Verify PR2 (build/test/vet) | this report | PR2 | DONE |

**All 18 tasks across Phases 1–6 are implemented and `[x]` in `tasks.md`. None unchecked.**

---

## 2. Runtime Evidence (real execution)

### `go build ./...`
```
build_exit=0
```
Whole module builds, including the TUI wiring against `internal/models`.

### `go test ./... -count=1` (full suite, TUI included)
```
ok  github.com/archon-ai/archon/cmd/archon          0.023s
ok  github.com/archon-ai/archon/internal/agent      0.003s
ok  github.com/archon-ai/archon/internal/config     0.005s
ok  github.com/archon-ai/archon/internal/initcmd    0.017s
ok  github.com/archon-ai/archon/internal/models     0.002s
ok  github.com/archon-ai/archon/internal/scaffold   0.007s
ok  github.com/archon-ai/archon/internal/status     0.002s
ok  github.com/archon-ai/archon/internal/tui        55.301s
ok  github.com/archon-ai/archon/internal/version    0.002s
ok  github.com/archon-ai/archon/skills              0.008s
test_exit=0
```
Entire suite green.

### PR2 TUI scenario tests (verbose)
```
=== RUN   TestModelsTab_DetectionCachedOncePerView
--- PASS: TestModelsTab_DetectionCachedOncePerView (0.00s)
=== RUN   TestModelsTab_CycleAndHintFromInjectedCatalog
--- PASS: TestModelsTab_CycleAndHintFromInjectedCatalog (0.00s)
=== RUN   TestModelsTab_FreeFormEntryUnchanged
--- PASS: TestModelsTab_FreeFormEntryUnchanged (0.00s)
PASS
ok  github.com/archon-ai/archon/internal/tui  0.004s
```

### PR1 engine scenario tests (verbose, re-run on the stacked branch)
```
--- PASS: TestParseModels (0.00s)        # 8 sub-cases incl. CRLF, bare provider/, namespaced ids
--- PASS: TestResolveModels_InstalledOpencodeShowsLiveCatalog (0.00s)
--- PASS: TestResolveModels_OnlyInstalledAgentsOffered (0.00s)
--- PASS: TestResolveModels_LiveEnumerationFallsBackToCurated (0.00s)  # error + empty sub-cases
PASS
ok  github.com/archon-ai/archon/internal/models  0.002s
```

### `go vet ./...`
```
vet_exit=0
```
Clean across the whole module.

### `gofmt -l internal/tui/`
```
internal/tui/agent_flow_test.go
internal/tui/agent_tab.go
internal/tui/judge_tab.go
internal/tui/mutation_tab.go
internal/tui/slider.go
internal/tui/slider_test.go
```
**The 4 PR2-modified TUI files are NOT in this list** — they are all formatted.
The six flagged files are PRE-EXISTING and untouched by PR2 (confirmed: PR2 touches
only `e2e_test.go`, `model.go`, `model_test.go`, `models_tab.go`). The pre-existing
formatting drift is outside this change's scope. WARNING below.

---

## 3. Spec Compliance Matrix (all 5 → passing tests)

Source: `openspec/changes/dynamic-model-detection/specs/harness-init/harness-init.feature`.

| Scenario (tag) | Mapped test | Layer | Result |
|----------------|-------------|-------|--------|
| Installed opencode shows the live catalog (@happy) | `TestResolveModels_InstalledOpencodeShowsLiveCatalog` — detector{opencode}=true + fake live list; asserts live present, curated `OpencodeModels` absent | engine (PR1) | PASS |
| Only installed agents' models are offered (@happy) | `TestResolveModels_OnlyInstalledAgentsOffered` — detector{opencode:false}; no opencode models, present-agent models remain | engine (PR1) | PASS |
| Live enumeration error falls back silently (@error) | `TestResolveModels_LiveEnumerationFallsBackToCurated` — opencode present, lister err + empty; output == `config.OpencodeModels` | engine (PR1) | PASS |
| Detection is cached once per Models view (@edge) | `TestModelsTab_DetectionCachedOncePerView` — `detectCount` proves detection runs exactly once at open; 5× cycle/type keeps count at 1; cycle reads `m.catalog`. Plus source: `models.Resolve()` invoked only at `model.go:88`, reload reuses `m.modelsTab.catalog` (`model.go:182`), never in init/cmd/scaffold | TUI (PR2) | PASS |
| Free-form entry and advisory behavior unchanged (@happy) | `TestModelsTab_FreeFormEntryUnchanged` — arbitrary value accepted; `config.Validate`/`NormalizeModel` behave as before; `internal/config/model.go` UNCHANGED in PR2 | TUI (PR2) | PASS |

All 5 spec scenarios are compliant — each maps to a test that PASSED at runtime.

---

## 4. Design Coherence (PR2)

| Design requirement | Implementation | Match |
|--------------------|----------------|-------|
| `newModelsTabState(cfg, catalog)` stores catalog on state | `catalog []string` field added; assigned `catalog: catalog` in constructor (`models_tab.go`) | YES |
| `cycleStaticModel` reads `m.catalog` (empty-lead) | `catalog := append([]string{""}, m.catalog...)` replaces `config.StaticModels()` | YES |
| hint reads `m.catalog`; label "Static:"→"Available:" | `view`: `"Available: " + strings.Join(m.catalog, ", ")` | YES |
| `models.Resolve()` computed once at open | `model.go:88` `catalog := models.Resolve()` inside `NewModel` | YES |
| Passed to BOTH call sites; reload reuses cached catalog (no re-detect) | `model.go:93` open uses `catalog`; `model.go:182` reload uses `m.modelsTab.catalog` | YES |
| Detection NOT during `archon init` | `grep` confirms no `models.Resolve`/`internal/models` import in `initcmd/`, `cmd/`, `scaffold/`; sole caller is `model.go:88` | YES |
| `internal/models/` UNCHANGED in PR2 | `git diff feat/dynamic-model-detection -- internal/models/` empty | YES |
| `internal/config/model.go` UNCHANGED in PR2 | `git diff feat/dynamic-model-detection -- internal/config/model.go` empty | YES |
| All `newModelsTabState` callers updated to new signature | `model.go` ×2, `model_test.go` ×3, `e2e_test.go` ×3, `models_tab_test.go` ×3 — all pass `(cfg, catalog)` | YES |

No deviations from `design.md` in PR2.

---

## 5. Issues

**CRITICAL:** None.

**WARNING:**
- (W1) `gofmt -l internal/tui/` flags six PRE-EXISTING files
  (`agent_flow_test.go`, `agent_tab.go`, `judge_tab.go`, `mutation_tab.go`,
  `slider.go`, `slider_test.go`). They are NOT modified by this change and the
  four PR2-touched files are correctly formatted, so this does not block the
  change. It is a latent repo hygiene issue worth a separate cleanup PR.

**SUGGESTION:**
- (S1) `LookPathDetector` still probes `codex`/`gemini` while `ResolveModels`
  consumes only `claude`/`opencode` on this branch — intentional (catalog-agnostic,
  composes with the Slice 1 Gemini/OpenAI catalog once merged). Carried over from
  PR1; no action.

---

## Verification Report

**Verdict: PASS (full change — PR1 + PR2).**

The whole `dynamic-model-detection` change is complete and verified with real
execution evidence:

- **Tasks:** all 18 tasks across Phases 1–6 are `[x]` — none unchecked.
- **Build/test/vet:** `go build ./...`, `go test ./... -count=1` (full suite incl.
  TUI), and `go vet ./...` all exit 0.
- **gofmt:** the four PR2-modified TUI files are clean; six unrelated pre-existing
  files are flagged (W1, non-blocking, out of scope).
- **Spec:** all 5 scenarios map to tests that PASSED — 3 engine (PR1) + 2 TUI (PR2).
- **Design coherence (PR2):** matches `design.md` with zero deviations. Catalog is
  stored on state; cycle and hint read `m.catalog`; label renamed to "Available:";
  `models.Resolve()` runs exactly once at view open (`model.go:88`) and the reload
  path reuses the cached catalog; detection never runs during `archon init`;
  `internal/models/` and `internal/config/model.go` are untouched in PR2; all
  `newModelsTabState` callers updated to the new signature.

No CRITICAL or blocking issues. One non-blocking WARNING (pre-existing gofmt drift
in unrelated files) and one carried-over SUGGESTION. The change is ready to proceed
to judge and then archive.
