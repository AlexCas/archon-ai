# Exploration: preflight-arrow-key-questions

## Project Type
**Web testing**: not-web

archon-ai is a Go CLI/TUI (Cobra commands + a Bubble Tea TUI under `internal/tui`).
There is no web framework, no `package.json`, no browser-facing routes, and no
`playwright`/`cypress`/`chromedp` dependency. Playwright group E must stay `disabled`,
which matches the preflight decided this session (E1 = No).

## Problem Statement

Today the ARCHON orchestrator asks the SDD session preflight as one large Spanish
text block (the "Antes de continuar con SDD…" block with groups A–E) and instructs
the user to reply with codes like `A1, B1, C1, D1` or `usar recomendado`. We want
the orchestrator to ask the preflight as **per-group arrow-key questions** — one
selectable question per group A–E — i.e. the AskUserQuestion-style interaction,
instead of a plain text block the user must decode and answer by typing codes.

The **choices and their semantics must not change**: same 4–5 groups, same options,
same recommended defaults. Only the *asking mechanism* changes.

## Current-State Findings (concrete references)

### 1. Where the A–E block lives and how it is produced/consumed

The block has a **single source of truth**: the Go template string in
`internal/initcmd/templates.go`.

- `internal/initcmd/templates.go:44-71` — the fenced `text` code block containing
  the exact Spanish A–E prompt ("Antes de continuar con SDD…", `A. Ritmo` … `E. Pruebas web`).
  Backticks are represented by the `§` placeholder (`backtickPlaceholder`, line 15) and
  substituted at render time (`renderTemplate`, `templates.go:198-213`).
- `internal/initcmd/templates.go:32-82` — the full `## SDD Session Preflight (HARD GATE)`
  section that wraps the block: the "Required choices" list (lines 36-40), the block
  itself (44-71), "Project type & web testing (group E)" (73-76), and "Hard gate rules"
  (78-82).
- The same string constant `orchestratorSections` (lines 17-144) feeds **both**
  `RenderClaudeMD` and `RenderAgentsMD` (`templates.go:178-180, 190-196`). `archon init`
  writes the rendered result to `CLAUDE.md` (claude agent) or `AGENTS.md` (other agents):
  path chosen at `internal/initcmd/init.go:175-180`, rendered at `init.go:268-272`,
  written at `init.go:277`.
- `/home/skollhowl/Projects/archon-ai/CLAUDE.md:27-71` is the **rendered artifact** of
  that template (the repo's own committed CLAUDE.md, backticks already substituted). It is
  not a separate source of truth — it is regenerated from `templates.go`. Keeping the two
  in sync (regenerating CLAUDE.md from the template) is part of the work.

So: the block is **embedded into a generated CLAUDE.md/AGENTS.md**. There is no separate
runtime prompt. The orchestrator LLM reads its instructions from CLAUDE.md/AGENTS.md at
session start and then follows them; the "prompt" the user sees is emitted by the LLM
because the instructions tell it to.

Other files that merely mention preflight but do **not** contain the choice block (out of scope):
- `README.md:215` — one-line description of the preflight feature.
- `skills/harness-workflow/SKILL.md:103` — refers to recording "preflight choices" in SESSION_STATUS.
- `openspec/changes/archive/**` — historical artifacts, not live sources.

### 2. How the four/five choices map to config

There is **no Go code that parses the user's A–E answers**. A search for the option
tokens (`interactive`, `auto`, `ask-always`, `single-pr-default`, `force-chained`,
`auto-forecast`, `execution_mode`, `artifact_store`, `review_budget`) across
`internal/` finds them **only** inside the template string and its test
(`templates.go:37,39`; `templates_test.go:217`) — never in runtime logic.

The mapping is entirely **LLM-orchestrator-interpreted**:
- Groups A–D are cached by the orchestrator "for this session" (per the Hard-gate rules,
  `templates.go:78-82`) and echoed into later phase prompts and into `SESSION_STATUS.md`.
  Nothing writes them into `.archon/config.yaml`.
- Group E is the one choice with a **config side effect**, but even that is not wired
  from the preflight answer to config in Go. The preflight instruction
  (`templates.go:73-76`) documents that group E *maps to* `playwright.enabled` in
  `.archon/config.yaml`, and that the same value is set by the `--playwright` init flag
  or the TUI Playwright tab. The actual writers are:
  - `internal/initcmd/init.go:85` → `buildConfig(..., opts.Playwright)` →
    `init.go:207,229-230` sets `config.Playwright{Enabled: playwright}`.
  - TUI Playwright tab: `internal/tui/model.go:96,318` (`newPlaywrightTabState(cfg.Playwright)`
    / `applyToConfig`).
  - Config type: `internal/config/config.go:31,50`.

  These are driven by the `--playwright` flag / TUI toggle, **not** by parsing a
  preflight "E2" answer. The preflight's group E is advisory text telling the LLM to
  set `playwright.enabled` consistently; the enforcement lives in init/TUI.

**Conclusion:** changing how the preflight is *asked* has zero coupling to any Go
config-writing code. The answers flow through the LLM and SESSION_STATUS, not through
a parser.

### 3. What "arrow-key questions" means here

There is no existing Go TUI prompt component for the preflight (the only match for
"interactive" in TUI code is a terminal check, `internal/tui/model.go:449`, unrelated).
The `archon tui` (`internal/tui`) is a **config editor** with tabs
(Agent, Models, Judge, Mutation Testing, Playwright) — it is not the preflight surface.

"Arrow-key questions" therefore refers to the **LLM's own AskUserQuestion tool** (the
harness UI that renders selectable options the user navigates with arrow keys). The
orchestrator already has this capability at runtime. The change is to the **generated
instructions** so the orchestrator asks the preflight by issuing per-group
AskUserQuestion prompts (one question per group A–E, each with its options and the
recommended default marked) instead of printing the text block and asking for codes.

