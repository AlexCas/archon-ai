# cli-installer Specification

## Purpose

The `archon` CLI bootstraps the SDD harness into any project by extracting 21 embedded gentle-ai skills and scaffolding per-project config, agent instructions, and rollback tracking.

## Requirements

### Requirement: Init Command

The `archon init` command MUST scaffold a project for SDD workflow in one step.

#### Scenario: First run extracts embedded skills

- GIVEN an empty project directory
- AND no `~/.config/opencode/skills/` directory exists
- WHEN `archon init` runs
- THEN the 21 skills are extracted from the embedded `skills/` directory into `~/.config/opencode/skills/`
- AND each skill's `SKILL.md` is written to its subdirectory

#### Scenario: Idempotent extraction on re-run

- GIVEN `~/.config/opencode/skills/` already contains extracted skills
- WHEN `archon init` runs again
- THEN existing skill files are NOT overwritten
- AND the CLI reports "skills already present" and continues

#### Scenario: Agent detection — single agent

- GIVEN a project with `.opencode/` directory present
- WHEN `archon init` runs
- THEN the CLI detects OpenCode as the active agent
- AND writes orchestrator instructions to `AGENTS.md`

#### Scenario: Multiple agents detected

- GIVEN a project with BOTH `.claude/` and `.opencode/` directories
- WHEN `archon init` runs
- THEN the CLI prompts the user to select the primary agent
- AND scaffolds config for the chosen agent only

### Requirement: Go Embed Skill Distribution

The 21 gentle-ai skills MUST be embedded in the Go binary via `embed.FS`.

#### Scenario: Binary contains embedded skills at build time

- GIVEN the `archon` binary is built
- WHEN the binary is inspected
- THEN it contains a `skills/` directory with 21 skill subdirectories inside the embedded filesystem

#### Scenario: Extraction preserves directory structure

- GIVEN the 21 embedded skills
- WHEN extraction runs
- THEN each skill is written to `{target-dir}/{skill-name}/SKILL.md`
- AND the directory structure `skills/sdd-init/SKILL.md` is preserved

### Requirement: Config Files

`archon init` MUST write `.archon/config.yaml`, `.archon/rollback.json`, and a project orchestrator template (`AGENTS.md` or `claude.md`).

#### Scenario: Config and template written after successful init

- GIVEN `archon init` completes successfully
- THEN `.archon/config.yaml` contains `agent`, `harness_version`, `skill_count`, and `created_at`
- AND `.archon/rollback.json` contains an array of every file and directory created
- AND `AGENTS.md` (or `claude.md`) is created with orchestrator delegation rules and the 21-skill reference

### Requirement: Rollback Command

`archon rollback` MUST cleanly remove all files created by the last `archon init`.

#### Scenario: Rollback removes harness-created files

- GIVEN `.archon/rollback.json` exists with 3 file paths
- WHEN `archon rollback` runs
- THEN all 3 paths are deleted
- AND `.archon/` directory is removed
- AND project AGENTS.md changes are reverted from backup

#### Scenario: Rollback with no prior init

- GIVEN no `.archon/rollback.json` exists
- WHEN `archon rollback` runs
- THEN the CLI reports "nothing to rollback" and exits cleanly

### Requirement: Version and Status Commands

#### Scenario: Version prints binary version

- GIVEN the `archon` binary
- WHEN `archon version` runs
- THEN the version, commit hash, and build date are printed

#### Scenario: Status shows project harness state

- GIVEN `.archon/config.yaml` exists
- WHEN `archon status` runs
- THEN agent, harness version, and skill count are displayed

### Requirement: Error Handling

#### Scenario: Write failure aborts cleanly

- GIVEN extraction or config write fails (disk full, permission denied, etc.)
- WHEN the error occurs
- THEN the CLI exits with a non-zero code and descriptive message
- AND no partial `.archon/config.yaml` is written

### Requirement: Update Strategy

#### Scenario: Version mismatch triggers prompt

- GIVEN skills v1.0.0 are extracted and `archon` v1.1.0 runs `init`
- THEN the CLI detects the version gap and prompts: "Update skills to v1.1.0? [y/N]"
- AND on confirmation, overwrites outdated skills with embedded ones
