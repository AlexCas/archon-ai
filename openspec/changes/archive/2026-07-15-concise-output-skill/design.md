# Design: Concise Chat Output Skill

## Technical Approach

Approach 3 (approved): one new behavior skill + a matching persona pointer in both the
template source of truth and the live `CLAUDE.md`. The skill (`skills/concise-output/SKILL.md`)
carries the full contract; the 2–4 line persona pointer guarantees the reflex fires on
every orchestrator chat reply. `CLAUDE.md` is edited SURGICALLY (never regenerated) to
protect PR-1's hand-edited Rule 2 and Phase Models sections. Maps to spec.md's six ADDED
requirements; no capability spec is modified.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| Where the contract lives | Full contract in SKILL.md; short pointer in persona | All-in-persona (Approach 2) | Keeps always-loaded persona lean; pointer only guarantees activation |
| Activation model | Passive reflex; `disable-model-invocation: true`, `user-invocable: false` | comment-writer-style user-invocable | Spec req "Passive Activation"; not a slash command |
| CLAUDE.md update | Surgical Edit between Leader Persona and Phase Order | `archon init` regen | Regen reverts PR-1 Rule 2 + Phase Models hand-edits (desync is out of scope) |
| Skill body language | English body; Spanish gate string quoted verbatim | Spanish body | Repo artifact default; verbatim gate string preserves the allow-list literal |
| Persona placement | New `## Concise Chat Output` H2 after Leader Persona, before Phase Order | Inside Rules list | Persona-level reflex belongs with persona sections; mirrors both files |

