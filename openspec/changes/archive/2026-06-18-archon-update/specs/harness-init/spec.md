# Delta for harness-init

## ADDED Requirements

### Requirement: Truthful skill inventory versions

The `skill_inventory` written to `.archon/config.yaml` at init MUST record each skill's
real version, read from that skill's `SKILL.md` `metadata.version` frontmatter. Init
MUST NOT write a hardcoded version. The shared skill-refresh routine that builds this
inventory is reused by `archon update`.

#### Scenario: Init records real frontmatter versions

```gherkin
Scenario: Init records real frontmatter versions
  Given embedded skills whose "SKILL.md" frontmatter declares versions like "2.0" and "3.0"
  When the user runs "archon init --agent claude"
  Then each "skill_inventory" entry records that skill's real frontmatter version
  And no inventory entry uses a hardcoded version
```

#### Scenario: Missing frontmatter version is handled

```gherkin
@edge
Scenario: Missing frontmatter version is handled
  Given an embedded skill whose "SKILL.md" declares no metadata.version
  When the user runs "archon init --agent claude"
  Then that skill is still recorded in "skill_inventory"
  And init does not abort
```
