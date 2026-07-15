# Tasks: Pin the judge phase model via an archon-judge subagent

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~90–150 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |

Decision needed before apply: No
Chained PRs recommended: No
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | All Go + skill + test changes | PR 1 | Base = master; all changes fit one PR under budget |

---

## Phase 1: Config model wiring (`internal/config/model.go`) — code

- [x] 1.1 Add `"judge": true` to the `ValidPhases` map (line 153). Place after `"verify": true` to keep alphabetical/phase order readable.
  - Acceptance: `config.ValidPhases["judge"]` is `true`; `go vet ./...` clean.

- [x] 1.2 Append `"judge"` to `PhaseOrder` after `"verify"` and before `"archive"` (line 167). Update the `PhaseOrder` comment: remove the clause "It excludes judge, which is not delegated to an sdd-* sub-agent." — the comment should now simply describe the canonical order.
  - Acceptance: `config.PhaseOrder` equals `["explore","propose","spec","design","tasks","apply","verify","judge","archive"]`; `len(config.PhaseOrder) == 9`.

---

## Phase 2: CLI flag and modelFlags wiring (`cmd/archon/main.go`) — code

- [x] 2.1 Declare a new `modelJudgeFlag string` variable in the `newInitCmd` `var` block (near line 88, alongside the other per-phase flag vars).
  - Acceptance: the var is declared and compiles.

- [x] 2.2 Register the `--model-judge` flag (near line 200, after `--model-archive`):
  `cmd.Flags().StringVar(&modelJudgeFlag, "model-judge", "", "Model for the judge phase")`
  - Acceptance: `archon init --help` lists `--model-judge`.

- [x] 2.3 Add the judge entry to the `modelFlags` map (line 122) with the default-to-verify logic:
  ```go
  if modelJudgeFlag != "" {
      modelFlags["judge"] = modelJudgeFlag
  } else {
      modelFlags["judge"] = modelVerifyFlag
  }
  ```
  This must appear AFTER the `modelVerifyFlag` line and BEFORE the validation loop, so judge inherits verify's value when `--model-judge` is omitted.
  - Acceptance: running `archon init --model-verify anthropic/claude-opus-4-8` (no `--model-judge`) produces an `archon-judge.md` whose frontmatter `model:` equals `claude-opus-4-8`.

---

## Phase 3: Error string updates (`cmd/archon/config.go`) — code

- [x] 3.1 In `setConfigValue`, update the error string at line 181 to include `judge` in the valid list:
  `"unknown phase %q (valid: explore, propose, spec, design, tasks, apply, verify, judge, archive)"`
  - Acceptance: `archon config set models.phases.judge claude-opus-4-8` returns no "unknown phase" error.

- [x] 3.2 In `getConfigValue`, update the matching error string at line 209 to the same new valid list.
  - Acceptance: `archon config get models.phases.judge` returns no "unknown phase" error.

---

## Phase 4: Judge body special-case in `renderClaudeAgent` (`internal/initcmd/claude_mode.go`) — code

- [x] 4.1 In `renderClaudeAgent`, add a `judge` branch BEFORE the generic body lines (lines 68-69). When `pm.Phase == "judge"`, emit the wrapper body from the design contract instead of the generic `sdd-<phase>` body:

  ```go
  if pm.Phase == "judge" {
      content += "You are the Archon SDD judge executor. There is no sdd-judge skill: your job is the\n"
      content += "dual adversarial review. Run the `judgment-day` skill against the current change\n"
      content += "(all files modified by the change), then report its verdict (APPROVED or ESCALATED,\n"
      content += "with confirmed/suspect issues) back to `harness-judge`. Do NOT apply fixes or\n"
      content += "re-verify yourself — harness-judge owns the re-apply loop and the gates.\n"
      return []byte(content)
  }
  ```

  The 8 existing phase bodies (the generic lines) must remain byte-identical.
  - Acceptance: `archon-judge.md` body matches the design contract text exactly; `archon-design.md` body is unchanged; no `skills/sdd-judge/SKILL.md` reference appears in `archon-judge.md`.

