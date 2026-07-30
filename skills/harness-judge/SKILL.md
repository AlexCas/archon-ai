---
name: harness-judge
description: "Trigger: judge phase, judgment gate, post-verify review. Orchestrate judgment-day, mutation testing gate, and auto-re-apply loop."
license: MIT
metadata:
  
  version: "1.0"
  scope: orchestrator-gate
---

## Purpose

Orchestrate the judge phase: invoke `judgment-day` for dual adversarial review, optionally run mutation testing, Playwright E2E tests, and the Impeccable design-language detection gate as quality gates, and automatically re-run `sdd-apply` with structured feedback on failure (up to 3 retries).

## Activation Contract

Load when the orchestrator reaches the `judge` phase after `sdd-verify` passes. This skill wraps existing skills — it does NOT reimplement review or apply logic.

The FIRST action on activation is to read the judge flag (Step 0). If the judge phase is disabled, this skill returns `skipped` immediately and the orchestrator advances to `archive`, then, after archive stages its commit on the branch (openspec/hybrid mode), the PR is opened (single-PR flow).

## Hard Rules

- The judge phase is configurable. ALWAYS read `.archon/config.yaml` → `judge.enabled` BEFORE doing anything else. Default: `true` (run when the section is absent). When `judge.enabled: false`, SKIP the entire judge phase — do NOT invoke `judgment-day`, mutation testing, Playwright, or the Impeccable gate — and return `skipped` so the orchestrator advances from verify straight to archive.
- ALWAYS delegate the dual review to the archon-judge subagent. Do NOT invoke judgment-day inline on the orchestrator's model.
- ALWAYS invoke `sdd-apply` for re-fixes. Do NOT apply fixes inline.
- ALWAYS invoke `sdd-verify` after each re-apply before re-judging.
- Mutation testing is OPT-IN. Read `.archon/config.yaml` → `mutation_testing.enabled`. Default: `false`. Skip entirely when disabled.
- Playwright E2E is OPT-IN and runs only for web projects. ALWAYS read `.archon/config.yaml` → `playwright.enabled` to decide whether to run it. Default: `false`. Skip entirely when disabled. These tests run AFTER verify and after `judgment-day` passes.
- The Impeccable detection gate is OPT-IN. ALWAYS read `.archon/config.yaml` → `impeccable.enabled` to decide whether to run it. Default: `false`. Skip entirely when disabled — no invocation, no "### Impeccable Gate" section, no result-table column. When enabled, it runs `npx impeccable detect --json .` AFTER `judgment-day` passes, parses the JSON payload for pass/fail (never relying on exit code alone), and applies `impeccable.severity` to decide whether deterministic violations or LLM-critique findings block the gate.
- Security gate is OPT-IN. Read `.archon/config.yaml` → `security.enabled`. Default: `false`. When `security.enabled` is true, treat any unresolved `@security` CRITICAL coverage gap (reported by `sdd-verify`) as a failing gate — do NOT advance to archive. When `security.enabled` is false, skip this check entirely and count it as `pass`.
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
- `judge.enabled: false`: STOP. Do NOT invoke `judgment-day` or any gate. Return the report below with **Verdict: `skipped`**, leave `state.yaml` for the orchestrator to advance to `archive`, and exit — then, after archive stages its commit on the branch (openspec/hybrid mode), the PR is opened (single-PR flow).

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
impeccable:
  enabled: false
  auto_install: false
  severity: block-deterministic
  product_path: ""
  design_path: ""
