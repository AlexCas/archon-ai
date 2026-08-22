# SESSION STATUS — local-model-router

- **Active change**: local-model-router
- **Branch**: feat/local-model-router (tracker — now contains A1 + A2 + B via merged PRs)
- **Current phase**: archive — in_progress
- **Last updated**: 2026-08-22T02:20:00Z

## Integrated Judge (Feature Branch Chain) — APPROVED-with-warnings

Report: `openspec/changes/local-model-router/judge-integrated.md` (the pre-rebase
`judge.md` was preserved, not overwritten). Zero confirmed blockers. Judge re-ran the
checks itself: `go build`, `go vet`, `go test -count=1 ./...` all exit 0 (13 packages).

All six unreviewed rebase-delta items PASS: rules render gap-free 1..11; the two
`orchestratorRules` blocks kept their own rule-3 wording with no cross-copy; the
`TestTemplates_FiveRules` guard genuinely fires on a 12th rule and `sharedRules` covers all
10 shared rules; `skill_count` 27 confirmed by counting SKILL.md files; the artifact set
survived the checkpoint drop intact; and the routing rule still precedes delegation in both
rendered docs.

The two judges split: A approved, B rejected by elevating W2, W4, and the AGENTS.md legacy
preflight format to FAIL. All three are pre-stated accepted limitations, so no blocker
survived — but B was right that they are real, and W4's fix is blocked on
[[templates-go-drift]].

Judge findings FIXED before archive in `b4b4489`:
- `skills/embed_test.go` `TestFS_ContainsSkills` listed `graphify` but not `sdd-router`, so
  the new skill's presence in the embedded FS went unasserted. A gap this branch introduced.
- `internal/initcmd/templates_test.go:401` comment still said "Rule 2 divergence" after
  routing shifted delegation to rule 3.

Pre-rebase verify/judge records remain trustworthy for `internal/route/*`,
`cmd/archon/main.go`, and `skills/sdd-router/SKILL.md` — the conflict never touched them.

## WAVE 1 — COMPLETE (2026-08-22)

Rebased onto `master` (`1cedadc`) after [[graphify-integration]] merged as #103, pushed all
four branches, opened issue **#104**, and merged three stacked PRs into this tracker:
**#107** B→a2, **#106** A2→a1, **#105** A1→tracker. Tracker HEAD `d6257eb`.
`go build ./...` and `go test ./...` pass on the tracker.

**Rebase conflict resolution — never judged, this is what the integrated judge must check:**
- Rule numbering is now **11 rules**. This change inserts `archon route` as rule 2 (pushing
  everything down one); Graphify's rule 8 became rule 9. The Claude and Opencode
  `orchestratorRules` blocks were renumbered **independently** — their rule-3 wording
  differs, so no text was copied between them.
- `TestTemplates_FiveRules` updated: `sharedRules` strings renumbered, and the trailing
  boundary guard moved to `"12. "` / "exactly 11 rules".
- `skill_count` resolved to **27** by counting actual skill directories (25 base + graphify
  + sdd-router), not by trusting either side of the conflict.
- The stack's **duplicate planning-artifact checkpoint was dropped** so the artifacts live
  only on this tracker. That is what made all three slice→tracker merges clean. Slice B is
  consequently 160 lines, not the 191 recorded below.

## Preflight Choices

- Execution mode: interactive
- Artifact store: openspec
- PR strategy: Feature Branch Chain (ask-always)
- Review budget: 400 lines
- Playwright: disabled
- Impeccable: disabled
- Mutation testing: disabled

## Phase History

| Phase | Status | Timestamp |
|-------|--------|-----------|
| explore | completed | 2026-08-01T18:41:00Z |
| propose | completed | 2026-08-01T18:45:00Z |
| spec | completed | 2026-08-01T22:15:00Z |
| design | completed | 2026-08-01T22:45:00Z |
| tasks | completed | 2026-08-01T23:30:00Z |
| apply | completed | 2026-08-02T00:10:00Z |
| verify | completed (PASS-with-warnings) | 2026-08-02T06:00:00Z |
| judge | completed (APPROVED-pass-with-warnings) | 2026-08-02T06:30:00Z |

## Key Artifacts

- Proposal: `openspec/changes/local-model-router/proposal.md`
- Spec: `openspec/changes/local-model-router/specs/local-model-router/spec.md`
- Feature file: `openspec/changes/local-model-router/specs/local-model-router/local-model-router.feature`
- Design: `openspec/changes/local-model-router/design.md`
- Tasks: `openspec/changes/local-model-router/tasks.md`
- Verify: `openspec/changes/local-model-router/verify.md`
- Judge: `openspec/changes/local-model-router/judge.md`
- State: `openspec/changes/local-model-router/state.yaml`

## Code Artifacts

- `internal/route/resolve.go` — pure Resolve(Input) Result; Normalize
- `internal/route/rules.go` — single-source verb/keyword data; D3 rule
- `internal/route/discover.go` — active-change discovery + read-only state read
- `internal/route/resolve_test.go` — 18 fixtures + keyword outline + normalize + implicit precedence
- `internal/route/discover_test.go` — discovery precedence + readState tolerant behavior
- `cmd/archon/main.go` — newRouteCmd wired into root
- `skills/sdd-router/SKILL.md` — model classifier contract
- `internal/initcmd/templates.go` — archon route rule 2 in both orchestratorRules blocks

## Warnings Carried Forward

- W1: Slice A code diff (~771 lines) exceeds 400-line budget; A1 (469) user-accepted as irreducible. Carry to PR.
- W2: archive/completed + control word → rule:"next" but phase stays archive (semantic oddity, gate catches it).
- W3: SKILL.md includes "tareas" in tasks keywords; rules.go intentionally excludes it. Undocumented safe divergence.
- W4: Root CLAUDE.md/AGENTS.md lack the archon route Rule 2 (pre-existing drift). Follow-up: archon init --force after tracker merges.

## Open Questions

None — all spec requirements verified, all tasks complete.

## Next Recommended Step — WAVE 2

1. **Integrated judge on this tracker** (in progress) — the per-slice judge ran on
   2026-08-02 against the PRE-rebase stack, so the conflict resolution above has never been
   judged. This gate covers it.
2. On pass: run `sdd-archive` on this tracker as ONE commit — spec merge into
   `openspec/specs/`, move the change folder to
   `openspec/changes/archive/2026-08-22-local-model-router/`, move this SESSION_STATUS.md
   into it, `archon map` regen — staged BEFORE the tracker PR merges to `master`.
3. Open the tracker PR to `master`, linking #104.

**WAVE 2 follow-up — the original plan is now UNSAFE.** W4 said to run
`archon init --force` after the tracker merges. Do NOT: `internal/initcmd/templates.go` is
behind the root docs on the archive-before-PR prose (verified 2026-08-17), so a full
regeneration reverts content merged in #96–#99. Bring `templates.go` current FIRST, then
regenerate. Tracked separately.
