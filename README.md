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

# Verify
archon status
```

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
1. Detects your AI agent (Claude, OpenCode, etc.)
2. Extracts 24 embedded skills into `~/.config/<agent>/skills/`
3. Creates `.archon/config.yaml` with harness metadata
4. Writes `CLAUDE.md` or `AGENTS.md` with orchestrator instructions
5. Creates `openspec/` directory structure

**Flags:**

```bash
archon init --agent claude      # Override auto-detection
archon init --force            # Re-initialize even if already done
archon init --dry-run          # Show what would happen without doing it
archon init --model claude-sonnet-4  # Default AI model for all SDD phases
```

### `tui` — Interactive Configuration

```bash
archon tui
```

Launch a terminal UI to configure:
- **Models tab**: Set AI models per SDD phase (explore, propose, spec, design, tasks, apply, verify, archive)
- **Mutation Testing tab**: Toggle mutation testing and set threshold
- **Agent tab**: Switch agent and re-run initialization

**Key bindings:**

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `Ctrl+S` | Save |
| `Ctrl+Q` | Quit |

When you save, the TUI automatically regenerates the orchestrator instructions (`CLAUDE.md` or `AGENTS.md`) so they stay in sync with your config.

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

`.archon/config.yaml` (auto-generated):

```yaml
harness_version: "0.1.0"
agent: claude
skill_count: 24
created_at: "2026-06-11T00:00:00Z"
mutation_testing:
  enabled: false
  threshold: 0.80
models:
  default: "claude-sonnet-4"
  phases:
    explore: "claude-sonnet-4"
    propose: "claude-sonnet-4"
    spec: "claude-sonnet-4"
    design: "claude-sonnet-4"
    tasks: "claude-sonnet-4"
    apply: "claude-sonnet-4"
    verify: "claude-sonnet-4"
    archive: "claude-sonnet-4"
skill_inventory:
  - name: sdd-init
    version: "1.0"
    source: embedded
  # ... 23 more
```

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
