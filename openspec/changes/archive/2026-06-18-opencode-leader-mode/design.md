# Design: Opencode archon-leader mode + configurable leader model (Slice 2)

## Technical Approach

Add `Leader string` to `ModelConfig` (`models.leader`), storing the full
`provider/model-id` verbatim. A single new writer `mergeOpencodeAgent`
(`internal/initcmd/opencode_mode.go`) does an additive read-modify-write of the
**project** `opencode.json`, setting only `agent["archon-leader"]`. Both
`initcmd.Run()` (after `writeTemplate`) and the TUI `saveConfig` path call this
same function, so init and TUI produce byte-identical output. The writer is
opencode-only and a no-op when `models.leader` is empty. `archon update` is left
untouched. Maps to specs `harness-init` (4 requirements, 8 scenarios).

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Leader storage | `Leader string` on `ModelConfig`, verbatim | Top-level section; catalog id + prefix map | Reuses `models:` block + TUI; spec mandates verbatim, advisory-only validation |
| Config write location | Project-root `opencode.json` (beside `AGENTS.md`) | Global `~/.config/opencode` | `{file:./AGENTS.md}` resolves relative to config file; per-project footprint; rollback-trackable |
| JSON determinism | Marshal a fixed-field Go struct for the `archon-leader` value; whole doc via `json.MarshalIndent` | `map[string]any` for the value | `encoding/json` sorts map keys AND struct fields emit in declaration order — both deterministic; struct also documents the shape and guards key set |
| Merge unit | One shared `mergeOpencodeAgent`, called by init + TUI | Separate init/TUI writers | Eliminates drift (spec: byte-identical parity) |
| Idempotency | Read existing doc into `map[string]any`, set only `agent.archon-leader`, re-marshal | Overwrite whole file | Preserves unrelated keys/agents; re-run yields identical bytes |

## Data Flow

    .archon/config.yaml (models.leader)
            │
       buildConfig → cfg.Models.Leader
            │
    Run(): writeTemplate → mergeOpencodeAgent(projectDir, leader) ──┐
                                                                    ├─→ opencode.json
    TUI saveConfig: applyToConfig → cfg.Save → mergeOpencodeAgent ──┘   (agent.archon-leader)
            │                                                              │
       buildRollbackManifest ───── registers opencode.json path ──────────┘

## File Changes

| File | Action | PR | Description |
|------|--------|----|-------------|
| `internal/config/model.go` | Modify | PR1 | Add `Leader string \`yaml:"leader,omitempty"\`` to `ModelConfig` |
| `internal/config/config.go` | Modify | PR1 | Copy `Leader` in `Clone()` (set on the `ModelConfig` literal) |
| `internal/config/config_test.go` | Modify | PR1 | Add `Leader` to `TestConfig_CloneRoundtrip` fixture |
| `internal/initcmd/opencode_mode.go` | Create | PR1 | `mergeOpencodeAgent` additive/atomic JSON merge + value struct |
| `internal/initcmd/init.go` | Modify | PR1 | Call after `writeTemplate` (opencode + non-empty leader gated); register path in `buildRollbackManifest` |
| `internal/initcmd/opencode_mode_test.go` | Create | PR1 | Create/preserve/idempotent/no-op writer tests |
| `internal/tui/models_tab.go` | Modify | PR2 | Leader `textinput` (opencode-only); `applyToConfig` sets `cfg.Models.Leader` |
| `internal/tui/model.go` | Modify | PR2 | `saveConfig` calls `mergeOpencodeAgent` for opencode after `Save` |
| `internal/tui/*_test.go` | Modify | PR2 | TUI==init parity test |

## Interfaces / Contracts

```go
// internal/initcmd/opencode_mode.go
const leaderAgentName = "archon-leader"

// archonLeaderAgent is the fixed shape written under agent.archon-leader.
// Struct field order is the deterministic JSON output order.
type archonLeaderAgent struct {
    Mode        string `json:"mode"`        // always "primary"
    Description string `json:"description"`
    Model       string `json:"model"`       // verbatim models.leader
    Prompt      string `json:"prompt"`      // "{file:./AGENTS.md}"
}

// mergeOpencodeAgent additively merges agent.archon-leader into
// <projectDir>/opencode.json. No-op (nil, "") when leader == "".
// Returns the written path ("" when nothing written) for rollback registration.
func mergeOpencodeAgent(projectDir, leader string) (written string, err error)
```

Algorithm: if `leader == ""` return `"", nil`. Read `opencode.json` if present
into `doc map[string]any` (else `doc = map{}`). Ensure `doc["agent"]` is a
`map[string]any` (preserve existing). Set `agent["archon-leader"]` to the
`archonLeaderAgent` value. `json.MarshalIndent(doc, "", "  ")` + trailing `\n`,
write `.tmp`, `os.Rename` (same atomic pattern as `config.Save`/`writeTemplate`).
Never set `default_agent`. Idempotent because map-key marshaling is sorted and
the value struct is fixed-order. Call site in `Run()`:

```go
if agentName == "opencode" {
    if p, err := mergeOpencodeAgent(opts.ProjectDir, cfg.Models.Leader); err != nil {
        return nil, fmt.Errorf("merge opencode agent: %w", err)
    } else if p != "" {
        // append p to the rollback manifest (re-write or extend before WriteManifest)
    }
}
```

Note: `buildRollbackManifest` runs before `writeTemplate` today; the
`opencode.json` path must be appended to the manifest and `WriteManifest()`
called after the merge succeeds (reorder so the merge feeds the manifest).

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Clone/roundtrip keeps `models.leader` verbatim | `config_test.go` fixture (S1) |
| Unit | Merge creates file with correct shape, no `default_agent` | temp dir, parse JSON (S2) |
| Unit | Merge preserves unrelated keys/agents | seed existing JSON, assert untouched (S3) |
| Unit | Idempotent: two runs → byte-identical bytes | `bytes.Equal` of two outputs (S4) |
| Unit | Non-opencode / empty leader write nothing | assert no file (S5, S6) |
| Unit | Rollback manifest lists `opencode.json` | inspect manifest (S2) |
| Integration | TUI save == init merge bytes | drive `saveConfig`, compare to `mergeOpencodeAgent` (S7) |
| Integration | `archon update` does not touch `opencode.json` | mtime/bytes unchanged (S8) |

## Migration / Rollout

No migration required (`models.leader` is `omitempty`; absent in existing
configs). Delivery is a chained, stacked-to-main PR pair: **PR1** = config field +
`mergeOpencodeAgent` + init wiring + writer/roundtrip tests (self-contained core,
mergeable alone). **PR2** = TUI leader field + save-path merge + parity test
(stacked on PR1). Rollback per-project via the registered `opencode.json` path in
`.archon/rollback.json`.

## Open Questions

- [ ] None blocking. (Description string text is cosmetic; pick a short fixed
  value like "Archon SDD orchestration leader".)
