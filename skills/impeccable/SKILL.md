---
name: impeccable
description: "Trigger: impeccable.enabled, design-language gate, frontend design quality. Orchestrate WHEN and HOW Archon calls the external `npx impeccable` tool across SDD phases."
license: MIT
metadata:
  
  version: "1.0"
  scope: reference
---

## Activation Contract

Load this skill from any phase skill (`sdd-design`, `sdd-apply`, `sdd-verify`,
`sdd-tasks`, `sdd-explore`, `sdd-spec`, `harness-judge`) when
`.archon/config.yaml` → `impeccable.enabled: true`. When the flag is absent or
`false`, no phase should read this file — treat the gate as fully inert.

This skill is a **thin orchestration layer**. It does NOT reimplement
[Impeccable](https://github.com/pbakaus/impeccable)'s 58-rule deterministic
detector or its LLM design critique. Every actual check runs inside the `npx
impeccable` tool or the AI agent's own slash-command execution — Archon only
decides when to call it and how to interpret the result.

## Two Invocation Surfaces (do not confuse them)

Impeccable exposes two distinct surfaces. Mixing them up is the most common
integration mistake:

| Surface | Commands | Invocation | Who runs it |
|---------|----------|-----------|-------------|
| **npx CLI** | `install`, `update`, `detect` | `npx impeccable <cmd>` | Shelled out by the harness (judge gate, auto_install) |
| **Agent slash-commands** | `init` + 23 design verbs (`craft`, `shape`, `critique`, `audit`, `polish`, `bolder`, `quieter`, `distill`, `harden`, `onboard`, `animate`, `colorize`, `typeset`, `layout`, `delight`, `overdrive`, `clarify`, `adapt`, `optimize`, `live`, `document`, `extract`, `pin`) | `/impeccable <cmd>` inside the AI coding tool | Run by the agent itself, NOT shelled out |

Never shell out to `npx impeccable init` or `npx impeccable <verb>` — those are
agent-run slash commands. Never expect the harness to run design verbs on its
own — only `install`, `update`, and `detect` are real shell commands.

## Per-Phase Invocation Map

| Phase | Action | Mechanism |
|-------|--------|-----------|
| `sdd-design` | Read `PRODUCT.md`/`DESIGN.md` (if present) as design input context | Read-only file access — no command |
| `sdd-tasks` | Emit an "Impeccable pass" task when the change touches frontend files | Prose task, no invocation |
| `sdd-apply` | Run `/impeccable <verb>` (craft/polish/harden/animate/...) on frontend-affecting changes | Agent slash command |
| `sdd-verify` | Advisory note if Impeccable hooks/artifacts are absent | No invocation — presence check only |
| `harness-judge` (Step 3c) | Run the detection gate | `npx impeccable detect --json .` |

## Detect Invocation (Judge Gate)

The judge gate is the ONLY place that executes Impeccable:

```
npx impeccable detect --json .
```

Run from the target-project root. `--json` is REQUIRED — it is the only
CI-friendly, parseable output format. Other useful flags: `--no-config` (raw
scan, ignore project context), `--no-inline-ignores` (bypass inline waivers).
`detect` also accepts a specific target (a dir, a file, or a URL) instead of
`.` when the gate needs to scope to changed files only.

### Exit code / JSON interpretation

Impeccable's README does not document a non-zero exit code on violations —
**never rely on exit code for pass/fail**. Use exit code ONLY to distinguish
"ran successfully" from "tool missing / crashed":

1. Parse the `--json` payload into two buckets: deterministic-detector
   violations and LLM-critique findings.
2. If the JSON is unparseable or the shape is unrecognized, treat all findings
   as advisory and note the parse failure — never hard-fail on a parsing
   problem alone.
3. A non-zero exit combined with no usable output means the tool crashed or
   was not found — return `blocked`, never a silent pass.

### Severity mapping

`impeccable.severity` decides how the two finding buckets map to a verdict:

| Severity | Deterministic violations | LLM critique findings |
|----------|--------------------------|------------------------|
| `block-deterministic` (default) | `fail` | advisory only |
| `block-all` | `fail` | `fail` |
| `advisory` | advisory only | advisory only |

## Node / npx Missing (blocked, never silent-pass)

If `node` or `npx` is not on PATH in the target project:

```
Impeccable requires Node.js and npx. Install Node.js or set
impeccable.enabled: false to skip this gate.
```

## auto_install Semantics

- `auto_install: false` (default): assume Impeccable is already installed. If
  `npx impeccable` reports the package is not found, return `blocked` with:
  ```
  Impeccable is not installed. Run 'npx impeccable install' (or set
  impeccable.auto_install: true), or set impeccable.enabled: false to skip
  this gate.
  ```
  Never install silently.
- `auto_install: true`: run `npx impeccable install` once before the first
  gate invocation, then continue with `detect`. This is a one-time setup
  action, not a repeated install on every run.

## PRODUCT.md / DESIGN.md Ownership

- Both files live at the **target-project root** (or `product_path`/
  `design_path` if configured) — Impeccable's own default location.
- They are generated by the agent running **`/impeccable init`** (a slash
  command, not `npx impeccable init`).
- Archon never owns, generates, or overwrites these files. `sdd-design` only
  READS them as input context; the SDD `design.md` artifact
  (`openspec/changes/<change>/design.md`) is a completely separate file and is
  never replaced by Impeccable's docs.

## Rules

- NEVER reimplement the deterministic detector or the LLM critique in Go code
  or skill prose — the only invocation mechanism for checks is `npx
  impeccable <subcommand>` or an agent-run `/impeccable <verb>` slash command.
- NEVER shell out to a slash-command name (`init` or any of the 23 design
  verbs) — those only work when run by the agent inside its own turn.
- NEVER rely on `detect`'s exit code alone for pass/fail — parse `--json`.
- NEVER silently install Impeccable — respect `auto_install`.
- When `impeccable.enabled: false`, this skill is not loaded and no phase
  changes behavior.
