# Archon AI — Harness for Spec-Driven Development

**One command. Zero manual config.**

`archon init` scaffolds the complete SDD workflow (Spec → Hard Spec → Gherkin → TDD → Judge) into any project. It auto-detects your AI agent, installs 25 skills, and writes the orchestrator instructions so you can run `sdd-explore`, `sdd-apply`, and `judgment-day` without touching a single config file.

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
| `archon update` | Refresh installed skills from the embedded set without touching your config or orchestrator file |
| `archon status` | Show current harness status (agent, version, skills) |
| `archon tui` | Interactive terminal UI for configuration |
| `archon config` | Get/set/list config values (`archon config set`, `get`, `list`) |
| `archon rollback` | Remove all files created by `archon init` |
| `archon version` | Print version |

### `init` — Scaffold SDD

```bash
archon init
```

What it does:
1. Detects your AI agent (Claude, OpenCode, etc.) — and **creates the agent folder if it doesn't exist yet**, so you can bootstrap a blank project
2. Extracts 25 embedded skills into `~/.config/<agent>/skills/`
3. Creates `.archon/config.yaml` with harness metadata
4. Writes `CLAUDE.md` or `AGENTS.md` with orchestrator instructions
5. Creates `openspec/` directory structure

If a `CLAUDE.md`/`AGENTS.md` already exists, init **asks before replacing it** and aborts if you decline (use `--force` to replace without prompting).

**Flags:**

```bash
archon init --agent claude          # Override auto-detection (creates the folder if missing)
archon init --playwright            # Enable Playwright web E2E generation + execution
archon init --impeccable            # Enable the Impeccable design-language gate
archon init --force                 # Replace an existing orchestrator file without prompting
archon init --dry-run               # Show what would happen without doing it
archon init --model claude-sonnet-4-6  # Default AI model for all SDD phases
```

> Prefer the interactive route? `archon tui` exposes all of these settings
> (and can run `init` for you) — see below.

### `update` — Refresh Skills

```bash
archon update
```

Upgrades the installed skills to the version embedded in your current `archon` binary,
**without** rewriting `CLAUDE.md`/`AGENTS.md` or resetting your config. It compares the
embedded skills against what's installed and classifies each as **added**, **changed**,
or **orphaned** (installed but no longer shipped). Your `models`, `playwright`,
`mutation_testing`, `judge`, `created_at`, and `agent` settings are preserved — only
`harness_version`, `skill_count`, and `skill_inventory` are updated. When there's no
gap, it reports "already up to date" and writes nothing.

> **Machine-wide scope:** skills live in a shared directory
> (`~/.config/opencode/skills/`) that projects symlink into. Refreshing it affects
> **every** project linked to that directory. `archon update` always prints the scope so
> you know what's impacted.

**Flags:**

```bash
archon update --check    # Report the diff (added/changed/orphaned) without writing anything
archon update --prune    # Also remove orphaned skills (kept by default)
archon update --agent claude  # Override the agent recorded in config
```

If a project's skills were installed as **real directories** instead of symlinks
(copy-mode, e.g. on systems without symlink support), `archon update` refreshes the
shared directory but **does not** re-link that project. It emits a warning telling you to
re-run `archon init` in that project to refresh its own copy.

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

The TUI opens on the **Agent** tab. Tabs, in order: **Agent · Models · Judge ·
Mutation Testing · Playwright · Security · Impeccable**.

**🤖 Agent** — Choose the AI agent (`opencode`, `claude`, `codex`, `agents`) and
run/re-run initialization. Selecting an agent creates its folder if missing; if an
orchestrator file already exists, the TUI asks before replacing it.

**🧠 Models** — Set the default (and optional leader) model plus an optional override
per SDD phase (explore, propose, spec, design, tasks, apply, verify, **judge**,
archive). Each row is a structured `provider/model` reference with an optional
reasoning **effort** (`default`, `low`, `medium`, `high`) for reasoning-capable models.
- Providers and models are sourced from your opencode catalog; pick a provider, then a
  model, then an effort — or drop into free-form entry to type any `provider/model`
  string. A live search filter narrows long model lists as you type.
- Empty phase fields inherit the default model.

**⚖️ Judge** — Toggle the judgment-day review gate that runs after `verify`.

**🧬 Mutation Testing** — Toggle mutation testing on/off and set the kill-rate
threshold with the slider. When enabled, the **judge** phase runs mutation testing
as a quality gate (must meet the threshold to pass).

**🎭 Playwright** — Toggle Playwright web E2E testing, and set the test directory and
base URL. When enabled on a web project, the harness generates Playwright specs from
your Gherkin scenarios and runs them after the **verify** and **judge** phases.

**🛡️ Security** — Toggle the opt-in security baseline and pick a profile (`cli` or
`web`). When enabled, the security baseline is woven into the SDD phases.

**🎨 Impeccable** — Toggle the opt-in Impeccable design-language gate, auto-install,
severity, and the product/design doc paths. When enabled, the harness runs
Impeccable design verbs during apply and the `npx impeccable detect` gate after
the judge phase.

#### Key bindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous section |
| `↑` / `↓` | Move between fields / list entries |
| `←` / `→` | Adjust the slider (Mutation Testing) |
| `Enter` / `Space` | Open a picker, confirm a choice, or toggle a switch |
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
harness_version: "0.9.0"
agent: claude
skill_count: 25
created_at: "2026-06-11T00:00:00Z"
mutation_testing:
  enabled: false
  threshold: 0.80
judge:
  enabled: true          # run the judgment-day review gate after verify
playwright:
  enabled: false         # generate + run Playwright E2E from Gherkin (web projects)
  test_dir: e2e
  base_url: http://localhost:3000
security:
  enabled: false         # weave the opt-in security baseline into the SDD phases
  profile: cli           # "cli" | "web"
impeccable:
  enabled: false         # run the Impeccable design-language gate after judge
  auto_install: false    # run `npx impeccable install` once before the first gate run
  severity: block-deterministic  # "block-deterministic" | "block-all" | "advisory"
models:
  default:               # structured provider/model with optional reasoning effort
    provider: anthropic
    model: claude-sonnet-4-6
  phases:
    explore:  { provider: anthropic, model: claude-opus-4-8 }
    propose:  { provider: anthropic, model: claude-opus-4-8 }
    spec:     { provider: anthropic, model: claude-sonnet-4-6 }
    design:   { provider: anthropic, model: claude-opus-4-8 }
    tasks:    { provider: anthropic, model: claude-sonnet-4-6 }
    apply:    { provider: anthropic, model: claude-sonnet-5 }
    verify:   { provider: anthropic, model: claude-opus-4-8 }
    judge:    { provider: anthropic, model: claude-sonnet-4-6 }
    archive:  { provider: anthropic, model: claude-haiku-4-5 }
skill_inventory:
  - name: sdd-init
    version: "1.0"
    source: embedded
  # ... 24 more
```

You can also set individual values from the CLI, e.g.
`archon config set playwright.enabled true` or `archon config set models.default claude-opus-4-8`,
and inspect the model configuration with `archon config list`.

## Architecture

```
archon CLI → cobra root
  ├── internal/initcmd (orchestrator + per-phase subagents)
  ├── internal/agent (detect: scan .opencode/, .claude/, etc.)
  ├── internal/scaffold (embed.FS → skill directories)
  ├── internal/config (read/write .archon/config.yaml)
  ├── internal/models (structured provider/model refs)
  ├── internal/opencode (read the opencode model catalog cache)
  ├── internal/status (harness status reporting)
  ├── internal/version (build version)
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
