# Delta for harness-init

## ADDED Requirements

### Requirement: Configurable leader model

`.archon/config.yaml` MUST hold a `models.leader` field carrying the FULL
`provider/model-id` value verbatim (e.g. `anthropic/claude-sonnet-4-...`), with no
prefix stripping or normalization applied to the stored value. The field MUST survive
`Config.Clone()` and config serialization round-trips. Any validation of the value is
advisory only and MUST NOT reject or rewrite it.

#### Scenario: Leader model survives clone and round-trip

```gherkin
Scenario: Leader model survives clone and round-trip
  Given "models.leader" set to "anthropic/claude-sonnet-4-20250514"
  When the config is cloned and serialized then reloaded
  Then "models.leader" equals "anthropic/claude-sonnet-4-20250514" verbatim
```

### Requirement: Opencode archon-leader agent merge at init

For the **opencode** agent, `archon init` MUST additively merge a primary agent named
`archon-leader` into the project `opencode.json` (creating the file when absent),
setting ONLY `agent.archon-leader` with: `mode: "primary"`, `prompt:
"{file:./AGENTS.md}"`, and `model` set to `models.leader`. The merge MUST be additive
and idempotent — it MUST NOT modify or remove unrelated keys, and re-running init MUST
produce the same result with no duplication or drift. The agent MUST NOT be set as the
default agent (no `default_agent` written). The written `opencode.json` path MUST be
registered in the rollback manifest.

#### Scenario: Init writes the archon-leader agent

```gherkin
Scenario: Init writes the archon-leader agent
  Given an opencode project with "models.leader" set and no "opencode.json"
  When the user runs "archon init --agent opencode"
  Then "opencode.json" sets "agent.archon-leader" with mode "primary"
  And its "prompt" is "{file:./AGENTS.md}" and "model" equals "models.leader"
  And the "opencode.json" path is registered for rollback
```

#### Scenario: Merge into an existing opencode.json preserves other keys

```gherkin
@edge
Scenario: Merge into an existing opencode.json preserves other keys
  Given an opencode project whose "opencode.json" has unrelated keys and agents
  When the user runs "archon init --agent opencode"
  Then "agent.archon-leader" is added
  And every pre-existing key and agent is left unchanged
  And no "default_agent" key is written
```

#### Scenario: Re-running init is idempotent

```gherkin
@edge
Scenario: Re-running init is idempotent
  Given an opencode project already initialized with "archon-leader"
  When the user runs "archon init --agent opencode" again
  Then "agent.archon-leader" appears exactly once with the same content
  And no other key drifts
```

### Requirement: Leader merge no-op guards

The archon-leader merge MUST be a no-op when the agent is not opencode OR when
`models.leader` is empty: no `opencode.json` is created and nothing is written.

#### Scenario: Non-opencode agent writes no opencode.json

```gherkin
@edge
Scenario: Non-opencode agent writes no opencode.json
  Given a project initialized with "archon init --agent claude"
  When init completes
  Then no "opencode.json" is created or modified
```

#### Scenario: Empty leader model writes nothing

```gherkin
@error
Scenario: Empty leader model writes nothing
  Given an opencode project with an empty "models.leader"
  When the user runs "archon init --agent opencode"
  Then no "opencode.json" is created
  And no archon-leader agent is written
```

### Requirement: TUI and update write-path parity

For an opencode project, saving the leader-model field in the TUI Models tab MUST
produce the SAME archon-leader merge result as `archon init`, with no divergence
between the two write paths. `archon update` MUST NOT write or rewrite the opencode
agent (update stays skill-only).

#### Scenario: TUI save matches init merge result

```gherkin
@happy
Scenario: TUI save matches init merge result
  Given an opencode project and a chosen leader model
  When the leader-model field is saved via the TUI Models tab
  Then the resulting "agent.archon-leader" equals what "archon init" would produce
```

#### Scenario: Update leaves the opencode agent untouched

```gherkin
@edge
Scenario: Update leaves the opencode agent untouched
  Given an opencode project with an existing "agent.archon-leader"
  When the user runs "archon update"
  Then "opencode.json" is not written or rewritten
```
