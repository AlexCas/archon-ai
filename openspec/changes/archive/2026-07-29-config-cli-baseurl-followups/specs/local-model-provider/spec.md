# Spec delta: config CLI base_url follow-ups (#90, #91)

<!-- proposal: ../../proposal.md | parent capability: openspec/changes/archive/2026-07-28-local-model-provider/specs/local-model-provider/spec.md -->

## Purpose

Close two CLI/status gaps that surfaced after the `local-model-provider` change landed.
This is a **delta spec** against the archived `local-model-provider` capability. It adds
requirements to the existing capability; REQ-1 through REQ-7 from the parent spec are
unchanged and remain binding.

---

## Delta Requirements

### Requirement: REQ-8 — Broaden the "none configured" emptiness guard

The guard that prints `(none configured)` for the Models block MUST treat a config as
"configured" if ANY of the following is non-empty: `Default.FullID()`, `Default.BaseURL`,
`Leader.FullID()`, `Leader.BaseURL`, or any entry in `Phases` (non-empty `FullID()` OR
non-empty `BaseURL`).

`(none configured)` MUST be printed ONLY when every field above is empty (a genuinely
empty config). The guard MUST be applied identically in BOTH surfaces:

- `archon config list` (`cmd/archon/config.go`, list sub-command block starting at line 127)
- `archon status` Models block (`internal/status/display.go`, line 69)

The existing `TestConfigCmd_ListEmpty` (`:508`) MUST continue to pass unchanged —
the all-empty fixture still triggers `(none configured)`.

#### Scenario: base_url-only default ref is treated as configured

```gherkin
Scenario: base_url-only default ref is treated as configured
  Given a config where models.default has no provider/model but BaseURL "http://localhost:11434/v1"
  When the user runs: archon config list
  Then the output does NOT contain "(none configured)"
  And the output contains "models.default.base_url = http://localhost:11434/v1"
```

#### Scenario: base_url-only phase ref is treated as configured

```gherkin
Scenario: base_url-only phase ref is treated as configured
  Given a config where models.phases.apply has no provider/model but BaseURL "http://localhost:11434/v1"
  When the user runs: archon config list
  Then the output does NOT contain "(none configured)"
  And the output contains "models.phases.apply.base_url = http://localhost:11434/v1"
```

#### Scenario: genuinely empty config still shows (none configured) — regression guard

```gherkin
Scenario: genuinely empty config still shows (none configured) — regression guard
  Given a config with no models keys set at all
  When the user runs: archon config list
  Then the output is exactly "(none configured)"
```

#### Scenario: status shows (none configured) for genuinely empty config

```gherkin
Scenario: status shows (none configured) for genuinely empty config
  Given a config with no models keys set at all
  When the user runs: archon status
  Then the Models block contains "(none configured)"
```

#### Scenario: status does NOT show (none configured) when only base_url is set

```gherkin
Scenario: status does NOT show (none configured) when only base_url is set
  Given a config where models.default has no provider/model but BaseURL "http://localhost:11434/v1"
  When the user runs: archon status
  Then the Models block does NOT contain "(none configured)"
```

---

### Requirement: REQ-9 — Suppress empty primary line when model id is absent

When rendering a ref for `config list` output, the primary model-id line
(`models.X = `) MUST be suppressed if `FullID() == ""`. The `base_url` line MUST
always be printed when `BaseURL != ""`, regardless of whether the model-id line is
printed. This rule applies to the default ref, every phase ref, and the leader ref
(after REQ-11 wires it).

The phase-loop primary line (`models.phases.<phase> = `) follows the same rule: emit
only when `FullID() != ""`.

#### Scenario: empty-id ref emits only the base_url line

```gherkin
Scenario: empty-id ref emits only the base_url line
  Given a config where models.default has BaseURL "http://localhost:11434/v1" and no provider/model
  When the user runs: archon config list
  Then the output contains "models.default.base_url = http://localhost:11434/v1"
  And the output does NOT contain "models.default = "
```

#### Scenario: ref with both id and base_url emits both lines

