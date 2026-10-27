---
name: harness-judge
description: "Trigger: judge phase, judgment gate, post-verify review. Orchestrate single focused review via archon-judge, mutation testing gate, and auto-re-apply loop."
license: MIT
metadata:
  
  version: "2.0"
  scope: orchestrator-gate
---

## Purpose

Orchestrate the judge phase: delegate a single focused review to `archon-judge`, optionally run mutation testing and Playwright E2E tests as quality gates, and automatically re-run `sdd-apply` with structured feedback on failure (up to 3 retries).

## Activation Contract

Load when the orchestrator reaches the `judge` phase after `sdd-verify` passes. This skill wraps existing skills — it does NOT reimplement review or apply logic.

The FIRST action on activation is Step 0: compute the full-track risk tier, then read the judge flag. If the judge phase is disabled, or the computed tier is skip-eligible and not lightened, this skill returns `skipped` immediately and the orchestrator advances to `archive`, then, after archive stages its commit on the branch, the PR is opened (single-PR flow).

## Hard Rules

- The judge phase is configurable. ALWAYS read `.archon/config.yaml` → `judge.enabled` as part of Step 0, before delegating any review. Default: `true` (run when the section is absent). When `judge.enabled: false`, SKIP the entire judge phase regardless of the computed risk tier — do NOT invoke `archon-judge`, mutation testing, or Playwright — and return `skipped` so the orchestrator advances from verify straight to archive.
- The judge phase only runs on the `full` track; a computed risk tier that is skip-eligible (full track, size small) MAY also skip or lighten it even when `judge.enabled` is not explicitly false. `bugfix` and `fast` never run judge regardless of tier.
- ALWAYS delegate a single focused review to the archon-judge subagent. Do NOT invoke judgment-day inline on the orchestrator's model. Do NOT run a dual or blind review.
- ALWAYS invoke `sdd-apply` for re-fixes. Do NOT apply fixes inline.
- ALWAYS invoke `sdd-verify` after each re-apply before re-judging.
- Mutation testing is OPT-IN. Read `.archon/config.yaml` → `mutation_testing.enabled`. Default: `false`. Skip entirely when disabled.
- Playwright E2E is OPT-IN and runs only for web projects. ALWAYS read `.archon/config.yaml` → `playwright.enabled` to decide whether to run it. Default: `false`. Skip entirely when disabled. These tests run AFTER verify and after the single judge review passes.
- Security gate is OPT-IN. Read `.archon/config.yaml` → `security.enabled`. Default: `false`. When `security.enabled` is true, treat any unresolved `@security` CRITICAL coverage gap (reported by `sdd-verify`) as a failing gate — do NOT advance to archive. When `security.enabled` is false, skip this check entirely and count it as `pass`.
- Maximum 3 retry cycles. The 4th failure returns `blocked` with `max_retries_exceeded: true`.
- NEVER skip the re-verify step between re-apply and re-judge.
- Accumulate all issues across retry cycles in the feedback block.

## Execution Steps

### Step 0: Judge Flag Gate (HARD GATE)

#### Step 0a: Compute the Risk Tier (full track only)

Before reading the `judge.enabled` flag, compute the per-change risk tier so the
flag check below has it available. This tiering applies to the `full` track only —
[[harness-fast-track]] and `[[harness-bugfix-track]]` never run judge regardless of
tier, which in practice means the orchestrator never invokes this skill for them at
all (neither track has a judge phase in its sequence — see [[harness-workflow]]).

1. Run `archon workflow status --change {change-name} --json` and read its `track`
   and `size` keys.
