# Exploration: concise-output-skill

## Project Type
**Web testing**: not-web

Go CLI/TUI (`archon` command, Bubbletea TUI under `internal/tui/`, cobra-style init
under `internal/initcmd/`). No web framework, no `package.json`, no browser surface,
no Playwright/Cypress. Playwright must stay disabled (`playwright.enabled: false` in
`.archon/config.yaml`). No group-E preflight needed.

## Problem Statement
The ARCHON orchestrator's **chat replies to the user** are too long and verbose. We
want a mechanism that makes those chat summaries shorter and more direct — but
**selectively**. It MUST preserve, uncut and complete:
- the Human Review Gate pause (`¿Quieres ajustar algo en esta fase antes de continuar?`),
- decision tables,
- risks / open-question lists,
- the substantive content of SDD artifacts shown to the user.

It should only trim narration and filler. Scope is the orchestrator's chat output to
the user — **not** the internal subagent handoff prompts, and it must not weaken any
gate, artifact completeness, or the Leader Persona's language/tone rules.

## Current State

### 1. How skills are defined and wired
- Skills live one-per-directory under `skills/<name>/SKILL.md`, embedded into the Go
  binary via `skills/embed.go:5` (`//go:embed */SKILL.md all:_shared`). There is **no
  hardcoded skill list** — `internal/scaffold/extract.go:10-56` walks every top-level
  dir that contains a `SKILL.md` and extracts it. So adding a new
  `skills/<name>/SKILL.md` is automatically picked up; the count is dynamic.
- On `archon init`/`update`, skills are extracted to the machine-global dir
  `~/.config/opencode/skills` and symlinked into the project
  (`internal/initcmd/refresh.go:30-57`). Inventory + version come from each skill's
  real `metadata.version` frontmatter, parsed by `internal/scaffold/version.go`.
- `.archon/config.yaml` carries `skill_count: 24` and a `skill_inventory:` list.
  **These are generated outputs**, not the source of truth — they are rebuilt from
  the embedded FS on init/update. A new skill bumps the count to 25 the next time
  init/update runs; the checked-in config values would then be stale until regen.
- Skill-authoring convention: `skills/skill-creator/SKILL.md`. Required frontmatter
  shape (`skill-creator/SKILL.md:53-64`): `name`, `description` (one quoted line,
  Trigger-first, <=250 chars), `license`, `metadata.version`. Body sections in order:
  Activation Contract, Hard Rules, Decision Gates, Execution Steps, Output Contract,
  References. Target 180–450 tokens, hard max 1000.
- The style guide it cites as normative, `docs/skill-style-guide.md`, **does not
  exist** (there is no `docs/` directory). So the **inline fallback rules** in
  skill-creator apply. `skill-registry/SKILL.md:50` and `skill-improver` reference the
  same missing file — a pre-existing dangling reference, not something this change
  needs to fix.
- Behavior/persona skills already exist and are the closest precedent for a "shape
  your output" skill: `skills/comment-writer/SKILL.md`,
  `skills/cognitive-doc-design/SKILL.md`. Both are plain markdown, `license: Apache-2.0`,
  `metadata.version: "1.0"`, "When to Use" + decision-table body, `user-invocable`
  (no `disable-model-invocation`), triggered by description keywords.

### 2. How the orchestrator's behavior is configured
- The orchestrator persona/rules live in **`CLAUDE.md`** at the repo root, generated
  from `internal/initcmd/templates.go`. The template is
  `orchestratorSections + orchestratorTrailer` (`templates.go:17-175`); backticks are
  written as `§` and substituted at render (`templates.go:15`, `:195`).
- The persona sections that the concise skill must respect are in this template:
  Leader Persona (`templates.go:19-27`), Human Review Gate (`:104-120`), Vague Request
  Guard (`:79-102`), Commit Attribution (`:133-138`).
