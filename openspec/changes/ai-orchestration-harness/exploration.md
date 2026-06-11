# Exploration: AI Orchestration Harness for Per-Project Developer Assistant

## Current State

The `archon-ai` project is a conceptual framework with no code yet. It has:
- `openspec/config.yaml` — SDD initialized (mode: openspec, strict TDD: false)
- `.atl/skill-registry.md` — indexes global skills from `~/.config/opencode/skills/`, `~/.claude/skills/`, etc.
- `idea-crear-un-harness-ia-basado-en-2026-06-10.md` — the original idea file

The global OpenCode skill directory already contains a comprehensive SDD workflow suite (10 skills covering the full lifecycle from init → explore → propose → spec → design → tasks → apply → verify → archive → onboard) plus supporting skills for PRs, commits, issues, comments, and skill management.

## Affected Areas

- `openspec/config.yaml` — will need new rules for harness-specific phases (judge, mutation testing)
- `.atl/skill-registry.md` — may need to index project-local skills if the harness installs them
- Project root — will need a new installer entry point and per-project skill directories
- New files: `openspec/changes/ai-orchestration-harness/exploration.md` (this document)

## Summary of Existing Skills

### SDD Workflow (10 skills) — Already Cover the Core Flow
| Skill | Phase | What it does | Maps to harness workflow |
|-------|-------|--------------|-------------------------|
| `sdd-init` | Setup | Detects stack, initializes openspec/Engram | Harness setup / project bootstrap |
| `sdd-explore` | Explore | Investigates codebase, identifies options | **Spec** phase (idea exploration) |
| `sdd-propose` | Propose | Writes intent, scope, approach | **Hard Spec** phase (scope definition) |
| `sdd-spec` | Spec | Writes Gherkin scenarios (Given/When/Then) | **Gherkin** phase |
| `sdd-design` | Design | Technical design, decisions, file changes | Design between spec and implementation |
| `sdd-tasks` | Tasks | Breaks work into ordered, verifiable tasks | Task planning before TDD |
| `sdd-apply` | Apply | Implements code (with optional strict TDD mode) | **TDD Implementation** phase |
| `sdd-verify` | Verify | Runs tests, proves compliance with specs | Verification before the judge |
| `sdd-archive` | Archive | Merges deltas, moves to archive | Cycle completion |
| `sdd-onboard` | Onboard | Walks user through a full SDD cycle | Developer onboarding |

### Supporting Skills (Already Exist)
| Skill | Purpose | Relevance to harness |
|-------|---------|---------------------|
| `judgment-day` | Dual adversarial review (2 blind judges) | **THE JUDGE** — fits perfectly |
| `chained-pr` | Split PRs >400 lines | Delivery guard |
| `branch-pr` | Issue-first PR creation | Delivery |
| `work-unit-commits` | Plan commits as work units | Delivery |
| `issue-creation` | GitHub issues with templates | External workflow |
| `comment-writer` | Warm, direct review comments | External workflow |
| `skill-creator` | Create new LLM-first skills | Creating harness-specific skills |
| `skill-improver` | Audit existing skills | Maintaining harness skills |
| `skill-registry` | Index available skills | Already used by the project |
| `go-testing` | Go testing patterns | Language-specific context (if harness is Go) |
| `cognitive-doc-design` | Low-cognitive-load docs | Documentation for the harness |

### Gap Analysis — What Is Missing

| Need | Existing? | Gap severity |
|------|-----------|--------------|
| **Per-project installer** | ❌ None | **Critical** — the user explicitly wants this |
| **Workflow orchestrator** | ❌ None | **Critical** — a meta-skill that enforces the sequence: explore → propose → spec → design → tasks → apply → verify → judge → archive |
| **Mutation testing integration** | ❌ None | **Medium** — the judge is supposed to use mutation testing, not just dual review |
| **Hard Spec phase** | ⚠️ Unclear | **Medium** — `sdd-propose` covers scope, but the idea mentions "Hard Spec" as a distinct artifact between Spec and Gherkin |
| **Playwright integration** | ❌ None | **Low** — open question, post-judge |
| **Project-local skill loading** | ⚠️ Partial | **High** — agents currently scan global dirs; per-project `.claude/skills/` support is unverified |

## Approaches

### 1. CLI Installer + Meta-Skill (Recommended)

**Description**: A CLI tool (`archon` or `npx archon-ai`) that developers run once per project. It:
- Creates per-project skill directories (`.claude/skills/`, `.opencode/skills/`, `.agents/skills/`)
- Symlinks or copies the relevant SDD skills from the global directory
- Runs `sdd-init` to create `openspec/`
- Installs a `harness-workflow` meta-skill that enforces the phase sequence
- Optionally installs a `harness-judge` skill that wraps `judgment-day` + mutation testing

- **Pros**: Full control, can enforce workflow, supports mutation testing as a subprocess, easy to update (symlinks), language-agnostic
- **Cons**: Requires installing the CLI first, agents must support project-local skill directories, Windows symlink issues
- **Effort**: Medium

