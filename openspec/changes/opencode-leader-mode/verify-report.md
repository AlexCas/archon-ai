# Verify Report: opencode-leader-mode (PR1 cycle — core)

**Change:** opencode-leader-mode
**Scope:** PR1 (core: config field + writer + init wiring + tests + `--leader` CLI flag). PR2 (TUI, Phases 5-7) is intentionally deferred and NOT verified here.
**Date:** 2026-06-18
**Verdict:** PASS

---

## 1. Task Completeness (PR1: Phases 1-4 + task 3.4)

| Task | Description | Status | Evidence |
|------|-------------|--------|----------|
| 1.1 | `Leader string \`yaml:"leader,omitempty"\`` on `ModelConfig` | DONE | `internal/config/model.go:11` |
| 1.2 | `Clone()` copies `Leader` on the `ModelConfig` literal | DONE | `internal/config/config.go:96` (`Leader: c.Models.Leader`) |
| 1.3 | `TestConfig_CloneRoundtrip` fixture sets `Leader` | DONE | `internal/config/config_test.go:199` (`anthropic/claude-sonnet-4-20250514`) |
| 2.1 | `leaderAgentName` const + `archonLeaderAgent` struct (json tags, declaration order = output order) | DONE | `internal/initcmd/opencode_mode.go:12,16-21` (Mode, Description, Model, Prompt) |
| 2.2 | `mergeOpencodeAgent(projectDir, leader)`: empty no-op, map read-modify-write, only `agent.archon-leader`, no `default_agent`, MarshalIndent + `\n`, atomic tmp+Rename | DONE | `internal/initcmd/opencode_mode.go:33-76` |
| 3.1 | `ModelLeader` in `Options`; threaded into `buildConfig` → `ModelConfig{Leader}` | DONE | `init.go:24, 85, 207, 234` |
| 3.2 | `Run()` calls merge after `writeTemplate`, opencode-gated, error wrapped `merge opencode agent: %w` | DONE | `init.go:101-109` |
| 3.3 | Rollback ordering fix: merge path appended to `CreatedPaths` BEFORE `WriteManifest()`; `WriteManifest()` after merge | DONE | `init.go:107, 111` (manifest built at 91, written at 111 after merge) |
| 3.4 | `--leader` CLI flag, advisory-validated, threaded into `Options.ModelLeader` | DONE | `cmd/archon/main.go:90, 141-143, 153, 195` |
| 4.1 | Test: merge creates correct shape + manifest registers path (S2) | DONE | `opencode_mode_test.go:46-121` (2 tests) |
| 4.2 | Test: preserves pre-existing keys/agents, no `default_agent` (S3) | DONE | `opencode_mode_test.go:125-174` |
| 4.3 | Test: idempotent byte-identical (S4) | DONE | `opencode_mode_test.go:177-199` |
| 4.4 | Test: non-opencode (S5) + empty leader (S6) write nothing | DONE | `opencode_mode_test.go:201-247` |
| 4.5 | `go build` + `go test` green | DONE | see §2 |

**PR2 (deferred, expected unchecked — NOT a finding):** Phase 5 (5.1, 5.2 TUI leader field), Phase 6 (6.1 save-path merge), Phase 7 (7.1, 7.2, 7.3 parity/update tests). Correctly left `[ ]` for the stacked PR2 cycle.

---

## 2. Build / Test / Vet / Gofmt Evidence (real output)

**`go build ./...`** → exit 0 (no output).

**`go test ./cmd/... ./internal/config/... ./internal/initcmd/...`** → exit 0
```
ok  	github.com/archon-ai/archon/cmd/archon	(cached)
ok  	github.com/archon-ai/archon/internal/config	(cached)
ok  	github.com/archon-ai/archon/internal/initcmd	(cached)
```

**Targeted verbose run (PR1 scenario tests)** → exit 0
```
--- PASS: TestMergeOpencodeAgent_CreatesAgent (0.00s)
--- PASS: TestRun_RegistersOpencodeJSONForRollback (0.00s)
--- PASS: TestMergeOpencodeAgent_PreservesExisting (0.00s)
--- PASS: TestMergeOpencodeAgent_Idempotent (0.00s)
--- PASS: TestMergeOpencodeAgent_EmptyLeaderWritesNothing (0.00s)
--- PASS: TestRun_NonOpencodeWritesNoOpencodeJSON (0.00s)
--- PASS: TestConfig_CloneRoundtrip (0.00s)
```

**`go vet ./cmd/... ./internal/config/... ./internal/initcmd/...`** → exit 0 (no output).

**`gofmt -l cmd/archon/main.go internal/initcmd/opencode_mode.go internal/initcmd/init.go internal/config/model.go internal/config/config.go`** → exit 0, empty list (all formatted).