**This is a docs/template change, not a Go feature.** Specifically: edit the template
string in `templates.go:32-82` and regenerate `CLAUDE.md`. The accompanying test
`internal/initcmd/templates_test.go` asserts on the block's literal content and will
need updating to match the new wording (see Risks).

### 4. Whether `archon tui` / `archon init` preflight UI must stay consistent

- `archon tui` has **no preflight UI**; it edits config tabs. Its Playwright tab remains
  the way to toggle `playwright.enabled`. No change needed, but the new preflight text
  should keep the sentence that says the TUI Playwright tab and `--playwright` flag set
  the same value (`templates.go:76`) so the three paths stay described consistently.
- `archon init` writes the CLAUDE.md/AGENTS.md that contains the instructions; because
  the block is generated, editing `templates.go` and regenerating is the correct and
  only wiring point. No init CLI flag corresponds to groups A–D.

## Approaches

1. **Docs/template-only (recommended)** — Rewrite the `## SDD Session Preflight`
   section in `templates.go` so it instructs the orchestrator to ask each group A–E as
   a separate AskUserQuestion (arrow-key selectable options, recommended option marked
   as default), instead of printing the code-based text block. Preserve every option and
   default verbatim. Regenerate `CLAUDE.md` (and conceptually AGENTS.md, though only
   CLAUDE.md is committed here) from the template. Update `templates_test.go` assertions.
   - Pros: smallest possible change; zero runtime/config coupling; matches how the
     feature actually works (LLM-interpreted instructions); trivially reviewable;
     no behavior change to init/TUI/config.
   - Cons: relies on the orchestrator honoring the instruction to use AskUserQuestion
     (no hard Go enforcement) — but that is already true of the entire preflight today.
   - Effort: Low.

2. **Add a Go preflight prompt component** — Build an interactive Bubble Tea / survey-style
   prompt in Go that renders the five groups and writes results somewhere.
   - Pros: hard-enforced UX; deterministic.
   - Cons: large; the orchestrator is the LLM harness, not the `archon` binary — the
     preflight happens inside the chat session, so a Go prompt in the CLI would not run
     at the moment the orchestrator needs it; introduces a second, divergent source of
     truth; blows the ~400-line budget; contradicts "keep it small". Not aligned with
     how preflight is consumed.
   - Effort: High.

3. **Both** — template change plus a Go component. Same downsides as (2) with no added
   benefit for this slice.
   - Effort: High.

## Recommendation

**Approach 1 (docs/template-only).** The preflight is an LLM-interpreted instruction
block generated into CLAUDE.md/AGENTS.md; the only lever that changes *how it is asked*
is the template wording. Rewrite the section to direct the orchestrator to use per-group
AskUserQuestion (arrow-key) prompts, keeping all options and recommended defaults
identical. Regenerate CLAUDE.md and update the template test. This fits the force-chained
/ ~400-line budget comfortably (expected well under 100 changed lines).

Concrete edit surface:
- `internal/initcmd/templates.go:32-82` — rewrite the preflight section.
- `/home/skollhowl/Projects/archon-ai/CLAUDE.md:27-82` — regenerate to match.
- `internal/initcmd/templates_test.go` — update assertions that pin the old block text
  (e.g. the `backtickChecks` around lines 216-224 reference `interactive`/`auto`/etc.,
  and any test asserting the literal Spanish block wording).

## Scope / Non-goals

In scope:
- Rewriting the preflight instruction so groups A–E are asked as per-group arrow-key
  (AskUserQuestion) selections with recommended defaults marked.
- Regenerating CLAUDE.md from the updated template.
- Updating `templates_test.go` to match the new wording.

Non-goals (MUST NOT change):
- The set of groups (A–E), their options, or the recommended defaults — semantics frozen.
- `.archon/config.yaml` schema or the Playwright wiring in init/TUI/config.
- Any Go runtime behavior of `archon init`, `archon tui`, or config parsing.
- Adding a Go interactive prompt component.
- README/harness-workflow prose beyond, at most, a one-line consistency touch-up (prefer
  none this slice).

## Open Questions (for Human Review Gate)

1. **Group D "Otro" (D3, custom line budget):** AskUserQuestion options are discrete.
   The current D3 = "Otro: preguntar el número después." Should the arrow-key question
   offer D1/D2/D3 as options and, when D3 is chosen, follow up with a free-text ask for
   the number? (Recommended: yes — keep D3 as an option that triggers a follow-up prompt.)
2. **Group E conditionality:** Today group E is only asked for NEW/blank/unknown projects
   (explore determines web vs not-web). Should the arrow-key flow skip group E entirely
   when explore already determined `not-web` (as here), or always ask all five? (Recommended:
   keep current conditional behavior — ask E only when project type is unknown.)
3. **Language:** Keep the questions in Spanish (matching today's block and the Leader
   Persona chat-language rule)? (Recommended: yes.)
4. **"usar recomendado" shortcut:** With per-group arrow-key questions, should we still
   support a single "usar recomendado" fast-path that accepts all defaults at once, or
   rely on each question defaulting to its recommended option? (Recommended: keep a note
   that selecting each default is equivalent; a global shortcut is optional.)
5. **AGENTS.md:** Only CLAUDE.md is committed in this repo, but the same template feeds
   AGENTS.md for non-claude agents. Confirm no separate AGENTS.md artifact needs
   regenerating here (it is produced at init time from the same string).

## Ready for Proposal
Yes. This is a low-effort, docs/template-only change with no runtime coupling. Recommend
the orchestrator present approach 1 and resolve open questions 1–2 (D3 follow-up and
group-E conditionality) with the user before the propose phase.