```gherkin
Scenario: ref with both id and base_url emits both lines
  Given a config where models.phases.apply has Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
  When the user runs: archon config list
  Then the output contains "models.phases.apply = ollama/llama3"
  And the output contains "models.phases.apply.base_url = http://localhost:11434/v1"
```

---

### Requirement: REQ-10 — status renders base_url lines and suppresses empty primary

`archon status` Models block MUST render a `base_url:` sub-line for every ref whose
`BaseURL != ""`. The primary model-id line MUST be suppressed when `FullID() == ""`.

These rules mirror REQ-8/REQ-9 on the `config list` surface and apply to the
default ref, all phase refs, and the leader ref (after REQ-11).

#### Scenario: status shows base_url for default ref

```gherkin
Scenario: status shows base_url for default ref
  Given a config where models.default has BaseURL "http://localhost:11434/v1" and Provider "ollama", Model "llama3"
  When the user runs: archon status
  Then the Models block contains "Default:"
  And the Models block contains "http://localhost:11434/v1"
```

#### Scenario: status suppresses empty primary for default ref with only base_url

```gherkin
Scenario: status suppresses empty primary for default ref with only base_url
  Given a config where models.default has BaseURL "http://localhost:11434/v1" and no provider/model
  When the user runs: archon status
  Then the Models block contains "http://localhost:11434/v1"
  And the Models block does NOT contain "Default:  \n" (no blank model-id value)
```

#### Scenario: status shows base_url for a phase ref

```gherkin
Scenario: status shows base_url for a phase ref
  Given a config where models.phases.apply has Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
  When the user runs: archon status
  Then the Models block contains "apply:"
  And the Models block contains "http://localhost:11434/v1"
```

---

### Requirement: REQ-11 — Wire models.leader through config set/get/list

`archon config set models.leader <provider>/<model>` MUST write the `Leader` ref on
`ModelConfig`. `archon config get models.leader` MUST return the stored `FullID()`.
`archon config set models.leader.base_url <url>` MUST write `Leader.BaseURL` without
altering the provider/model fields. `archon config get models.leader.base_url` MUST
return the stored value (empty string if unset).

`archon config list` MUST include a `models.leader = <id>` line (only when
`FullID() != ""`) and a `models.leader.base_url = <url>` line (only when
`BaseURL != ""`), rendered between the default block and the phases block.

`baseURLRefForKey` MUST handle `"models.leader.base_url"` so that advisory
`ValidateBaseURL` runs on the leader ref after a set. The validation MUST remain
advisory (warn-to-stderr, non-blocking) — identical contract as REQ-3 in the parent
spec.

The supported-keys error strings in `setConfigValue` (`:273`) and `getConfigValue`
(`:320`) MUST be updated to list the new leader keys:
`models.leader, models.leader.base_url`.

#### Scenario: set and get models.leader round-trip

```gherkin
Scenario: set and get models.leader round-trip
  Given an archon project with no leader configured
  When the user runs: archon config set models.leader ollama/llama3
  Then archon config get models.leader prints "ollama/llama3"
  And the provider and BaseURL fields of the leader ref are unchanged
```

#### Scenario: set and get models.leader.base_url round-trip

```gherkin
Scenario: set and get models.leader.base_url round-trip
  Given an archon project with no leader base_url configured
  When the user runs: archon config set models.leader.base_url http://localhost:11434/v1
  Then archon config get models.leader.base_url prints "http://localhost:11434/v1"
  And the provider and model fields of the leader ref are unchanged
```

#### Scenario: config list shows leader block

```gherkin
Scenario: config list shows leader block
  Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
  When the user runs: archon config list
  Then the output contains "models.leader = ollama/llama3"
  And the output contains "models.leader.base_url = http://localhost:11434/v1"
```

#### Scenario: leader base_url set triggers advisory validation — non-blocking on bad URL

```gherkin
Scenario: leader base_url set triggers advisory validation — non-blocking on bad URL
  Given an archon project with no leader configured
  When the user runs: archon config set models.leader.base_url ftp://bad-url
  Then the command exits 0
  And stderr contains a warning about an invalid http/https URL
  And archon config get models.leader.base_url prints "ftp://bad-url"
```