### 2. Pure Skill-Based (No CLI)

**Description**: The harness is just a set of skills placed in the project. The user manually copies them or uses a shell script. The meta-skill (`harness-workflow`) orchestrates by reading `openspec/changes/{name}/state.yaml`.

- **Pros**: Native to the skill ecosystem, no external dependencies, the AI agent runs it directly
- **Cons**: No enforcement mechanism, manual setup, skill staleness risk
- **Effort**: Low

### 3. Configuration-Only (`.archon/`)

**Description**: A `.archon/config.yaml` specifies which skills to use and the workflow settings. The orchestrator reads this file and behaves accordingly.

- **Pros**: Simple, declarative, per-project config
- **Cons**: Doesn't solve the skill installation problem, requires orchestrator support for `.archon/` config
- **Effort**: Low

### 4. Package Manager / Plugin Model

**Description**: A package manager for skills (e.g., `archon add sdd-init`, `archon add judgment-day`). Skills are "published" and installed per-project.

- **Pros**: Scalable, could become an ecosystem
- **Cons**: Over-engineering for an MVP, requires a registry, maintenance burden
- **Effort**: High

## Recommendation

**Approach 1: CLI Installer + Meta-Skill**.

The harness should be a lightweight CLI tool that:
1. `archon init` — detects the AI agent (Claude, OpenCode, etc.), creates project-local skill directories, symlinks relevant global skills, runs `sdd-init`, and writes `.archon/config.yaml`
2. `archon workflow` — reads the state of the current change and tells the user (or orchestrator) what the next phase is
3. `archon judge` — runs `judgment-day` + optionally runs mutation testing tools

Why this approach:
- It satisfies the **per-project** constraint without touching global config
- It **reuses existing skills** by symlinking them (no reinvention)
- It provides the **installer** the user explicitly wants
- It gives a natural hook for **mutation testing** and **Playwright** as CLI commands

## Risks

1. **Agent skill resolution**: Claude Code / OpenCode / Gemini CLI may not support loading skills from project-local directories. If they only read global paths, the per-project constraint is impossible. **This is the #1 blocker to verify.**
2. **Hard Spec ambiguity**: The idea describes "Spec → Hard Spec → Gherkin." The existing SDD flow goes `explore → propose → spec`. Is "Hard Spec" just a stricter `proposal.md` or a new artifact? If it's a new artifact, we need a new skill or phase.
3. **Mutation testing complexity**: Mutation testing is slow and noisy. Running it as a gate in the daily workflow could be a friction point. We need to choose the right tool (Stryker for JS, mutmut for Python, gremlins for Go) and make it configurable.
4. **Symlink portability**: Windows and some CI environments don't support symlinks well. The installer may need a copy fallback.
5. **Reinventing CI/CD**: The harness must stay as a **developer assistant** workflow, not become a CI/CD pipeline. The boundary is the local dev loop.

## Open Questions

1. **Do Claude/OpenCode support project-local `.claude/skills/` or `.opencode/skills/`?** We need to test this or read the agent documentation.
2. **What exactly is "Hard Spec"?** Is it a new artifact, or is it just the `proposal.md` with stricter constraints?
3. **Which mutation testing framework?** The project currently has no code. We need to decide the harness implementation language (likely Node.js/TypeScript for ecosystem fit, or Go for the existing `go-testing` skill).
4. **Playwright: after mutation testing or as part of verify?** The idea says "after mutation testing" but the SDD verify phase already runs tests.
5. **Should the harness be a Go CLI, a Node.js CLI, or a shell script?** Given the existing `go-testing` skill and the user's skill ecosystem, a Node.js CLI might have better ecosystem reach, but Go is simpler for distribution.
6. **How does the "judge fails → back to TDD" loop work mechanically?** Does the orchestrator re-run `sdd-apply` with the judge's feedback as a prompt? Or does the meta-skill handle this loop?

## Next Steps

1. **Verify agent skill loading** — Research whether Claude Code, OpenCode, and Gemini CLI support project-local skill directories. This is the most critical blocker.
2. **Clarify "Hard Spec"** — Ask the user (or propose a definition) for what distinguishes "Hard Spec" from the existing `proposal.md` and `spec.md`.
3. **Choose harness implementation language** — Decide whether the CLI will be Node.js, Go, Python, or shell. This affects the mutation testing framework choice.
4. **Create a Proposal** — Once the above is clarified, run `sdd-propose` for the first slice: the installer + meta-skill.
5. **Prototype the installer** — A minimal shell script or Node.js script that symlinks global SDD skills into a project-local directory.

## Ready for Proposal

**Yes**, but with a prerequisite: we need to verify whether AI agents support project-local skill directories. If they do, the proposal can proceed immediately. If they don't, the architecture needs to pivot (e.g., using a global skill that points to project config, or lobbying for agent support).

The orchestrator should:
- Ask the user to clarify the "Hard Spec" concept
- Verify the agent's project-local skill loading capability
- Then proceed to `sdd-propose` for `ai-orchestration-harness`
