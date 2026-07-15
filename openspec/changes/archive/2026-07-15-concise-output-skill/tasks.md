# Tasks: Concise Chat Output Skill

## Change
concise-output-skill

## Phase
tasks

## Ordered Task List

### T-1 — Create `skills/concise-output/` directory and `SKILL.md` [x]

**File:** `skills/concise-output/SKILL.md` (new, ~75 lines)

Create the file with the following frontmatter (exact fields, exact values):

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

Body sections in order (English; target 180–450 tokens):

1. **When to Use** — scope = the orchestrator's CHAT output to the user ONLY; explicitly NOT subagent handoff prompts and NOT SDD artifact bodies; NOT the `detail_level` axis.
2. **Hard Rules** (table) — concise-by-default: lead with the point; prefer tight bullets or 1–3 short paragraphs; drop narration, preamble, and recap of visible work.
3. **Preserve-Verbatim Allow-List** (numbered, NEVER trim / always complete):
   1. Human Review Gate question, quoted verbatim: `¿Quieres ajustar algo en esta fase antes de continuar?`
   2. Decision tables shown to the user (e.g., preflight A–E option groups).
   3. Risks and open-question lists from any SDD phase.
   4. Substantive content of SDD artifacts shown to the user (proposal/spec/design/tasks).
4. **Must-Not-Weaken** — defer to these on any conflict: Leader Persona language rule (reply in user's language; neutral Spanish), warm/direct tone, Human Review Gate SHOW+ASK, Vague Request Guard, Commit Attribution.
5. **Decision Gate** — "When in doubt, keep it": if unsure whether content is filler or load-bearing, KEEP it. Concise never overrides a gate or drops the allow-list.
6. **Examples** — one trimmed status update (before/after); one phase-end reply showing the verbatim gate string retained.

**Checkable:** file exists at `skills/concise-output/SKILL.md`.

---

### T-2 — Insert `## Concise Chat Output` persona section into `internal/initcmd/templates.go` [x]

**File:** `internal/initcmd/templates.go`

In `orchestratorSections`, insert the following block immediately AFTER the `**Behavior**:` paragraph (ending `"...acknowledge it with evidence and correct it."`) and BEFORE the existing `## Phase Order` line. Write backticks as `§` per the file convention.

Block to insert (backticks as `§`):

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

**Checkable:** `grep -n "## Concise Chat Output" internal/initcmd/templates.go` returns a line.

---

### T-3 — SURGICAL insert of `## Concise Chat Output` into `CLAUDE.md` [x]

**File:** `CLAUDE.md`

Using the Edit tool, insert the identical section (with LITERAL backticks around `concise-output`) between the `**Behavior**:` paragraph and `## Phase Order`. The `old_string` must span the end of the Behavior paragraph through the `## Phase Order` heading to guarantee uniqueness; `new_string` splices the new section between them.

Do NOT run `archon init` / `RenderClaudeMD`. Do NOT regenerate the file.

The inserted block (literal backticks — this is the on-disk file, not a template):

```
## Concise Chat Output

Your chat replies to the user are concise by DEFAULT: lead with the actionable point,
prefer a tight bullet list or 1–3 short paragraphs, and drop narration, preamble, and
recap of work already visible. This applies ONLY to chat output — never to subagent
handoff prompts or SDD artifact bodies. See the `concise-output` skill for the full
contract.

PRESERVE VERBATIM, always complete (never trim): the Human Review Gate question
"¿Quieres ajustar algo en esta fase antes de continuar?", decision tables, risks and
open-question lists, and the substantive content of SDD artifacts shown to the user.
Concise must NOT weaken the Leader Persona language/tone rules or any gate. When in
doubt, keep it.
```

**Checkable:** `grep -n "## Concise Chat Output" CLAUDE.md` returns a line.

---

## Verification Tasks

### V-1 — Mechanical: Skill file exists and frontmatter is well-formed (Gherkin #11)

```bash
# File exists
test -f skills/concise-output/SKILL.md && echo "EXISTS" || echo "MISSING"

# name field
grep -c 'name: concise-output' skills/concise-output/SKILL.md

# description length ≤ 250 chars
python3 -c "
import re, sys
txt = open('skills/concise-output/SKILL.md').read()
m = re.search(r'^description:\s*\"(.+?)\"', txt, re.MULTILINE)
desc = m.group(1) if m else ''
print(len(desc), '<= 250:', len(desc) <= 250)
"

# license field
grep -c 'license: Apache-2.0' skills/concise-output/SKILL.md

# metadata.version field
grep -c 'version: "1.0"' skills/concise-output/SKILL.md
```

Pass condition: file exists; all greps return 1; description length <= 250.

---

### V-2 — Mechanical: Passive activation frontmatter (Gherkin #9)

```bash
grep -c 'user-invocable: false' skills/concise-output/SKILL.md
grep -c 'disable-model-invocation: true' skills/concise-output/SKILL.md
```

Pass condition: both greps return 1.

---

### V-3 — Mechanical: Verbatim Spanish gate string present in skill (Gherkin #2, grep-verifiable)

```bash
grep -c '¿Quieres ajustar algo en esta fase antes de continuar?' skills/concise-output/SKILL.md
```

Pass condition: grep returns 1.

---

### V-4 — Mechanical: Persona pointer present in `CLAUDE.md` (Gherkin #12)

```bash
grep -n '## Concise Chat Output' CLAUDE.md
grep -n 'concise-output' CLAUDE.md
```

Pass condition: both greps return at least one match.

---

### V-5 — Mechanical: Persona pointer present in `templates.go` (Gherkin #13)

```bash
grep -n '## Concise Chat Output' internal/initcmd/templates.go
grep -n 'concise-output' internal/initcmd/templates.go
```

Pass condition: both greps return at least one match.

---

### V-6 — CLAUDE.md post-edit invariant: Rule 2 and Phase Models intact

After T-3, verify these two sections remain byte-identical to their pre-edit content:

```bash
# Rule 2 text
grep -c "invoking its \`archon-<phase>\` subagent" CLAUDE.md || \
grep -c "archon-<phase>" CLAUDE.md

# Phase Models "binding hard gate" text
grep -c "binding hard gate" CLAUDE.md
```

Pass condition: each grep returns >= 1. If either returns 0, T-3 overshot; revert and redo.

---

### V-7 — Existing unit tests still pass

```bash
go test ./skills/... ./internal/scaffold/...
```

Pass condition: exit 0. No test edits are needed; `embed.go` uses `//go:embed */SKILL.md` glob so the new `skills/concise-output/SKILL.md` is picked up automatically.

---

### V-8 — Behavior review checklist (Gherkin #1–#8, #10 — doc/behavior, not mechanical)

Review the completed `skills/concise-output/SKILL.md` against each scenario:

| # | Scenario | Check |
|---|----------|-------|
| 1 | Verbose narration trimmed | SKILL.md Hard Rules section covers lead-with-point, no preamble |
| 2 | Gate question preserved verbatim | Preserve-Verbatim list item 1 quotes exact Spanish string |
| 3 | Decision tables preserved | Preserve-Verbatim list item 2 covers tables |
| 4 | Risks/open-questions preserved | Preserve-Verbatim list item 3 covers risks |
| 5 | SDD artifact content not truncated | Preserve-Verbatim list item 4 covers artifact content |
| 6 | Spanish + tone honored | Must-Not-Weaken includes Leader Persona language rule |
| 7 | Gate not bypassed | Must-Not-Weaken includes Human Review Gate SHOW+ASK |
| 8 | Ambiguous content kept | Decision Gate section present |
| 10 | Scope excludes subagent prompts | When to Use section explicitly states NOT subagent handoff prompts |

Pass condition: reviewer confirms each row is covered by the skill body.

---

## Out of Scope (do NOT create tasks for)

- Template ↔ CLAUDE.md Rule 2 / Phase Models desync back-port.
- `.archon/config.yaml` `skill_count` 24 → 25 / `skill_inventory` bump (regenerates on next `archon update`).
- `skills/embed_test.go` `expectedSkills` list extension (the existing test checks a fixed allowlist of known skills and does not fail on new additions — a new skill directory is automatically embedded by the glob; `embed_test.go` will pick it up in `foundSkills` without any modification).

---

## Embed Mechanism Note (for Apply)

`skills/embed.go` uses `//go:embed */SKILL.md all:_shared`. A new directory `skills/concise-output/SKILL.md` is matched by the `*/SKILL.md` glob automatically — no changes to `embed.go` or any registry file are required. `embed_test.go` checks an `expectedSkills` allowlist of 10 known skills; `concise-output` is not in that list but the test does not fail when extra skills are present (it only fails if an *expected* skill is *missing*). The test is therefore green without modification. This is confirmed as out of scope per the design.

---

## Risks / Open Questions

| Risk | Status |
|------|--------|
| T-3 (CLAUDE.md surgical edit) overshoots the anchor and corrupts Rule 2 or Phase Models | Mitigated by V-6 post-edit invariant check. If V-6 fails, revert T-3 and narrow the `old_string`. |
| Skill body exceeds 450-token budget | Low. Design target is 180–450 tokens; enforced by review during V-8. |
| `go test ./skills/...` failure if embed_test.go is modified inadvertently | N/A — embed_test.go is not touched. V-7 catches regressions. |
| templates.go backtick-as-§ convention violated | Enforced by T-2 constraint; V-5 grep confirms the section is present. |

---

## Estimated Line Count vs. Budget

| File | Estimated lines |
|------|----------------|
| `skills/concise-output/SKILL.md` (new) | ~75 |
| `internal/initcmd/templates.go` (insert) | ~14 |
| `CLAUDE.md` (insert) | ~12 |
| **Total** | **~101 of 400-line budget** |