- **Template ↔ CLAUDE.md desync (confirmed, load-bearing):** the on-disk `CLAUDE.md`
  has hand-edits that were NOT back-ported to `templates.go`:
  - `CLAUDE.md:125` (Rule 2): *"You MUST delegate each phase by invoking its
    `archon-<phase>` subagent … the subagent's frontmatter model is the gate"* —
    but `templates.go:144` still says *"Delegate each phase to sdd-* sub-agent"*.
  - `CLAUDE.md:141-142` Phase Models: *"Each phase runs in its named `archon-<phase>`
    subagent. The `model` field … is the binding hard gate"* — but `templates.go:160`
    still says *"Advisory: … request the model below … This is a preference, not a
    hard gate"*.
  Consequence: **regenerating `CLAUDE.md` from `templates.go` today would REVERT these
  hand-edits.** Any wiring that touches CLAUDE.md must be a surgical edit, not a regen.

### 3. Existing verbosity / concise conventions
- `skills/_shared/persistence-contract.md:158` defines a `detail_level:
  concise | standard | deep` knob the orchestrator may pass to subagents. **Important
  distinction:** this controls a subagent's *returned artifact/analysis verbosity*, and
  it explicitly "does NOT affect what gets persisted." It is a **different axis** from
  this change (orchestrator → user *chat* verbosity). The new skill should not be
  confused with or fold into `detail_level`.
