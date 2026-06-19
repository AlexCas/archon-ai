# Verify Report: opencode-leader-mode (FINAL — whole change, PR1 + PR2)

**Change:** opencode-leader-mode
**Scope:** FINAL verify covering the WHOLE change. PR1 (core: config field + writer + init wiring + `--leader` CLI flag + tests) was verified + judged PASS and delivered as PR #41 on the parent branch. This cycle adds PR2 (TUI leader field + save-path merge + parity/update tests, Phases 5-7) and confirms all 7 phases complete and all 8 spec scenarios covered by passing tests — i.e. archive readiness.
**Branch:** `feat/opencode-leader-mode-tui` (stacked on the PR1 branch).
**Artifact store:** openspec
**Date:** 2026-06-18
**Verdict:** PASS

---

## 1. Task Completeness (all 7 phases)

| Task | Description | Status | Evidence |
|------|-------------|--------|----------|
| 1.1 | `Leader string \`yaml:"leader,omitempty"\`` on `ModelConfig` | DONE | `internal/config/model.go` |
| 1.2 | `Clone()` copies `Leader` on the `ModelConfig` literal | DONE | `internal/config/config.go` (`Leader: c.Models.Leader`) |
| 1.3 | `TestConfig_CloneRoundtrip` fixture sets `Leader` (S1) | DONE | `internal/config/config_test.go` (`anthropic/claude-sonnet-4-20250514`) |
| 2.1 | `leaderAgentName` const + `archonLeaderAgent` struct (json tags, declaration order = output order) | DONE | `internal/initcmd/opencode_mode.go` (Mode, Description, Model, Prompt) |
| 2.2 | `mergeOpencodeAgent(projectDir, leader)`: empty no-op, read-modify-write, only `agent.archon-leader`, no `default_agent`, MarshalIndent + `\n`, atomic tmp+Rename | DONE | `internal/initcmd/opencode_mode.go` |
| 3.1 | `ModelLeader` in `Options`; threaded into `buildConfig` → `ModelConfig{Leader}` | DONE | `internal/initcmd/init.go` |
| 3.2 | `Run()` calls merge after `writeTemplate`, opencode-gated, error wrapped `merge opencode agent: %w` | DONE | `internal/initcmd/init.go` |
| 3.3 | Rollback ordering fix: merge path appended to `CreatedPaths` BEFORE `WriteManifest()` | DONE | `internal/initcmd/init.go` |
| 3.4 | `--leader` CLI flag, advisory-validated, threaded into `Options.ModelLeader` | DONE | `cmd/archon/main.go` |
| 4.1 | Test: merge creates correct shape + manifest registers path (S2) | DONE | `internal/initcmd/opencode_mode_test.go` |
| 4.2 | Test: preserves pre-existing keys/agents, no `default_agent` (S3) | DONE | `internal/initcmd/opencode_mode_test.go` |
| 4.3 | Test: idempotent byte-identical (S4) | DONE | `internal/initcmd/opencode_mode_test.go` |
| 4.4 | Test: non-opencode (S5) + empty leader (S6) write nothing | DONE | `internal/initcmd/opencode_mode_test.go` |
| 4.5 | `go build` + `go test ./internal/config/... ./internal/initcmd/...` green | DONE | see §2 |
| 5.1 | TUI leader `textinput`, rendered + focus-traversed only when `cfg.Agent == "opencode"` | DONE | `internal/tui/models_tab.go` (`leaderInput`, `leaderEnabled`, `leaderInputIndex()`, appended to `inputs` only for opencode) |
| 5.2 | `applyToConfig` sets `cfg.Models.Leader` from the input (opencode only) | DONE | `internal/tui/models_tab.go` (`applyToConfig`, guarded by `leaderInputIndex() >= 0`) |
| 6.1 | `saveConfig` calls `mergeOpencodeAgent` (via exported wrapper) when `cfg.Agent == "opencode"`; surfaces errors; `archon update` untouched | DONE | `internal/tui/model.go` (`saveConfig`, calls `initcmd.MergeOpencodeAgent`) |
| 7.1 | TUI==init parity test (S7) | DONE | `internal/tui/model_test.go` (`TestSaveConfig_OpencodeLeaderMatchesInitMerge`) |
| 7.2 | `archon update` leaves opencode.json unwritten (mtime/bytes unchanged) (S8) | DONE | `internal/initcmd/update_test.go` (`TestUpdate_LeavesOpencodeJSONUntouched`) |
| 7.3 | `go build ./...` + `go test ./...` green | DONE | see §2 |

