# Judge Report: graphify-integration

<!-- [[graphify-integration]] · [proposal](proposal.md) · [spec](specs/graphify-integration/spec.md) ·
     [design](design.md) · [tasks](tasks.md) · [verify](verify.md) -->

> Phase: judge · Round: 1 · Mode: interactive · Store: openspec · Playwright: off ·
> Impeccable: off · Mutation testing: off.
> Diff baseline: `9ed159c..HEAD` (6 apply commits, 19 files, 579 ins / 54 del).
> **Verdict: APPROVED.**

## Judgment Day — graphify-integration

| Dimension | Judge A | Judge B |
|---|---|---|
| CRITICAL issues | 0 | 0 |
| WARNING-real | 1 (W-01: hardcoded path) | 1 (W-B: mode-d drift) |
| INFO | 2 | 2 |
| Scope violations | 0 | 0 |
| Commit hygiene | PASS | PASS |
| go test | all 12 packages pass | all 12 packages pass |
| gofmt | clean | clean |
| Per-judge verdict | APPROVE | APPROVE |

**Confirmed issues (both judges):** 0  
**Suspect issues (one judge, independently verified by me):** 2 — W-01, W-B  
**Contradictions:** 0

## Build / Test / Vet / gofmt (real output, run by judge executor)

| Command | Result |
|---|---|
| `go build ./...` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `go test ./...` | all 12 packages ok (see below) |
| `gofmt -l` on 12 touched Go files | empty — clean |

```
ok  github.com/archon-ai/archon/cmd/archon
ok  github.com/archon-ai/archon/internal/agent
ok  github.com/archon-ai/archon/internal/config
ok  github.com/archon-ai/archon/internal/initcmd
ok  github.com/archon-ai/archon/internal/mapgen
ok  github.com/archon-ai/archon/internal/models
ok  github.com/archon-ai/archon/internal/opencode
ok  github.com/archon-ai/archon/internal/scaffold
ok  github.com/archon-ai/archon/internal/status
ok  github.com/archon-ai/archon/internal/tui
ok  github.com/archon-ai/archon/internal/version
ok  github.com/archon-ai/archon/skills
```

## Blockers

None. No CRITICAL issues were found by either judge or by independent verification.

## Suspect Warnings (carry into PR body)

### W-01 — `skills/sdd-tasks/SKILL.md:169` hardcodes default path instead of configured `output_dir`

**Found by:** Judge A. Independently confirmed by judge executor.

The `sdd-tasks` conditional block (line 169) reads:

```
read Leiden community data from `graph.json`/`GRAPH_REPORT.md` in
`.archon/graphify/` — **read-only file access...
```

`.archon/graphify/` is the default value of `graphify.output_dir`, not the user-configured value. Any project that sets a non-default `output_dir` (e.g., `archon config set graphify.output_dir custom/graph`) will have `sdd-tasks` looking at the wrong location — it silently falls back to heuristic slice boundaries even though a valid graph exists at the configured path. The graphify SKILL.md consistently uses the canonical `<output_dir>` placeholder throughout (§3, §4, Artifact Layout). The sdd-tasks edit is inconsistent with the skill's own convention.

**Impact:** Graceful — the R-07 fallback handles missing community data. No phase fails. But users with custom `output_dir` get heuristic rather than graph-informed slice boundaries without any diagnostic.

**What a follow-up apply must change:** Replace `.archon/graphify/` at `skills/sdd-tasks/SKILL.md:169` with "the configured `graphify.output_dir`" or `<output_dir>/`, consistent with the graphify skill's own vocabulary.

---

### W-B — `skills/graphify/SKILL.md:85` mode (d) definition diverges from spec R-07 mode (d)

**Found by:** Judge B. Independently confirmed by judge executor.

Spec `R-07` defines mode (d) as "`graph.json` **absent or unreadable**" — a disjunction that covers (1) file absent and (2) file exists but cannot be read (e.g., IO error, permissions denied).

`SKILL.md` §6 mode (d) reads: "`graph.json` **absent and binary unavailable**" — a conjunction that fires when the file is absent AND the binary is also absent (so extraction cannot remedy the absence). An existing `graph.json` that fails with an OS-level read error (permission denied) matches neither mode (d) in the skill (file is not absent) nor mode (e) (not a parse or schema error) — it falls through all eight defined modes.

**Impact:** Graceful — the advisory-degradation architecture means any unhandled failure would result in baseline grep/read in practice. No phase breaks. But the spec-to-skill label drift is a maintenance hazard: a future maintainer tracing spec-to-skill compliance will find a material mismatch on the first explicitly labeled failure mode.

**What a follow-up apply must change:** Align `SKILL.md` §6 mode (d) with the spec by changing "absent and binary unavailable" to "absent or unreadable (IO error, permissions denied)" and updating the advisory note text accordingly (the current note text already uses "unavailable and cannot extract" which is closer to the absent-and-binary-absent semantics). Alternatively, split into two rows and update the spec's mode-d definition to match the skill's conjunction, adding an explicit IO-error mode. Either direction resolves the drift.

## INFO (no action required)

