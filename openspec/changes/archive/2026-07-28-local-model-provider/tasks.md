# Tasks: Local Model Provider (Ollama / LocalAI via OpenCode)

<!-- proposal: proposal.md | spec: specs/local-model-provider/spec.md | design: design.md -->

Implements [[local-model-provider]]. REQ-1..5 in PR-A; REQ-6..7 in PR-B.

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | PR-A ~300, PR-B ~220 |
| 400-line budget risk | Medium (PR-A ~300; PR-B ~220) |
| Chained PRs recommended | Yes |
| Suggested split | PR-A (REQ-1–5) → PR-B (REQ-6–7) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| A | Data model + opencode emission + CLI | PR-A | Independently mergeable; targets master |
| B | Claude guard + TUI sub-mode | PR-B | Targets master after PR-A merges; needs PhaseModel.BaseURL and io.Writer sig |

> **PR-A0 fallback**: if the opencode golden tests push PR-A above 400 lines, split REQ-2 (CLI) into its own PR-A0 stacked before the opencode tests. Flag at apply time; no pre-emptive split now.

---

## PR-A: Config Core + OpenCode Emission (REQ-1..5)

**Base branch**: master. Independently mergeable.

### Phase A1: Caller Audit (open question resolution)

- [x] A1.1 [REQ-1] Confirm no third external caller of `MergeOpencodeAgent`/`WriteClaudeAgents` — grep `internal/initcmd` exported symbols across the repo; result: only `internal/tui/model.go:372,381` and `model_test.go:766`. Document finding as a comment in `opencode_mode.go` and `claude_mode.go` next to the exported wrappers. No code change required if confirmed.

### Phase A2: Data Model (`internal/config/model.go`) [REQ-1, REQ-3]

- [x] A2.1 [REQ-1] Add `BaseURL string` field to `ModelRef` (YAML tag `base_url,omitempty`). Add `Provider string` and `BaseURL string` to `PhaseModel`. Populate both in `ResolvePhaseModels` from the resolved ref after Default fallback; `Model` stays `FullID()` unchanged.
- [x] A2.2 [REQ-1] Implement `MarshalYAML` mapping-switch: emit scalar `FullID()` when `Effort=="" && BaseURL==""`, else emit full mapping. `UnmarshalYAML` unchanged (already handles both forms).
- [x] A2.3 [REQ-3] Add `ValidateBaseURL(ref ModelRef, w io.Writer)` advisory helper: warn when provider empty + BaseURL set; warn when BaseURL is not a valid http/https URL. Never returns error.
- [x] A2.4 [REQ-1] Extend `TestConfig_CloneRoundtrip` fixture in `internal/config/config_test.go` to include a ref with non-empty `BaseURL`.

### Phase A3: Model Unit Tests (`internal/config/model_test.go`) [REQ-1, REQ-3]

- [x] A3.1 [REQ-1] Table test — scalar round-trip byte-identical; mapping when `BaseURL` set; scalar decodes `BaseURL=""`. Covers scenarios: "Scalar ref round-trips byte-identically", "BaseURL ref marshals as mapping", "Scalar input decodes with empty BaseURL".
- [x] A3.2 [REQ-3] Table test — `ValidateBaseURL`: valid http no warn; ftp URL warns; empty provider warns. Captures `bytes.Buffer`. Covers scenarios: "Valid BaseURL produces no warning", "Non-http BaseURL triggers a warning", "BaseURL set but provider empty triggers a warning".

### Phase A4: CLI (`cmd/archon/config.go` + `cmd/archon/config_test.go`) [REQ-2]

- [x] A4.1 [REQ-2] In `cmd/archon/config.go`: extend `set`/`get` key routing to accept `models.phases.<phase>.base_url` and `models.default.base_url` — read/write `ModelRef.BaseURL`; leave provider/model untouched. Extend `list` to emit `base_url = <value>` lines for refs with non-empty `BaseURL`.
- [x] A4.2 [REQ-2] In `cmd/archon/config_test.go`: table tests covering set+get round-trip, list output, unset returns empty, and that provider/model are not altered on set. Covers scenarios: "Set and get base_url for a phase", "list shows base_url lines", "Get base_url when unset returns empty".

