# Delta for harness-workflow

## ADDED Requirements

### Requirement: Automatic map.md Regen on Phase Transition

After every successful phase transition, the harness-workflow transition path MUST
trigger `archon map` to regenerate `openspec/map.md`. This MUST happen after the
`state.yaml` update so the new phase and status are reflected in the generated index.
The regen MUST be transparent to the orchestrator: a regen failure SHOULD be reported
as a warning but MUST NOT block the phase transition from being recorded.

#### Scenario: map.md is regenerated after a successful phase transition

```gherkin
@happy
Scenario: map.md is regenerated after a successful phase transition
  Given a change my-feature transitioning from spec to design
  When harness-workflow approves the transition and updates state.yaml
  Then archon map is invoked to regenerate openspec/map.md
  And map.md reflects my-feature with phase=design and status=in_progress
```

#### Scenario: Regen failure does not block the transition

```gherkin
@error
Scenario: Regen failure does not block the transition
  Given archon map exits non-zero during a phase transition regen
  When harness-workflow detects the regen failure
  Then the phase transition is still recorded in state.yaml
  And a warning about the regen failure is surfaced to the orchestrator
  And the transition is not rolled back
```

#### Scenario: Regen runs after state.yaml is written

```gherkin
@edge
Scenario: Regen runs after state.yaml is written
  Given a phase transition from tasks to apply
  When harness-workflow processes the transition
  Then state.yaml is updated before archon map is invoked
  And the generated map.md shows the updated phase and status
```
