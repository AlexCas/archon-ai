# Delta for harness-init

## MODIFIED Requirements

### Requirement: Dynamic model selection with free-form fallback

Model configuration MUST offer a model catalog that is DYNAMIC: it reflects the
agent CLIs detected on the system `PATH`. Models belonging to an agent CLI that is
NOT installed MUST be hidden from the offered/cyclable catalog. When the `opencode`
CLI is installed, the opencode portion of the catalog MUST be enumerated live from
that CLI instead of the static curated list. ONLY when `opencode` IS installed but
live enumeration fails (error, timeout, or unparseable output) MUST the offered
catalog fall back silently to the curated opencode list; detection MUST NEVER break
or block the TUI. (An absent `opencode` CLI does NOT trigger the curated fallback —
its models are hidden per the filter rule above.) Detection MUST
run once when the Models view is opened and be cached for that session — NOT per
keystroke and NOT during `archon init`. Claude and any remote-only provider stay
curated (no local enumeration). Free-form entry MUST always remain accepted (any
string), and `NormalizeModel`, `Validate`, and free-form acceptance behavior MUST be
unchanged by this feature.
(Previously: the catalog was a fixed static list of Claude + Opencode models offered
regardless of which agent CLIs were installed.)

| Aspect | Behavior |
|--------|----------|
| Catalog source | Detected agent CLIs on PATH (cached once per Models view) |
| Uninstalled agent | Its models hidden from the offered catalog |
| opencode installed | Live-enumerated catalog replaces curated opencode list |
| Enumeration failure | opencode installed but enumeration fails/times out/unparseable → silent fallback to curated opencode list; TUI never blocked |
| Claude / remote-only | Stays curated; no local enumeration |
| Free-form / Validate / NormalizeModel | Unchanged; any string accepted with advisory warning |

#### Scenario: Installed opencode shows the live catalog

```gherkin
@happy
Scenario: Installed opencode shows the live catalog
  Given the "opencode" CLI is installed on PATH
  When the user opens the Models view
  Then the offered opencode models are enumerated live from the opencode CLI
  And the stale curated opencode list is not shown
```

#### Scenario: Only installed agents' models are offered

```gherkin
@happy
Scenario: Only installed agents' models are offered
  Given the "opencode" CLI is not installed on PATH
  When the user cycles the catalog in the Models view
  Then no opencode models appear in the offered catalog
  And models for installed agents remain offered
```

#### Scenario: Detection is cached once per Models view

```gherkin
@edge
Scenario: Detection is cached once per Models view
  Given the Models view has been opened and detection has run once
  When the user cycles models and types repeatedly
  Then detection does not run again for that session
  And it never runs during "archon init"
```

#### Scenario: Live enumeration error falls back silently

```gherkin
@error
Scenario: Live enumeration error falls back silently
  Given the "opencode" CLI is installed but enumeration fails, times out, or returns unparseable output
  When the user opens the Models view
  Then the curated opencode list is offered as a silent fallback
  And the TUI is neither blocked nor shown an error
```

#### Scenario: Free-form entry and advisory behavior unchanged

```gherkin
@happy
Scenario: Free-form entry and advisory behavior unchanged
  Given the Models view default model input is empty
  When the user types "some-custom-model"
  Then the value is accepted and a non-blocking warning may be shown
  And NormalizeModel and Validate behave exactly as before this feature
```
