# harness-update Specification

## Purpose

Provide a safe, version-aware `archon update` that refreshes installed skills from the
embedded set without rewriting the orchestrator template or resetting user config.

## Requirements

### Requirement: Skill refresh preserves user config and template

`archon update` MUST refresh installed skills from the embedded set. It MUST NOT modify
the orchestrator template (`CLAUDE.md`/`AGENTS.md`). It MUST preserve `models`,
`playwright`, `mutation_testing`, `judge`, `created_at`, and `agent` in
`.archon/config.yaml`; it MAY update only `harness_version`, `skill_count`, and
`skill_inventory`.

#### Scenario: Update refreshes skills without touching template or user config

```gherkin
Scenario: Update refreshes skills without touching template or user config
  Given an initialized project with a customized "CLAUDE.md" and user config values
  When the user runs "archon update"
  Then the installed skills are refreshed from the embedded set
  And "CLAUDE.md" is left unchanged
  And "models", "playwright", "mutation_testing", "judge", "created_at", and "agent" are preserved
  And only "harness_version", "skill_count", and "skill_inventory" may change
```

### Requirement: Version-gap classification

`archon update` MUST compute the gap between embedded and installed skills, classifying
each skill as added, changed, or orphaned (installed but no longer embedded). When no
gaps exist, it MUST report "already up to date" and write nothing.

#### Scenario: Update classifies the version gap

```gherkin
Scenario: Update classifies the version gap
  Given installed skills differing from the embedded set
  When the user runs "archon update"
  Then each skill is classified as added, changed, or orphaned
  And the classification is reported to the user
```

#### Scenario: No gaps reports already up to date

```gherkin
@edge
Scenario: No gaps reports already up to date
  Given installed skills matching the embedded set
  When the user runs "archon update"
  Then the command reports "already up to date"
  And nothing is written
```

### Requirement: Dry-run check flag

`archon update --check` MUST report the diff (added, changed, orphaned) and MUST write
NOTHING.

#### Scenario: Check reports the diff without writing

```gherkin
Scenario: Check reports the diff without writing
  Given installed skills differing from the embedded set
  When the user runs "archon update --check"
  Then the added, changed, and orphaned skills are reported
  And no files are written
```

### Requirement: Opt-in orphan pruning

`archon update --prune` MUST remove orphaned skills. Without `--prune`, orphaned skills
MUST be reported but kept.

#### Scenario: Prune removes orphaned skills

```gherkin
Scenario: Prune removes orphaned skills
  Given an installed skill that is no longer embedded
  When the user runs "archon update --prune"
  Then the orphaned skill is removed
```

#### Scenario: Orphans are kept without prune

```gherkin
@edge
Scenario: Orphans are kept without prune
  Given an installed skill that is no longer embedded
  When the user runs "archon update"
  Then the orphaned skill is reported
  And the orphaned skill is kept
```

### Requirement: Copy-mode warning and machine-wide scope

When a project's installed skill path is a real directory rather than a symlink
(copy-mode), `archon update` MUST emit a WARNING that the project needs its own update
and MUST NOT auto re-link. The output SHOULD make clear that refreshing the shared
machine-wide skills directory affects all symlinked projects.

#### Scenario: Copy-mode install warns without re-linking

```gherkin
@edge
Scenario: Copy-mode install warns without re-linking
  Given a project whose installed skill path is a real directory, not a symlink
  When the user runs "archon update"
  Then a warning is emitted that the project needs its own update
  And the skill path is not re-linked automatically
```

#### Scenario: Output states machine-wide scope

```gherkin
Scenario: Output states machine-wide scope
  Given a symlinked project
  When the user runs "archon update"
  Then the output makes clear the refresh affects all symlinked projects
```

### Requirement: Update error handling

If `.archon/config.yaml` is missing (e.g. running update before any init), `archon
update` MUST report an actionable error and write nothing.

#### Scenario: Update before init reports an actionable error

```gherkin
@error
Scenario: Update before init reports an actionable error
  Given a project with no ".archon/config.yaml"
  When the user runs "archon update"
  Then an actionable error is reported telling the user to run init first
  And nothing is written
```
