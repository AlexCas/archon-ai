---
name: harness-judge
description: "Trigger: judge phase, judgment gate, post-verify review. Orchestrate judgment-day, mutation testing gate, and auto-re-apply loop."
license: MIT
metadata:
  
  version: "1.0"
  scope: orchestrator-gate
---

## Purpose

Orchestrate the judge phase: invoke `judgment-day` for dual adversarial review, optionally run mutation testing and Playwright E2E tests as quality gates, and automatically re-run `sdd-apply` with structured feedback on failure (up to 3 retries).

## Activation Contract

Load when the orchestrator reaches the `judge` phase after `sdd-verify` passes. This skill wraps existing skills — it does NOT reimplement review or apply logic.

The FIRST action on activation is to read the judge flag (Step 0). If the judge phase is disabled, this skill returns `skipped` immediately and the orchestrator advances to `archive`.

## Hard Rules

- The judge phase is configurable. ALWAYS read `.archon/config.yaml` → `judge.enabled` BEFORE doing anything else. Default: `true` (run when the section is absent). When `judge.enabled: false`, SKIP the entire judge phase — do NOT invoke `judgment-day`, mutation testing, or Playwright — and return `skipped` so the orchestrator advances from verify straight to archive.
- ALWAYS delegate the dual review to the archon-judge subagent. Do NOT invoke judgment-day inline on the orchestrator's model.
- ALWAYS invoke `sdd-apply` for re-fixes. Do NOT apply fixes inline.
- ALWAYS invoke `sdd-verify` after each re-apply before re-judging.
- Mutation testing is OPT-IN. Read `.archon/config.yaml` → `mutation_testing.enabled`. Default: `false`. Skip entirely when disabled.
- Playwright E2E is OPT-IN and runs only for web projects. ALWAYS read `.archon/config.yaml` → `playwright.enabled` to decide whether to run it. Default: `false`. Skip entirely when disabled. These tests run AFTER verify and after `judgment-day` passes.
- Maximum 3 retry cycles. The 4th failure returns `blocked` with `max_retries_exceeded: true`.
- NEVER skip the re-verify step between re-apply and re-judge.
- Accumulate all issues across retry cycles in the feedback block.

## Execution Steps

### Step 0: Judge Flag Gate (HARD GATE)

Read `.archon/config.yaml` → `judge.enabled` BEFORE any other action:

```yaml
judge:
  enabled: true
```

- `judge.enabled: true` (or the `judge` section is absent → default `true`): proceed to Step 1.
- `judge.enabled: false`: STOP. Do NOT invoke `judgment-day` or any gate. Return the report below with **Verdict: `skipped`**, leave `state.yaml` for the orchestrator to advance to `archive`, and exit.

```markdown
## Judge Phase Report

**Change**: {change-name}
**Verdict**: skipped
**Reason**: judge phase disabled via `.archon/config.yaml` → `judge.enabled: false`
```

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
- If the file or either gate section is missing, default that gate to `enabled: false`.

### Step 2: Delegate Dual Review to archon-judge

Delegate the dual adversarial review to the `archon-judge` subagent (whose
frontmatter `model:` is the binding hard gate):
- Target: the current change (all files modified by the change)
- Criteria: spec compliance, design coherence, code quality

The archon-judge subagent invokes `judgment-day` internally and reports its verdict.
Capture the verdict from the subagent's output:
- `pass` → both judges approve with no confirmed CRITICAL or real WARNING issues
- `fail` → one or more confirmed issues found

### Step 3: Mutation Testing Gate (conditional)

**Only if `mutation_testing.enabled: true`:**

1. Identify changed Go files from the change's apply-progress or git diff
2. Run: `{tool} unleash --threshold {threshold} --output json {files...}`
3. Parse the JSON output for mutation score and surviving mutants
4. If score < threshold → gate fails, collect surviving mutants as issues
5. If score >= threshold → gate passes

**If `mutation_testing.enabled: false`:** Skip this step entirely.

### Step 3b: Playwright E2E Gate (conditional)

**Only if `playwright.enabled: true` AND `judgment-day` passed:**

These are the web end-to-end tests generated from the Gherkin `.feature` files
(generated during `sdd-apply`). They run AFTER verify and AFTER judgment-day, per
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
requires judgment-day to pass AND every enabled gate to pass.

| judgment-day | mutation gate | playwright gate | result |
|---|---|---|---|
| pass | pass (or skipped) | pass (or skipped) | `pass` → advance to archive |
| pass | fail | any | `fail` → enter re-apply loop |
| pass | pass (or skipped) | fail | `fail` → enter re-apply loop |
| fail | any | any | `fail` → enter re-apply loop |

### Step 5: On Pass

1. Update `openspec/changes/{change-name}/state.yaml`: set `phase: judge, status: completed`
2. Return success verdict
3. Orchestrator may proceed to `archive` phase

### Step 6: On Fail — Auto Re-Apply Loop

If `retry_count < 3`:

1. Build structured feedback block (see format below)
2. Invoke `sdd-apply` with the feedback as input prompt
3. After `sdd-apply` completes, invoke `sdd-verify`
4. If `sdd-verify` passes, return to Step 2 (re-judge)
5. If `sdd-verify` fails, include verify failures in next feedback block and return to Step 2
6. Increment `retry_count`

If `retry_count == 3` (4th failure):

1. Return `blocked` with all accumulated issues
2. Set `max_retries_exceeded: true`
3. Do NOT attempt further retries

## Structured Feedback Format

When judgment fails, produce feedback that `sdd-apply` can consume directly:

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

- Attempt: {retry_count + 1} of 3
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

### Judgment-Day Result
- Judge A: {APPROVED | ISSUES FOUND}
- Judge B: {APPROVED | ISSUES FOUND}
- Confirmed issues: {count}
- Suspect issues: {count}

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
| `judgment-day` skill unavailable | `blocked` — report: `judgment-day skill not found` |
| `sdd-apply` fails during re-apply | Count as retry attempt; include failure in next feedback |
| `sdd-verify` fails after re-apply | Include verify failures in feedback; count as retry attempt |
| `.archon/config.yaml` missing | Default to `judge.enabled: true`, `mutation_testing.enabled: false`, `playwright.enabled: false`; warn in report |
| Mutation tool not installed | `blocked` — report: `{tool} not found in PATH — install or disable mutation_testing` |
| Playwright not installed | `blocked` — report: `playwright not found — install (npx playwright install) or disable playwright` |
| App/dev server unreachable at base_url | `blocked` — report: `cannot reach {base_url} — start the app or set playwright.base_url` |
| `state.yaml` missing or corrupt | `blocked` — report: `state.yaml not found — run harness-workflow first` |

## Rules

- The judge phase only runs when `.archon/config.yaml` → `judge.enabled` is `true` (default). When `false`, the phase is skipped entirely (Step 0). The flag is toggled from the TUI's "Judge" tab.
- This skill does NOT implement dual review logic — delegate to `judgment-day`.
- This skill does NOT implement fix logic — delegate to `sdd-apply`.
- This skill does NOT implement verification logic — delegate to `sdd-verify`.
- The orchestrator does NOT pause between retries — the loop is fully automatic.
- After max retries, the orchestrator MUST surface accumulated issues to the user.
- Mutation testing and the Playwright E2E gate run ONLY after `judgment-day` passes. Overall `pass` requires judgment-day AND every enabled gate to pass.
- The Playwright gate executes the specs generated from Gherkin `.feature` files; it never authors product code.
- Each retry cycle counts as ONE attempt regardless of how many sub-steps (apply → verify → judge) it contains.