#### Scenario: unknown models.leader key is rejected with updated error listing

```gherkin
Scenario: unknown models.leader key is rejected with updated error listing
  Given an archon project
  When the user runs: archon config set models.leader.typo somevalue
  Then the command exits non-zero
  And stderr contains "models.leader" and "models.leader.base_url" in the supported-keys message
```

#### Scenario: unknown models.leader key in get is rejected

```gherkin
Scenario: unknown models.leader key in get is rejected
  Given an archon project
  When the user runs: archon config get models.leader.typo
  Then the command exits non-zero
  And stderr contains "models.leader" and "models.leader.base_url" in the supported-keys message
```

---

### Requirement: REQ-12 — status renders leader block symmetrically with default

`archon status` Models block MUST render a `Leader:` sub-block when either
`Leader.FullID() != ""` OR `Leader.BaseURL != ""`. The sub-block MUST be symmetric
with the `Default:` block: a model-id line (suppressed when `FullID() == ""`) and a
`base_url:` line (printed only when `BaseURL != ""`). The leader block MUST appear
after the default block and before the phases block.

#### Scenario: status shows Leader block with id and base_url

```gherkin
Scenario: status shows Leader block with id and base_url
  Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
  When the user runs: archon status
  Then the Models block contains "Leader:"
  And the Models block contains "ollama/llama3"
  And the Models block contains "http://localhost:11434/v1"
```

#### Scenario: status shows Leader block when only base_url is set

```gherkin
Scenario: status shows Leader block when only base_url is set
  Given a config where models.leader has no provider/model but BaseURL "http://localhost:11434/v1"
  When the user runs: archon status
  Then the Models block contains "Leader:"
  And the Models block contains "http://localhost:11434/v1"
  And the Models block does NOT contain a blank leader model-id line
```

#### Scenario: status omits Leader block when leader is genuinely empty

```gherkin
Scenario: status omits Leader block when leader is genuinely empty
  Given a config where models.leader is completely unset
  When the user runs: archon status
  Then the Models block does NOT contain "Leader:"
```

#### Scenario: config list and status render leader symmetrically

```gherkin
Scenario: config list and status render leader symmetrically
  Given a config where models.leader is Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
  When the user runs both: archon config list and archon status
  Then config list contains "models.leader = ollama/llama3" and "models.leader.base_url = http://localhost:11434/v1"
  And status contains "Leader:" with the same model id and base_url values
```

---

## Key Invariants (delta)

1. **Guard strictly additive**: the new emptiness check only broadens "configured" — it never causes a configured ref to revert to `(none configured)`.
2. **Primary-line suppression**: `models.X = ` (or `X:` in status) is printed IFF `FullID() != ""`. The `base_url` line is independent.
3. **Advisory-only**: all `ValidateBaseURL` calls introduced in REQ-11 remain warn-to-stderr and non-blocking. `config set` exits 0 regardless of base_url validity.
4. **Symmetric surfaces**: `config list` and `status` MUST agree on which sub-blocks appear; a ref visible in one surface MUST be visible in the other.
5. **No schema change**: `ModelRef.BaseURL` and `ModelConfig.Leader` already exist. No `Clone()`, `Validate()`, or YAML-tag change.
6. **`TestConfigCmd_ListEmpty` regression guard**: the all-empty config MUST still produce exactly `(none configured)` after the guard change.

## PR Mapping Summary

All requirements in this delta ship in a single PR (estimated ~265 lines, under the 400-line budget).

| REQ | Description | Surface |
|-----|-------------|---------|
| REQ-8 | Broaden emptiness guard — base_url-only ref counts as configured | `config list` + `status` |
| REQ-9 | Suppress empty primary line when model id absent | `config list` |
| REQ-10 | status renders base_url lines; suppresses empty primary | `status` |
| REQ-11 | Wire models.leader through config set/get/list + advisory validation | `config list` |
| REQ-12 | status renders Leader block symmetric to Default | `status` |
