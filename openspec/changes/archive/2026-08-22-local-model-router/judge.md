# Judge Record — local-model-router

- **Change**: local-model-router
- **Branch (integrated)**: `feat/lmr-slice-b` (linearly contains A1 + A2 + B)
- **Judge type**: Integrated judge, Feature Branch Chain
- **Verdict**: **APPROVED (pass-with-warnings)**
- **Iteration**: 1 of 3 (no re-apply required — zero CRITICAL confirmed)
- **Mutation testing**: skipped (disabled in session preflight)
- **Playwright**: skipped (disabled in session preflight)
- **Impeccable**: skipped (disabled in session preflight)

---

## Dual Adversarial Review

### Reviewer A — Skeptic Pass

Examined: `Resolve` precedence and normalization correctness; D3 single-source-of-truth
and drift prevention; discovery fallback robustness; read-only invariant; CLI exit-code
and stdout/stderr contract; SKILL.md classifier contract; templates wiring; spec/artifact
consistency.

**Methodology**: read all source files, ran targeted edge-case probes beyond the 18
fixtures, compared verb sets against spec D3, checked terminal-phase behavior, verified
keyword collision handling, probed `wordBoundaryMatch` on short conjunction `"e"`, and
confirmed `os.DirFS` path-traversal safety.

### Reviewer B — Independent Advocate Pass

Approached from the opposite angle: looked for what A might have over-penalized, checked
whether each WARNING had a legitimate deferral justification, and re-verified that no
finding classified below CRITICAL is actually load-bearing for correctness.

---

## Findings

### CRITICAL
None.

### WARNING

**W1 (carry-forward from verify — review budget)**
Slice A cumulative code diff is ~771 insertions. The resolver+test commit alone is ~488
lines, above the 400-line session budget. The task plan pre-forecast this, proposed a
test-split commit, and the user explicitly accepted Slice A1 at 469 lines as an
irreducible tested unit. The functional split into A1 (469: resolver+rules+fixture tests)
and A2 (303: discover+CLI+tests) is sound — A1 is cohesive and reviewable. Not
re-litigated as a blocker. Carry forward to PR-split / review decision.

**W2 (archive/completed + control word — terminal-state sentinel behavior)**
File: `/home/skollhowl/Projects/archon-ai/internal/route/resolve.go`, `resolveControl`
(~line 88).

When the current phase is `archive` and status is `completed`, a control word (e.g.
"Continuemos") produces `{"phase":"archive","rule":"next","path":"code"}` — the `next`
rule fires but the phase cannot actually advance (no successor). The leader would route to
`archon-archive` and harness-workflow would immediately block with "phase already
completed." The behavior is not incorrect per spec (harness-workflow is the legality gate,
not the router), but the `rule:"next"` label is semantically misleading for a terminal
state — a consumer reading the JSON would expect "next" to mean the phase advanced.
Acceptable as a WARNING: the gate catches it before any damage, no spec text mandates a
different output, and adding a special sentinel here would introduce a second phase-list
or hard-code knowledge of terminality that the spec expressly defers to harness-workflow.

**W3 (SKILL.md keyword table includes "tareas" — undocumented divergence from rules.go)**
File: `/home/skollhowl/Projects/archon-ai/skills/sdd-router/SKILL.md`, tasks row.

The model classifier's keyword table (SKILL.md line 36) includes `"tareas"` under tasks,
while `rules.go` intentionally excludes bare `"tareas"` from the code keyword table
(documented comment at line 37–39) to prevent `"Implementa las tareas"` from
multi-matching. The divergence is safe — the model classifier is only consulted when no
code rule fires, so the dangerous case (`"implementa"` always fires the code `apply`
keyword first) never reaches the model. However, the SKILL.md does not mention that its
keyword table intentionally differs from the code table in this one entry. A weak model
reading SKILL.md alongside the code might treat them as mirrors. Acceptable deferral: the
D3 single-source requirement (design A2) applies only to `judgeVerbs`/`verifyVerbs`; the
SKILL.md table is a separate model-guidance artifact. Add a note in a future SKILL.md
pass.

