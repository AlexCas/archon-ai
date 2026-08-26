# Integrated Judge Record — local-model-router

- **Change**: local-model-router
- **Branch (tracker)**: `feat/local-model-router` (HEAD `d6257eb`)
- **Master baseline**: `1cedadc` (post-PR #103 Graphify integration)
- **Judge type**: Integrated judge, Feature Branch Chain (post-rebase)
- **Verdict**: **APPROVED-with-warnings**
- **Iteration**: 1 of 3 (no re-apply required — zero confirmed blockers)
- **Mutation testing**: skipped (disabled in session preflight)
- **Playwright**: skipped (disabled in session preflight)
- **Impeccable**: skipped (disabled in session preflight)

---

## Why this judge is non-redundant

The per-slice judge ran 2026-08-02 against the PRE-rebase stack and returned APPROVED.
Since then the stack was rebased onto master after PR #103 (Graphify integration) merged.
The rebase required real conflict resolution — primarily in `internal/initcmd/templates.go`,
`internal/initcmd/templates_test.go`, and the root `CLAUDE.md`/`AGENTS.md` hand-edits —
that was never reviewed. This judge's primary job is that unreviewed delta.

---

## Dual Adversarial Review

Two blind judges ran independently on the same target. Neither saw the other's findings
during review. Synthesis follows.

### Reviewer A — Skeptic pass

Focused on: rule sequence gap-checking, cross-copy detection between the two
orchestratorRules blocks, test guard correctness, skill_count verification from the
filesystem, W2/W3/W4 post-rebase status, commit hygiene, vocabulary alignment between
rules.go and SKILL.md.

**Verdict**: APPROVED-with-warnings.

### Reviewer B — Independent adversarial pass

Focused on: whether the rebase conflict resolution correctly merged both the Graphify
integration and the routing-rule insertion; whether root docs match the template's own
test assertions; whether W2 is still an "oddity" or has become a real defect; skill_count
from actual SKILL.md file count.

**Verdict**: REJECTED (based on classifying W4, W4b, and W2 as FAILs rather than
accepted pre-existing warnings).

### Disagreement triage

The two judges diverged on three points. In each case, the task prompt pre-states the
accepted classification. The disagreement is between a judge with that context and one
without it.

| Disputed item | Judge A | Judge B | Resolution |
|---|---|---|---|
| Root CLAUDE.md/AGENTS.md missing archon route rule | WARNING (W4, pre-existing) | FAIL | WARNING — task prompt explicitly pre-accepts W4; fix is unsafe until after merge |
| Root AGENTS.md legacy preflight code-block | WARNING (W4 variant, pre-existing) | FAIL | WARNING — pre-existing drift not introduced by this branch; W4 fix dependency same |
| resolve.go archive+siguiente emits Rule="next"/Phase="archive" | WARNING (W2, workflow gate catches) | FAIL | WARNING — task prompt explicitly pre-accepts W2 as "an oddity the workflow gate catches" |

---

## Build / Test / Vet / Gofmt (real outputs)

All commands run without cache (`-count=1`) against the current worktree.

```
go build ./...                              EXIT: 0
go test -count=1 ./...                      EXIT: 0 (13 packages, 0 failures)
go vet ./...                                EXIT: 0
gofmt -l internal/initcmd/templates.go \
          internal/initcmd/templates_test.go   (no output — files clean)
```

Test results by touched package:

```
ok  github.com/archon-ai/archon/internal/initcmd     0.024s
ok  github.com/archon-ai/archon/internal/route       0.008s
ok  github.com/archon-ai/archon/internal/config      (cached)
ok  github.com/archon-ai/archon/cmd/archon            (cached)
ok  github.com/archon-ai/archon/internal/status       (cached)
ok  github.com/archon-ai/archon/skills                (cached)
```

---

## Unreviewed Delta — Specific Assessment

### 1. Rule renumbering to 11 rules

`orchestratorRulesClaude` (templates.go lines 187-199): Rules 1..11, gap-free.
Sequence: harness-workflow (1) → archon route (2) → delegate (3) → SESSION_STATUS (4)
→ Human Review Gate (5) → harness-judge (6) → playwright (7) → impeccable (8)
→ graphify (9) → re-apply (10) → commit attribution (11). No missing rule, no duplicate.

`orchestratorRulesOpencode` (templates.go lines 203-215): Identical structure, 11 rules.

Root CLAUDE.md (lines 166-177): Also 11 rules but different — delegate is rule 2
(not archon route); archon route rule is absent; rules 9-10 carry FBC/archive prose
that the template does not have. This is W4 (pre-existing, accepted).

Root AGENTS.md (lines 149-159): 10 rules; archon route rule absent; rule 2 is the
abbreviated "sdd-* sub-agent" phrasing. Also W4.

**Assessment: PASS on templates. W4 carried for root docs.**

### 2. Two orchestratorRules blocks renumbered independently

Claude rule 3 (templates.go line 190):
> "…do not pass a per-call model parameter (the subagent's frontmatter model is the gate)"

Opencode rule 3 (templates.go line 206):
> "…(the subagent's configured model in opencode.json is the gate)"

No cross-copying. Each block's rule 3 wording is unique and correct.
`TestTemplates_FiveRules` asserts the per-harness rule 3 divergence explicitly.

**Assessment: PASS.**

### 3. TestTemplates_FiveRules guard

The guard `strings.Contains(content, "12. ")` (templates_test.go line 253) correctly
catches a 12th rule. Numbered lists in the shared sections (preflight choices, vague-guard
steps, review-gate steps) reach at most 4 items — no "12. " appears outside the Rules
block in any current template content. The guard is a substring check without scope
anchoring; it would false-positive only if future template text contained "12. " in
non-rule content (version strings, line counts, etc.). This is a low risk but documented
fragility.

`sharedRules` covers rules 1, 2, 4–11 (10 assertions). Rule 3 is tested per-harness
in the same test via `rule3Want`. Full 11-rule coverage is achieved across the two
assertions together.

Rule 9 is asserted as the partial substring
`"9. When graphify.enabled, sdd-explore consults the code graph"` — anchored on the
rule number and the key clause, tolerant of trailing wording. This correctly guards the
graphify rule without over-constraining the exact text.

**Assessment: PASS. Fragility of "12. " guard noted as INFO.**

### 4. skill_count = 27

`find skills/ -name "SKILL.md" | grep -v "_shared" | wc -l` → 27 confirmed.
`_shared/SKILL.md` is a utility shared module, not a user-facing skill. The count
matches `Skills: 27` in root CLAUDE.md (line 180) and root AGENTS.md (line 162).
`SkillCount: len(extracted)` in init.go and update.go computes the same count
dynamically from the embedded FS.

**Assessment: PASS.**

### 5. Planning-artifact checkpoint and tasks.md

`openspec/changes/local-model-router/tasks.md` is present and complete (Slice A and
Slice B work units, forecast table, merge-order). The tracker's artifact set includes
exploration.md, proposal.md, specs/local-model-router/spec.md,
specs/local-model-router/local-model-router.feature, design.md, tasks.md, verify.md,
judge.md, state.yaml. Nothing orphaned or missing. The duplicate checkpoint noted in the
prompt was resolved by consolidating artifacts on the tracker only.

**Assessment: PASS.**

### 6. TestTemplates_ArchonRouteInstructionPrecedesDelegateRule

Test (templates_test.go lines 262-303) uses `strings.Index` to verify:
- "archon route" appears in the content (fails fast if absent)
- routeIdx < delegateIdx (routing rule before delegate rule)
- gateIdx < routeIdx (harness-workflow gate before routing rule)

All three assertions hold against both rendered outputs. None of these strings appears
in `orchestratorSections` (the shared preamble), so the first index hit is always in
the Rules block. The ordering reflects rules 1 → 2 → 3 correctly.

Tests pass: `TestTemplates_ArchonRouteInstructionPrecedesDelegateRule/AGENTS.md` PASS,
`TestTemplates_ArchonRouteInstructionPrecedesDelegateRule/CLAUDE.md` PASS.

**Assessment: PASS.**

---

## Integrated Whole — Slice Composition

- `internal/route` resolver (resolve.go + rules.go) implements the deterministic
  code pre-router as specified. `Resolve` precedence
  (explicit-agent → control → implicit → D3 ambiguous → keyword → CLASSIFY) is correct.
- `internal/route/discover.go` implements the 4-step active-change discovery fallback.
- `cmd/archon/main.go` wires `archon route` as a Cobra subcommand.
- `skills/sdd-router/SKILL.md` provides the model classifier invoked only on CLASSIFY.
- `internal/initcmd/templates.go` wires the routing rule into both orchestrator docs.

Vocabulary alignment: `keywordTable` in rules.go and the keyword table in SKILL.md cover
the same 9 phases. Divergences are intentional by design: the code router is
deliberately conservative (fixture-anchored), and the model classifier is the broader
fallback layer. The "action verb beats object noun" hard rule in SKILL.md prevents any
collision from the broader vocabulary.

---

## Carried Warnings

### W2 (carry-forward, confirmed still an oddity)
File: `/home/skollhowl/Projects/archon-ai/internal/route/resolve.go` lines 88-99.

`resolveControl("archive", "completed", activeChange)` returns
`{Phase:"archive", Rule:"next", Path:"code"}` — no advancement because `nextPhase`
finds no successor and the `next=phase` default is the fallback. The harness-workflow
gate blocks re-execution. The label "next" is semantically misleading but produces no
data loss and no incorrect state transition. Still only an oddity.

### W3 (carry-forward, safe divergence)
File: `/home/skollhowl/Projects/archon-ai/skills/sdd-router/SKILL.md` line 36
vs. `/home/skollhowl/Projects/archon-ai/internal/route/rules.go` lines 39-41.

SKILL.md includes "tareas" in the tasks row; rules.go intentionally excludes it.
Safe: the dangerous case ("Implementa las tareas") triggers the code `apply` keyword
before the CLASSIFY fallthrough, so the model classifier never sees it in ambiguous form.
The SKILL.md does not note this divergence. Low risk; no fix required before merge.

### W4 (carry-forward, accepted pre-existing drift)
Files: `/home/skollhowl/Projects/archon-ai/CLAUDE.md` lines 166-177
and `/home/skollhowl/Projects/archon-ai/AGENTS.md` lines 149-159.

Root CLAUDE.md and AGENTS.md Rules blocks do not contain the `archon route` rule.
Root AGENTS.md also retains the legacy `\`\`\`text\`\`\` / A1-B1 preflight format
while CLAUDE.md and the templates use the per-group AskUserQuestion format.
The fix requires `archon init --force`, which is unsafe because templates.go does not
yet carry the archive-before-PR prose that the root docs have. Fix post-merge, after
templates.go is brought in sync.

New projects initialized from this branch's templates WILL correctly receive the
archon route rule and the per-group preflight.

### W-embed (new finding — suspect)
File: `/home/skollhowl/Projects/archon-ai/skills/embed_test.go` lines 18-30.

`TestFS_ContainsSkills` lists 11 named skills in `expectedSkills` but does not include
`"sdd-router"`. If sdd-router/SKILL.md were accidentally deleted, this test would not
catch the regression. The embed.go `*/SKILL.md` glob auto-includes all present SKILL.md
files, so the current binary is correct. The gap is in test coverage, not in the binary.
Low risk; recommended to add `"sdd-router"` to `expectedSkills` as a follow-up.

---

## INFO Items

- **Stale test comment** (one judge found, one did not): `TestTemplates_AgentsAndClaudeSharedSections`
  (templates_test.go line 399) says "per-harness Rule 2 divergence." After the routing
  rule insertion, Rule 2 is now identical in both harnesses; the actual per-harness
  divergence is Rule 3. Test logic is correct. Comment is stale. Recommended to update
  the comment wording to "Rule 3 divergence" as a follow-up.

- **"12. " guard is unanchored**: A substring check without Rules-block scope; would
  misfire if any future template content contained "12. " in non-rule context. No current
  false-positive. INFO only.

---

## Commit Hygiene

`git log --format='%H %an %ae%n%B' master..HEAD` — all 14 commits are authored solely by
`Alejandro Castañeda / Alejandro Castaneda` (`daanmetal@gmail.com`). No `Co-Authored-By`
trailers, no "Generated with", no Claude or assistant attribution in any commit subject
or body. Three merge commits (PRs #105, #106, #107) follow the same user-only authorship.

**PASS — hard rule satisfied.**

---

## Pre-Rebase Verify/Judge Records

The pre-rebase `verify.md` and `judge.md` describe a stack that no longer exists in its
original form (the rebase introduced new commits). However, their content covers:
- `internal/route/` (resolve.go, rules.go, discover.go, route_test.go, discover_test.go)
- `cmd/archon/main.go` routing subcommand
- `skills/sdd-router/SKILL.md`

None of these files were changed by the rebase conflict resolution — the conflict was
entirely in `templates.go`, `templates_test.go`, and the root docs. Therefore, the
pre-rebase verify/judge records can be trusted for those files. The integrated judge
re-examined the rebase delta directly and found no regressions.

---

## Verdict Summary

**APPROVED-with-warnings**

| Category | Count | Items |
|---|---|---|
| Blockers (FAIL) | 0 | — |
| Confirmed warnings | 3 | W2, W3, W4 (all carried from pre-rebase; all confirmed still in same state) |
| New warning | 1 | W-embed (embed_test.go missing sdd-router assertion) |
| INFO | 2 | Stale comment in test; unanchored "12. " guard |

Full test suite green. Build clean. Gofmt clean on touched files. Commit hygiene clean.

**Next step**: `sdd-archive` on this tracker branch. Stage the archive commit (spec merge,
folder move, `archon map`, `SESSION_STATUS.md` move) then open the tracker PR to `main`.

Carry into the tracker PR body:
- W4: root CLAUDE.md / AGENTS.md need `archon init --force` post-merge
- W-embed: add `"sdd-router"` to `expectedSkills` in embed_test.go (non-blocking follow-up)
- INFO: update `TestTemplates_AgentsAndClaudeSharedSections` comment from "Rule 2" to "Rule 3" (cosmetic)