---

## 3. Spec Compliance Matrix (specs/harness-init/harness-init.feature)

| Scenario | Requirement | Covering test | Result |
|----------|-------------|---------------|--------|
| S1 — Leader survives clone/round-trip | `models.leader` verbatim through Clone + serialize/reload | `TestConfig_CloneRoundtrip` (config_test.go:181, DeepEqual incl. Leader) | PASS |
| S2 — Init writes archon-leader (mode primary, prompt `{file:./AGENTS.md}`, model=leader) + rollback registration | `TestMergeOpencodeAgent_CreatesAgent` + `TestRun_RegistersOpencodeJSONForRollback` | PASS |
| S3 — Merge preserves other keys/agents, no `default_agent` | `TestMergeOpencodeAgent_PreservesExisting` | PASS |
| S4 — Idempotent byte-identical | `TestMergeOpencodeAgent_Idempotent` (`bytes.Equal`) | PASS |
| S5 — Non-opencode agent writes no opencode.json | `TestRun_NonOpencodeWritesNoOpencodeJSON` | PASS |
| S6 — Empty leader writes nothing | `TestMergeOpencodeAgent_EmptyLeaderWritesNothing` | PASS |
| S7 — TUI save == init merge | (PR2 scope) | Not yet covered — deferred to PR2 |
| S8 — `archon update` leaves opencode.json untouched | (PR2 scope) | Not yet covered — deferred to PR2 |

All six PR1-relevant scenarios (S1-S6) map to a test that PASSED at runtime. S7/S8 require the TUI/save path landing in PR2 and are correctly out of scope here.

---

## 4. Design Coherence (vs design.md)

| Design intent | Implementation | Status |
|---------------|----------------|--------|
| Single shared `mergeOpencodeAgent` | One function in `opencode_mode.go`, called from `Run()` (init). TUI call site is PR2. | Coherent (PR1 portion) |
| Additive / atomic write | Read into `map[string]any`, set only `agent.archon-leader`, `.tmp` + `os.Rename` | Coherent (opencode_mode.go:40-73) |
| Deterministic struct (declaration order = output) | `archonLeaderAgent{Mode, Description, Model, Prompt}` fixed-field struct; doc via `MarshalIndent` (sorted map keys) + `\n` | Coherent |
| No `default_agent` | Never written; asserted by S2/S3 tests | Coherent |
| Rollback ordering fix (WriteManifest AFTER merge) | Manifest built at init.go:91, merge path appended at :107, `WriteManifest()` at :111 | Coherent |
| Opencode-gated | `if agentName == "opencode"` guards the merge call (init.go:101) | Coherent |
| No-op on empty leader | `if leader == "" { return "", nil }` (opencode_mode.go:34-36) | Coherent |
| Description string (cosmetic, fixed) | `"Archon SDD orchestration leader"` | Coherent (matches design open-question suggestion) |
| `--leader` CLI flag (task 3.4, NOT in original design File Changes) | Added by orchestrator; flag → `modelLeaderFlag` → advisory `config.Validate` → `Options.ModelLeader` → `buildConfig` → `cfg.Models.Leader` → merge. End-to-end wired and advisory-validated. | Coherent extension; correctly wired |

Deviation note: task 3.4 (`--leader` flag) is an addition beyond design.md's File Changes table, which listed only the config/writer/init files for PR1. The flag is consistent with the existing `--model*` flag pattern, advisory-validated (warns, never rejects, per `config.Validate`), and threads cleanly to `cfg.Models.Leader`. This is a complete, beneficial addition — not a regression.

---

## 5. Issues

### CRITICAL
None.

### WARNING
None.

### SUGGESTION
- (Non-blocking) There is no direct unit test asserting the `--leader` CLI flag wiring end-to-end (flag value reaching `opencode.json`). The path is exercised indirectly via `Options.ModelLeader` in `TestRun_RegistersOpencodeJSONForRollback`, and the flag binding is trivial cobra plumbing. A `cmd`-level test could be added in PR2 alongside the TUI parity test for completeness.

---

## Final Verdict: PASS

All PR1 tasks (Phases 1-4 + task 3.4) are implemented in code and verified. `go build`, `go test`, `go vet`, and `gofmt` all pass with real runtime evidence. Spec scenarios S1-S6 each map to a test that passed at runtime. Design coherence holds across all PR1 invariants (shared writer, additive/atomic write, deterministic struct, no `default_agent`, rollback ordering fix, opencode-gating, empty-leader no-op), and the orchestrator's `--leader` flag is correctly wired and advisory-validated. S7/S8 and Phases 5-7 are intentionally deferred to PR2 and are not blocking.