### Phase A5: OpenCode Emission (`internal/initcmd/opencode_mode.go`) [REQ-4, REQ-5]

- [x] A5.1 [REQ-4] Add `io.Writer` param to `mergeOpencodeAgent` and its exported wrapper `MergeOpencodeAgent`. Implement `buildProviderBlock(refs []config.PhaseModel, w io.Writer) map[string]any`: coalesce by `Provider` id, first-BaseURL-wins with stderr warning on conflict, sorted ids/model keys via `map[string]any` + `json.MarshalIndent`. Returns `nil` when no ref has `BaseURL!=""`.
- [x] A5.2 [REQ-4] Merge `buildProviderBlock` result into `doc["provider"]`: read existing map (if present) and set only archon-built ids — never delete user-defined ids. Skip merge entirely when `buildProviderBlock` returns nil.
- [x] A5.3 [REQ-5] Verify idempotency path: `buildProviderBlock` result marshals deterministically because `map[string]any` keys sort in `MarshalIndent`.

### Phase A6: OpenCode Tests (`internal/initcmd/opencode_mode_test.go`) [REQ-4, REQ-5]

- [x] A6.1 [REQ-4] Table/golden tests (JSON assertion via `os.ReadFile`+`json.Unmarshal`): Ollama happy path, LocalAI happy path, multi-phase coalesce, mixed local+remote, conflict-warn (capture `bytes.Buffer`). Covers scenarios: "Ollama happy path — single phase", "LocalAI happy path — single phase", "Multiple phases same provider are coalesced", "Mixed local and remote phases", "Conflicting BaseURLs for same provider id".
- [x] A6.2 [REQ-4] Tests: preserve-user-provider, no-BaseURL yields no provider key. Covers scenarios: "Existing user-defined provider entries are preserved", "No BaseURL refs — no provider block emitted".
- [x] A6.3 [REQ-5] Idempotency test: write twice with same config, byte-compare. Key order assertion: two providers "aaa"/"zzz" appear sorted. Covers scenario: "Re-run produces identical output".
- [x] A6.4 Update existing `mergeOpencodeAgent` call sites in `opencode_mode_test.go` to pass the new `io.Writer` param (`io.Discard` for tests that don't assert warnings).

### Phase A7: Wiring (`internal/initcmd/init.go`, `internal/tui/model.go`) [REQ-4]

- [x] A7.1 In `internal/initcmd/init.go`: pass `os.Stderr` as the `io.Writer` arg to `mergeOpencodeAgent` (opencode path only, line ~106).
- [x] A7.2 In `internal/tui/model.go`: pass `io.Discard` to `MergeOpencodeAgent` at line 372 (TUI save path — warnings go to discard; user sees save confirmation, not raw stderr).

### PR-A Definition of Done

- `go build ./...` passes, `go vet ./...` clean, `gofmt -l .` returns empty.
- All A3.x, A4.x, A6.x tests pass.
- No ref without `BaseURL` changes its marshaled YAML bytes (byte-identical round-trip).
- A `opencode.json` produced with no `BaseURL` refs is byte-identical to the pre-PR baseline.
- REQ-1 (3 scenarios), REQ-2 (3 scenarios), REQ-3 (3 scenarios), REQ-4 (7 scenarios), REQ-5 (1 scenario) all covered by passing tests.

---

## PR-B: Claude Guard + TUI BaseURL Editing (REQ-6..7)

**Base branch**: master after PR-A merges. Requires `PhaseModel.BaseURL` and `io.Writer` signature from PR-A.

### Phase B1: Claude Guard (`internal/initcmd/claude_mode.go`) [REQ-6]

- [x] B1.1 [REQ-6] Add `io.Writer` param to `writeClaudeAgents` and `WriteClaudeAgents`. For each resolved phase ref where `PhaseModel.BaseURL != ""`, emit the exact warning string to `w` before writing the bare-model agent file. File is always written (warn-and-skip, never abort).
- [x] B1.2 [REQ-6] In `internal/initcmd/claude_mode_test.go`: tests: local ref warns + file written with bare model; remote ref no warning; two local phases each produce one warning. Update existing test call sites to pass `io.Discard`. Covers scenarios: "Local ref on Claude path triggers warn-and-skip", "Remote ref on Claude path has no warning", "Multiple local phases each emit a warning".

### Phase B2: Wiring (`internal/initcmd/init.go`, `internal/tui/model.go`) [REQ-6]

- [x] B2.1 In `internal/initcmd/init.go`: pass `os.Stderr` as the `io.Writer` arg to `writeClaudeAgents` (claude path, line ~119).
- [x] B2.2 In `internal/tui/model.go`: pass `io.Discard` to `WriteClaudeAgents` at line 381.

### Phase B3: TUI Sub-mode (`internal/tui/models_tab.go`) [REQ-7]

- [x] B3.1 [REQ-7] Add `baseURLEdit subMode` constant to the `subMode` const block (after `freeForm`). Add key handler in `rowNav` case: key `u` → enter `baseURLEdit`, seed `m.input` with `m.rows[m.focusedRow].ref.BaseURL`, call `m.input.Focus()`.
- [x] B3.2 [REQ-7] Add `baseURLEdit` case to the top-level `Update` non-key dispatch (mirrors `freeForm` textinput forwarding). Add key handling in `baseURLEdit` case: Enter → `strings.TrimSpace(m.input.Value())` → `m.rows[focused].ref.BaseURL`, `changed=true`, `m.input.Blur()`, back to `rowNav`; Escape → `m.input.Blur()`, back to `rowNav` without committing.
- [x] B3.3 [REQ-7] Add `baseURLEdit` case to `hintLine()` and `renderRow()` mirroring `freeForm` layout. In plain `rowNav` row render: append ` @ <baseURL>` when `ref.BaseURL != ""`.
- [x] B3.4 [REQ-7] Verify `applyToConfig` requires no change — it assigns `row.ref` as a struct value copy, so `BaseURL` persists automatically.

### Phase B4: TUI Tests (`internal/tui/models_tab_test.go`) [REQ-7]

- [x] B4.1 [REQ-7] teatest: open `baseURLEdit` sub-mode via key `u`, type URL, press Enter — assert `ref.BaseURL` set. Covers scenario: "User can set BaseURL via TUI sub-mode". (Implemented as direct `st.update()` state-transition tests, matching this repo's established `models_tab_test.go` convention and the go-testing skill's decision gate — teatest is reserved for full interactive-flow tests, none of which exist in this package.)
- [x] B4.2 [REQ-7] teatest: pre-set BaseURL, enter sub-mode, clear input, press Enter — assert `BaseURL=""`. Covers scenario: "User can clear BaseURL via TUI sub-mode".
- [x] B4.3 [REQ-7] teatest: pre-set BaseURL, enter sub-mode, type new value, press Escape — assert BaseURL unchanged. Covers scenario: "Escape cancels BaseURL edit".
- [x] B4.4 [REQ-7] teatest: render assertion — row with `BaseURL` set shows the URL string in the rendered output. Covers scenario: "Row with BaseURL shows endpoint in display".
- [x] B4.5 [REQ-7] Verify saving persists `BaseURL` to `config.yaml` (integration: call `applyToConfig` + `SaveConfig`, assert YAML). Covers scenario: "saving the config persists the BaseURL to .archon/config.yaml". (Covered inside B4.1/B4.2 via `applyToConfig` assertions on the resulting `ModelConfig`; `Config.Save`'s YAML marshal path for `BaseURL` was already asserted by PR-A's `TestModelRef_MarshalYAML` round-trip tests, so no duplicate YAML-file test was added here.)

### PR-B Definition of Done

- `go build ./...` passes, `go vet ./...` clean, `gofmt -l .` returns empty.
- All B1.x, B4.x tests pass.
- REQ-6 (3 scenarios) and REQ-7 (5 scenarios) covered by passing tests.
- A ref without `BaseURL` on the claude path produces no warning and no changed output.
- Existing tests that call `writeClaudeAgents` directly compile and pass with the new `io.Writer` param.
