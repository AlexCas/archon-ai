# Exploration: Graphify Slice B — graph diff in verify + edge evidence in judge

> Scope note: Slice B was deferred from the Slice A design
> (`openspec/changes/archive/2026-08-17-graphify-integration/`) with a single
> one-line description and **no further detail**. This exploration concretizes
> what it means against the real code. Pure research — nothing implemented.

## Project Type

**Web testing**: not-web (Go CLI/TUI orchestrator; no browser surface, no
`package.json` UI layer). No Playwright recommendation, no Impeccable
recommendation.

## What Slice B was deferred as

From the archived Slice A artifacts, verbatim:

- exploration.md L284–285: *"Slice B: structural graph diff in `sdd-verify`;
  edge (`EXTRACTED`/`INFERRED`) evidence in `harness-judge`."*
- proposal.md L33–34 (Out of Scope): *"Slice B (graph diff in verify, edge
  evidence in judge); Slice C (bridge `internal/mapgen` spec graph ↔ code graph
  in archive)."*

That is the entire specification inherited. Everything below is the concretization.

## Current State

### How Graphify is wired today (post Slice A / PR #103)

- **Opt-in, default-off, advisory / never-blocking.** `graphify.enabled` in
  `.archon/config.yaml` gates everything; when the block is absent (it is
  absent in the repo's own `.archon/config.yaml` today — `skill_count: 26`, no
  `graphify:` block) the feature is byte-identical-inert.
- **`sdd-explore` is the SOLE extraction site.** It shells `graphify extract`
  (fresh/absent/stale per `skills/graphify/SKILL.md` §3–§4), reads
  `graph.json` / `GRAPH_REPORT.md`, and writes a tracked ≤40-line excerpt to
  `openspec/changes/<change>/graph-report.excerpt.md`.
- **`sdd-tasks` is strictly file-read** (Leiden communities for slice
  boundaries) — MUST NOT shell any `graphify` command, even if the binary is
  present and `graph.json` is missing (SKILL.md §5).
- **`sdd-verify` and `harness-judge` have ZERO Graphify awareness today.**
  Grep-confirmed: neither skill references graphify, graph.json, EXTRACTED, or
  INFERRED. Slice B is the *first* time these two phases would touch Graphify.

### Data surface Graphify exposes (from the skill/design contract)

Empirical caveat: the `graphify` binary is **not installed** on this machine and
no `graph.json` exists in the repo, so the shape below is drawn from
`skills/graphify/SKILL.md`, the archived design.md, and the archived
exploration.md — not from a live artifact. Any Slice B design MUST validate the
real schema before committing to a diff format.

- **`graph.json`** — machine-readable code graph: nodes (symbols/files) + edges.
  Edges carry a **provenance tag: `EXTRACTED`** (deterministic tree-sitter AST —
  call/import/def edges) vs **`INFERRED`** (LLM-assisted, present ONLY when
  `semantic: true`). Default `semantic: false` → EXTRACTED edges only, zero LLM
  calls (SKILL.md §8).
- **`GRAPH_REPORT.md`** — human-readable: node/edge counts, EXTRACTED/INFERRED
  tallies, top Leiden communities.
- **`graph.html`** — visualization (irrelevant to Slice B).
- **CLI verbs: `graphify extract | query | path | explain`.** There is **no
  native `diff` verb** documented anywhere. A "graph diff" therefore does not
  exist in the tool — it must be computed harness/skill-side by comparing two
  `graph.json` snapshots (set difference of nodes and edges).
- Artifacts live in gitignored `.archon/graphify/` (`.gitignore:10` covers
  `.archon/`). Staleness = `mtime(graph.json) < HEAD committer time`.

### How sdd-verify works today (`skills/sdd-verify/SKILL.md`)

Reads all status `contextFiles`; maps each Gherkin scenario to a covering test;
requires real runtime test evidence (static analysis alone is never
verification); emits a compliance matrix and a verdict `PASS` /
`PASS WITH WARNINGS` / `FAIL`. It already carries **conditional, config-gated
checks** as precedent:
- **Impeccable presence check** — advisory NOTE only (never CRITICAL), skipped
  when `impeccable.enabled: false`.
- **Security coverage check** — CRITICAL when a `@security` scenario lacks a
  covering test, skipped when `security.enabled: false`.

The Impeccable presence check is the exact structural template for an advisory,
never-blocking Graphify diff note in verify.

### How judge works today (`skills/harness-judge/SKILL.md` + `judgment-day`)

`harness-judge` delegates dual blind adversarial review to `archon-judge`
(→ `judgment-day`), then runs optional gates (mutation / Playwright / Impeccable)
that CAN block and feed a re-apply loop (max 3 retries). Step 4's result table
is the pass/fail contract: judgment-day AND every *enabled* gate must pass.
Confirmed issues become a structured feedback block consumed by `sdd-apply`.

