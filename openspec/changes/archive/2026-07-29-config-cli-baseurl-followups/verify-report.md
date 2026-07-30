# Verification Report: config CLI base_url follow-ups (#90, #91)

<!-- proposal: proposal.md | spec: specs/local-model-provider/spec.md | design: design.md | tasks: tasks.md -->

- **Change**: `config-cli-baseurl-followups`
- **Mode**: full spec-driven verify (proposal + [spec](specs/local-model-provider/spec.md) + [design](design.md) + [tasks](tasks.md) all present)
- **Capability**: [[local-model-provider]] delta (REQ-8..REQ-12)
- **Verdict**: **PASS** — ready for judge

## Completeness

| Dimension | Result |
|-----------|--------|
| Tasks checked | 26/26 (`tasks.md` Phases 1–9 all `[x]`) |
| Unchecked implementation tasks | none |
| Requirements verified | REQ-8, REQ-9, REQ-10, REQ-11, REQ-12 |
| Gherkin scenarios | 24/24 covered by passing tests |

## Build / Test Evidence

| Command | Exit | Result |
|---------|------|--------|
| `gofmt -l` (6 touched files) | 0 | empty output — all formatted |
| `go vet ./...` | 0 | no issues |
| `go build ./...` | 0 | compiles clean |
| `go test ./...` | 0 | all packages `ok` |
| `go test -count=1 ... ./internal/config/... ./cmd/archon/... ./internal/status/...` | 0 | fresh run, all relevant tests PASS |

## Spec Compliance Matrix

| REQ | Scenario(s) | Evidence (code) | Covering test | Status |
|-----|-------------|-----------------|---------------|--------|
| REQ-8 | base_url-only default/phase configured; genuinely-empty still `(none configured)`; status both cases | `internal/config/model.go:127-140` (`HasAny`), guard swap `cmd/archon/config.go:127`, `internal/status/display.go:69` | `TestModelConfig_HasAny` (8 rows), `TestConfigCmd_ListBaseURLOnlyIsConfigured`, `TestConfigCmd_ListEmpty` (regression, unchanged), `TestDisplay_ModelsNoneConfigured` | PASS |
| REQ-9 | empty-id ref emits only base_url line; both-set emits both | primary suppression `cmd/archon/config.go:132-137,153-158` | `TestConfigCmd_ListSuppressesEmptyPrimary` (2 sub) | PASS |
| REQ-10 | status base_url for default/phase; suppress empty primary | `internal/status/display.go:72-77,100-105` | `TestDisplay_ModelsBaseURLLines` (3 sub) | PASS |
| REQ-11 | leader set/get/list round-trip; advisory non-blocking; unknown key; supported-keys updated in set+get | set arms `config.go:200-204`, get arms `:300-303`, `baseURLRefForKey` leader case `:174-175`, list block `:139-144`, error strings `:290,:341` | `TestConfigCmd_LeaderSetGet`, `TestConfigCmd_ListShowsLeaderBlock`, `TestConfigCmd_LeaderBaseURLAdvisory`, `TestConfigCmd_LeaderUnknownKey` | PASS |
| REQ-12 | status Leader block symmetric; present on base_url-only; omitted when empty | `internal/status/display.go:79-91` | `TestDisplay_LeaderBlock` (3 sub), `TestConfigCmd_LeaderSymmetry` | PASS |

## Deviations (both spec-correct, NOT defects)

1. **status base_url-only leader renders on the `Leader:` label line** (`display.go:85-90`): the else-branch prints `Leader:  <BaseURL>` rather than a suppressed id line + `leader base_url:` sub-line as the design task-6.1 snippet showed. The REQ-12 Gherkin scenario "status shows Leader block when only base_url is set" requires: `Leader:` present, URL present, no blank leader model-id line. All three hold. Spec wins over the design snippet. Verified by `TestDisplay_LeaderBlock/leader with only base_url has no blank id line`.
2. **`TestConfigCmd_LeaderSymmetry` lives in `cmd/archon/config_test.go`** (not `display_test.go`) because package `status` cannot import package `main`; `cmd/archon` imports `internal/status` and calls `status.Format(cfg)` directly. Coverage is equivalent (arguably stronger — full CLI set→save→reload→status round-trip). Matches the design testing table's own name.

## Correctness / Invariants

| Invariant | Result |
|-----------|--------|
| Guard strictly additive; `(none configured)` only when genuinely empty | Confirmed — `TestConfigCmd_ListEmpty` unchanged & passing |
| Primary-line suppression IFF `FullID() != ""`; base_url line independent | Confirmed on both surfaces |
| Advisory-only `ValidateBaseURL` (warn-to-stderr, non-blocking, exit 0, value saved) | Confirmed — `config.go:56-70` writes then warns then `Save()` returns nil; `TestConfigCmd_LeaderBaseURLAdvisory` |
| Symmetric surfaces (list ↔ status) | Confirmed — `TestConfigCmd_LeaderSymmetry` |
| No schema/Clone/Validate change | Confirmed — only additive `HasAny()` in `internal/config` (`config.go` not in diff; `Clone`/`Validate` untouched) |
| Ordering default → leader → phases | Confirmed both surfaces (`config.go:132-160`, `display.go:72-107`) |

## Design Coherence

| Decision | Result |
|----------|--------|
| Two surfaces agree via single `HasAny()` source of truth | Coherent |
| No `internal/config` schema/Clone/Validate change | Coherent (design line 14 invariant honored) |
| Deviation #1 (Leader label rendering) | design snippet superseded by spec Gherkin; WARNING-class, no spec break |

## Issues

- **CRITICAL**: none.
- **WARNING**: none blocking.
- **SUGGESTION**: the feature scenario `leader guard included in broadened emptiness check` has no dedicated end-to-end CLI test (leader-base_url-only → `config list` omits `(none configured)`). It is proven compositionally by `TestModelConfig_HasAny/leader_BaseURL_only`→true plus the `!HasAny()` list guard. Optional to add for completeness; not a blocker.

## Verdict

**PASS** — implementation matches REQ-8..REQ-12 and all 24 Gherkin scenarios with passing runtime tests; both apply-flagged deviations are spec-correct. Ready for the judge phase.
