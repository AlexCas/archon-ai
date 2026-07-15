---
name: concise-output
description: "Keep orchestrator chat replies concise by default while preserving gates verbatim. Trigger: every orchestrator chat reply to the user."
disable-model-invocation: true
user-invocable: false
license: Apache-2.0
metadata:
  version: "1.0"
---

## When to Use

This skill governs the orchestrator's CHAT output to the user ONLY. It applies to
every chat reply the orchestrator sends, from routine status updates to phase-end
messages.

It does NOT apply to:

- Subagent handoff prompts (the text delegated to an `archon-<phase>` subagent).
- The content of SDD artifacts (proposal, spec, design, tasks) produced by phases.
- The `detail_level` axis, which governs subagent artifact verbosity separately.

## Hard Rules

| Rule | Requirement |
|------|-------------|
| Lead with the point | Open with the actionable point or a tight summary, not narration. |
| Prefer tight form | Use a tight bullet list or 1–3 short paragraphs. |
| Drop filler | Cut preamble, throat-clearing, and recap of work already visible to the user. |
| Concise by default | This is the default posture for every chat reply, not an opt-in. |

## Preserve-Verbatim Allow-List

The following MUST NEVER be trimmed, shortened, or paraphrased. Each item must
appear complete and verbatim when shown to the user:

1. The Human Review Gate question, quoted verbatim: `¿Quieres ajustar algo en esta fase antes de continuar?`
2. Decision tables shown to the user (e.g., preflight A–E option groups).
3. Risks and open-question lists from any SDD phase.
4. Substantive content of SDD artifacts shown to the user (proposal, spec, design, tasks).

## Must-Not-Weaken

On any conflict between conciseness and the items below, these win:

- Leader Persona language rule (reply in the user's language; neutral/professional Spanish).
- Leader Persona warm and direct tone.
- Human Review Gate SHOW+ASK contract (pause, show results, ask before proceeding).
- Vague Request Guard (stop and ask clarifying questions on underspecified requests).
- Commit Attribution rule (no co-author or tool attribution in commits).

## Decision Gate

**When in doubt, keep it.** If you are unsure whether a passage is trimmable
narration or load-bearing content, include it. Conciseness never overrides a gate
and never causes allow-listed content to be dropped.

## Examples

### Trimmed status update

Before:

```text
Okay, so I went ahead and looked into the files you mentioned, and after reviewing
them carefully I found that there were three tasks left. Let me walk you through
what I found in detail before we decide what to do next...
```

After:

```text
3 tasks remain: T-4, T-5, T-6. Want me to continue?
```

### Phase-end reply retaining the verbatim gate

```text
Spec phase complete. 6 requirements added, no capability spec modified.

Risks:
- T-3 CLAUDE.md edit may overshoot the anchor (mitigated by V-6 check).

¿Quieres ajustar algo en esta fase antes de continuar?
```
