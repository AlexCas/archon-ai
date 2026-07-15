# Design: AI Orchestration Harness CLI

## Architecture Overview

```
archon CLI → cobra root
  ├── internal/init (orchestrator)
  ├── internal/agent (detect: scan .opencode/, .claude/, etc.)
  ├── internal/scaffold (embed.FS → ~/.config/opencode/skills/)
  └── internal/config (read/write .archon/config.yaml + rollback.json)
```

## Architecture Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| CLI | `cobra` v1.8 | Subcommand auto-gen, flags, help |
| Config | YAML (`yaml.v3`) | Matches openspec; human-readable |
| Embed | `//go:embed */SKILL.md` | Zero deps; single binary |
| Agent priority | `.opencode/` > `.claude/` > `.agents/` > `.codex/` | Likelihood order; prompt on conflict |
| Symlink fallback | Copy on `EPERM`/`EINVAL` | Windows without dev mode |
| Rollback | `.tmp` + `os.Rename` | POSIX-atomic |
| State format | YAML | Same family as config |
| Errors | Sentinel + `%w`; `os.Exit(1)` at boundary | Test via `errors.Is/As` |

## Go Package Layout

```
cmd/archon/main.go           # cobra root, ldflags
internal/
  agent/detect.go            # detect()+resolve() → agent, config dir
  config/config.go           # Load/Save .archon/config.yaml
  config/rollback.go         # Track, WriteManifest, Cleanup
  scaffold/extract.go        # embed.FS iterate → ~/.config/opencode/skills/
  scaffold/symlink.go        # SymlinkOrCopy to project-local
  init/init.go               # Orchestrator: detect→extract→symlink→config→template
  init/templates.go          # Embedded AGENTS.md/CLAUDE.md template
  version/info.go            # Ldflags: Version, Commit, Date
  status/display.go          # Read config, format output
skills/                      # embed.FS root (23 entries)
  embed.go                   # //go:embed */SKILL.md
  _shared/                   # Shared SDD references
  sdd-init/ ... sdd-archive/ # 10 SDD workflow skills
  judgment-day/ ...          # 11 supporting skills
  harness-workflow/          # New meta-skill
  harness-judge/             # New meta-skill
```

## Go Embed & Extraction

`skills/embed.go`:
```go
//go:embed */SKILL.md
var FS embed.FS
```

Extraction: `fs.ReadDir` → 23 dirs → `fs.ReadFile` → `os.MkdirAll` + `os.WriteFile`. **Idempotency**: compare `metadata.version` in YAML frontmatter; skip if installed ≥ embedded. Version gap → prompt `"Update skills? [y/N]"`.

## CLI Commands

```
archon init       [--agent <name>] [--force]
archon rollback   [--dry-run]
archon version
archon status
```

**`init` flow**:
1. `agent.Detect()` — scan `.opencode/`, `.claude/`, `.agents/`, `.codex/`; if multi-agent → interactive prompt
2. `scaffold.Extract()` → 23 skills to global dir
3. `scaffold.SymlinkOrCopy()` → project-local dir for agent
4. `config.Write()` → `.archon/config.yaml`
5. `config.RollbackTrack()` → record every path
6. Write orchestrator template to `AGENTS.md` / `CLAUDE.md`

## Config Schema

```yaml
# .archon/config.yaml
harness_version: "1.0.0"
agent: opencode
skill_count: 23
created_at: "2026-06-10T00:00:00Z"
mutation_testing:
  enabled: false
  tool: gremlins
  threshold: 0.80
skill_inventory:
  - {name: sdd-init, version: "2.0", source: embedded}
  # ...22 more
```

## Rollback Schema

```json
{
  "version": "1.0.0",
  "paths": [".archon/config.yaml", ".archon/rollback.json", "..."],
  "original_agents_md_backup": "AGENTS.md.backup.20260610"
}
```
Paths removed in reverse order. `AGENTS.md` restored from backup.

## Meta-Skill: harness-workflow

**State machine** (strictly linear):
```
explore → propose → spec → design → tasks → apply → verify → judge → archive
```
Each phase: `in_progress` | `completed`. Transitions:
- `completed(N)` → `in_progress(N+1)` ✓
- `completed(N)` → `completed(N+2)` ✗ (blocked — reports missing phases)

`openspec/changes/{name}/state.yaml`:
```yaml
phase: design
status: completed
history:
  - {phase: explore, status: completed, ts: ...}
  - {phase: propose, status: completed, ts: ...}
```

SKILL.md gate: orchestrator checks `state.yaml` before every SDD skill invocation; blocks invalid transitions.

## Meta-Skill: harness-judge

```
judge ──pass──▶ archive
  │ fail
  ▼ (retry < 3)
feedback ──▶ sdd-apply ──▶ verify ──▶ judge
  │ retry == 3
  ▼
blocked
```

1. Invoke `judgment-day` dual review
2. If `mutation_testing.enabled` → run `gremlins` on changed Go files
3. Pass → advance. Fail → structured feedback (`## Issues`, `## Action Required`) → auto re-apply → auto verify → re-judge
4. Max 3 cycles; 4th failure → `blocked`

## Mutation Integration

`gremlins unleash --threshold 0.80 --output json {files...}` → parse surviving mutants → map to files/requirements via line numbers → include in feedback block. Skipped when `enabled: false`.

## Orchestrator Template

```markdown
# ARCHON AI Orchestrator
Phase order: explore→propose→spec→design→tasks→apply→verify→judge→archive
Rules:
1. Check harness-workflow before any phase
2. Delegate each phase to sdd-* sub-agent
3. After verify, invoke harness-judge
4. On judge fail: re-apply with feedback (max 3)
Skills: 23 (embedded via archon init)
Config: .archon/config.yaml
```

## Error Handling

| Condition | Behavior |
|-----------|----------|
| Permission denied / disk full | `os.IsPermission` → exit(1) |
| Agent not detected | Interactive prompt; no init without resolution |
| Corrupt state.yaml | Exit with `"run archon init --force"` |
| Missing embedded skill | Fatal: corrupt build |
| Judge max retries | `blocked` + accumulated issues; user resolves |

No partial writes — temp file + rename for all YAML/JSON persistence.

## Testing

| Layer | Tooling |
|-------|---------|
| Unit | Table-driven + `fstest.MapFS` for all `internal/` packages |
| Integration | Golden files (`gotest.tools/golden`); `archon init` in `t.TempDir()` |
| E2E | `init → status → rollback` cycle; `state.yaml` transitions |

Inject `HomeDir` via config struct — never `os.Getenv("HOME")` directly.

## Build & Distribution

```bash
go build -ldflags="-X 'version.Version=...' -X 'version.Commit=...' -X 'version.Date=...'" ./cmd/archon
go install github.com/alexcasdev/archon-ai/cmd/archon@v1.0.0
```

GoReleaser → cross-compile (linux/darwin amd64+arm64) + Homebrew formula on tag push.