```

- `judge.enabled` controls whether the whole phase runs (see Step 0). Default: `true`.
- If the file or either gate section is missing, default that gate to `enabled: false`.
- `impeccable.enabled` gates Step 3c. Default: `false`. If the section is missing, treat it as disabled.

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

### Step 3c: Impeccable Detection Gate (conditional)

**Only if `impeccable.enabled: true` AND `judgment-day` passed:**

This is the design-language quality gate backed by the external `npx impeccable`
tool (see the `impeccable` skill). It runs AFTER verify and AFTER judgment-day,
mirroring the Playwright gate's placement in the flow.

1. Read `.archon/config.yaml` → `impeccable.enabled`. If not `true` → skip entirely
   (no invocation, no "### Impeccable Gate" section, no result-table column).
2. Check that `node` and `npx` are on PATH in the target project. Missing → return
   `blocked` with the actionable message (see below). Never silent-pass.
3. If `auto_install: true` AND Impeccable is not yet installed → run
   `npx impeccable install` once, then continue. If `auto_install: false` and the
   package is missing (npx reports not-found) → return `blocked` with the install
   instruction (package-missing variant, below). No silent install.
4. Run `npx impeccable detect --json .` from the target-project root.
5. Interpret results — do NOT rely on exit code for pass/fail:
   - Parse the `--json` payload into deterministic-detector violations vs
     LLM-critique findings.
   - Exit code is used ONLY to detect tool crash / not-found → `blocked`.
   - Unrecognized/unparseable JSON → treat findings as advisory, note the parse
     failure, do NOT hard-fail.
6. Apply `impeccable.severity`:
   - `block-deterministic` (default): deterministic violations > 0 → gate `fail`;
     LLM-critique findings are reported as advisory (non-blocking).
   - `block-all`: any finding from either category (deterministic or LLM-critique)
     → gate `fail`.
   - `advisory`: all findings are advisory only; gate always returns `pass`.
7. Emit the "### Impeccable Gate" section (see Output Contract).
8. Fold the gate status into the overall judge result table as a new column;
   `fail`/`blocked` degrade the overall verdict exactly like the Playwright column.

**Blocked messages (verbatim):**

- Node/npx missing:
  `Impeccable requires Node.js and npx. Install Node.js or set impeccable.enabled: false to skip this gate.`
- Package missing, `auto_install: false`:
  `Impeccable is not installed. Run 'npx impeccable install' (or set impeccable.auto_install: true), or set impeccable.enabled: false to skip this gate.`

**If `impeccable.enabled: false`:** Skip this step entirely.

### Step 4: Evaluate Result

A gate that is disabled or skipped counts as `pass` for that column. Overall `pass`
requires judgment-day to pass AND every enabled gate to pass.

| judgment-day | mutation gate | playwright gate | impeccable gate | result |
|---|---|---|---|---|
| pass | pass (or skipped) | pass (or skipped) | pass (or skipped) | `pass` → advance to archive, then, after archive stages its commit on the branch (openspec/hybrid mode), the PR is opened (single-PR flow) |
| pass | fail | any | any | `fail` → enter re-apply loop |
| pass | pass (or skipped) | fail | any | `fail` → enter re-apply loop |
| pass | pass (or skipped) | pass (or skipped) | fail or blocked | `fail` → enter re-apply loop |
| fail | any | any | any | `fail` → enter re-apply loop |

### Step 5: On Pass

1. Update `openspec/changes/{change-name}/state.yaml`: set `phase: judge, status: completed`
2. Return success verdict
3. Orchestrator may proceed to `archive` phase, then, after archive stages its commit on the branch (openspec/hybrid mode), the PR is opened (single-PR flow)

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

### Impeccable Gate
- Status: {pass | fail | blocked | skipped}
- Severity mode: {block-deterministic | block-all | advisory} (if run)
- Deterministic violations: {n} (if run)
- Advisory findings (LLM critique): {n} (if run)
- Details:
  - {rule id / description per violation, when n > 0}
- (blocked only) Reason: {node/npx missing | package not installed | detect crashed}

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
| `.archon/config.yaml` missing | Default to `judge.enabled: true`, `mutation_testing.enabled: false`, `playwright.enabled: false`, `impeccable.enabled: false`; warn in report |
| Mutation tool not installed | `blocked` — report: `{tool} not found in PATH — install or disable mutation_testing` |
| Playwright not installed | `blocked` — report: `playwright not found — install (npx playwright install) or disable playwright` |
| App/dev server unreachable at base_url | `blocked` — report: `cannot reach {base_url} — start the app or set playwright.base_url` |
| Impeccable: node/npx missing | `blocked` — report: `Impeccable requires Node.js and npx. Install Node.js or set impeccable.enabled: false to skip this gate.` |
| Impeccable: package not installed + `auto_install: false` | `blocked` — report: `Impeccable is not installed. Run 'npx impeccable install' (or set impeccable.auto_install: true), or set impeccable.enabled: false to skip this gate.` |
| Impeccable: config absent | Default to `impeccable.enabled: false` (gate skipped) |
| `state.yaml` missing or corrupt | `blocked` — report: `state.yaml not found — run harness-workflow first` |

## Rules

- The judge phase only runs when `.archon/config.yaml` → `judge.enabled` is `true` (default). When `false`, the phase is skipped entirely (Step 0). The flag is toggled from the TUI's "Judge" tab.
- This skill does NOT implement dual review logic — delegate to `judgment-day`.
- This skill does NOT implement fix logic — delegate to `sdd-apply`.
- This skill does NOT implement verification logic — delegate to `sdd-verify`.
- The orchestrator does NOT pause between retries — the loop is fully automatic.
- After max retries, the orchestrator MUST surface accumulated issues to the user.
- Mutation testing, the Playwright E2E gate, and the Impeccable detection gate run ONLY after `judgment-day` passes. Overall `pass` requires judgment-day AND every enabled gate to pass.
- The Playwright gate executes the specs generated from Gherkin `.feature` files; it never authors product code.
- The Impeccable gate ALWAYS parses the `npx impeccable detect --json .` payload for pass/fail — it NEVER relies on exit code alone (exit code is only used to detect tool crash / not-found). Deterministic-detector violations and LLM-critique findings are mapped to the verdict according to `impeccable.severity` (`block-deterministic` default, `block-all`, or `advisory`).
- Each retry cycle counts as ONE attempt regardless of how many sub-steps (apply → verify → judge) it contains.