Crucial constraint for Slice B: **Impeccable/Playwright/mutation are blocking
gates with a verdict column. Graphify is advisory / never-blocking.** So edge
evidence must NOT become a fourth column in the Step 4 table and must NOT enter
the re-apply loop as a fail trigger. It can only *enrich the reasoning* of
findings the judges already reached by normal review.

## The two features, concretized

### 1. Graph diff in `sdd-verify`

Purpose: after apply, surface *unexpected structural changes* — e.g. a function
that lost all its callers, a new cross-package import edge, a deleted symbol
still referenced elsewhere — as advisory context beside the compliance matrix.

Mechanics required (none exist today):
- **Two snapshots.** A "before" graph (pre-apply baseline) and an "after" graph
  (post-apply). Graphify overwrites a single `output_dir/graph.json` on each
  `extract`, so the baseline must be **copied aside** (e.g.
  `.archon/graphify/graph.baseline.json`) before re-extraction, or the
  explore-time graph must be preserved as the baseline.
- **Who extracts the "after".** verify runs after apply. Either verify
  re-extracts (making verify a **second extraction site** — a relaxation of the
  Slice A "sole extraction site" invariant, SKILL.md §3), or a pre-apply hook
  snapshots the baseline and verify re-extracts the after. This is the central
  design decision.
- **The diff itself** = set difference over nodes and edges (added/removed/
  changed), computed skill-side or in a small Go helper, since the tool has no
  `diff` verb.
- **Output**: an advisory `### Code Graph Diff (advisory)` NOTE in the
  verification report — NEVER CRITICAL, NEVER changes the PASS/FAIL verdict on
  its own. Modeled exactly on the existing Impeccable presence NOTE.

### 2. Edge evidence in `harness-judge` / `judgment-day`

Purpose: let judges *cite specific edges* as structural backing for findings
they already made — e.g. "func `X` is dead: its only inbound call-edge `Y→X`
was removed (see graph)". This grounds review comments in deterministic AST
facts instead of prose assertion.

Mechanics:
- Judges query the post-apply graph via `graphify query` / `graphify explain`
  (real CLI verbs) or read `graph.json` to look up callers/callees of a symbol
  under review.
- Edge provenance matters: **EXTRACTED** edges are trustworthy deterministic
  facts; **INFERRED** edges (only present when `semantic: true`) are
  LLM-guessed and must be labeled as such if cited. In the default
  `semantic: false` mode, only EXTRACTED edges exist.
- **Strictly supporting, never a gate.** No new column in the Step 4 result
  table; no new `fail` trigger; edge evidence never enters the re-apply loop by
  itself. It only enriches the *description* of issues judges independently
  confirm.

## Affected Areas

- `skills/sdd-verify/SKILL.md` — new conditional, advisory "Code Graph Diff"
  step gated on `graphify.enabled`, modeled on the existing Impeccable presence
  NOTE. (~15–20 lines of prose.)
- `skills/harness-judge/SKILL.md` and/or `skills/judgment-day/SKILL.md` — a
  gated note that judges MAY cite EXTRACTED-edge evidence; explicit "never a
  gate, never in the result table, never a re-apply trigger" guardrail.
- `skills/graphify/SKILL.md` — extend the Per-Phase Invocation Map (§3) with
  `sdd-verify` (and possibly `harness-judge`) rows, plus baseline-snapshot rules
  and diff degradation modes in the §6 table. Today §3 has only explore + tasks
  rows and §5 asserts explore is the *sole* extraction site — Slice B must
  amend that invariant deliberately.
- Baseline snapshot storage under gitignored `.archon/graphify/` (no new Go if
  skill-only; a small Go helper only if a snapshot/copy or set-diff is done in
  Go).
- Possibly `internal/config/config.go` — ONLY if Slice B needs a new knob
  (e.g. `graphify.diff` on/off). Prefer NOT to add one (advisory features
  resist over-config, per the Slice A exploration Q2).

## Approaches

1. **Skill-only, both features, minimal (recommended).** All behavior in skill
   prose (mirrors Slice A, where 100% of runtime behavior is skill-driven and
   the Go code only carries config). verify writes an advisory diff NOTE from a
   baseline snapshot + re-extract; judge cites EXTRACTED edges to enrich
   existing findings. No new Go, no new config knob, no new gate.
   - Pros: matches the Slice A precedent exactly; zero blocking surface; trivial
     revert; fits one small PR; preserves never-blocking by construction.
   - Cons: diff correctness depends on node/edge-ID stability (unverified);
     "second extraction site" relaxation needs explicit sign-off.
   - Effort: Low–Medium.

