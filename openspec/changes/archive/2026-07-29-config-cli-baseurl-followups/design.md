# Design: config CLI base_url follow-ups (#90, #91)

<!-- proposal: proposal.md | spec: specs/local-model-provider/spec.md -->

Implements the [[local-model-provider]] delta (REQ-8..REQ-12). See
[proposal](proposal.md) and [spec](specs/local-model-provider/spec.md).

## Technical Approach

Two render surfaces — `archon config list` (`cmd/archon/config.go`) and the
`archon status` Models block (`internal/status/display.go`) — must agree on
(a) when to print `(none configured)`, (b) suppressing an empty primary line,
(c) always printing a `base_url` line, and (d) a new leader ref between default
and phases. No `internal/config/` schema, `Clone`, or `Validate` change: the
`Leader` field, `ModelRef.BaseURL`, and advisory `ValidateBaseURL` already
exist. All new work is presentation + key routing in the two `main`/`status`
files.

## Architecture Decisions

| Decision | Options | Choice + rationale |
|----------|---------|--------------------|
| Emptiness guard | (a) inline expanded boolean at each site; (b) shared `hasAnyModelConfig(cfg.Models)` helper in `internal/config` | **(b) helper** on `ModelConfig` as method `func (mc ModelConfig) HasAny() bool`. Two call sites (`config.go:127`, `display.go:69`) must stay identical (Invariant 4); a single method makes divergence impossible and is the natural home given `Leader`+`Phases`+`Default` all live on `ModelConfig`. Cheap, pure, no state. |
| Primary-suppression + base_url render | (a) per-site inline; (b) shared render helper | **(a) per-site inline**, NOT a shared helper. The two surfaces differ in format string, indentation, and label (`models.default = %s` vs `Default:  %s`; `models.phases.%s` vs `%-8s`). A shared helper would need a format-spec/label parameter object and buy little — the suppression logic is one `if FullID()!=""` guard plus one `if BaseURL!=""` guard, repeated. Keeping it inline matches the file's existing style (the phase loop already inlines this) and avoids a cross-package helper carrying UI format strings. |
| Leader key routing | fold into existing `switch` | Add explicit `case "models.leader"` / `case "models.leader.base_url"` arms in `setConfigValue`/`getConfigValue`, mirroring the `models.default` arms exactly. No prefix parsing needed (leader is a single ref, not a map). |
| `baseURLRefForKey` leader arm | new `case` | Add `case key == "models.leader.base_url": return cfg.Models.Leader, true`. Reuses the existing set-time advisory path (`config.go:62-66`) unchanged — validation stays advisory/non-blocking automatically. |

## Data Flow

    config set models.leader[.base_url]
        └─ setConfigValue → writes Leader / Leader.BaseURL
           └─ if key endsWith .base_url: baseURLRefForKey → ValidateBaseURL(stderr)  [advisory]
              └─ cfg.Save()

    config list / status Display
        └─ Models.HasAny()? no → "(none configured)"
                            yes → renderRef(Default) · renderRef(Leader) · renderRef(each phase)
                                  renderRef = [FullID()!="" → primary line] + [BaseURL!="" → base_url line]

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modify | Add `func (mc ModelConfig) HasAny() bool` — true if `Default`/`Leader` non-empty (FullID OR BaseURL) or any phase non-empty. |
| `cmd/archon/config.go` | Modify | List: swap guard for `!cfg.Models.HasAny()`; suppress `models.phases.<p> = ` when `FullID()==""`; add leader block between default and phases. Add `models.leader`/`.base_url` arms to `setConfigValue`, `getConfigValue`, `baseURLRefForKey`; update both supported-keys strings. |
| `internal/status/display.go` | Modify | Swap guard for `!cfg.Models.HasAny()`; render `base_url:` sub-line for default & phases; suppress empty primary; add `Leader:` block between Default and phases. |
| `internal/config/model_test.go` | Modify | Unit tests for `HasAny()` truth table. |
| `cmd/archon/config_test.go` | Modify | Add REQ-8/9/11 tests (below). |
| `internal/status/display_test.go` | Modify | Add REQ-8/10/12 tests (below). |

