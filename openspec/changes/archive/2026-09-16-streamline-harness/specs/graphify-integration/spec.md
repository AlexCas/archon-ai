# Delta for graphify-integration (RETIRED)

<!-- [[graphify-integration]] · [proposal](../../proposal.md) · [exploration](../../exploration.md) -->

The `graphify-integration` capability is RETIRED in full. Graphify is removed from the
harness (Go config/CLI/init/TUI/status plumbing, the `skills/graphify/` skill, the
`embed_test.go` expected-skill entry, and all phase-skill hooks). The live capability
spec at `openspec/specs/graphify-integration/spec.md` and its `.feature` file are
deleted; this spec-level note records the retirement. The physical spec deletion happens
in the apply phase.

## REMOVED Requirements

### Requirement: Graphify Integration (entire capability)

(Reason: Graphify is dead surface — never enabled in this repo, adds a config block,
a CLI flag, a TUI tab, a `status` block, a skill, and conditional hooks in every phase
skill for a feature that is never exercised. All of R-01 through R-19 — config struct,
`archon config` get/set keys, `archon status` block, `--graphify` init flag, preflight
group G, the `skills/graphify/SKILL.md` orchestration skill, the advisory/inertness
contracts, `auto_install`/`semantic` semantics, staleness re-extraction, `sdd-explore`
and `sdd-tasks` consumption, the tracked excerpt, version-pin advisory, naming
discipline, and the TUI tab — are removed as a unit.)

(Migration: None. Graphify was opt-in and default-off; no consuming project depended on
it. An existing `.archon/config.yaml` carrying a `graphify.*` block loads without error
(leftover keys are ignored by the lenient loader); `archon config set/get graphify.*`
becomes an unknown-key error by design. Removed Go plumbing is an additive-free deletion
and is restored verbatim by reverting the change. If graph-informed comprehension is
wanted again, it can be re-added as an external overlay skill rather than baked into the
core phase skills.)

#### Scenario: Graphify config surface is gone

```gherkin
@happy
Scenario: Graphify config surface is gone
  Given the harness after this change is applied
  When the user runs "archon config set graphify.enabled true"
  Then the command fails with an unknown-key error
  And no graphify block is exposed by the config schema
```

#### Scenario: An existing graphify block loads without error

```gherkin
@edge
Scenario: An existing graphify block loads without error
  Given an existing .archon/config.yaml that still contains a graphify block
  When the config is loaded
  Then the load succeeds
  And the leftover graphify keys are ignored
```

#### Scenario: The graphify skill and spec no longer exist

```gherkin
@happy
Scenario: The graphify skill and spec no longer exist
  Given the harness after this change is applied
  When the embedded skills and live specs are enumerated
  Then skills/graphify/ is not present
  And openspec/specs/graphify-integration/ is not present
  And the embed_test expected-skill list does not include "graphify"
```
