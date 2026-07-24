# Exploration: Obsidian-Style Vault for SDD Specs

## Project Type

**Web testing**: not-web

This is a Go CLI + a set of Markdown skills (`skills/`) that orchestrate SDD. There is
no browser surface. Playwright is not applicable (matches session preflight group E = no).

## Current State

The SDD artifacts live under `openspec/` in a flat, folder-per-thing layout:

```
openspec/
├── config.yaml
├── specs/                         <- source of truth (main specs)
│   └── {capability}/
│       ├── spec.md
│       └── {capability}.feature   <- Gherkin (executable use cases)
└── changes/
    ├── archive/
    │   └── YYYY-MM-DD-{change}/    <- completed changes (audit trail, immutable)
    │       ├── proposal.md
    │       ├── exploration.md
    │       ├── design.md
    │       ├── tasks.md
    │       ├── verify-report.md
    │       ├── SESSION_STATUS.md
    │       ├── state.yaml
    │       └── specs/{cap}/{spec.md, {cap}.feature}
    └── {change}/                  <- active change (same shape, no date prefix)
```

Verified against the live repo: **23 capabilities** under `specs/`, **20 archived
changes**, plus active changes (`ai-orchestration-harness`, `issue-16-leader-personality`,
`opencode-phase-subagents`, this one).

### Which skills WRITE artifacts (and would need to emit/maintain links)

| Skill | Writes | Would need to emit/maintain links |
|-------|--------|-----------------------------------|
| `sdd-explore` | `changes/{change}/exploration.md` | link to `map.md`, sibling change artifacts |
| `sdd-propose` | `changes/{change}/proposal.md` | link to `map.md`, spec capabilities it touches (its Capabilities section already names them) |
| `sdd-spec` | `changes/{change}/specs/{cap}/spec.md` + `{cap}.feature` | link to proposal, main spec, and reference its own `.feature` |
| `sdd-design` | `changes/{change}/design.md` | link to proposal + specs |
| `sdd-tasks` | `changes/{change}/tasks.md` | link to design + specs |
| `sdd-verify` | `changes/{change}/verify-report.md` | link to specs/tasks |
| `sdd-archive` | merges deltas into `specs/{cap}/spec.md`; **MOVES** `changes/{change}/` → `changes/archive/YYYY-MM-DD-{change}/` | **rewrite every link that pointed into the moved folder**, register/rename capabilities in `map.md` |
| `sdd-init` | scaffolds `openspec/{specs,changes,changes/archive}` | would create the initial `map.md` |

The persistence contract for these skills lives in `skills/_shared/openspec-convention.md`
and `skills/_shared/sdd-phase-common.md` (Sections B/C).

### Confirmed facts (verified, not assumed)

- **No cross-links exist today.** `grep` for `[[` and for relative markdown links
  `](../` across `openspec/**/*.md` returns **0** matches. References between
  capabilities are bare prose (e.g. the word `harness-workflow`).
- **No `map.md` exists** anywhere in the repo.
- **Go code never parses the markdown body.** `internal/` + `cmd/` reference only
  *paths* (`state.yaml`, `openspec/config.yaml`, artifact template strings). No
  markdown/wikilink parser is imported (no goldmark/blackfriday). So the link syntax
  we pick cannot break the Go build or the harness. The real consumers are the
  `sdd-*` skills + `_shared/openspec-convention.md` — flexible prose readers.
- **`.feature` files stay put.** They already live beside `spec.md`; the plan is that
  `.md` files merely *reference* them. No change to `.feature` content or location.
- The **archive move is real and frequent**: 20 folders already carry the
  `YYYY-MM-DD-` prefix, and each move relocates 6–8 `.md` files one directory deeper
  and under a renamed parent.

## Link Mechanism Options

### A. Obsidian `[[wikilinks]]`

Resolve by note *name*, location-independent.

- Pros: A change's artifacts keep working after the archive move **without any
  rewrite** — `[[harness-workflow]]` resolves the same before and after the folder is
  relocated. This directly attacks the hardest problem (link integrity on archive).
  Human-navigable in Obsidian; agent can grep the bare token. Backlink-friendly.
- Cons: Not clickable on GitHub's web renderer (renders as literal `[[...]]`).
  Requires a naming discipline: note names must be unique vault-wide, but we have
  many files literally named `spec.md` / `proposal.md`, so a wikilink to `spec.md`
  is ambiguous. Needs either unique display names (`[[harness-workflow spec]]`) or
  path-qualified wikilinks (`[[specs/harness-workflow/spec|harness-workflow]]`),
  which reintroduces path fragility on the archive move.

