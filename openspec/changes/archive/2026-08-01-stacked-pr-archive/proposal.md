# Proposal: Stacked-to-Main Archive Ownership (converge to Feature Branch Chain)

<!-- Link convention: [[capability]] wikilinks for capability identity, relative
     links for intra-change navigation. Rule: skills/_shared/spec-vault.md. -->

## Intent

Slice 2b of #93 closes the last archive-ownership gap. Single-PR and Feature Branch
Chain (FBC) both have a legitimate home for the archive commit because a **single
owning ref** (i) holds the whole change and (ii) is still un-merged when the
integrated judge passes. **Stacked-to-Main deletes that ref by design**: slices ship
to `main` independently, in order (`skills/chained-pr/references/chaining-details.md:44-57`,
`skills/sdd-apply/SKILL.md:97`). "Ship each slice independently" and "one un-merged
owning ref at judge time" are mutually exclusive — so there is no clean, judge-legal,
zero-extra-PR, zero-pollution home for the archive commit in a true Stacked-to-Main
flow. Explore ruled out every candidate: polluting the last/foundation slice (A/C),
a separate archive-only PR (B — the exact #93 pain), a direct-to-`main` commit
(E — violates the branch-PR contract, `skills/branch-pr/SKILL.md:21-24`), and a
"synthetic integration branch" (D — which simply **is** FBC). Today the spec, skills,
and `CLAUDE.md` carry a silent DEFERRAL (`spec.md:281-287`, `chained-pr/SKILL.md:31-33`,
`chaining-details.md:56-57`, `CLAUDE.md:35-36`). This proposal replaces that deferral
with the approved decision.

## Scope

### In Scope
- **Converge-to-FBC rule.** When SDD archive-before-PR applies (mode `openspec`/`hybrid`)
  AND the chosen chain strategy is Stacked-to-Main, the change MUST converge to a
  Feature Branch Chain — add a tracker/integration branch that accumulates all slices
  before `main` — BEFORE the archive step. The shipped "Terminal Phase Ordering
  (Feature Branch Chain)" rule then owns the archive commit (staged on the tracker,
  after the integrated judge passes, before the tracker merges to `main`).
- Replace the deferral passages in all four locations with this decision (no silent gap).
- Docs-and-skills only; explore confirmed **zero Go** enforces archive position.

### Out of Scope (Non-goals)
- No new archive mechanics — reuse the FBC requirement verbatim.
- No support for pure independent-shipping Stacked-to-Main + archive-before-PR: that
  combination is precisely the **unsupported** one; the orchestrator converts to FBC.
- No `archon archive` CLI. Single-PR and FBC behavior unchanged.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `harness-workflow`: MODIFIED — the "Terminal Phase Ordering (Feature Branch Chain)"
  requirement's Stacked-to-Main deferral scenario (`spec.md:281-287`) is replaced by a
  converge-to-FBC scenario; the FBC archive rule now governs Stacked-to-Main after
  convergence. Single-PR and FBC requirements + scenarios stay verbatim and additive.

## Approach

The rule is a **conversion, not a new mechanism**: Stacked-to-Main + archive-before-PR
→ add tracker branch → the change is now FBC → slice-2's tracker rule owns the archive.
Zero new archive mechanics. The unresolved open question is *when* the orchestrator
converges (see Risks / Open Questions) — proposed default: decide at chain-strategy
selection in `sdd-tasks`, so a user who wants archive-before-PR never lands in pure
Stacked-to-Main.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `openspec/specs/harness-workflow/spec.md` | Modified | Replace deferral scenario (281-287) with converge-to-FBC decision |
| `skills/chained-pr/SKILL.md` | Modified | Replace deferral (31-33) with converge rule |
| `skills/chained-pr/references/chaining-details.md` | Modified | Replace deferral note (56-57) |
| `skills/sdd-archive/SKILL.md` | Modified | Note convergence precondition for the archive step |
| `skills/branch-pr/SKILL.md` | Modified | Cross-ref: Stacked+archive converges to FBC before PR |
| `skills/sdd-apply/SKILL.md` | Modified | Chain-strategy handling reflects convergence |
| `skills/_shared/session-status-contract.md` | Modified | Confirm root-residency holds through convergence |
| `skills/harness-workflow/SKILL.md` | Modified | Replace deferral pointer (31-32) |
| `CLAUDE.md` | Modified | Replace deferral pointer (35-36) |

## Invariants to Preserve

- Single-PR requirement + scenarios **verbatim** (`spec.md:112-181`).
- FBC requirement + scenarios **verbatim and ADDITIVE** (`spec.md:183-279`).
- Judge-gating (archive never before judge passes).
- Archive-internal order + single commit + `archon map --check` STOP gate.
- SESSION_STATUS root-residency + `[[session-status-contract]]`.
- branch-PR contract: every change ships via a PR (`branch-pr/SKILL.md:21-24`).

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Convergence timing ambiguous (tasks vs. archive time) | Med | Resolve at gate; default: at `sdd-tasks` strategy selection |
| Late conversion strands already-merged slices on `main` | Med | Prefer up-front choice so archive users never pick pure-stacked |
| Three phase-order copies drift | Low | Mirror slice-2 pointer-line decision (full rule in spec + chained-pr; pointers elsewhere) |

## Rollback Plan

Docs/skills-only change with no Go. Revert the change branch (or the specific edits)
to restore the deferral passages; no runtime, data, or CLI state to unwind.

## Dependencies

- Slice 2 (#97) FBC "Terminal Phase Ordering" rule — MERGED; this slice reuses it.

## Success Criteria

- [ ] No silent deferral remains: all four locations state the converge-to-FBC decision.
- [ ] Stacked-to-Main + archive-before-PR is documented as converging to FBC before archive.
- [ ] Pure independent-shipping Stacked-to-Main + archive-before-PR is documented unsupported.
- [ ] Single-PR and FBC requirements + scenarios unchanged (verbatim/additive).
- [ ] Zero new archive mechanics introduced.

## Open Questions (Human Review Gate)

1. **Convergence timing** — at chain-strategy selection in `sdd-tasks`, or lazily at
   archive time? (Proposed default: `sdd-tasks`.)
2. **Up-front vs. late** — force the choice up-front so an archive-before-PR user never
   selects pure Stacked-to-Main, vs. allow a late Stacked→FBC conversion?
3. **Copy fan-out** — do all three phase-order copies (`spec`, `harness-workflow/SKILL`,
   `CLAUDE.md`) need the full converge narrative, or only `spec` + `chained-pr` carry it
   while the rest keep pointer lines (mirroring slice-2's decision)?