## Interfaces / Contracts

```go
// HasAny reports whether any ref carries a model id OR a base_url. It is the
// single source of truth for the "(none configured)" guard on both the
// `config list` and `status` surfaces (Invariant 4).
func (mc ModelConfig) HasAny() bool {
    if mc.Default.FullID() != "" || mc.Default.BaseURL != "" {
        return true
    }
    if mc.Leader.FullID() != "" || mc.Leader.BaseURL != "" {
        return true
    }
    for _, r := range mc.Phases {
        if r.FullID() != "" || r.BaseURL != "" {
            return true
        }
    }
    return false
}
```

Key strings & ordering (both surfaces, top→bottom): **default → leader →
phases (sorted)**. List keys: `models.leader = <id>`,
`models.leader.base_url = <url>`. Status labels: `Leader:` (same
`    Default:  %s` / two-space alignment as default). Updated supported-keys
suffix in both `setConfigValue` (`:273`) and `getConfigValue` (`:320`):
insert `models.leader, models.leader.base_url` immediately after
`models.default.base_url`.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (`model_test.go`) | `HasAny()` truth table | Table test: all-empty→false; default-id→true; default-baseURL-only→true; leader-id→true; leader-baseURL-only→true; phase-id→true; phase-baseURL-only→true. |
| CLI (`config_test.go`) | REQ-8 | `TestConfigCmd_ListBaseURLOnlyIsConfigured`: set only `models.default.base_url`, assert output lacks `(none configured)` and contains `models.default.base_url = ...`. Same for `models.phases.apply.base_url`. |
| CLI | REQ-8 regression | `TestConfigCmd_ListEmpty` (`:508`) MUST pass unchanged — all-empty still prints exactly `(none configured)`. |
| CLI | REQ-9 | `TestConfigCmd_ListSuppressesEmptyPrimary`: base_url-only default → contains `models.default.base_url = `, does NOT contain `models.default = `. Both-set phase → contains both `models.phases.apply = ollama/llama3` and `.base_url` lines. |
| CLI | REQ-11 | `TestConfigCmd_LeaderSetGet` (round-trip id + base_url, provider/model preserved); `TestConfigCmd_ListShowsLeaderBlock` (both lines, ordered before phases); `TestConfigCmd_LeaderBaseURLAdvisory` (`ftp://bad-url` exits 0, stderr warns, get returns value); `TestConfigCmd_LeaderUnknownKey` set+get reject `models.leader.typo`, stderr lists `models.leader` and `models.leader.base_url`. |
| status (`display_test.go`) | REQ-8/10 | `TestDisplay_ModelsNoneConfigured` (all-empty→`(none configured)`); base_url-only default → no `(none configured)`, contains url; default & phase `base_url:` sub-lines; empty-primary suppression (no blank model-id line). |
| status | REQ-12 | `TestDisplay_LeaderBlock`: id+base_url → `Leader:` + both values, ordered after Default before phases; base_url-only → `Leader:` + url, no blank id line; leader-empty → no `Leader:`. |
| symmetry | REQ-12 | `TestConfigCmd_LeaderSymmetry` asserts list + `status.Format()` agree on leader id/base_url. |

## Migration / Rollout

No migration required. Additive, backward-compatible presentation change; no
config-file format change, no default-behavior change for existing configs.

## Size Estimate

~265 changed lines (spec PR mapping), well under the 400-line budget:
`HasAny()` + tests ~30; guard swaps ~4; leader render/routing/error-strings
~60; suppression edits ~15; the balance is test cases. The only refactor that
would push this up is factoring a shared cross-package render helper — rejected
above, which keeps the estimate flat. Ships as a single PR.

## Open Questions

- [ ] None blocking. Minor: status leader label alignment uses the same
  two-space `Default:  ` style for visual symmetry (per REQ-12 "symmetric with
  Default") rather than the `%-8s` phase alignment — confirm at review.
