## Exploration: Adapting ODD's proportionality philosophy into archon's phase-gate harness

### Project Type
**Web testing**: not-web (Go CLI / orchestration harness, no browser surface)

### Current State

archon's phase order is DATA, not hardcoded logic: `config.PhaseOrder` (the 9-phase
`full` track) and `workflow.BugfixTrack` (`internal/workflow/workflow.go:30`) are
both plain `[]string` slices. `CheckTransition` (`internal/workflow/workflow.go:142`)
is track-agnostic index arithmetic over whichever slice `trackOrder()` resolves —
same-phase/in_progress → resuming, same-phase/completed → blocked,
next/completed → allowed, >1 ahead → blocked (missing phases listed),
backward → blocked. This is Go-tested (`internal/workflow/workflow_test.go`) and
exposed via `archon workflow status|advance|complete|retry`; `harness-workflow/SKILL.md`
requires calling those commands rather than reimplementing the table.

`judge` already has a binary, per-project `judge.enabled` config toggle read by
`harness-judge/SKILL.md` Step 0 — the closest existing thing to a proportional
review switch, but static, not computed per change.
`openspec/specs/harness-workflow/spec.md` "Single Judge in the Full-Track Terminal
Ordering" locks in exactly one judge on the full track as a spec-level guarantee.

