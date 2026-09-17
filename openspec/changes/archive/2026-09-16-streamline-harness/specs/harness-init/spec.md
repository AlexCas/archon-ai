# Delta for harness-init

<!-- [[harness-init]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

This delta removes the Impeccable config/CLI/init/TUI/status surface, removes Graphify's
init surface (its capability is retired in `[[graphify-integration]]`), collapses the
artifact-store choice to OpenSpec-only (dropping engram/both), and pins the resulting TUI
tab set. All requirements are post-removal invariants. Playwright, Security, Models, and
the orchestrator-file guards are unchanged.

## ADDED Requirements

### Requirement: No Impeccable config or CLI surface

The config schema MUST NOT expose an `impeccable` block, and `archon` MUST NOT provide an
`--impeccable` init flag. `archon config set/get impeccable.<field>` MUST return an
unknown-key error. An existing `.archon/config.yaml` that still contains an `impeccable`
block MUST load without error (leftover keys are ignored by the lenient loader).

#### Scenario: Setting an impeccable key is an unknown-key error

```gherkin
@error
Scenario: Setting an impeccable key is an unknown-key error
  Given the harness after this change is applied
  When the user runs "archon config set impeccable.enabled true"
  Then the command fails with an unknown-key error
  And the supported-keys message does not list any impeccable key
```

#### Scenario: Existing impeccable block loads without error

```gherkin
@edge
Scenario: Existing impeccable block loads without error
  Given an existing .archon/config.yaml that still contains an impeccable block
  When the config is loaded
  Then the load succeeds
  And the leftover impeccable keys are ignored
```

#### Scenario: No impeccable init flag

```gherkin
@happy
Scenario: No impeccable init flag
  Given the harness after this change is applied
  When "archon init --impeccable" is run
  Then init reports an unknown flag error
  And no impeccable value is written to the emitted config
```

### Requirement: No Graphify init surface

`archon` MUST NOT provide a `--graphify` init flag, and the emitted config MUST NOT
contain a `graphify` block. This is the init-surface consequence of retiring the
`[[graphify-integration]]` capability. `archon config set/get graphify.<field>` MUST
return an unknown-key error, and an existing config carrying a `graphify` block MUST load
without error.

#### Scenario: No graphify init flag

```gherkin
@happy
Scenario: No graphify init flag
  Given the harness after this change is applied
  When "archon init --graphify" is run
  Then init reports an unknown flag error
  And the emitted config contains no graphify block
```

### Requirement: TUI exposes six tabs with Security last

The `archon tui` MUST expose exactly six tabs — Agent, Models, Judge, Mutation Testing,
Playwright, Security — with Security as the last tab. The Impeccable and Graphify tabs
MUST NOT be present. Tab navigation (next/previous, and wrap-around from the first tab)
MUST treat Security as the terminal tab in the ordering.

#### Scenario: Tab set has no Impeccable or Graphify tab

```gherkin
@happy
Scenario: Tab set has no Impeccable or Graphify tab
  Given the harness after this change is applied
  When the TUI tab labels are enumerated
  Then the labels are Agent, Models, Judge, Mutation Testing, Playwright, Security
  And no "Impeccable" or "Graphify" tab is present
```

#### Scenario: Shift-Tab from the first tab wraps to Security

```gherkin
@edge
Scenario: Shift-Tab from the first tab wraps to Security
  Given the TUI is focused on the Agent tab
  When the user presses Shift+Tab to move to the previous tab
  Then the focused tab is Security
```

### Requirement: Artifact store is always OpenSpec

The harness MUST treat OpenSpec as the sole artifact store. The preflight MUST NOT offer
an artifact-store choice (no `engram`, no `both`, no `none`); the standard default states
Artefactos=OpenSpec as prose. No init flag, config field, or TUI control MUST select a
non-OpenSpec artifact store.

#### Scenario: Preflight offers no artifact-store choice

```gherkin
@happy
Scenario: Preflight offers no artifact-store choice
  Given a fresh SDD session
  When the orchestrator applies the standard preflight default
  Then the artifact store is OpenSpec
  And the user is not asked to choose engram, both, or none
```

#### Scenario: No control selects a non-OpenSpec store

```gherkin
@edge
Scenario: No control selects a non-OpenSpec store
  Given the harness after this change is applied
  When the init flags, config keys, and TUI controls are enumerated
  Then none of them selects engram, both, or none as an artifact store
```
