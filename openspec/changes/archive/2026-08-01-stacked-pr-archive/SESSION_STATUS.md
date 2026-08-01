# Session Status

- **Active change**: `stacked-pr-archive`
- **Source**: slice 2b of #93 — Stacked-to-Main archive ownership (deferred from slice 2 / #97)
- **Current phase**: archive — PENDING (2026-08-01); re-judge APPROVED on retry 1; nit-1 folded (469297c); delivered single-PR. Archive MUST reconcile partial-tracked change folder (mv + git add -A)
- **Branch**: feat/stacked-pr-archive (from origin/master); single-PR flow
- **Judge retry**: 1 of 3. C6 (untracked planning artifacts) deferred to archive (git add whole folder)
- **Branch**: feat/stacked-pr-archive (from origin/master); single-PR flow
- **Note**: apply committed tasks.md+state.yaml inside the change folder (d300404); proposal/design/specs still untracked — archive folder-move will reconcile

## Preflight decisions (reused from this session)
- A. Ritmo: interactive
- B. Artefactos: openspec
- C. PRs: ask-always
- D. Review budget: 400 lines
- E. Playwright: disabled
- F. Impeccable: disabled

## Completed phases
- explore — 2026-07-31 (VERDICT: no clean home for Stacked-to-Main archive. The strategy's "each slice ships independently to main" property is mutually exclusive with "one un-merged owning ref holds the whole change at judge time". Design space collapses to: pollute (A/C), add-a-PR (B), break branch-PR policy (E), redefine-as-FBC (D), or document-limitation (F). 3 framed options for user decision. Docs-only, no Go.)

- propose — 2026-07-31 (proposal.md: Option 1 converge-to-FBC rule; replaces deferral in 4 locations; harness-workflow MODIFIED; docs-only)
- spec — 2026-07-31 (delta spec + .feature; NEW requirement "Stacked-to-Main Archive Convergence" (5 scenarios) + MODIFIED FBC requirement with deferral scenario removed; single-PR + FBC bodies verbatim; converge decided up-front at sdd-tasks; state.yaml at spec/completed)
- design — 2026-08-01 (design.md; narrow convergence layer, zero new archive mechanics; crux = sdd-tasks step 4a before caching strategy + sdd-apply blocked backstop; deferral replaced in 5 locations 1:1; pointers last; V1-V7 checks; decisions: keep backstop O, include N, skip D/J)
- tasks — 2026-08-01 (tasks.md; 6 work units for single PR; ~205 line est. within 400 budget; skip D/J include N; state.yaml at tasks/completed)
- apply — 2026-08-01 (6 commits on feat/stacked-pr-archive: 1d76b0c spec, ff19c4a sdd-tasks step 4a, 21e0ffc chained-pr, d6b7449 sdd-apply+branch-pr, bfcf203 pointers, d300404 bookkeeping; 8 doc files ~160 lines + tasks.md/state.yaml committed; no Go)
- verify — 2026-08-01 (PASS 9/9: no Go, single-PR+FBC byte-unchanged except removed deferral scenario, 5 scenarios mapped, crux step 4a precedes cache, deferral zero remaining, exact pointer header, go build OK. Note: real diff +613/-16 incl tasks.md; archive must git add whole folder to reconcile partial-tracking)
- judge — 2026-08-01 (FAIL, retry 1: C1 CRITICAL step 4a nested in High-budget conditional (gate skipped for low/medium-budget chain selection); C2 sdd-apply backstop needs self-detect fallback; C3/C4 spec scenario actor wrong (harness-workflow→orchestrator/sdd-tasks/sdd-apply); C5 add 'none' to safe-out; C6 stage planning artifacts (defer to archive))
- apply (judge-fix retry 1) — 2026-08-01 (commit a860fa3: C1 gate promoted to standalone unconditional '### Archive-before-PR Convergence Gate (MANDATORY)'; C2 self-detect fallback; C3/C4 actor fixes in both spec files; C5 'none' added; 3 files +42/-22; delta spec edited in-tree (untracked, reconciled at archive); no Go; single-PR+FBC bytes intact)
- re-judge — 2026-08-01 (APPROVED: C1-C5 resolved with file:line evidence, 0 blockers, no regressions, go build clean, spec files consistent; 2 INFO nits — nit-1 folded in 469297c)

## Options framed for the propose gate (USER decides)
- Option 1 — Converge to FBC before archive (redirect): reuse shipped FBC tracker rule, zero new mechanics, judge-legal, no extra PR, no pollution; COST = lose independent-shipping when archive is in play.
- Option 2 — Separate post-integration archive commit/PR: preserves independent shipping; COST = reintroduces the #93 extra-archive-only-PR pain for this one flow (documented, not silent).
- Option 3 — Discourage/not-supported: steer users to single-PR or FBC when archive is enabled; most honest, smallest surface; COST = no path for stacks+archive.

## Key artifacts / paths
- Predecessors (archived): openspec/changes/archive/2026-07-30-archive-before-pr/ (slice 1, #96),
  openspec/changes/archive/2026-07-31-chained-pr-archive/ (slice 2 Feature Branch Chain, #97)
- Shipped rules: openspec/specs/harness-workflow/spec.md — "Terminal Phase Ordering (Single-PR Flow)"
  and "Terminal Phase Ordering (Feature Branch Chain)"
- Change dir (planned): openspec/changes/stacked-pr-archive/

## Open questions (the hard ones for this slice)
- Stacked-to-Main has NO tracker branch and earlier stacked PRs can merge to `main` BEFORE the
  integrated judge runs — so there may be no single pre-merge point to own the archive.
- Is there ANY judge-legal, non-polluting option? Or must the integrated-judge timing model change
  for stacks (e.g., judge before any stacked PR merges)?
- Least-bad candidates from slice-2 explore: (#3) archive on last slice pre-main-merge (fragile);
  (#4) separate archive PR (but that IS the #93 pain). Frame the real trade-off for the user.

## Next recommended step
- Run explore (deep, honest about the no-clean-home problem); then Human Review Gate before propose.