### B. Standard relative links `[text](../specs/x/spec.md)`

- Pros: Clickable on GitHub and in every Markdown viewer; unambiguous (paths are
  unique); an agent can `grep`/resolve them deterministically and even validate that
  the target file exists.
- Cons: **Every link into a change folder breaks the moment `sdd-archive` moves it**
  (`changes/{c}/…` → `changes/archive/DATE-{c}/…`), and links from the moved files
  back out to `specs/` shift by one directory level. Requires a deterministic rewrite
  pass on archive.

### C. Hybrid — wikilinks for capability identity, relative links within a change

Use `[[capability]]` for the stable capability graph (change → the capabilities it
touches; `map.md` → capabilities), and relative links for intra-change navigation
(proposal ↔ spec ↔ design ↔ tasks) since those files always move together as a unit.

- Pros: The links that survive an archive move unchanged (capability identity, and
  intra-change relative links, which stay relatively positioned) need no rewrite;
  only the change→spec and map→change edges cross the move boundary.
- Cons: Two conventions to teach the skills; still needs one rewrite rule for the
  boundary-crossing edges.

### Backlinks without Obsidian running

Obsidian's backlink pane is an app feature; nothing here runs Obsidian. For an
agent, "backlinks" must be materialized as either (a) a generated section/file
(`map.md` lists, for each capability, the changes that reference it), or (b) a
`grep` over the vault for the capability token. Option (b) is only reliable if the
reference token is unambiguous — another argument for stable capability names.

## `map.md` (Map of Content / MOC)

The entry node at `openspec/map.md`. It should contain:

- **Capabilities index** — every `specs/{cap}/` with a link and one-line purpose.
- **Active changes** — every non-archived `changes/{c}/` with phase/status (readable
  from `state.yaml`) and links to its artifacts.
- **Archive index** — archived changes grouped by date, linking to their folders.
- **Backlink map** (optional but high-value for the agent): for each capability, the
  changes that touched it.

Generation options:
- **Deterministic harness step (recommended)** — a Go subcommand (e.g. `archon map`
  or a step folded into existing commands) walks `openspec/`, reads `state.yaml` for
  phase/status, and regenerates `map.md` between managed markers. Deterministic,
  cheap, no LLM drift, re-runnable after any change. This is the same class of work
  the Go code already does (path-based scaffolding), so it fits the codebase.
- **Agent-authored** — a skill writes/updates `map.md` each phase. Flexible but
  non-deterministic, easy to drift, and costs tokens on every transition.

## Link-Integrity Maintenance (the hardest part)

The **archive move is the top risk.** `sdd-archive` (1) merges deltas into
`specs/{cap}/spec.md`, and (2) moves `changes/{c}/` → `changes/archive/DATE-{c}/`.
Any link pointing *into* the moved folder, and any relative link *inside* the moved
files pointing *out* to `specs/`, changes meaning.

Two maintenance models:

- **Deterministic (recommended).** A Go/harness step owns link integrity:
  - On archive, rewrite the boundary-crossing edges and regenerate `map.md`.
  - On create/rename of a capability, update `map.md` and (if wikilinks) leave
    identity links untouched.
  - Provide a `--check` mode that fails if any relative link resolves to a missing
    file — a cheap CI-style guard.
  This removes the fragile step from the LLM's hands. Con: new Go code to write and
  test (previously Go never touched markdown bodies — this is the one real code
  expansion, and it must be surgical: rewrite only within managed markers / known
  link patterns, never free-form editing of artifact prose).

- **Agent-maintained.** Extend `sdd-archive` (and propose/spec) skills to fix links.
  Pro: no Go change. Con: non-deterministic on the highest-consequence, immutable
  audit-trail operation; a missed rewrite silently rots the graph and the archive is
  "never modified" after the fact.

Choosing **wikilinks for capability identity (Option A/C)** shrinks this problem: the
only edges that must be rewritten on archive are map→change and change→spec, not the
whole graph.

## Impact Surface

- **Skills that change**: `sdd-init` (create `map.md`), `sdd-propose`, `sdd-spec`,
  `sdd-design`, `sdd-tasks`, `sdd-verify` (emit links in their artifact templates),
  `sdd-archive` (link rewrite + map regen). Plus the shared conventions
  `skills/_shared/openspec-convention.md` and `skills/_shared/sdd-phase-common.md`
  (document the vault shape, `map.md`, and the link convention as the single source
  of truth all phase skills inherit).
