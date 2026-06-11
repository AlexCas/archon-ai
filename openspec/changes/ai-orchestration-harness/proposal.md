# Proposal: AI Orchestration Harness CLI

## Intent

Developers need a single-command installer that scaffolds the full AI workflow (Spec → Hard Spec → Gherkin → TDD → Judge) into any project. Today, configuring the 21 skills, project-local directories, and orchestrator instructions is entirely manual. The harness eliminates that friction: one command, plug-and-play.

## Scope

### In Scope
- `archon` CLI binary with `init` command: detects agent, scaffolds skill dirs, symlinks global skills, runs `sdd-init`, writes `AGENTS.md`/`claude.md`, creates `.archon/config.yaml`
- `harness-workflow` meta-skill: enforces phase sequence (explore → propose → spec → design → tasks → apply → verify → judge → archive)
- `harness-judge` skill: wraps `judgment-day` + mutation testing gate; loops back to `sdd-apply` on failure
- Agent detection: scan for `.claude/`, `.opencode/`, `.agents/` to determine which skill directories to create

### Out of Scope
- Playwright integration (post-judge, deferred)
- Package manager / plugin model (`archon add <skill>`)
- CI/CD pipeline (dev-loop only, not a CI gate)
- Mutation testing framework selection (configurable, not hardcoded)
- Reimplementing existing SDD/judgment-day skills — harness references them, not replaces them

## Capabilities

### New Capabilities
- `cli-installer`: Go CLI (`archon init`) that detects agent, scaffolds project-local skill dirs, symlinks skills, bootstraps openspec, and writes orchestrator config
- `harness-workflow`: Meta-skill that reads change state and enforces phase transition rules
- `harness-judge`: Skill wrapping judgment-day + configurable mutation testing gate with failure loop

### Modified Capabilities
- None (project has no existing specs)

## Approach

Go CLI binary distributed via `brew install` or `go install`. Single entry point `archon init`:

1. **Detect**: scan for `.claude/`, `.opencode/`, `.agents/` dirs and config files to identify the active AI agent
2. **Scaffold**: create project-local skill directories per detected agent; symlink (copy fallback on Windows) the 21 global skills
3. **Bootstrap**: invoke `sdd-init` to create `openspec/` structure
4. **Configure**: write/merge orchestrator instructions into the project's `AGENTS.md` or `claude.md`
5. **Track**: write `.archon/config.yaml` with harness version, detected agent, and skill inventory

Two additional skills installed alongside:
- `harness-workflow` — reads `openspec/changes/{name}/state.yaml`, enforces phase order, blocks invalid transitions
- `harness-judge` — orchestrates `judgment-day` + mutation testing; on failure, feeds judge feedback back into `sdd-apply`

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/archon/` | New | CLI entry point and subcommands |
| `internal/agent/` | New | Agent detection and skill resolution |
| `internal/scaffold/` | New | Directory creation, symlink logic |
| `internal/config/` | New | `.archon/config.yaml` read/write |
| `skills/harness-workflow/` | New | Meta-skill SKILL.md |
| `skills/harness-judge/` | New | Judge orchestration SKILL.md |
| Project root `AGENTS.md` | Modified | Orchestrator instructions appended |
| `.archon/` | New | Harness configuration directory |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| AI agents don't load project-local `.claude/skills/` or `.opencode/skills/` | High | Verify before implementation; fallback: global skill that reads `.archon/config.yaml` |
| Symlink portability on Windows | Medium | Copy fallback when symlinks fail |
| Meta-skill phase enforcement gets bypassed by user prompts | Medium | `state.yaml` persists phase; orchestrator instructions block out-of-order invocations |
| Mutation testing as gate slows daily workflow | Medium | Make mutation testing configurable; default off in `.archon/config.yaml` |

## Rollback Plan

Remove `.archon/`, project-local skill directories, and revert `AGENTS.md` changes. The harness is non-destructive: it only creates files and symlinks — no global config modifications. `archon init` writes a `.archon/rollback.json` mapping every created path for clean removal via `archon rollback`.

## Dependencies

- Verify AI agent project-local skill loading capability (Claude Code, OpenCode, Gemini CLI)
- Go 1.22+ for CLI build
- Existing 21 skills installed in global directories

## Success Criteria

- [ ] `archon init` runs in an empty project and produces a working `openspec/` + skill dirs + orchestrator config
- [ ] The developer can immediately invoke SDD skills from the project without manual configuration
- [ ] `harness-workflow` prevents out-of-order phase transitions
- [ ] `harness-judge` runs judgment-day and optionally mutation testing; on failure, loops back to sdd-apply
- [ ] `archon rollback` removes all harness-created files cleanly