- **Judge A INFO-1:** `TestDisplay_Graphify` "disabled" subtest uses a zero-value `Graphify{}` rather than a Load()-seeded config. Correct by behavior; `TestGraphify_DefaultsAbsentBlock` covers Load() defaults. Test-design note only.
- **Judge A INFO-2:** `Graphify.Version`/`Graphify.OutputDir` lack `yaml:",omitempty"`, unlike Playwright. Since defaults are always pre-seeded, generated configs always emit a `graphify:` block even when disabled. Consistent with the spec's always-pinned-defaults intent; style divergence from Playwright.
- **Judge B INFO-1:** `orchestratorRulesClaude` and `orchestratorRulesOpencode` in `templates.go` rule 9 lack the Feature Branch Chain parenthetical present in root `CLAUDE.md` rule 9. This is the existing templates-behind-root-docs drift, documented as a scoped-out follow-up. No new gap introduced by this change.
- **Judge B INFO-2:** No test for a partial graphify YAML block (e.g., `graphify: { enabled: true }` omitting version/output_dir). yaml.v3 merge semantics correctly leave pre-seeded fields untouched in this case. Gap is in test coverage, not in runtime behavior.

## Scrutiny Points — Independent Verification

1. **SKILL.md as primary deliverable (R-07..R-16 in prose):**
   - §5 sdd-tasks read-only: constraint stated in three independent locations (§3 invocation map row, §5 dedicated subsection with absolute language "even when the binary is present on PATH and `graph.json` is absent," Rules block "NEVER let `sdd-tasks` shell any `graphify` command, under any condition"). Structurally unambiguous. Judge B independently confirmed the three-place count.
   - §4 staleness: bidirectional definition on one line ("Stale iff `mtime < HEAD_time`. Fresh iff `mtime >= HEAD_time`"). No config knob stated explicitly. Cannot be read backward.
   - §2 surface separation: "NEVER" appears in both the §2 table and the Rules block for the `/graphify` slash-command prohibition. The documented Impeccable failure mode is cross-referenced.
   - §6 advisory table: all eight rows end in baseline grep/read. "blocked" appears only in "NEVER return `blocked`." Eight modes present; mode (d) label drift noted in W-B.
   - §8 semantic:false: "MUST NOT invoke any LLM API via Graphify" — correct, scoped language.
   - Assessment: the skill prose is unambiguous on its most critical constraints. W-B is a label-alignment maintenance hazard, not an ambiguity that causes incorrect agent behavior.

2. **Inertness (R-08):** §1 activation contract: when `graphify.enabled: false` "no phase reads this file, no `graphify` command is shelled, `output_dir` is never created, and no phase output changes." Reinforced by the Rules block final bullet. `sdd-explore` Step 3d says "When `graphify.enabled: false`, skip this step entirely — no command runs, no directory is created." No escape path exists. PASS.

3. **R-01 no-validation rule:** `internal/config/config.go` — no `Validate*` helper, no severity field, no Load() validation for Graphify. Version/OutputDir pre-seeded before `yaml.Unmarshal` (lines 136–137), mirroring `c.Judge.Enabled = true`. Clone() value-copies the struct (no pointer fields). PASS.

4. **Narrowed Phase Models guard:** `TestTemplates_ClaudePhaseModelsIsHardGate` scoped to `phaseModelsBlock(t, content)`. "advisory" appears in the rendered CLAUDE.md only at rule 8 and the group-G mapping paragraph — both before the Phase Models section. The Phase Models section itself is advisory-free; the guard still catches any regression. The narrowing matches the test's stated intent. PASS.

5. **Cross-file consistency:** Rule 8 text verbatim-identical across all four locations (templates.go Claude, templates.go Opencode, CLAUDE.md, AGENTS.md). Rule counts internally self-consistent per file (templates 10, CLAUDE.md 11, AGENTS.md 10). Group G question and mapping present and coherent in all three root locations. PASS.

6. **Commit hygiene:** Six conventional-commit subjects; all authored by Alejandro Castaneda `<daanmetal@gmail.com>`; zero `Co-Authored-By`, "Generated with", or agent/tool attribution in any message body. PASS.

7. **Scope discipline:** Diff touches exactly the 19 approved files. No TUI tab, no sdd-verify/harness-judge/mapgen edit, no .gitignore edit, no skill-registry source edit, no graph-report.excerpt.md committed. PASS.

## verify.md Cross-Check

Verify's PASS report is accurate and not overstated. Every claim in the executive summary was independently checked:

- "Rule renumbering internally consistent per file": confirmed (templates 10, CLAUDE 11, AGENTS 10, each self-consistent). ✓
- "No stale 'six' survives": confirmed. ✓
- "R-01 carries no validation/severity and defaults by pre-seed": confirmed. ✓
- "Commit hygiene clean, six conventional subjects, no attribution": confirmed. ✓
- "sdd-tasks read-only encoded in three places": confirmed (independently by both judges and by me). ✓
- "narrowed Phase-Models guard still enforces its real purpose": confirmed. ✓

Two items that verify did not catch (W-01 and W-B) are genuine gaps in the verify review, not errors — both have graceful fallbacks that prevented verify's runtime checks from failing. The verify verdict of PASS stands; these warnings are carry-forward notes for the PR.

## Skill Resolution

`skills/judgment-day/SKILL.md` — loaded. No other skill registry path matched the target files.

## Final Verdict

**JUDGMENT: APPROVED**

Both judges voted APPROVE independently. No CRITICAL issues found. Two suspect warnings (W-01, W-B) are confirmed real by independent verification but cause no correctness failure — the advisory-degradation architecture guarantees graceful fallback in both cases. Carry W-01 and W-B into the PR body as known follow-up items. The change is ready to advance to archive.