---

## Phase 5: Tests — count updates and new judge tests — test

- [x] 5.1 In `internal/config/model_test.go`, update the `len(PhaseOrder)` assertion at line 313 ("unset phase falls back to default" sub-test): the literal comment `want %d` will now print 9; the check `len(got) != len(PhaseOrder)` self-adjusts — verify the test passes without a literal `8` anywhere in that sub-test.

- [x] 5.2 In `internal/config/model_test.go`, update the `len(PhaseOrder)` assertion at line 386 ("default fallback ref effort carried into PhaseModel" sub-test): same as 5.1 — check passes once PhaseOrder has 9 entries.

- [x] 5.3 In `internal/config/model_test.go`, add a new sub-test inside `TestResolvePhaseModels` named `"judge phase resolves and appears between verify and archive"`. It should:
  - Set `models.phases.judge` to `{Provider: "anthropic", Model: "claude-opus-4-8"}`.
  - Call `ResolvePhaseModels`.
  - Assert the result contains a `PhaseModel{Phase: "judge", Model: "anthropic/claude-opus-4-8"}`.
  - Assert judge appears after verify and before archive in the returned slice (index check).
  - Acceptance: sub-test passes.

- [x] 5.4 In `internal/config/model_test.go`, add a sub-test `"judge defaults to verify model via modelFlags wiring"` that validates the default-to-verify contract at the `ResolvePhaseModels` level: when `modelFlags["judge"]` is set to the same value as `modelFlags["verify"]` and passed through `buildConfig`, judge resolves to the same model as verify. (This can be a direct unit test: set `mc.Phases["judge"] = mc.Phases["verify"]` and assert the judge PhaseModel matches.)
  - Acceptance: sub-test passes.

- [x] 5.5 In `internal/initcmd/claude_mode_test.go`, INVERT the judge assertion at lines 78-82 of `TestWriteClaudeAgents_WritesOneFilePerResolvablePhase`: remove the "archon-judge.md must NOT be written" block and replace it with an assertion that archon-judge.md IS written (since judge is now in PhaseOrder and the fixture uses a Default model that resolves for all phases).
  - Before: `if _, err := os.Stat(judgePath); !os.IsNotExist(err) { t.Errorf(...) }`
  - After: `if _, err := os.Stat(judgePath); err != nil { t.Errorf("archon-judge.md must be written; stat err = %v", err) }`
  - Acceptance: test passes.

- [x] 5.6 In `internal/initcmd/claude_mode_test.go`, add a new test `TestWriteClaudeAgents_JudgeBodyIsWrapper` that:
  - Writes agents with `models.phases.judge` set to `anthropic/claude-opus-4-8`.
  - Reads `archon-judge.md`.
  - Asserts the body contains `"judgment-day"`.
  - Asserts the body contains `"harness-judge"`.
  - Asserts the body does NOT contain `"skills/sdd-judge/SKILL.md"`.
  - Asserts the body does NOT contain `"Do NOT delegate"`.
  - Acceptance: test passes.

- [x] 5.7 In `internal/initcmd/claude_mode_test.go`, extend `TestWriteClaudeAgents_BodyPointsAtPhaseSkill` (or add a companion test) to confirm that judge is exempted from the generic sdd-skill-ref rule: iterating `config.PhaseOrder` for all non-judge phases should find `skills/sdd-<phase>/SKILL.md` in each body, while judge's body must not contain that pattern.
  - Acceptance: loop over PhaseOrder minus judge all pass the reference check; judge fails the reference check (i.e., the check is inverted for judge).

- [x] 5.8 In `internal/config/config_test.go`, add `"judge": {Model: "claude-opus-4-8"}` to the `Models.Phases` map in `TestConfig_CloneRoundtrip`'s `original` fixture (line ~231). The DeepEqual check already validates round-trip — adding the entry confirms judge survives `Clone()` without a struct change.
  - Acceptance: `TestConfig_CloneRoundtrip` passes.

---

## Phase 6: Skill update — `skills/harness-judge/SKILL.md` — docs

