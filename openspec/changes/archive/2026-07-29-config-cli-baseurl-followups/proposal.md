# Proposal: config CLI base_url follow-ups (#90, #91)

<!-- Follows the archived `local-model-provider` change, which introduced
     `ModelRef.BaseURL` and the OpenCode provider block. This change closes two
     CLI/status gaps that surfaced after that work landed. -->

## Intent

Two gaps remain after `local-model-provider` (archived) wired `ModelRef.BaseURL`:

- **#90 (bug):** `archon config list` and `archon status` share a guard that
  reports `(none configured)` whenever the model id is empty — even when a ref
  carries only a `base_url`. The endpoint is silently hidden.
- **#91 (feat):** `models.leader.base_url` (the OpenCode leader ref) cannot be
  set/get/listed via `archon config`; only `models.default` and
  `models.phases.<phase>` are wired. `status` never shows the leader at all.

Users configuring a local/leader endpoint get no feedback that it took effect.

## Scope

### In Scope
- Fix the `(none configured)` guard in BOTH surfaces so a ref with only a
  `base_url` is treated as configured and its endpoint is shown.
- Suppress the empty primary line (`models.X = `) when a ref's `FullID()==""`;
  print only the `base_url` line. Applies to default, phases, and leader.
- Wire `models.leader` (model id + `base_url`) through `config set/get/list`.
- Add `models.leader` (id + `base_url`) to the `archon status` Models block.

### Out of Scope
- Any `internal/config/` schema, `Clone`, or `Validate` change — `ModelRef` /
  `Leader` already carry everything needed (verified at explore).
- Making `ValidateBaseURL` blocking — it stays advisory (warn, never error).
- OpenCode emission, TUI Models tab, or new provider surfaces.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `local-model-provider`: extend the CLI/status surface requirements so
  base_url-only refs render, and add the leader ref to `config set/get/list`
  and the status Models block. (Delta spec against the archived capability.)

## Approach

- **Guard fix (both surfaces):** replace the `FullID()=="" && len(Phases)==0`
  emptiness test with one that also accounts for any ref (default, leader, phase)
  carrying a non-empty `base_url`. `(none configured)` remains ONLY when every
  ref is genuinely empty.
- **base_url-only rendering:** guard each primary line on `FullID()!=""`; always
  print the `base_url` line when set. Mirrors the existing default/phase pattern
  in `config.go:135`.
- **Leader wiring:** extend `setConfigValue`/`getConfigValue`, `baseURLRefForKey`,
  and the supported-keys error strings with `models.leader` and
  `models.leader.base_url`; add a Leader block to the status Models section.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/archon/config.go` | Modified | Fix list guard (`:127-137`); leader keys in `setConfigValue`/`getConfigValue`/`baseURLRefForKey`; supported-keys strings (`:273`,`:320`) |
| `internal/status/display.go` | Modified | Fix Models guard (`:69-85`); render base_url lines; add Leader block |
| `cmd/archon/config_test.go` | Modified | New cases; `TestConfigCmd_ListEmpty` (`:508`) MUST still pass |
| `internal/status/display_test.go` | Modified | Base_url + leader render cases |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `(none configured)` regresses for a genuinely empty config | Med | Keep `TestConfigCmd_ListEmpty` green; new guard is strictly additive (only broadens "configured", never the reverse) |
| Leader wiring makes validation blocking | Low | Reuse advisory `ValidateBaseURL`; assert non-blocking path in a test |
| Two surfaces drift (list vs status render differently) | Low | Add symmetric tests for both; identical guard logic |

## Rollback Plan

Additive rendering + guard broadening. Revert the single PR; no persisted-state
or schema change, so configs behave exactly as before. No migration.

## Dependencies

- `local-model-provider` (archived) — provides `ModelRef.BaseURL` and
  `ModelConfig.Leader`.

## Success Criteria

- [ ] A config whose only setting is a `base_url` (any ref) shows the endpoint,
      NOT `(none configured)`, in both `config list` and `status`.
- [ ] No empty `models.X = ` line is printed when a ref's model id is empty.
- [ ] `config set/get/list` handle `models.leader` and `models.leader.base_url`.
- [ ] `status` shows the leader id + base_url.
- [ ] A genuinely empty config still shows `(none configured)` in both surfaces.

## Size Estimate — single PR, under 400 lines

| Bucket | Est. lines |
|--------|-----------|
| `config.go` (guard + leader keys + error strings) | ~55 |
| `display.go` (guard + base_url + leader block) | ~30 |
| Tests (both surfaces) | ~180 |
| **Total** | **~265** |

Cohesive, single capability delta, no schema work → **one PR, ~265 lines, under
the 400-line budget**. No chained split needed.
