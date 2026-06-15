# Archon AI — Harness for Spec-Driven Development

**One command. Zero manual config.**

`archon init` scaffolds the complete SDD workflow (Spec → Hard Spec → Gherkin → TDD → Judge) into any project. It auto-detects your AI agent, installs 24 skills, and writes the orchestrator instructions so you can run `sdd-explore`, `sdd-apply`, and `judgment-day` without touching a single config file.

---

## Quick path

```bash
# Install with Homebrew (macOS/Linux)
brew install alexcas/archon-ai/archon

# Or with Scoop (Windows)
scoop bucket add archon https://github.com/AlexCas/scoop-bucket
scoop install archon

# Or with Go
go install github.com/archon-ai/archon/cmd/archon@latest

# Initialize in your project
archon init

# Configure everything interactively (recommended)
archon tui

# Verify
archon status
```

> **Tip:** `archon tui` is the easiest way to configure the harness — models,
> mutation testing, Playwright, and the agent — and it can even initialize a
> blank project for you. See [`tui` — Interactive Configuration](#tui--interactive-configuration-recommended).

## Install

### Homebrew (macOS/Linux)

```bash
brew tap alexcas/archon-ai
brew install archon
```

### Scoop (Windows)

```bash
scoop bucket add archon https://github.com/AlexCas/scoop-bucket
scoop install archon
```

### Go

```bash
go install github.com/archon-ai/archon/cmd/archon@latest
```

### Update

```bash
# Homebrew
brew upgrade archon

# Scoop
scoop update archon

# Go
go install github.com/archon-ai/archon/cmd/archon@latest
```

## Commands

| Command | Purpose |
|---------|---------|
| `archon init` | Scaffold SDD skills, orchestrator instructions, and config in the current project |
| `archon status` | Show current harness status (agent, version, skills) |
| `archon tui` | Interactive terminal UI for configuration |
| `archon rollback` | Remove all files created by `archon init` |
| `archon version` | Print version |

### `init` — Scaffold SDD

```bash
archon init
```

What it does:
1. Detects your AI agent (Claude, OpenCode, etc.) — and **creates the agent folder if it doesn't exist yet**, so you can bootstrap a blank project
2. Extracts 24 embedded skills into `~/.config/<agent>/skills/`
3. Creates `.archon/config.yaml` with harness metadata
4. Writes `CLAUDE.md` or `AGENTS.md` with orchestrator instructions
5. Creates `openspec/` directory structure

If a `CLAUDE.md`/`AGENTS.md` already exists, init **asks before replacing it** and aborts if you decline (use `--force` to replace without prompting).

**Flags:**

```bash
archon init --agent claude          # Override auto-detection (creates the folder if missing)
archon init --playwright            # Enable Playwright web E2E generation + execution
archon init --force                 # Replace an existing orchestrator file without prompting
archon init --dry-run               # Show what would happen without doing it
archon init --model claude-sonnet-4-6  # Default AI model for all SDD phases
```

> Prefer the interactive route? `archon tui` exposes all of these settings
> (and can run `init` for you) — see below.

### `tui` — Interactive Configuration (recommended)

```bash
archon tui
```

The fastest way to configure the harness — no hand-editing YAML. `archon tui` can
even **initialize a project from scratch**: if no `.archon/config.yaml` exists yet,
open the **Agent** tab and pick your agent; it creates the agent folder, `.archon/`,
and the orchestrator file for you.

Move between tabs with `Tab` / `Shift+Tab`, edit a tab, then press `Ctrl+S` to save.
Saving also regenerates `CLAUDE.md`/`AGENTS.md` so the orchestrator instructions stay
in sync with your config.

#### Sections

**🧠 Models** — Set the default AI model and an optional override per SDD phase
(explore, propose, spec, design, tasks, apply, verify, archive).
- Cycle the built-in catalog with `Ctrl+N` / `Ctrl+P`:
  - **Claude:** `claude-opus-4-8`, `claude-sonnet-4-6`, `claude-haiku-4-5`
  - **Opencode Go:** `deepseek-v4-flash`, `deepseek-v4-pro`, `glm-5`, `glm-5.1`, `kimi-k2.5`, `kimi-k2.6`, `qwen3.6-plus`, `qwen3.7-plus`
- Or type any model name freely — unknown models are accepted with a warning.
- Empty phase fields inherit the default model.

**🧬 Mutation Testing** — Toggle mutation testing on/off and set the kill-rate
threshold with the slider. When enabled, the **judge** phase runs mutation testing
as a quality gate (must meet the threshold to pass).

**🎭 Playwright** — Toggle Playwright web E2E testing, and set the test directory and
base URL. When enabled on a web project, the harness generates Playwright specs from
your Gherkin scenarios and runs them after the **verify** and **judge** phases.

**🤖 Agent** — Choose the AI agent (`opencode`, `claude`, `codex`, `agents`) and
run/re-run initialization. Selecting an agent creates its folder if missing; if an
orchestrator file already exists, the TUI asks before replacing it.

#### Key bindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous section |
| `↑` / `↓` | Move between fields |
| `←` / `→` | Adjust the slider (Mutation Testing) |
| `Enter` / `Space` | Toggle a switch / confirm |
| `Ctrl+N` / `Ctrl+P` | Cycle the model catalog (Models) |
| `Ctrl+S` | Save (regenerates the orchestrator file) |
| `Ctrl+Q` | Quit |

### `rollback` — Clean Removal

```bash
archon rollback
```

Removes all files created by `archon init` (`.archon/`, orchestrator instructions, project-local skill directories). Safe to run anytime — it tracks every created path.

```bash
archon rollback --dry-run  # Show what would be removed
```

## SDD Workflow

After running `archon init`, your project is ready for the full SDD cycle:

```
explore → propose → spec → design → tasks → apply → verify → judge → archive
```

### Control Gates

The orchestrator instructions (`CLAUDE.md`) enforce three mandatory control gates:

1. **SDD Session Preflight** — Forces execution mode, artifact store, PR strategy, and review budget before any SDD command
2. **Vague Request Guard** — Prevents vague requests like "add auth" from being delegated without clarification
3. **Human Review Gate** — Pauses after every phase with editable artifacts (propose, spec, design, tasks) for human approval

### Strict TDD Mode

If your project has a test runner (`go test`, `pytest`, `jest`, etc.), the harness detects it and activates Strict TDD Mode. Tests must be written before implementation during the `sdd-apply` phase.

## Configuration

Edit it the easy way with [`archon tui`](#tui--interactive-configuration-recommended),
or by hand. `.archon/config.yaml` (auto-generated):

```yaml
harness_version: "0.3.0"
agent: claude
skill_count: 24
created_at: "2026-06-11T00:00:00Z"
mutation_testing:
  enabled: false
  threshold: 0.80
playwright:
  enabled: false        # generate + run Playwright E2E from Gherkin (web projects)
  test_dir: e2e
  base_url: http://localhost:3000
models:
  default: "claude-sonnet-4-6"
  phases:
    explore: "claude-sonnet-4-6"
    propose: "claude-sonnet-4-6"
    spec: "claude-sonnet-4-6"
    design: "claude-sonnet-4-6"
    tasks: "claude-sonnet-4-6"
    apply: "claude-sonnet-4-6"
    verify: "claude-sonnet-4-6"
    archive: "claude-sonnet-4-6"
skill_inventory:
  - name: sdd-init
    version: "1.0"
    source: embedded
  # ... 23 more
```

You can also set individual values from the CLI, e.g.
`archon config set playwright.enabled true` or `archon config set models.default claude-opus-4-8`.

## Architecture

```
archon CLI → cobra root
  ├── internal/init (orchestrator)
  ├── internal/agent (detect: scan .opencode/, .claude/, etc.)
  ├── internal/scaffold (embed.FS → skill directories)
  ├── internal/config (read/write .archon/config.yaml)
  └── internal/tui (interactive terminal UI)
```

## Requirements

- Go 1.25+ (for building from source)
- macOS or Linux (Windows support planned)
- Terminal with color support (for TUI)

## Checklist

- [ ] `brew install archon` or `go install` works
- [ ] `archon init` scaffolds skills without errors
- [ ] `archon status` shows agent and version
- [ ] `archon tui` launches and saves config
- [ ] `archon rollback` removes all created files

## Next step

Read the orchestrator instructions in your project: `cat CLAUDE.md` or `cat AGENTS.md`. Then run:

```bash
archon tui
```

Configure your models and start the SDD workflow.

---

**License**: MIT
**Repository**: https://github.com/AlexCas/archon-ai