2. Compute the tier:
   - `track` != `full` → tiering does not apply; judge is not run for this track at
     all (governed by that track's own spec, not this gate).
   - `track: full` and `size` is `substantial` or `unknown` (including absent/empty,
     which reads as `unknown`) → tier **requires judge**.
   - `track: full` and `size: small` → tier is **skip-eligible**.

#### Step 0b: Judge Flag Gate

Read `.archon/config.yaml` → `judge.enabled`:

```yaml
judge:
  enabled: true
```

`judge.enabled: false` ALWAYS wins and short-circuits everything else, regardless of
the tier computed in Step 0a — this override is unchanged from before tiering
existed:

- `judge.enabled: false`: STOP. Do NOT invoke `archon-judge` or any gate — not even
  when the Step 0a tier would have required judge. Return the report below with
  **Verdict: `skipped`**, leave `state.yaml` for the orchestrator to advance to
  `archive`, and exit — then, after archive stages its commit on the branch, the PR
  is opened (single-PR flow).

  ```markdown
  ## Judge Phase Report

  **Change**: {change-name}
  **Verdict**: skipped
  **Reason**: judge phase disabled via `.archon/config.yaml` → `judge.enabled: false`
  ```

- `judge.enabled: true` (or the `judge` section is absent → default `true`): the
  Step 0a tier now decides, for the `full` track (the only track that reaches this
  skill at all — `bugfix` and `fast` have no judge phase in their sequence, so the
  orchestrator never invokes `harness-judge` for them; if you somehow got here on a
  non-full track, treat it as an orchestrator routing bug and report `blocked`
  rather than guessing a tier):
  - Tier **requires judge** (full + substantial/unknown) → proceed to Step 1.
  - Tier **skip-eligible** (full + small) → the single judge review MAY be skipped or
    lightened; report `skipped` (mirroring the format above, with Reason:
    `skip-eligible risk tier (track: full, size: small)`) and let the orchestrator
    advance to `archive`, OR proceed to Step 1 with a lightened review — either is
    acceptable, but do NOT silently fall through without recording which path was
    taken in the report.

### Step 1: Read Configuration

Read `.archon/config.yaml` for the judge flag and the gate settings:

```yaml
judge:
  enabled: true
mutation_testing:
  enabled: false
  tool: gremlins
  threshold: 0.80
playwright:
  enabled: false
  test_dir: e2e
  base_url: http://localhost:3000
```

- `judge.enabled` controls whether the whole phase runs (see Step 0). Default: `true`.
- If the file or a gate section is missing, default that gate to `enabled: false`.

### Step 2: Delegate Single Review to archon-judge

Delegate a single focused review to the `archon-judge` subagent (whose
frontmatter `model:` is the binding hard gate):
- Target: the current change (all files modified by the change plus the change's spec and design)
- Criteria: spec compliance and design coherence (checked directly by archon-judge)
  plus code quality (checked by delegating to the native `code-review` skill, effort `high`,
  scoped to the files this change modified)

The archon-judge subagent runs `code-review` for the code-quality half of the review and
checks spec/design compliance itself, then reports one combined verdict directly.
Capture the verdict from the subagent's output:
- `pass` → code-review reports no `correctness`-category findings AND spec/design compliance holds
- `fail` → one or more `correctness`-category findings from code-review, or a spec/design gap

`code-review` findings outside the `correctness` category (simplification, efficiency, reuse)
are non-blocking notes — surface them in the report but do NOT fail the gate on them.

Note: `judgment-day` is NOT invoked. The SDD path uses this single focused review only.
`judgment-day` remains a standalone opt-in the user may invoke outside the SDD pipeline.

### Step 3: Mutation Testing Gate (conditional)

**Only if `mutation_testing.enabled: true`:**

1. Identify changed Go files from the change's apply-progress or git diff
2. Run: `{tool} unleash --threshold {threshold} --output json {files...}`
3. Parse the JSON output for mutation score and surviving mutants
4. If score < threshold → gate fails, collect surviving mutants as issues
5. If score >= threshold → gate passes

**If `mutation_testing.enabled: false`:** Skip this step entirely.

### Step 3b: Playwright E2E Gate (conditional)

**Only if `playwright.enabled: true` AND the single judge review passed:**

These are the web end-to-end tests generated from the Gherkin `.feature` files
(generated during `sdd-apply`). They run AFTER verify and AFTER the single judge review, per
the harness flow.

1. Ensure the app/dev server is reachable at `playwright.base_url` (start it if the
   project provides a documented command; otherwise report it as a blocker).
2. Run the Playwright suite in `playwright.test_dir`, e.g. `npx playwright test --reporter=json`.
3. Parse results. Any failing scenario → gate fails; collect failures (scenario name,
   `.feature` source, error) as issues.
4. All scenarios pass → gate passes.

**If `playwright.enabled: false`:** Skip this step entirely.

### Step 4: Evaluate Result

A gate that is disabled or skipped counts as `pass` for that column. Overall `pass`
requires the single judge review to pass AND every enabled gate to pass.

| single judge | mutation gate | playwright gate | result |
|---|---|---|---|
| pass | pass (or skipped) | pass (or skipped) | `pass` → advance to archive, then, after archive stages its commit on the branch, the PR is opened (single-PR flow) |
| pass | fail | any | `fail` → enter re-apply loop |
| pass | pass (or skipped) | fail | `fail` → enter re-apply loop |
| fail | any | any | `fail` → enter re-apply loop |

### Step 5: On Pass

1. Run `archon workflow complete --change {change-name}` (marks `judge` phase `completed`)
2. Run `archon workflow retry reset --change {change-name}` — zero the persisted retry
   count now that the cycle is done, so it never leaks into an unrelated later use of
   judge for this change.
3. Return success verdict
4. Orchestrator may proceed to `archive` phase, then, after archive stages its commit on the branch, the PR is opened (single-PR flow)

### Step 6: On Fail — Auto Re-Apply Loop

Run `archon workflow retry increment --change {change-name}` FIRST — it is the
persisted source of truth for the retry count (see `internal/workflow`), not an
in-conversation count-as-you-go variable, so it survives a context compaction or a
closed-and-reopened session mid-loop.

- If it succeeds (reports `retry_count: N/3`):
  1. Build structured feedback block (see format below)
  2. Invoke `sdd-apply` with the feedback as input prompt
  3. After `sdd-apply` completes, invoke `sdd-verify`
  4. If `sdd-verify` passes, return to Step 2 (re-judge)
  5. If `sdd-verify` fails, include verify failures in next feedback block and return to Step 2
- If it fails (`max_retries_exceeded: true`, the 4th attempt):
  1. Return `blocked` with all accumulated issues
  2. Set `max_retries_exceeded: true`
  3. Do NOT attempt further retries, and do NOT call `archon workflow retry increment` again

## Structured Feedback Format

When the single judge review fails, produce feedback that `sdd-apply` can consume directly:

```markdown
## Issues

- {issue_1_description}
- {issue_2_description}
- {mutation_survivor_1}: {mutant_description} (file:line)

## Action Required

- Fix {issue_1}: {specific directive} → `path/to/file.ext:{line}` (requirement: {req_name})
- Fix {issue_2}: {specific directive} → `path/to/file.ext:{line}` (requirement: {req_name})
- Kill mutant {mutant_id}: {what the mutant changed} → `path/to/file.ext:{line}`

## Retry Context

- Attempt: {retry_count} of 3 (the value `archon workflow retry increment` just returned)
- Previous issues resolved: {count}
- Remaining issues: {count}
```

Each directive in `## Action Required` MUST:
- Map to a specific file path and line number when available
- Reference the spec requirement it relates to
- Be a single actionable instruction (not a vague suggestion)

## Mutation Feedback Format

When mutation testing fails, append surviving mutants to the Issues section:

```markdown
## Issues

### Mutation Testing (score: {actual} / threshold: {threshold})

- Surviving mutant `{mutant_id}` in `{file}:{line}`: {mutation_type} — {description}
- Surviving mutant `{mutant_id}` in `{file}:{line}`: {mutation_type} — {description}
```

## Output Contract

Return `## Judge Phase Report`:

```markdown
## Judge Phase Report

**Change**: {change-name}
**Verdict**: {pass | fail | blocked}
**Retry**: {attempt} / 3

### Single Judge Result
- Review: {APPROVED | ISSUES FOUND}
- Confirmed issues: {count} (code-review `correctness` findings + spec/design gaps)
- Non-blocking notes: {count} (code-review simplification/efficiency/reuse findings, if any)

### Mutation Gate
- Status: {passed | failed | skipped}
- Score: {actual} / {threshold} (if run)

### Playwright Gate
- Status: {passed | failed | skipped}
- Scenarios: {passed}/{total} (if run)

### Accumulated Issues
{running total across all retry cycles}

### State Update
- Phase: judge
- Status: {completed | in_progress}
```

If blocked:

```markdown
### BLOCKED
- max_retries_exceeded: true
- Total issues unresolved: {count}
- Recommendation: manual review required
```

## Error Handling

| Condition | Behavior |
|---|---|
| `archon-judge` subagent unavailable | `blocked` — report: `archon-judge subagent not found` |
| `sdd-apply` fails during re-apply | Count as retry attempt; include failure in next feedback |
| `sdd-verify` fails after re-apply | Include verify failures in feedback; count as retry attempt |
| `.archon/config.yaml` missing | Default to `judge.enabled: true`, `mutation_testing.enabled: false`, `playwright.enabled: false`; warn in report |
| Mutation tool not installed | `blocked` — report: `{tool} not found in PATH — install or disable mutation_testing` |
| Playwright not installed | `blocked` — report: `playwright not found — install (npx playwright install) or disable playwright` |
| App/dev server unreachable at base_url | `blocked` — report: `cannot reach {base_url} — start the app or set playwright.base_url` |
| `state.yaml` missing or corrupt | `blocked` — report: `state.yaml not found — run harness-workflow first` |

## Rules

- The judge phase only runs when `.archon/config.yaml` → `judge.enabled` is `true` (default). When `false`, the phase is skipped entirely (Step 0). The flag is toggled from the TUI's "Judge" tab.
- This skill does NOT implement review logic — delegate a single focused review to `archon-judge`.
- This skill does NOT implement fix logic — delegate to `sdd-apply`.
- This skill does NOT implement verification logic — delegate to `sdd-verify`.
- The orchestrator does NOT pause between retries — the loop is fully automatic.
- After max retries, the orchestrator MUST surface accumulated issues to the user.
- Mutation testing and the Playwright E2E gate run ONLY after the single judge review passes. Overall `pass` requires the single judge review AND every enabled gate to pass.
- The Playwright gate executes the specs generated from Gherkin `.feature` files; it never authors product code.
- Each retry cycle counts as ONE attempt regardless of how many sub-steps (apply → verify → judge) it contains.
- `judgment-day` is a standalone opt-in skill — it is NOT invoked by the SDD judge path. Users may invoke it independently outside the SDD pipeline.