## Data Flow

    archon init/update ──embed glob (skills/*/SKILL.md)──→ ~/.config/opencode/skills
         │                                                        │
    templates.go (source) ──render (unchanged path)──→ CLAUDE.md (surgical edit, NOT regen)
         │                                                        │
         └──────── orchestrator loads persona pointer ───────────┘
                              │
                    chat reply composed: concise-by-default,
                    preserve-verbatim allow-list enforced

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `skills/concise-output/SKILL.md` | Create | New passive behavior skill; auto-embedded via glob |
| `internal/initcmd/templates.go` | Modify | Add `## Concise Chat Output` subsection to `orchestratorSections` (backticks as `§`) |
| `CLAUDE.md` | Modify | Surgical insert of identical section after Leader Persona; NO regen |

### Deliverable 1 — `skills/concise-output/SKILL.md` content plan

Frontmatter (comment-writer style + passive gates from sdd-design frontmatter):

```yaml
---
name: concise-output
description: "Keep orchestrator chat replies concise by default while preserving gates verbatim. Trigger: every orchestrator chat reply to the user."
disable-model-invocation: true
user-invocable: false
license: Apache-2.0
metadata:
  version: "1.0"
---
```

`description` is 1 quoted line, Trigger-first clause included, <=250 chars. English body,
sections in order:

1. **When to Use** — scope = the orchestrator's CHAT output to the user ONLY. Explicitly
   NOT subagent handoff prompts and NOT SDD artifact bodies; not the `detail_level` axis.
2. **Hard Rules** (table) — concise-by-default (lead with the point; tight bullets or 1–3
   short paragraphs; drop narration/preamble/recap of visible work); no em dashes optional.
3. **Preserve-Verbatim Allow-List** (numbered, NEVER trim / always complete):
   1. Human Review Gate question, quoted verbatim:
      `¿Quieres ajustar algo en esta fase antes de continuar?`
   2. Decision tables shown to the user (e.g., preflight A–E option groups).
   3. Risks and open-question lists from any SDD phase.
   4. Substantive content of SDD artifacts shown to the user (proposal/spec/design/tasks).
4. **Must-Not-Weaken** — defer to these on any conflict: Leader Persona language rule
   (reply in user's language; neutral Spanish), warm/direct tone, Human Review Gate
   SHOW+ASK, Vague Request Guard, Commit Attribution.
5. **Decision Gate** — "When in doubt, keep it": if unsure whether content is filler or
   load-bearing, KEEP it. Concise never overrides a gate or drops the allow-list.
6. **Examples** — one trimmed status update (before/after); one phase-end reply showing the
   verbatim gate string retained.

Target 180–450 tokens (skill-creator budget).

### Deliverable 2 — `internal/initcmd/templates.go` persona section

Insert into `orchestratorSections`, immediately AFTER the Leader Persona **Behavior**
paragraph (line 27) and BEFORE `## Phase Order` (line 29). Backticks written as `§`:

```
## Concise Chat Output

Your chat replies to the user are concise by DEFAULT: lead with the actionable point,
prefer a tight bullet list or 1–3 short paragraphs, and drop narration, preamble, and
recap of work already visible. This applies ONLY to chat output — never to subagent
handoff prompts or SDD artifact bodies. See the §concise-output§ skill for the full
contract.

PRESERVE VERBATIM, always complete (never trim): the Human Review Gate question
"¿Quieres ajustar algo en esta fase antes de continuar?", decision tables, risks and
open-question lists, and the substantive content of SDD artifacts shown to the user.
Concise must NOT weaken the Leader Persona language/tone rules or any gate. When in
doubt, keep it.
```

### Deliverable 3 — `CLAUDE.md` surgical edit strategy

The identical section (with LITERAL backticks around `concise-output`) is inserted in the
on-disk `CLAUDE.md` at the SAME anchor: after the Leader Persona `**Behavior**:` paragraph
ending "...acknowledge it with evidence and correct it." and before `## Phase Order`.

- Method: `Edit` tool, single insertion. `old_string` = the `**Behavior**:` paragraph +
  the following `## Phase Order` line; `new_string` = same + the new section spliced between.
- NO `archon init` / `RenderClaudeMD` regen. Verify/apply must not run a regen command.
- Post-edit invariants to check: `CLAUDE.md` Rule 2 wording ("invoking its `archon-<phase>`
  subagent … the subagent's frontmatter model is the gate") and the Phase Models section
  ("binding hard gate") remain byte-identical.

## Interfaces / Contracts

No Go interface, API, or type changes. The only "contract" is the SKILL.md frontmatter
schema (name/description/license/metadata.version) already enforced by convention, and the
`## Concise Chat Output` heading text, which the verify checks grep for.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit (existing) | Embed + extract + version still green after new skill | `go test ./skills/... ./internal/scaffold/...` (no test edits) |
| Mechanical verify | File exists; frontmatter has name=concise-output, description<=250, license, metadata.version | Direct file read / grep |
| Mechanical verify | Persona pointer present in `CLAUDE.md` AND `templates.go` (heading `## Concise Chat Output`) | grep both files |
| Doc/behavior review | All runtime LLM-behavior scenarios | Review checklist — not unit-testable |

## Gherkin (13) → Verification Mapping

Mechanically enforced (~3): file-existence/frontmatter, CLAUDE.md pointer, templates.go
pointer. Everything else is doc/behavior review (per spec.md Enforcement Notes).

| # | Scenario | Enforcement |
|---|----------|-------------|
| 1 | Verbose narration is trimmed | Behavior review |
| 2 | Human Review Gate question preserved verbatim | Behavior review (string also grep-checkable in skill + templates.go) |
| 3 | Decision tables preserved complete | Behavior review |
| 4 | Risks / open-question lists preserved complete | Behavior review |
| 5 | SDD artifact content not truncated | Behavior review |
| 6 | Spanish + tone honored under concise mode | Behavior review |
| 7 | Human Review Gate not bypassed | Behavior review |
| 8 | Ambiguous content kept, not trimmed | Behavior review |
| 9 | Skill passive, not invocable | Mechanical-ish: grep frontmatter `user-invocable: false` / `disable-model-invocation: true` |
| 10 | Skill does not alter subagent handoff prompts | Behavior review (skill scope text asserts it) |
| 11 | **Skill file exists + well-formed frontmatter** | **Mechanical** — file read + frontmatter fields |
| 12 | **Persona pointer present in CLAUDE.md** | **Mechanical** — grep `## Concise Chat Output` + skill name |
| 13 | **Persona pointer present in templates.go** | **Mechanical** — grep `## Concise Chat Output`, consistent |

The ~3 fully mechanical scenarios are 11, 12, 13; scenario 9 is a lightweight frontmatter
grep (near-mechanical). Scenario 2's literal string is grep-verifiable in the two source
files even though the runtime behavior is review-only.

## Migration / Rollout

No migration. Single-commit revert restores prior behavior (delete skill file + revert the
two text edits). `.archon/config.yaml` `skill_count` (24→25) / `skill_inventory` drift is
expected and regenerates on the next `archon update` — NOT hand-edited here.

## Line Estimate vs. 400-Line Budget

| File | Est. lines |
|------|-----------|
| `skills/concise-output/SKILL.md` (new) | ~75 |
| `internal/initcmd/templates.go` (insert) | ~14 |
| `CLAUDE.md` (insert) | ~12 |
| **Total changed** | **~101** |

Well under the 400-line review budget; single PR (per this session's single-PR strategy).

## Open Questions

- None blocking. Deferred (out of scope, confirmed at explore/propose gates):
  template↔CLAUDE.md Rule 2 / Phase Models desync back-port; `.archon/config.yaml`
  skill_count/skill_inventory hand-bump; `skills/embed_test.go` expected-list extension.