**W4 (carry-forward from verify — pre-existing root CLAUDE.md/AGENTS.md drift)**
Root `/home/skollhowl/Projects/archon-ai/CLAUDE.md` and
`/home/skollhowl/Projects/archon-ai/AGENTS.md` do not contain the `archon route` Rule 2
that the template now generates. The golden test (`TestOrchestratorRulesRoutingOrder`)
asserts the rule is present and correctly ordered in the rendered output — so new `archon
init` runs produce correct orchestrators. The root files were not regenerated as part of
this change. Deferring is acceptable: updating them requires `archon init --force` on
the live project (a deliberate action with user approval), and the change's value
(template correctness) is already delivered. Carry forward as a recommended follow-up.

### SUGGESTION

**S1 (carry-forward from verify — `_shared/SKILL.md` in skill_count)**
`skill_count=27` includes `_shared/SKILL.md`. Consistent across all three sources. Future
cleanup candidate; not in scope here.

**S2 (wordBoundaryMatch compiles a new regexp per word, per call)**
File: `/home/skollhowl/Projects/archon-ai/internal/route/rules.go`, `wordBoundaryMatch`
(line 97–100).

Approximately 74 regexp compilations per `Resolve` call. For a CLI tool called once per
user interaction, this is not a performance problem. Future optimization: pre-compile a
single alternation regexp per word list, or use a compiled map. Not actionable in this
change.

**S3 (SKILL.md output format for ASK slightly inconsistent with CLI stderr format)**
File: `/home/skollhowl/Projects/archon-ai/skills/sdd-router/SKILL.md`, line 67.

The model classifier's output contract says `→ Router: ASK` for unclassifiable messages.
The code pre-router's CLI stderr format for D3-ambiguous is `→ Router: archon-ASK  (rule:
ambiguous, active-change: ...)`. These are outputs from different paths (model vs code)
and do not interact. A weak model reading both in the same context might interpret them as
the same format. No functional risk; the leader parses JSON stdout (not stderr) for all
decisions. Note for a future SKILL.md clarity pass.

---

## Re-Judge Result

No CRITICAL findings were identified. No re-apply iteration was triggered. The implementation is correct on all contract dimensions:

- `Resolve` precedence (explicit-agent > control > implicit > D3 > keyword > CLASSIFY): correct and tested.
- D3 single-source-of-truth (`judgeVerbs`/`verifyVerbs` shared by D3 check and keyword table rows): enforced by variable sharing, cannot drift.
- Discovery fallback (flag > SESSION_STATUS.md > sole-folder > none): correct, tolerant of missing/corrupt files, never hard-fails.
- Read-only invariant: the router performs no writes on any code path; `os.DirFS` sandboxes reads; confirmed by SHA-256 check in verify.
- CLI exit-code contract: exit 0 for all resolved outputs including ASK and CLASSIFY; exit 1 only on internal/usage errors.
- stdout/stderr contract: JSON always on stdout, human echo always on stderr; `--json=false` is a documented no-op.
- SKILL.md classifier contract: invoked only on CLASSIFY signal; emits one line then stops; prohibits state arithmetic; provider-neutral.
- Templates wiring: `archon route` rule inserted as Rule 2 after the harness-workflow gate and before delegate, in both `orchestratorRulesClaude` and `orchestratorRulesOpencode`; asserted by `TestOrchestratorRulesRoutingOrder`.
- `config.PhaseOrder` is the sole phase-sequence source; no second list in `internal/route`.
- All gates (preflight, vague-request, human-review) remain untouched and in their original positions.
- `skill_count=27` consistent across config.yaml, embedded FS, and rendered CLAUDE.md.

---

## Gates Skipped

| Gate | Reason |
|------|--------|
| Mutation testing | Disabled in session preflight |
| Playwright | Disabled in session preflight |
| Impeccable | Disabled in session preflight |

---

## Verdict Summary

**APPROVED — pass-with-warnings**

Four warnings carried forward (W1–W4); none block merge. Zero CRITICAL. Zero re-apply
iterations. Full test suite green (13 packages, 0 failures). 18/18 fixtures conform.
13/13 spec requirements verified. Read-only, gate-integrity, and D3 single-source
invariants hold.

**Next recommended**: proceed to `sdd-archive` on the tracker branch. Stage the archive
commit (spec merge, folder move, `archon map`, `SESSION_STATUS.md` move), then open the
tracker PR to `main`. Carry W1 into the PR description for reviewers; address W4 (root
CLAUDE.md/AGENTS.md update) as a follow-up `archon init --force` after the tracker merges.