**Completeness: 19/19 tasks `[x]` across all 7 phases. No unchecked tasks. Archive-ready.**

---

## 2. Build / Test / Vet / Gofmt Evidence (real runtime output)

**`go build ./...`** → exit 0 (no output).

**`go test ./...`** → exit 0
```
ok  	github.com/archon-ai/archon/cmd/archon
ok  	github.com/archon-ai/archon/internal/agent
ok  	github.com/archon-ai/archon/internal/config
ok  	github.com/archon-ai/archon/internal/initcmd
ok  	github.com/archon-ai/archon/internal/scaffold
ok  	github.com/archon-ai/archon/internal/status
ok  	github.com/archon-ai/archon/internal/tui
ok  	github.com/archon-ai/archon/internal/version
ok  	github.com/archon-ai/archon/skills
```

**Targeted verbose — S1 (config roundtrip)** → exit 0
```
--- PASS: TestConfig_CloneRoundtrip (0.00s)
--- PASS: TestConfig_Roundtrip (0.00s)
```

**Targeted verbose — initcmd (S2-S6, S8)** → exit 0
```
--- PASS: TestMergeOpencodeAgent_CreatesAgent (0.00s)          # S2
--- PASS: TestRun_RegistersOpencodeJSONForRollback (0.00s)     # S2 (rollback registration)
--- PASS: TestMergeOpencodeAgent_PreservesExisting (0.00s)     # S3
--- PASS: TestMergeOpencodeAgent_Idempotent (0.00s)            # S4
--- PASS: TestMergeOpencodeAgent_EmptyLeaderWritesNothing (0.00s) # S6
--- PASS: TestRun_NonOpencodeWritesNoOpencodeJSON (0.00s)      # S5
--- PASS: TestUpdate_LeavesOpencodeJSONUntouched (0.00s)       # S8
ok  	github.com/archon-ai/archon/internal/initcmd	0.014s
```

**Targeted verbose — TUI (S7)** → exit 0
```
--- PASS: TestSaveConfig_OpencodeLeaderMatchesInitMerge (0.00s) # S7
ok  	github.com/archon-ai/archon/internal/tui	0.003s
```

**Full TUI / config / cmd suites (non-cached, regression check)** → exit 0
```
ok  	github.com/archon-ai/archon/internal/tui	0.540s
ok  	github.com/archon-ai/archon/internal/config	0.003s
ok  	github.com/archon-ai/archon/cmd/archon	0.018s
```

**`go vet ./...`** → exit 0 (no output).

**`gofmt -l internal/tui/models_tab.go internal/tui/model.go internal/tui/model_test.go internal/initcmd/opencode_mode.go internal/initcmd/update_test.go`** → exit 0, empty list (all formatted).

---

## 3. Spec Compliance Matrix — ALL 8 scenarios (specs/harness-init/harness-init.feature)

A scenario is compliant ONLY if a covering test PASSED at runtime. All 8 below passed.