Preflight already has a narrow, named "deviation" escape hatch (prose-only, by
design): a trivially small change can get a single-PR suggestion instead of the
default Feature Branch Chain, surfaced as one scoped question
(`openspec/specs/harness-workflow/spec.md` "Preflight is a Standard Default
Expressed as Prose").

`sdd-explore` Step 3b (`skills/sdd-explore/SKILL.md`) is existing precedent for a
phase sub-step that classifies something (`web`/`not-web`/`unknown`) and feeds that
classification into a later gate decision (`playwright.enabled`) — structurally
identical to what an ODD-style size/risk classification step would need to do.
`sdd-router` is deliberately single-responsibility (phase-family only, never size).

CLAUDE.md Rule 3 (mandatory per-phase sub-agent delegation) and Rule 6 (judge
mandatory after every verify) are **prose-only**: the only code-enforced hooks in
this repo are the `state.yaml` guard and the commit-attribution hook
(`internal/initcmd/claude_hooks.go`). Nothing in code blocks inline phase work or
forces a judge invocation beyond the `judge.enabled` flag.

No existing "proportional"/"risk-tiered task size" logic was found anywhere in
`skills/` or `internal/` (the only "risk" concept in the repo is
`openspec/specs/propose-security-risk/spec.md`, a security-risk classification
during `propose`, unrelated to task-size/ceremony proportionality). The `bugfix`
track (5 phases, no judge) is the one existing precedent for "some changes need
less ceremony."

**Side finding (now fixed as a prerequisite, see below):** while opening this very
change, `archon workflow advance --to explore` failed with
`state.yaml not found` for a brand-new change — there was no bootstrap path for the
very first phase of a change that never had a `state.yaml`. Fixed in
`internal/workflow/workflow.go` (`CheckTransition` now treats `currentPhase == ""`
as "no state yet" and allows only the track's first phase) and
`cmd/archon/main.go` (`advance` bootstraps a fresh `State` on `ErrStateNotFound`
instead of erroring, with a new `--track` flag). Covered by
`TestCheckTransition_Bootstrap*` and `TestWorkflowAdvance_Bootstrap*`.

### Affected Areas
- `internal/workflow/workflow.go` — `trackOrder()` is where a third track value
  would be added; `BugfixTrack` is the template to copy.
- `internal/workflow/workflow_test.go` — would need a new track's transition-table
  tests mirroring the existing bugfix ones.
- `skills/harness-workflow/SKILL.md` — "Phase Sequence" section and the
  track-resolution rule document the track enum; a third track needs an entry here.
- `openspec/specs/harness-workflow/spec.md` — "Phase State Machine" / "Phase
  Skipping Prevention" requirements name exactly two tracks; a new track needs a
  spec delta, same shape as the existing `harness-bugfix-track` delta.
- `skills/sdd-explore/SKILL.md` Step 3b — structural precedent for where a
  size/risk classification step could be inserted (a Step 3c reporting
  `small`/`substantial`/`unknown`, the same shape as `web`/`not-web`/`unknown`).
- `skills/sdd-router/SKILL.md` — single-responsibility (phase-family only); would
  need either a sibling classifier or an argued exception to its own Hard Rules to
  also emit a size signal.
- `skills/harness-judge/SKILL.md` Step 0 — the existing `judge.enabled` gate is the
  natural seam for a risk-tiered skip; today it's static, not computed per-change.
- `CLAUDE.md` Rule 3, Rule 6, and the "SDD Session Preflight" deviation rule — the
  three human-facing contract points that would need editing; none is backed by a
  code hook today.
- `internal/initcmd/templates.go` — generates CLAUDE.md's Rule wording and the
  preflight deviation prose; needs corresponding edits if the rules change.

### Approaches

1. **Extend `track` with a third value (`fast`/`inline`), reusing the existing
   track mechanism**
   - Pros: minimal code change, fully reuses `CheckTransition`/`Advance`/state.yaml
     machinery and its Go test coverage; preserves Rule 3 (still phase-gated, just a
     shorter list); auditable (every fast-track run still has a `state.yaml` +
     history).
   - Cons: still phase-fixed, just a shorter fixed ladder — doesn't capture ODD's
     "zero planning artifact for small, understood work" spirit; doesn't touch the
     Rule 3/6 tension (judge-free only because judge isn't in the list, same
     shortcut `bugfix` already uses).
   - Effort: Low.

2. **Add a sizing/risk classification step upstream (sdd-explore Step 3b-style, or
   in `archon route`) that short-circuits track selection** — *recommended*
   - Insert a size-classification sub-step (`small`/`substantial`/`unknown`) into
     `sdd-explore`, mirroring the web-detection pattern, feeding the `fast`/`inline`
     track from Approach 1 on `small`, defaulting to the existing preflight
     deviation-rule pattern (one scoped question) rather than inventing new UX.
   - Pros: captures ODD's actual "size the work before committing to ceremony"
     step; reuses already-existing patterns (Step 3b classify-then-gate, the
     preflight deviation rule, track-as-data) instead of inventing new mechanism;
     keeps `sdd-router` single-responsibility by living in `sdd-explore`.
   - Cons: extends `sdd-explore`'s structured output contract (a new "Size" row);
     still needs Approach 1's track underneath; higher-stakes judgment call than
     web/not-web since it affects ceremony, not just a test-generation toggle.
   - Effort: Medium.

3. **Risk-tier the judge phase only: `judge.enabled` from boolean to a per-change
   computed tier**
   - Pros: smallest, most surgical change; only touches one phase; directly
     addresses Rule 6's friction without touching Rule 3.
   - Cons: directly collides with `openspec/specs/harness-workflow/spec.md`'s
     "Single Judge in the Full-Track Terminal Ordering" requirement, which
     currently guarantees judge always runs on the full track with no exception —
     this needs a MODIFIED spec delta, a bigger governance lift than it looks;
     doesn't address Rule 3 at all.
   - Effort: Low (code) / Medium (spec governance — contradicts an existing hard
     guarantee rather than extending an open one).

### Recommendation

Approach 2 on top of Approach 1: add the `fast`/`inline` track (Approach 1) as the
mechanism, and a sizing step in `sdd-explore` (Approach 2) as the judgment call that
selects it. This is the only combination that reproduces ODD's actual "measure
before you commit to ceremony," and it assembles patterns already proven in this
codebase rather than inventing new ones. The user has confirmed **both** this
sizing/fast-track work (Rule 3 side) **and** judge risk-tiering (Approach 3 / Rule 6
side) are in scope for this change — bundling both widens the spec-review surface
(two spec deltas: a new `harness-fast-track` sibling spec, and a MODIFIED delta to
`harness-workflow/spec.md`'s "Single Judge" requirement) but was an explicit,
informed choice.

### Risks
- Rule 3's "always delegate" conflict is prose/spec-only (no code hook) — low
  mechanical risk to relax, but also means today's guarantee was already only as
  strong as the orchestrating model's compliance.
- Rule 6 / "Single Judge in the Full-Track Terminal Ordering" is spec-governed, not
  just prose — needs an explicit MODIFIED requirement delta, scenario-backed, not
  just an ADDED one.
- A new track needs its own spec delta (mirroring `harness-bugfix-track`'s
  precedent) or spec/code drift results — exactly the anti-pattern
  `harness-workflow/spec.md`'s own preamble warns against.
- Sizing classification (`small`/`substantial`) is inherently fuzzier than
  web/not-web — a wrong call under-ceremonies a change that needed review. This is
  the core judgment-quality tradeoff ODD itself accepts, and archon would be
  importing it.
- `judgment-day` is a separate, already-removed-from-the-SDD-pipeline dual-review
  path; risk-tiering work must not reintroduce or conflate it with `harness-judge`.

### Ready for Proposal
Yes. Scope confirmed with the user: both the fast-track/sizing mechanism (Rule 3
side) and judge risk-tiering (Rule 6 side) are in scope for the `sdd-propose` phase.
The bootstrap gap found while opening this change (`archon workflow advance`
couldn't create a first `state.yaml`) has been fixed and tested as a prerequisite,
committed separately from this change's artifacts.
