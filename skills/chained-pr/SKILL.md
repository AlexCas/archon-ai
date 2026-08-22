---
name: chained-pr
description: "Trigger: PRs over 400 lines, stacked PRs, review slices. Split oversized changes into chained PRs that protect review focus."
license: Apache-2.0
metadata:
  
  version: "1.0"
---

## Activation Contract

Load this skill when a planned PR may exceed **400 changed lines**, SDD forecasts `400-line budget risk: High` or `Chained PRs recommended: Yes`, or the user asks for chained/stacked PRs, review slices, or reviewer-load control.

## Hard Rules

- Split PRs over **400 changed lines** unless a maintainer explicitly accepts `size:exception`.
- Keep each PR reviewable in about **≤60 minutes**.
- Use one deliverable work unit per PR; keep tests/docs with the unit they verify.
- State start, end, prior dependencies, follow-up work, and out-of-scope items in every chained PR.
- Every child PR must include a dependency diagram marking the current PR with `📍`.
- In Feature Branch Chain, create a draft/no-merge tracker PR; child PR #1 targets the tracker branch, later children target the immediate parent branch. The tracker PR is also the archive commit's home (see Execution Steps 7–8).
- Treat polluted diffs as base bugs: retarget or rebase until only the current work unit appears.
- Do not mix chain strategies after the user chooses one.
- **Feature Branch Chain archive ownership:** the archive commit (spec merge,
  folder move, `archon map`, `SESSION_STATUS.md` move) is staged on the
  **tracker branch**, AFTER the integrated judge passes on the tracker and
  BEFORE the tracker PR merges to `main`. It travels inside the tracker PR — never
  a child PR diff, never a separate archive-only PR. Archive-internal order and
  one-commit staging are unchanged from the single-PR flow (`[[harness-workflow]]`
  "Terminal Phase Ordering (Feature Branch Chain)").
- **Stacked-to-Main + archive-before-PR converges to Feature Branch Chain.** When
  archive-before-PR is in effect (artifact store `openspec`/`hybrid`), pure
  Stacked-to-Main is unsupported: it ships slices independently to `main`, leaving
  no un-merged ref to own the archive commit. At `sdd-tasks` strategy selection the
  orchestrator MUST select **Feature Branch Chain** instead and notify the user.
  After convergence the change is FBC and the Feature Branch Chain archive rule
  above governs — zero new archive mechanics. A late Stacked→FBC conversion (after
  slices already merged to `main`) strands those slices and is NOT sanctioned; there
  is no recovery procedure. When archive-before-PR is NOT in effect (`engram`),
  Stacked-to-Main is unaffected and does not converge. Full rule:
  `[[harness-workflow]]` "Stacked-to-Main Archive Convergence".

## Decision Gates

| Condition | Action |
|---|---|
| PR ≤400 changed lines and focused | Keep single PR. |
| PR >400, each slice can land independently, archive-before-PR NOT in effect (`engram`) | Use Stacked PRs to main. |
| PR >400, each slice independent, BUT archive-before-PR in effect (`openspec`/`hybrid`) | Converge to Feature Branch Chain (Stacked-to-Main unsupported here). |
| PR >400, feature must integrate before main | Use Feature Branch Chain with tracker. |
| Generated/vendor/migration diff cannot split cleanly | Ask maintainer for `size:exception`. |
| SDD provides `delivery_strategy` | Follow it before apply/PR creation. |

## Execution Steps

1. Estimate changed lines and identify independent work units. When
   `graphify.enabled` and Leiden community data is present in `graph.json`,
   it MAY inform work-unit identification — advisory only, never required.
2. Ask for a chain strategy when none is cached and the budget is exceeded.
3. Create branches/PRs using the chosen strategy only.
4. Add Chain Context to each PR without replacing the repo PR template.
5. Verify each PR independently: CI/tests/docs/manual checks, rollback scope, and clean diff.
6. Keep the tracker PR draft/no-merge until all child PRs are reviewed and merged
   into the tracker branch.
7. After the LAST child PR merges into the tracker branch, run the **integrated
   judge on the tracker branch** (evaluate all change files as a unified whole via
   `[[harness-judge]]`). Detect the "last child merged" state explicitly, not by
   waiting on a GitHub event: in interactive mode confirm it with the user; in auto
   mode verify on the tracker branch that every child PR is merged (e.g. `gh pr view`
   per child, or that the tracker contains each child's commits) before triggering
   the integrated judge. The integrated judge is the gate for BOTH the archive step
   and the tracker merge. If it fails, re-apply on the tracker and re-judge before
   proceeding (bounded by the same max-3-retries cap as Rule 8, applied to the
   integrated judge on the tracker) — do NOT archive and do NOT merge the tracker.
8. Once the integrated judge passes, stage the archive commit on the tracker branch
   (spec merge → folder move → `archon map --backfill`/`--check` STOP gate →
   `SESSION_STATUS.md` move, one commit), THEN merge the tracker PR to `main`. See
   `[[sdd-archive]]` for the archive mechanics.

## Output Contract

Return the chosen strategy, PR order, current PR boundary, dependency diagram, review budget (`additions + deletions`), verification plan, and any `size:exception` rationale.

## References

- [references/chaining-details.md](references/chaining-details.md) — strategy diagrams, PR body section, branch commands, and reviewer guidance.