2. **Verify-diff only; defer edge evidence.** Ship the advisory diff NOTE now;
   push judge edge-evidence to a Slice B.2.
   - Pros: smallest first step; isolates the extraction-site + baseline decision.
   - Cons: leaves half of the deferred scope open; two PRs for one deferred line.
   - Effort: Low.

3. **Judge edge-evidence only; no diff.** No baseline snapshot needed — judges
   just query the single post-apply graph. Sidesteps the hardest problem
   (baseline retention + ID stability).
   - Pros: cheapest, no second extraction, no invariant change; pure enrichment.
   - Cons: drops the "graph diff in verify" half; less value.
   - Effort: Low.

4. **Go-plumbed diff with a `graphify.diff` config knob + Go set-diff helper.**
   - Pros: deterministic Go diff, unit-testable, config-toggleable.
   - Cons: over-config for an advisory feature (contradicts Slice A Q2); more
     surface; still can't block; heaviest revert. Not recommended.
   - Effort: Medium–High.

## Recommendation

**Approach 1 (skill-only, both features, minimal)** — but ship it only after
resolving two blockers as part of propose/design, not before:

1. **Node/edge-ID stability** must be empirically confirmed (install the pinned
   `graphifyy==0.9.45`, extract the same tree twice, diff). If IDs embed
   volatile line numbers or hashes, a naive set-diff is noisy and the diff
   feature is not viable as specified — fall back to Approach 3 (edge evidence
   only) until upstream offers stable IDs or a native `diff`.
2. **Extraction-site invariant.** Slice A hard-codes "sdd-explore is the sole
   extraction site." A verify-time re-extract deliberately relaxes that. Get
   explicit sign-off and amend `skills/graphify/SKILL.md` §3/§5 in the same
   change, or route the "after" extraction through explore-owned tooling.

Preserve the hard rule throughout: the diff is an advisory NOTE, edge evidence
is supporting citation only — **neither may block a phase, fail a gate, return
`blocked`, or enter the judge re-apply loop.** All new failure modes degrade to
one-line advisory notes + baseline behavior, exactly like SKILL.md §6.

## Risks

- **No native `diff` verb** — the harness must compute the diff over
  `graph.json`; fragile against young `0.9.x` schema drift (mitigated by the
  `v0.9.45` pin + advisory parse-fail fallback).
- **Node/edge-ID stability unknown** (binary not installed; could not verify).
  If IDs are unstable, diffs are noisy/useless — the top de-risking item.
- **Double extraction cost / performance.** A second tree-sitter pass at
  verify-time (plus each re-apply retry re-diffs) multiplies cost on large
  repos. Contained by `semantic: false` (pure local, no LLM) but still real
  wall-clock on big trees.
- **Baseline retention.** Single `output_dir/graph.json` is overwritten on
  re-extract; without a copy-aside snapshot there is no "before" to diff against.
- **Extraction-site invariant relaxation** — verify as a second extraction site
  contradicts Slice A SKILL.md §3/§5; must be amended deliberately, not silently.
- **INFERRED edges need `semantic: true`** (LLM). In the default mode edge
  evidence is EXTRACTED-only; the deferred description implies both tags but
  only one is available by default.
- **Advisory pressure on the judge.** Graph evidence must stay *supporting*;
  risk that judges over-weight it or that it silently creeps into the Step 4
  result table and becomes a de-facto blocking gate.
- **First Graphify coupling into verify/judge** — these phases are Graphify-free
  today; Slice B must keep them fully inert when `graphify.enabled: false`.

## Open Questions

- Are Graphify node/edge IDs stable across two extractions of the same tree?
  (Empirical — requires installing `graphifyy==0.9.45`; not done here.)
- Does verify re-extract the "after" graph (second extraction site), or reuse an
  explore-owned snapshot? Where is the baseline stored
  (`.archon/graphify/graph.baseline.json`?) and who writes it?
- Skill-only vs a small Go set-diff helper — does anything here warrant Go
  plumbing, or does it all live in skill prose like Slice A?
- Should edge evidence be limited to EXTRACTED edges, or also cite INFERRED
  edges (only when `semantic: true`) with an explicit "inferred" label?
- Re-apply loop interaction: re-diff on every retry (cost) or diff once at the
  first verify only?
- Any new config knob (`graphify.diff`) — or keep zero-config per the Slice A
  anti-over-config stance?

## Ready for Proposal

**Yes, with conditions.** The scope is now concrete (skill-only, advisory,
Approach 1) and the deferred one-liner is fully unpacked. But two blockers —
node/edge-ID stability and the extraction-site invariant relaxation — should be
resolved (or explicitly accepted with a fallback to Approach 3) during propose/
design before committing. Orchestrator should surface both to the user at the
Human Review Gate.
