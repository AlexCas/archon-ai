# Tasks: config CLI base_url follow-ups (#90, #91)

<!-- proposal: proposal.md | spec: specs/local-model-provider/spec.md | design: design.md -->

Implements the [[local-model-provider]] delta (REQ-8..REQ-12). All five requirements
ship in a single PR (~265 estimated lines, under the 400-line budget).

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~265 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | single-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: single-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | HasAny() + guard swaps + leader routing + render + all tests | PR-1 | Base = master; additive presentation only; no schema change |

---

## Phase 1: HasAny() helper (`internal/config/model.go`) [REQ-8]

- [x] 1.1 [REQ-8] Add `func (mc ModelConfig) HasAny() bool` to `internal/config/model.go` immediately after the `ModelConfig` type definition. Returns `true` if any of the following is non-empty: `Default.FullID()`, `Default.BaseURL`, `Leader.FullID()`, `Leader.BaseURL`, or any entry in `Phases` (`FullID() != ""` OR `BaseURL != ""`). Returns `false` otherwise. Add the doc comment from `design.md` verbatim.

## Phase 2: HasAny() unit tests (`internal/config/model_test.go`) [REQ-8]

- [x] 2.1 [REQ-8] Add a table-driven test `TestModelConfig_HasAny` in `internal/config/model_test.go`. Required rows: all-empty → false; default-id-only → true; default-BaseURL-only → true; leader-id-only → true; leader-BaseURL-only → true; phase-id-only → true; phase-BaseURL-only → true; all-fields-set → true.

## Phase 3: config list — guard swap and suppression (`cmd/archon/config.go`) [REQ-8, REQ-9]

- [x] 3.1 [REQ-8] Replace the emptiness guard at `config.go:127-128`. Change `cfg.Models.Default.FullID() == "" && len(cfg.Models.Phases) == 0` to `!cfg.Models.HasAny()`. The `(none configured)` branch body is unchanged.
- [x] 3.2 [REQ-9] Wrap the default primary line (`config.go:132-133`) in `if cfg.Models.Default.FullID() != ""` so the `models.default = ` line is suppressed when `FullID()` is empty. The `base_url` line at `:135-137` is already conditional on `BaseURL != ""`; leave it unchanged.
- [x] 3.3 [REQ-9] In the phase loop (`config.go:144-150`), wrap the primary line `fmt.Fprintf(stdout, "models.phases.%s = %s\n", ...)` in `if cfg.Models.Phases[phase].FullID() != ""` so it is suppressed when the phase ref has only a `base_url`.

## Phase 4: config list — leader block (`cmd/archon/config.go`) [REQ-11]

- [x] 4.1 [REQ-11] Insert a leader render block between the default block and the phases loop (after the `Default.BaseURL` block, before `if len(cfg.Models.Phases) > 0`). Logic: if `cfg.Models.Leader.FullID() != ""`, emit `models.leader = <FullID()>\n`; if `cfg.Models.Leader.BaseURL != ""`, emit `models.leader.base_url = <BaseURL>\n`. No output when both are empty.
- [x] 4.2 [REQ-11] Add `case "models.leader":` arm to `setConfigValue` (after the `models.default.base_url` case, before `case strings.HasPrefix(key, "models.phases.")`). Parse value with `config.ParseModelRef(value)` and assign to `cfg.Models.Leader`. Mirror the `models.default` arm exactly.
- [x] 4.3 [REQ-11] Add `case "models.leader.base_url":` arm to `setConfigValue` (immediately after 4.2). Assign `value` to `cfg.Models.Leader.BaseURL`. Mirror the `models.default.base_url` arm.
- [x] 4.4 [REQ-11] Add `case "models.leader":` arm to `getConfigValue` (after `models.default.base_url`, before phases prefix). Return `cfg.Models.Leader.FullID(), nil`.
- [x] 4.5 [REQ-11] Add `case "models.leader.base_url":` arm to `getConfigValue` (immediately after 4.4). Return `cfg.Models.Leader.BaseURL, nil`.
- [x] 4.6 [REQ-11] Add `case key == "models.leader.base_url": return cfg.Models.Leader, true` to `baseURLRefForKey` (`config.go:161-171`), between the `models.default.base_url` case and the `models.phases.` prefix case. This routes advisory `ValidateBaseURL` through the existing set-time path at `:62-66` — no new advisory call needed.
- [x] 4.7 [REQ-11] Update the supported-keys error string in `setConfigValue` (`:273`): insert `models.leader, models.leader.base_url` immediately after `models.default.base_url`. Update the equivalent string in `getConfigValue` (`:320`) the same way.

## Phase 5: status — guard swap, suppression, and base_url lines (`internal/status/display.go`) [REQ-8, REQ-10]

- [x] 5.1 [REQ-8] Replace the emptiness guard at `display.go:69`. Change `cfg.Models.Default.FullID() == "" && len(cfg.Models.Phases) == 0` to `!cfg.Models.HasAny()`. The `(none configured)` branch body is unchanged.
- [x] 5.2 [REQ-10] Wrap the Default primary line (`display.go:72-73`) in `if cfg.Models.Default.FullID() != ""`. Immediately after, add: `if cfg.Models.Default.BaseURL != "" { fmt.Fprintf(w, "    default base_url:  %s\n", cfg.Models.Default.BaseURL) }`. This mirrors the `config list` surface.
- [x] 5.3 [REQ-10] In the phase loop (`display.go:76-83`), wrap the primary line `fmt.Fprintf(w, "    %-8s %s\n", phase+":", ...)` in `if cfg.Models.Phases[phase].FullID() != ""`. After that primary line, add: `if baseURL := cfg.Models.Phases[phase].BaseURL; baseURL != "" { fmt.Fprintf(w, "    %-8s %s\n", "base_url:", baseURL) }`.

