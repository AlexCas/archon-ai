---
name: harness-judge
description: "Trigger: judge phase, judgment gate, post-verify review. Orchestrate judgment-day, mutation testing gate, and auto-re-apply loop."
license: MIT
metadata:
  
  version: "1.0"
  scope: orchestrator-gate
---

## Purpose

Orchestrate the judge phase: invoke `judgment-day` for dual adversarial review, optionally run mutation testing as a quality gate, and automatically re-run `sdd-apply` with structured feedback on failure (up to 3 retries).

## Activation Contract

Load when the orchestrator reaches the `judge` phase after `sdd-verify` passes. This skill wraps existing skills — it does NOT reimplement review or apply logic.

## Hard Rules

- ALWAYS invoke `judgment-day` skill for dual review. Do NOT perform review inline.
- ALWAYS invoke `sdd-apply` for re-fixes. Do NOT apply fixes inline.
- ALWAYS invoke `sdd-verify` after each re-apply before re-judging.
- Mutation testing is OPT-IN. Read `.archon/config.yaml` → `mutation_testing.enabled`. Default: `false`. Skip entirely when disabled.
- Maximum 3 retry cycles. The 4th failure returns `blocked` with `max_retries_exceeded: true`.
- NEVER skip the re-verify step between re-apply and re-judge.
- Accumulate all issues across retry cycles in the feedback block.

## Execution Steps

### Step 1: Read Configuration

Read `.archon/config.yaml` for mutation testing settings:

```yaml
mutation_testing:
  enabled: false
  tool: gremlins
  threshold: 0.80
```

If the file or section is missing, default to `enabled: false`.

### Step 2: Invoke Judgment-Day

Delegate to the `judgment-day` skill:
- Target: the current change (all files modified by the change)
- Criteria: spec compliance, design coherence, code quality

Capture the verdict:
- `pass` → both judges approve with no confirmed CRITICAL or real WARNING issues
- `fail` → one or more confirmed issues found

### Step 3: Mutation Testing Gate (conditional)

**Only if `mutation_testing.enabled: true`:**

1. Identify changed Go files from the change's apply-progress or git diff
2. Run: `{tool} unleash --threshold {threshold} --output json {files...}`
3. Parse the JSON output for mutation score and surviving mutants
4. If score < threshold → gate fails, collect surviving mutants as issues
5. If score >= threshold → gate passes

**If `mutation_testing.enabled: false`:** Skip this step entirely. Only `judgment-day` verdict determines the result.

### Step 4: Evaluate Result

| judgment-day | mutation gate | result |
|---|---|---|
| pass | pass (or skipped) | `pass` → advance to archive |
| pass | fail | `fail` → enter re-apply loop |
| fail | pass (or skipped) | `fail` → enter re-apply loop |
| fail | fail | `fail` → enter re-apply loop |

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
| `.archon/config.yaml` missing | Default to `mutation_testing.enabled: false`; warn in report |
| Mutation tool not installed | `blocked` — report: `{tool} not found in PATH — install or disable mutation_testing` |
| `state.yaml` missing or corrupt | `blocked` — report: `state.yaml not found — run harness-workflow first` |

## Rules

- This skill does NOT implement dual review logic — delegate to `judgment-day`.
- This skill does NOT implement fix logic — delegate to `sdd-apply`.
- This skill does NOT implement verification logic — delegate to `sdd-verify`.
- The orchestrator does NOT pause between retries — the loop is fully automatic.
- After max retries, the orchestrator MUST surface accumulated issues to the user.
- Mutation testing runs ONLY after `judgment-day` passes (both gates must pass for overall pass).
- Each retry cycle counts as ONE attempt regardless of how many sub-steps (apply → verify → judge) it contains.
