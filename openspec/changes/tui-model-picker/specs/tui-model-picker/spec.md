# tui-model-picker Specification

## Purpose

Replace the free-form text rows in the TUI Models tab with an in-tab, per-row two-step
provider→model picker driven by the opencode cache (via the 3a catalog helpers), populating
`config.ModelRef{Provider, Model}` directly — while preserving untouched legacy values, keeping
a free-form escape hatch always available, and warning in-UI (never stderr) when the cache is
present but unreadable.

## Requirements

### Requirement: Provider→model pick sets the ModelRef

Focusing a row and pressing Enter MUST open a provider list (`DetectAvailableProviders`); choosing
a provider and pressing Enter MUST open that provider's model list (`FilterModelsForSDD`); choosing
a model and pressing Enter MUST set the row's `ModelRef` to the picked provider+model, mark the row
changed, and return to row navigation.

#### Scenario: Pick provider then model sets the ModelRef
```gherkin
Scenario: Pick provider then model sets the ModelRef
  Given the Models tab with a non-empty provider catalog
  When the user focuses a row, presses Enter, selects a provider, presses Enter, selects a model, presses Enter
  Then the row's ModelRef is the picked provider+model and the tab returns to row navigation
```

#### Scenario: opencode bare key and slashed key map without double-prefix
```gherkin
@edge
Scenario: opencode bare key and slashed key map without double-prefix
  Given the built-in "opencode" provider with a bare model key and another provider with an already-slashed key
  When each is picked
  Then the opencode ref is {Provider:"opencode", Model:key} (FullID "opencode/key")
  And the slashed-key ref's FullID equals the key verbatim (no double provider prefix)
```

#### Scenario: A provider with no SDD models falls back to free-form
```gherkin
@edge
Scenario: A provider with no SDD models falls back to free-form
  Given a selectable provider that has no tool_call models
  When the user selects it
  Then free-form entry opens instead of an empty model list
```

### Requirement: Esc cancels without changing the row

Pressing Esc in model selection MUST return to provider selection; pressing Esc in provider
selection MUST return to row navigation. In both cases the row's `ModelRef` and changed flag MUST
be unchanged.

#### Scenario: Esc steps back and cancels
```gherkin
Scenario: Esc steps back and cancels
  Given a row in model selection
  When the user presses Esc
  Then the tab returns to provider selection
  And pressing Esc again returns to row navigation with the row's ModelRef unchanged
```

### Requirement: Free-form entry is always available

In row navigation, pressing the free-form key (`e`) on the focused row MUST open a text input
seeded with the row's current `FullID()`. Confirming with Enter MUST set the row's `ModelRef` to
`ParseModelRef(value)` and mark it changed; Esc MUST cancel leaving the row unchanged. When no
providers are available (absent cache), pressing Enter on a row MUST open free-form directly.

#### Scenario: Free-form toggles and parses on confirm
```gherkin
Scenario: Free-form toggles and parses on confirm
  Given any focused row
  When the user presses "e", types "x/y", and presses Enter
  Then the row's ModelRef equals ParseModelRef("x/y") and the row is changed
```

#### Scenario: Free-form Esc cancels
```gherkin
@edge
Scenario: Free-form Esc cancels
  Given a row in free-form entry
  When the user presses Esc
  Then the row's ModelRef is unchanged
```

#### Scenario: Absent cache opens free-form directly
```gherkin
@edge
Scenario: Absent cache opens free-form directly
  Given an empty provider catalog and no cache error
  When the user presses Enter on a row
  Then free-form entry opens with no provider list and no warning is shown
```

### Requirement: Untouched legacy values are preserved

A row the user does not re-pick or edit MUST be written back to config verbatim from its seeded
`ModelRef`, including a legacy bare alias (`Provider==""`). A provider MUST NEVER be guessed for an
untouched value.

#### Scenario: Untouched legacy ModelRef preserved verbatim
```gherkin
Scenario: Untouched legacy ModelRef preserved verbatim
  Given a row seeded from config with a legacy bare alias (Provider empty)
  When the user navigates without re-picking or editing that row and saves
  Then applyToConfig writes that row's ModelRef verbatim (provider stays empty)
```

#### Scenario: Clearing a phase row deletes the phase entry
```gherkin
@edge
Scenario: Clearing a phase row deletes the phase entry
  Given a phase row with a value
  When the user free-forms it to empty, confirms, and saves
  Then the phase key is absent from cfg.Models.Phases
```

### Requirement: Corrupt cache warns in-UI; absent cache is silent

When `LoadModelsOrEmpty` returns an error (cache present but unreadable), the tab MUST render an
inline warning and MUST NOT write to stderr. When the cache is absent (empty map, no error), no
warning is shown.

#### Scenario: Corrupt cache shows an inline warning
```gherkin
@error
Scenario: Corrupt cache shows an inline warning
  Given the opencode cache load returned an error
  When the Models tab renders
  Then an inline warning is shown in the view and nothing is written to stderr
```

### Requirement: Leader row uses the picker only for opencode

When `agent == "opencode"` a Leader row MUST be present and support the same provider→model picker
and free-form entry. For a non-opencode agent no Leader row is shown and `cfg.Models.Leader` MUST
be left as loaded on save.

#### Scenario: Leader row present and editable for opencode
```gherkin
Scenario: Leader row present and editable for opencode
  Given agent is "opencode"
  When the tab renders
  Then a Leader row is present and supports the picker and free-form
  And for a non-opencode agent no Leader row is shown and cfg.Models.Leader is left as loaded
```

### Requirement: Deterministic sorted lists and bounded navigation

Provider lists MUST appear in sorted id order and model lists in sorted Name order, identically
across runs (no map-iteration leakage). Row navigation (Up/Down) MUST clamp within
`[0, len(rows)-1]`; Tab MUST remain the global tab switch (rows are navigated with Up/Down).

#### Scenario: Lists are sorted and navigation clamps
```gherkin
Scenario: Lists are sorted and navigation clamps
  Given multiple providers and models
  When the picker renders
  Then providers appear in sorted id order and models in sorted Name order
  And pressing Up at the first row or Down at the last row keeps focusedRow within bounds
```
