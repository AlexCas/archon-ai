# Delta for cli-installer

## MODIFIED Requirements

### Requirement: Init Command

The `archon init` command MUST scaffold a project for SDD workflow in one step, including Leader persona instructions in the generated template.

(Previously: scaffolded templates without persona guidance.)

#### Scenario: First run extracts embedded skills

- GIVEN an empty project directory
- AND no `~/.config/opencode/skills/` directory exists
- WHEN `archon init` runs
- THEN the 21 skills are extracted into `~/.config/opencode/skills/`
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
- AND the generated template includes a `## Leader Persona` section

#### Scenario: Multiple agents detected

- GIVEN a project with BOTH `.claude/` and `.opencode/` directories
- WHEN `archon init` runs
- THEN the CLI prompts the user to select the primary agent
- AND scaffolds config for the chosen agent only

### Requirement: Config Files

`archon init` MUST write `.archon/config.yaml`, `.archon/rollback.json`, and a project orchestrator template (`AGENTS.md` or `claude.md`) containing delegation rules, skill reference, and persona instructions.

(Previously: template contained delegation rules and skill reference without persona.)

#### Scenario: Config and template written after successful init

- GIVEN `archon init` completes successfully
- THEN `.archon/config.yaml` contains `agent`, `harness_version`, `skill_count`, and `created_at`
- AND `.archon/rollback.json` contains an array of every file and directory created
- AND `AGENTS.md` (or `claude.md`) contains orchestrator delegation rules, `## Leader Persona` section, and the 21-skill reference