- Per-skill "keep it short" guidance already exists and is the tonal precedent to
  match: `comment-writer/SKILL.md:27` ("Prefer 1 to 3 short paragraphs or a tight
  bullet list"), `cognitive-doc-design/SKILL.md:27`, `skill-creator/SKILL.md:27`
  (token budgets). None of these govern orchestrator chat output.
- No existing rule tells the orchestrator to trim its own chat replies. This is a
  genuine gap, not a duplicate.
- **Conflict surfaces the skill must NOT weaken:** Leader Persona language rule (reply
  in the user's language, neutral Spanish — `CLAUDE.md:21-23`), warm/direct tone
  (`:25`), and the Human Review Gate SHOW+ASK contract (`:104-120`). "Concise" must
  never drop the Spanish gate question, the decision tables, or the risks list.

### 4. What "a skill" concretely is here + registration/tests
- A skill = a directory `skills/<name>/` containing `SKILL.md` (markdown + YAML
  frontmatter), optional `assets/` and `references/`. Discovery is automatic via the
  embed glob + `scaffold.Extract` dir walk — **no registry file edit required for the
  harness to load it**. `.atl/skill-registry.md` (from the `skill-registry` skill) is
  an *index for delegators*, regenerated on demand, not a gate.
- Registration touchpoints a new skill implies: none required for embedding; the
  `skill_count`/`skill_inventory` in `.archon/config.yaml` are regenerated outputs.
- Tests that validate skills: `skills/embed_test.go` (asserts the embedded FS is
  non-empty and lists expected skills — a new skill does not break it),
  `internal/scaffold/extract_test.go` and `version_test.go` (extraction + frontmatter
  version parsing). No test enforces the frontmatter *schema* of every skill, so a new
  well-formed skill needs no test change; adding the new name to `embed_test.go`'s
  expected list is optional and low-value.

## Affected Areas
- `skills/concise-output/SKILL.md` — NEW skill file (the primary deliverable of apply).
- `CLAUDE.md` — surgical edit to add a short "Concise Chat Output" persona rule that
  points at / summarizes the skill (regen is unsafe due to the desync above).
- `internal/initcmd/templates.go` — the persona source-of-truth; a matching section
  here keeps future regens correct. Editing it alone does NOT update the live CLAUDE.md.
- `.archon/config.yaml` (`skill_count`, `skill_inventory`) — will drift; refreshed by
  a later `archon update`, out of scope to hand-edit here.
- `skills/embed_test.go` — optional: extend expected-skills list.

## Approaches

1. **Standalone skill file only** (`skills/concise-output/SKILL.md`) — author a new
   behavior skill following the comment-writer/cognitive-doc-design pattern; rely on
   automatic embedding + description-trigger loading.
   - Pros: matches existing convention; smallest, most reviewable diff; no CLAUDE.md
     desync risk; automatically embedded and versioned.
   - Cons: the orchestrator persona (`CLAUDE.md`) does not *point* at it, so activation
     depends on trigger-matching alone — a persona-level behavior may not reliably fire
     on every chat reply without an explicit rule.
   - Effort: Low.

2. **Persona section in `templates.go` + surgical `CLAUDE.md` edit only** — add a
   "Concise Chat Output" rule directly into the orchestrator persona, no separate skill.
   - Pros: always in-context for the orchestrator; no discovery dependency.
   - Cons: bloats the always-loaded persona (the exact thing we're fighting); no
     reusable skill; still must handle the template↔CLAUDE.md desync carefully.
   - Effort: Low-Medium.

3. **Both: standalone skill file + a short persona pointer** (surgical `CLAUDE.md`
   edit AND matching `templates.go` edit) — the skill carries the full contract
   (what to trim, what to preserve verbatim); the persona adds a 2–4 line rule that
   names the skill and its hard preservation list, so it reliably activates on chat
   replies.
   - Pros: reliable activation + detailed contract lives in the right place (skill);
     keeps the persona pointer tiny; template edit keeps future regens honest.
   - Cons: touches three files; must carefully make CLAUDE.md a surgical edit (not
     regen) to avoid reverting the archon-<phase> hand-edits; slightly larger diff.
   - Effort: Medium (still well within the ~400-line budget).

## Recommendation
**Approach 3 (both), with a tightly-scoped skill and a minimal persona pointer.**

Rationale: the behavior we want is a *persona-level* reflex on every chat reply, so a
persona pointer is needed for reliable activation; but the detailed contract (the
trim/preserve rules, the verbatim-preservation list) belongs in a skill to keep the
always-loaded persona lean. Author `skills/concise-output/SKILL.md` in the
comment-writer style, with a **Hard Rules / preservation allow-list** that explicitly
protects: the Human Review Gate question (Spanish, verbatim), decision tables,
risks/open-question lists, SDD artifact content, and the Leader Persona
language/tone rules. Add the same section to `templates.go` (source of truth) and a
**surgical** edit to `CLAUDE.md` (never a regen, to preserve the archon-<phase>
hand-edits). Do not touch subagent handoff prompts or `detail_level`.

## Risks
- **Over-trimming gates:** the biggest risk is the skill weakening the Human Review
  Gate or dropping decision tables/risks. Mitigate with an explicit verbatim-preserve
  allow-list in Hard Rules and a Decision Gate that says "when in doubt, keep it."
- **CLAUDE.md regen reverts hand-edits:** must be a surgical edit. Flag for the
  proposal/design: whether to also back-port the archon-<phase> desync now is a
  separate concern — recommend NOT bundling it into this ~400-line PR-3.
- **Config drift:** `skill_count` (24→25) and `skill_inventory` in
  `.archon/config.yaml` become stale until an `archon update`; harmless (regenerated),
  but note it so reviewers don't expect a hand-edit.
- **Trigger reliability:** a skill alone may not fire on plain chat turns; the persona
  pointer mitigates this (reason for Approach 3 over 1).
- **Language interaction:** concise rules must not cause English leakage into Spanish
  chat replies; the skill must defer to the Leader Persona language rule.

## Open Questions (for Human Review Gate)
1. **Persona wiring:** approve Approach 3 (skill + `templates.go` + surgical
   `CLAUDE.md`), or prefer skill-only (Approach 1) to keep PR-3 minimal and rely on
   trigger activation?
2. **Desync back-port:** leave the `templates.go` "sdd-* sub-agent / Advisory model"
   desync alone in this PR (recommended), or fold a back-port in?
3. **Config values:** leave `skill_count`/`skill_inventory` to be regenerated by a
   later `archon update` (recommended), or hand-bump them in this PR for accuracy?
4. **Skill invocability:** should `concise-output` be `user-invocable` (like
   comment-writer) so the user can also invoke it directly, or purely a passive
   persona reflex?
5. **Language of the skill body:** English (repo artifact default), with the
   preserved Spanish gate string quoted verbatim — confirm.

## Ready for Proposal
Yes. The mechanism, the concrete files, the recommended approach, and the scope
boundaries are clear. The orchestrator should present the recommendation (Approach 3),
surface the five open questions above, and ask the Human Review Gate question before
moving to propose.