| Scenario | Requirement | Covering test (PASSED) |
|----------|-------------|------------------------|
| S1 — Leader survives clone/round-trip | `models.leader` verbatim through Clone + serialize/reload | `TestConfig_CloneRoundtrip` (+ `TestConfig_Roundtrip`), `internal/config/config_test.go` |
| S2 — Init writes archon-leader (mode `primary`, prompt `{file:./AGENTS.md}`, model=leader) + rollback registration | `TestMergeOpencodeAgent_CreatesAgent` + `TestRun_RegistersOpencodeJSONForRollback`, `internal/initcmd/opencode_mode_test.go` |
| S3 — Merge preserves other keys/agents, no `default_agent` | `TestMergeOpencodeAgent_PreservesExisting`, `internal/initcmd/opencode_mode_test.go` |
| S4 — Idempotent byte-identical | `TestMergeOpencodeAgent_Idempotent` (`bytes.Equal`), `internal/initcmd/opencode_mode_test.go` |
| S5 — Non-opencode agent writes no opencode.json | `TestRun_NonOpencodeWritesNoOpencodeJSON`, `internal/initcmd/opencode_mode_test.go` |
| S6 — Empty leader writes nothing | `TestMergeOpencodeAgent_EmptyLeaderWritesNothing`, `internal/initcmd/opencode_mode_test.go` |
| S7 — TUI save == init merge | `TestSaveConfig_OpencodeLeaderMatchesInitMerge` (`bytes.Equal` of opencode.json from `saveConfig()` vs direct `initcmd.MergeOpencodeAgent`), `internal/tui/model_test.go` |
| S8 — `archon update` leaves opencode.json untouched | `TestUpdate_LeavesOpencodeJSONUntouched` (asserts byte AND mtime unchanged after `Update()`), `internal/initcmd/update_test.go` |

**All 8 scenarios map to a test that PASSED at runtime. Full coverage.**

---

## 4. Design Coherence (vs design.md — full change, emphasis on PR2 seams)

| Design intent | Implementation | Status |
|---------------|----------------|--------|
| Single shared `mergeOpencodeAgent` writer for init AND TUI | TUI calls exported `initcmd.MergeOpencodeAgent`, a thin wrapper that delegates to `mergeOpencodeAgent`. One writer implementation. | Coherent |
| Exported integration seam | `func MergeOpencodeAgent(projectDir, leader string) (string, error)` in `opencode_mode.go`, documented as the seam for the TUI save path | Coherent |
| TUI leader input is opencode-only (rendered + focus-traversed only when agent==opencode) | `leaderEnabled := cfg.Agent == "opencode"`; input appended to `inputs` only when enabled; `view()` and `applyToConfig` guard on `leaderInputIndex() >= 0`. Focus ring (`focusNext`/`focusPrev` mod `len(m.inputs)`) naturally includes the leader input only when present. | Coherent |
| `applyToConfig` sets `cfg.Models.Leader` only for opencode | Guarded by `leaderInputIndex() >= 0`; non-opencode leaves `cfg.Models.Leader` as loaded | Coherent |
| `saveConfig` produces byte-identical output to init | After `cfg.Save()`, opencode-gated call to the shared writer; S7 parity test proves byte-equality at runtime | Coherent |
| `archon update` untouched | No merge call added to the update path; S8 test proves opencode.json bytes + mtime unchanged after `Update()` | Coherent |
| Additive / atomic write, deterministic struct, no `default_agent`, opencode-gated, empty-leader no-op (PR1 invariants, still hold) | Unchanged from PR1; covered by S2-S6 | Coherent |

No design deviations introduced by PR2. The `--leader` CLI flag (PR1 task 3.4) remains a beneficial addition beyond design.md's original File Changes table, consistent with the existing `--model*` flag pattern and advisory-validated.

---

## 5. Issues

### CRITICAL
None.

### WARNING
None.

### SUGGESTION
- (Non-blocking) No direct `cmd`-level test asserts the `--leader` CLI flag end-to-end (flag value reaching `opencode.json`); the path is exercised indirectly via `Options.ModelLeader` and is trivial cobra plumbing. Optional future hardening.

---

## Final Verdict: PASS

The whole change is complete and verified. All 19 tasks across all 7 phases are `[x]`. `go build ./...`, `go test ./...`, `go vet ./...`, and `gofmt -l` (PR2 files) all pass with real runtime evidence. All 8 spec scenarios (S1-S8) map to a test that PASSED at runtime — including the PR2 additions S7 (TUI==init byte parity) and S8 (update leaves opencode.json untouched). Design coherence holds across all PR1 and PR2 invariants, with the exported `MergeOpencodeAgent` wrapper serving as the single shared-writer seam guaranteeing TUI==init byte-identical output. No CRITICAL or WARNING issues. The change is archive-ready.