- **Go code**: untouched **only if** `map.md`/link maintenance is agent-driven. If we
  take the recommended deterministic path, Go gains one new, self-contained module
  (map generation + link rewrite/check) plus a scaffold line in `initcmd` to seed
  `map.md`. No existing Go behavior changes; no markdown-body parsing of artifact
  prose beyond the managed regions.
- **`.feature` files**: no change. `.md` references them; content and location stay.
- **CLAUDE.md / templates**: the phase-order and convention prose may add a one-line
  pointer to `map.md` as the entry node. `internal/initcmd/templates.go` seeds the
  orchestrator template but does not need structural change.

## Approaches

1. **Deterministic vault + wikilink identity (Recommended)** — Adopt `map.md` as a
   harness-generated MOC; use `[[capability]]` wikilinks for the stable capability
   graph and relative links for intra-change nav; a Go step regenerates `map.md` and
   rewrites the few boundary-crossing edges on archive, with a `--check` guard.
   - Pros: Solves link integrity at the source (deterministic), minimizes rewrite
     surface, keeps the audit trail trustworthy, cheap to re-run, no LLM drift on the
     highest-consequence operation.
   - Cons: New (small, well-scoped) Go module + tests; two link conventions to teach.
   - Effort: Medium.

2. **Relative links only, agent-maintained** — All links are relative markdown;
   skills (esp. `sdd-archive`) rewrite them and regenerate `map.md` by hand.
   - Pros: No Go change; clickable on GitHub; unambiguous targets.
   - Cons: Highest link-rot risk exactly where it hurts most (immutable archive);
     token cost every phase; drift-prone.
   - Effort: Medium (skill work) but fragile.

3. **Wikilinks only, no harness, `map.md` agent-authored** — Pure Obsidian vault;
   agent maintains everything.
   - Pros: Lowest code footprint; nice in Obsidian.
   - Cons: Ambiguous `[[spec.md]]` names; not clickable on GitHub; no determinism;
     backlinks depend on unambiguous tokens.
   - Effort: Low but least reliable for the stated "fast and reliable for an AI agent"
     goal.

## Recommendation

**Approach 1.** It directly serves the user's stated priority — navigation that is
*fast and reliable for an AI agent* — by making the graph deterministic where it
matters (the MOC and the archive move) instead of trusting an LLM to keep an
immutable audit trail's links fresh. Wikilink identity for capabilities plus relative
intra-change links minimizes the rewrite surface; the Go step is small and isolated
and never touches artifact prose outside managed regions.

## Review-Budget / Slicing Flag

This will almost certainly exceed the **400-line** review budget if attempted as one
PR: it touches 6–7 skill files, 2 shared convention modules, adds a Go module with
tests, and seeds `map.md`. **Flag for chained-PR / multi-slice.** Natural slices:

1. Convention + `map.md` shape + `sdd-init` seeding (docs/skills only).
2. Deterministic `map.md` generation + `--check` guard (Go, isolated).
3. Link emission in phase skills (propose/spec/design/tasks/verify).
4. Archive link-rewrite (Go) — the riskiest; land last, with tests over a fixture move.

## Open Questions

1. **Link convention**: wikilinks-for-identity + relative-intra-change (recommended),
   or relative-only for GitHub clickability? This drives everything downstream.
2. **Maintenance owner**: deterministic Go step (recommended) or agent-maintained
   skills? Are we comfortable adding a small markdown-writing Go module, given Go has
   never touched artifact bodies before?
3. **`map.md` scope**: capabilities + active changes + archive index only, or also a
   materialized backlink map (capability → referencing changes)?
4. **Archived-artifact links**: should links inside already-archived changes be
   rewritten too (backfill the 20 existing folders), or is the vault convention only
   forward-looking for new changes? Backfilling mutates the audit trail.
5. **Where does map regen fire**: a standalone `archon map` command, or auto-run on
   every phase transition / archive?
6. **Managed-markers policy**: is it acceptable for a Go step to edit `.md` files only
   within explicit `<!-- MAP:START/END -->`-style markers, leaving all authored prose
   untouched?

## Ready for Proposal

**Yes, with a decision gate.** The exploration is complete, but the proposal needs the
user's answers to Open Questions 1 and 2 (link convention + maintenance owner) before
scoping — they determine whether this is a skills-only change or a skills + Go change,
and how it slices into chained PRs under the 400-line budget.