## Phase 6: status — leader block (`internal/status/display.go`) [REQ-12]

- [x] 6.1 [REQ-12] Insert a `Leader:` block between the Default block and the phases block (after the Default/base_url render, before `if len(cfg.Models.Phases) > 0`). Logic: render the block only when `cfg.Models.Leader.FullID() != "" || cfg.Models.Leader.BaseURL != ""`. Within the block: if `cfg.Models.Leader.FullID() != ""`, emit `    Leader:  <FullID()>\n`; if `cfg.Models.Leader.BaseURL != ""`, emit `    leader base_url:  <BaseURL>\n`. Use the same two-space `Default:  ` alignment style (symmetric with Default per REQ-12).

## Phase 7: config_test.go — REQ-8, REQ-9, REQ-11 tests (`cmd/archon/config_test.go`)

- [x] 7.1 [REQ-8] Add `TestConfigCmd_ListBaseURLOnlyIsConfigured`: set only `models.default.base_url http://localhost:11434/v1`, run `config list`, assert output does NOT contain `(none configured)` and DOES contain `models.default.base_url = http://localhost:11434/v1`. Add a parallel sub-case for `models.phases.apply.base_url`.
- [x] 7.2 [REQ-8] Verify `TestConfigCmd_ListEmpty` (`:508`) still passes unchanged — all-empty config produces exactly `(none configured)`.
- [x] 7.3 [REQ-9] Add `TestConfigCmd_ListSuppressesEmptyPrimary`: (a) base_url-only default → output contains `models.default.base_url = `, does NOT contain `models.default = `; (b) phase with both id and base_url → output contains both `models.phases.apply = ollama/llama3` and `models.phases.apply.base_url = `.
- [x] 7.4 [REQ-11] Add `TestConfigCmd_LeaderSetGet`: set `models.leader ollama/llama3`, then `models.leader.base_url http://localhost:11434/v1`; assert `get models.leader` returns `ollama/llama3` and `get models.leader.base_url` returns `http://localhost:11434/v1`. Assert that setting `models.leader.base_url` does not alter the provider/model fields.
- [x] 7.5 [REQ-11] Add `TestConfigCmd_ListShowsLeaderBlock`: configure both leader id and base_url; run `config list`; assert `models.leader = ollama/llama3` appears, `models.leader.base_url = http://localhost:11434/v1` appears, and the leader block appears before any `models.phases.` line.
- [x] 7.6 [REQ-11] Add `TestConfigCmd_LeaderBaseURLAdvisory`: set `models.leader.base_url ftp://bad-url`; assert exit code 0; assert stderr contains a warning; assert `get models.leader.base_url` returns `ftp://bad-url` (value is stored despite advisory warning).
- [x] 7.7 [REQ-11] Add `TestConfigCmd_LeaderUnknownKey`: run `set models.leader.typo somevalue` and `get models.leader.typo`; assert both exit non-zero; assert stderr contains `models.leader` and `models.leader.base_url` in the supported-keys message.

## Phase 8: display_test.go — REQ-8, REQ-10, REQ-12 tests (`internal/status/display_test.go`)

- [x] 8.1 [REQ-8] Add `TestDisplay_ModelsNoneConfigured`: all-empty config → Models block contains `(none configured)`. Add sub-case: base_url-only default → Models block does NOT contain `(none configured)`, DOES contain the URL.
- [x] 8.2 [REQ-10] Add `TestDisplay_ModelsBaseURLLines`: (a) default with both id and base_url → output contains `Default:` label and URL; (b) default with only base_url → no blank id line, URL present; (c) phase with base_url → output contains phase label and URL.
- [x] 8.3 [REQ-12] Add `TestDisplay_LeaderBlock`: (a) leader with id+base_url → `Leader:` present, id and base_url values present, block appears after Default before phases; (b) leader base_url-only → `Leader:` present, no blank id line, URL present; (c) genuinely empty leader → `Leader:` absent.
- [x] 8.4 [REQ-12] Add `TestConfigCmd_LeaderSymmetry` (symmetry cross-check) in `cmd/archon/config_test.go` — placed here rather than `display_test.go` because `internal/status` (package `status`) cannot call back into `cmd/archon` (package `main`); `cmd/archon` can import `internal/status` directly. Configures leader id+base_url; runs `config list` and `status.Format()`; asserts both surfaces show the same leader id and base_url values.

## Phase 9: Verification

- [x] 9.1 Run `go build ./...` — must compile cleanly with no errors.
- [x] 9.2 Run `go test ./internal/config/... ./cmd/archon/... ./internal/status/...` — all tests must pass, including the `TestConfigCmd_ListEmpty` regression guard.
- [x] 9.3 Run `gofmt -l ./internal/config/model.go ./cmd/archon/config.go ./internal/status/display.go ./internal/config/model_test.go ./cmd/archon/config_test.go ./internal/status/display_test.go` — must return empty (no unformatted files).
- [x] 9.4 Run `go vet ./...` — must report no issues.

---

## Definition of Done

- `go build ./...` passes, `go vet ./...` clean, `gofmt -l .` returns empty.
- All tests in Phases 2, 7, and 8 pass, including `TestConfigCmd_ListEmpty` (`:508`).
- `config list` and `status` agree on which sub-blocks appear for default, leader, and phases (Invariant 4 / REQ-12 symmetry).
- A ref without `BaseURL` and without leader set produces no change to existing `config list` or `status` output (backward-compatible).
- REQ-8 (guard broaden), REQ-9 (primary suppression in list), REQ-10 (base_url lines in status), REQ-11 (leader set/get/list/advisory), REQ-12 (Leader block in status) all covered by passing tests.