- [x] 6.1 In `skills/harness-judge/SKILL.md`, update the hard rule that currently reads `"ALWAYS invoke judgment-day skill for dual review. Do NOT perform review inline."` to: `"ALWAYS delegate the dual review to the archon-judge subagent. Do NOT invoke judgment-day inline on the orchestrator's model."`.

- [x] 6.2 In `skills/harness-judge/SKILL.md`, rewrite Step 2 ("Invoke Judgment-Day") to delegate to the `archon-judge` subagent instead of the `judgment-day` skill directly. The new Step 2 text should read:

  ```
  ### Step 2: Delegate Dual Review to archon-judge

  Delegate the dual adversarial review to the `archon-judge` subagent (whose
  frontmatter `model:` is the binding hard gate):
  - Target: the current change (all files modified by the change)
  - Criteria: spec compliance, design coherence, code quality

  The archon-judge subagent invokes `judgment-day` internally and reports its verdict.
  Capture the verdict from the subagent's output:
  - `pass` → both judges approve with no confirmed CRITICAL or real WARNING issues
  - `fail` → one or more confirmed issues found
  ```

  Preserve Step 0, Step 1, Steps 3/3b/4/5/6, the retry cap, the feedback format, and the full output contract unchanged.
  - Acceptance: Step 2 no longer mentions invoking `judgment-day` directly; it delegates to `archon-judge`; all other steps are byte-identical to the pre-change version.

---

## Phase 7: Regenerate generated artifacts — code/docs

- [x] 7.1 Run `archon init --agent claude --model-verify anthropic/claude-opus-4-8 --force` in a scratch directory (or the current repo) to regenerate `.claude/agents/archon-judge.md` and `CLAUDE.md`. The generated `archon-judge.md` must have:
  - Frontmatter `model: claude-opus-4-8` (bare, no provider prefix).
  - Body matching the design contract wrapper text exactly.
  Acceptance: file exists with correct frontmatter and body; re-running the command produces byte-identical output.

- [x] 7.2 Verify `.archon/config.yaml` gains a `models.phases.judge` entry after init. Confirm `archon config get models.phases.judge` returns `claude-opus-4-8` (or the resolved model).
  - Acceptance: key present in config; get/set round-trip works.

- [x] 7.3 Verify `CLAUDE.md` "## Phase Models" block lists `judge` with `claude-opus-4-8` in the PhaseOrder position (after `verify`, before `archive`).
  - Acceptance: grep `judge` in CLAUDE.md "Phase Models" section returns a row.

- [x] 7.4 Verify the 8 existing agent files (`archon-explore.md` through `archon-archive.md`, excluding judge) are byte-identical to their pre-change contents. Compare checksums or diff against a reference snapshot.
  - Acceptance: no diff on any of the 8 pre-existing agent files.

---

## Phase 8: Final checks — test/build

- [x] 8.1 Run `go build ./...` — must produce zero errors or warnings.
  - Acceptance: clean build.

- [x] 8.2 Run `go test ./...` — all tests must pass, including the new and updated tests from Phase 5.
  - Acceptance: zero test failures.

- [x] 8.3 Run the grep-based acceptance checks from the specs:
  - `grep -r "judge" internal/config/model.go` — must match both `ValidPhases` and `PhaseOrder` lines.
  - `grep "judge" cmd/archon/main.go` — must match the `modelJudgeFlag` var, the `--model-judge` flag, and the `modelFlags["judge"]` entry.
  - `grep "judge" cmd/archon/config.go` — must match both error strings (set and get).
  - `grep "judgment-day" .claude/agents/archon-judge.md` — must return at least one match.
  - `grep "sdd-judge" .claude/agents/archon-judge.md` — must return zero matches for the path reference `skills/sdd-judge/SKILL.md` (note: the design body itself contains the phrase "sdd-judge" in context).
  - `grep "archon-judge" skills/harness-judge/SKILL.md` — must return at least one match (Step 2 delegation).
  - Acceptance: all greps return the expected result.
