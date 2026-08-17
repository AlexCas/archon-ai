# Verification Report: Graphify Integration — Slice A

<!-- [[graphify-integration]] · [proposal](proposal.md) · [spec](specs/graphify-integration/spec.md) ·
     [design](design.md) · [tasks](tasks.md) · [feature](specs/graphify-integration/graphify-integration.feature) -->

> Phase: verify · Mode: interactive · Store: openspec · Playwright: off · Impeccable: off ·
> Mutation testing: off. Baseline diff: `9ed159c..HEAD` (6 apply commits).
> **Verdict: PASS.**

## Executive Summary

The implementation matches the approved contract. All six machine-testable requirements
(R-01..R-06) pass at runtime; the eleven agent-executed / repo-check requirements
(R-07..R-18) are satisfied by document review (T-19) and a documented binary-missing
dry-run (T-20), which is the coverage model the design and tasks phases explicitly
inherited — not a new gap. Every hard scrutiny point in the verify brief checks out: rule
renumbering is internally consistent per file (templates 10 / CLAUDE 11 / AGENTS 10), the
narrowed Phase-Models guard still enforces its real purpose, no stale "six" survives, R-01
carries no validation/severity and defaults by pre-seed, `output_dir` takes any string, the
pin is `v0.9.45` with zero `v8` reappearance, all scope guards held, and commit hygiene is
clean (six conventional subjects, no attribution trailers).

## Build / Test / Vet / gofmt Evidence (real output)

| Command | Result |
|---|---|
| `go build ./...` | exit 0, no output |
| `go vet ./...` | exit 0, no output |
| `go test ./...` | all 12 packages `ok` |
| `gofmt -l` on the 12 touched Go files | empty (clean) |

Targeted verbose run of the requirement-bearing tests (all PASS):

- `TestConfig_Load` incl. new `graphify_all_fields` subtest; `TestConfig_CloneRoundtrip`;
  `TestGraphify_DefaultsAbsentBlock` — **R-01**
- `TestConfigCmd_GraphifySetGet` (5-key round-trip); `TestConfigCmd_UnknownKeyListsGraphifyKeys` — **R-02**
- `TestDisplay_Graphify` (disabled → only `Enabled:`; enabled → version/output_dir) — **R-03**
- `TestBuildConfig_GraphifyFlag` (`--graphify` → `enabled:true`, defaults seeded) — **R-04**
- `TestTemplates_ContainSDDSessionPreflight` (group G, mapping, `seven`, `A–G`);
  `TestTemplates_FiveRules` (rule 8 renumbered, boundary guard now "no rule 11");
  `TestTemplates_ClaudePhaseModelsIsHardGate` (narrowed to Phase Models block) — **R-05**
- `TestFS_ContainsSkills` includes `graphify` — **R-06**

Note: `gofmt -l` has pre-existing unrelated drift elsewhere in the repo; the check above is
scoped to the files this change touches, all of which are clean.

## Per-Requirement Status (R-01..R-18)

| Req | Verification | Status |
|---|---|---|
| R-01 Config struct | Go tests pass; scalar shape, no `severity`, no `Validate*`, no `Load()` validation; version/output_dir pre-seeded **before** `yaml.Unmarshal` beside `Judge.Enabled`; `Clone()` value-copies | PASS |
| R-02 config get/set | Go tests pass; 5 set + 5 get cases; **both** error strings list the five `graphify.*` keys; `version`/`output_dir` assign verbatim (no format validation) | PASS |
| R-03 status block | Go test passes; block renders after Impeccable, before Models; disabled → `Enabled:` only | PASS |
| R-04 `--graphify` flag | Go test passes; flag registered, `Options.Graphify` → `buildConfig` → config, defaults seeded | PASS |
| R-05 preflight group G | Go tests pass; group G question + mapping in templates.go and both root docs; `A–G`/`seven` consistent; rule 8 inserted, renumbered | PASS |
| R-06 skill file | Go test passes; `skills/graphify/SKILL.md` present, dynamic count (no Go constant) | PASS |
| R-07 advisory absolute | T-19 doc review: all 8 modes (a–h) present, one advisory note each, always fall back, never `blocked`/fail | PASS (doc) |
| R-08 inertness | T-19: Activation Contract §1 + Rules bullet; T-20: `output_dir` not created with binary absent | PASS (doc + dry-run) |
| R-09 surface separation | T-19: §2 table — CLI only shellable; `/graphify` NEVER; MCP deferred w/ auth caveat | PASS (doc) |
| R-10 auto_install | T-19 §7 + T-20: false+missing → note naming install cmds, no silent install; true → install once | PASS (doc + dry-run) |
| R-11 semantic no-LLM | T-19 §8: `semantic:false` = zero LLM calls, pure AST, still queryable | PASS (doc) |
| R-12 staleness | T-19 §4: mtime-vs-HEAD-`%ct`, stale `<` / fresh `≥`, auto re-extract, exact note string | PASS (doc) |
| R-13 sdd-explore consumption | T-19 §3 + sdd-explore Step 3d: fresh/absent/stale paths, no `/graphify`, no MCP | PASS (doc) |
| R-14 sdd-tasks read-only | T-19 §3 row + §5 subsection + sdd-tasks two edit sites: never shells, even binary-present/graph-absent | PASS (doc) |
| R-15 tracked excerpt | T-19 Artifact Layout + sdd-explore Step 3d: ≤40 line/2 KB excerpt refreshed on every (re-)extraction | PASS (doc) |
| R-16 version mismatch | T-19 §9: one advisory note, continue without blocking | PASS (doc) |
| R-17 naming discipline | grep: "spec graph"/"vault graph" appear only in §10 forbidding them for Graphify; "code graph" used throughout | PASS |
| R-18 no .gitignore edit | `.gitignore` untouched in diff; line 9 `# Local directories`, line 10 `.archon/`; `git status` shows no `.archon/graphify` | PASS |

## Gherkin Scenario Mapping

26 scenario blocks (incl. 3 Scenario Outlines). R-01..R-06 scenarios map to the passing Go
tests above. R-07..R-16 scenarios describe agent-executed skill behaviour with no Go exec
wrapper to mock in this repo (identical to `npx impeccable`, only shelled from skill prose);
they are covered by T-19 document review and the T-20 dry-run. R-17/R-18 scenarios are
repo/content checks verified directly. This coverage split is declared in `tasks.md`
("Coverage Declaration") and `design.md` §Testing Strategy — inherited, not a new finding.

## T-19 — SKILL.md Contract Checklist (design §3)

`skills/graphify/SKILL.md` (149 lines) reviewed against the design §3 checklist. Every
contract is present and unambiguous:

- **Inertness (R-08)** — §1 Activation Contract: load only when `graphify.enabled: true`;
  when absent/false "no phase reads this file, no `graphify` command is shelled,
  `output_dir` … is never created, and no phase output changes." Reinforced by the final
  Rules bullet. CLEAR.
- **Invocation surfaces (R-09)** — §2 table: Shell CLI `graphify extract|query|path|explain`
  = "the only shellable surface"; `/graphify` = "**NEVER** … agent-run slash command, not a
  shell command"; MCP `python -m graphify.serve` = "**Deferred** … headless/cron auth caveat
  if adopted later." All three surfaces distinguished; the Impeccable failure mode is named.
  CLEAR.
- **Advisory absolute (R-07)** — §6 table: all eight modes a–h, each a single-line advisory
  note + fallback; header states "**Never** return `blocked`, **never** fail the phase,
  **never** halt the SDD flow." CLEAR.
- **sdd-tasks read-only (R-14)** — encoded in **two** places: the §3 map row
  ("file read — never shell") and a dedicated §5 subsection that explicitly forbids shelling
  "even when the binary is present on PATH and `graph.json` is absent," and states this is
  "structural, not a policy choice re-derived per run." Judgement: this is **structurally
  obvious, not merely asserted** — the two-site encoding plus the single-extraction-site
  invariant ("`sdd-explore` is the sole extraction site") makes the constraint hard to miss
  on a partial read. Reinforced a third time in the Rules block. CLEAR.
- **Staleness (R-12)** — §4: reference `git show -s --format=%ct HEAD`; subject
  `mtime(graph.json)`; "Stale iff `mtime < HEAD_time`. Fresh iff `≥`"; fresh → reuse,
  do NOT re-extract; stale+binary → auto re-extract + refresh excerpt + emit exactly
  `graph may be stale — re-extracting` (byte-matches feature line 163); stale+no-binary →
  mode (f); "no config knob." Matches spec R-12 exactly. CLEAR.
- **auto_install (R-10)** — §7: false+missing → note naming `uv tool install graphifyy` /
  `pipx install graphifyy`, never silent; true+missing → install once, "not a repeated
  install on every run." CLEAR.
- **semantic:false = zero LLM (R-11)** — §8: MUST NOT invoke any LLM API, "pure local,
  deterministic tree-sitter AST … no network calls, no model cost," still queryable. CLEAR.
- **version-mismatch (R-16)** — §9: single advisory note, continue, "Never fail the phase
  over a version drift alone." CLEAR.
- **naming (R-17)** — §10: "code graph" exclusively for Graphify output; "spec graph"/"vault
  graph" reserved for `internal/mapgen`'s `openspec/map.md`; disjointness preserves Slice C.
  grep confirms no misuse. CLEAR.

Package name is the PyPI `graphifyy` (double-y) in install commands; the CLI/binary is
`graphify`; version pin `v0.9.45`. Consistent throughout. No `v8` anywhere.

## T-20 — Binary-Missing Dry-Run Record (R-07 a/b, R-10)

Executed in the repo working tree; **nothing was installed**. Exact commands and observations:

- `which graphify` → `which: no graphify in (…PATH…)`, exit 1.
- `command -v graphify` → exit 1 (absent).
- `graphify --version` → `bash: line 5: graphify: command not found`, exit 127.
- Runtime probe: `python3` present (`/home/linuxbrew/.linuxbrew/bin/python3`), `uv` present,
  `pipx` absent. (So the live environment is R-07 mode (a) — binary missing but a runtime
  exists; mode (b) — no Python/uv/pipx at all — is not reproducible here since python3/uv
  exist, and is verified by document review of §6 row b instead.)
- `ls -la .archon/graphify` → `No such file or directory` — `output_dir` was **not** created
  (consistent with R-08 inertness and the note-and-continue contract; no side effect from a
  probe).

Documented behaviour confirmed against SKILL.md §6 row (a) and §7: with the binary absent and
`auto_install: false` (default), the consuming phase MUST emit the single advisory note
`graphify unavailable: binary not on PATH; proceeding with baseline grep/read`, name the
install commands (`uv tool install graphifyy` / `pipx install graphifyy`), NOT install, and
continue with baseline grep/read — never failing the phase. The skill prose matches this
exactly. Note-and-continue is the documented and correct behaviour.

## Scrutiny Points (verify brief)

1. **Rule renumbering** — templates.go both Rules blocks end at **10**; root `CLAUDE.md` at
   **11**; root `AGENTS.md` at **10**. Each file internally consistent; graphify rule
   inserted as 8 in all four blocks with clean renumbering after. No rule text cross-copied.
   `TestTemplates_FiveRules` updated coherently (shared strings renumbered 8→9, 9→10; boundary
   guard moved from "no rule 10" to "no rule 11"). The CLAUDE 11 / templates 10 asymmetry is
   the known archive-before-PR drift (rule 10 in root CLAUDE only), explicitly documented as a
   scoped-out follow-up in tasks.md — correct, not a bug. CONFIRMED.
2. **Narrowed guard** — `TestTemplates_ClaudePhaseModelsIsHardGate` now scopes the `advisory`
   assertion via `phaseModelsBlock(t, content)`. The helper slices from `## Phase Models` to
   the next `## ` header (or EOF), so it captures the whole block. "advisory" appears in the
   rendered CLAUDE.md only at the group-G mapping (line 98) and rule 8 (line 174), both
   **before** the Phase Models section (line 185); the section itself is advisory-free and the
   guard still fails if that changes. The narrowing matches the test's own stated intent and
   does not weaken its real purpose. CONFIRMED.
3. **Preflight group count** — no stale "six"/"seis" anywhere in the three files; `seven` and
   `A–G` present and consistent across `CLAUDE.md`, `AGENTS.md`, `templates.go`; group G
   question text and the `graphify.enabled` mapping paragraph present in all three. CONFIRMED.
4. **R-01 no validation** — no `Validate*` helper, no `Load()` validation, no `severity` field;
   `version`/`output_dir` defaulted by pre-seeding the two exported constants **before**
   `yaml.Unmarshal` (mirrors `c.Judge.Enabled = true`); `Clone()` value-copies the struct.
   CONFIRMED.
5. **output_dir any string** — `setConfigValue` assigns `graphify.version`/`output_dir`
   verbatim; no path-format validation crept in. CONFIRMED.
6. **Pin v0.9.45** — appears in config constant, buildConfig, SKILL.md, spec/design. Grep for
   `\bv8\b` across all changed code/docs: zero hits. CONFIRMED.
7. **Scope guards** — no TUI file, no `sdd-verify`/`harness-judge`/`mapgen` edit, no
   `.gitignore` edit, no `skills/skill-registry/SKILL.md` source edit, and
   `graph-report.excerpt.md` was NOT created. Diff touches exactly the 19 approved files.
   CONFIRMED.
8. **Commit hygiene** — six conventional-commit subjects (`feat(config)`, `feat(cli)`,
   `feat(templates)`, `feat(docs)`, `feat(skills)`, `fix(templates)`); author uniformly
   `Alejandro Castaneda <daanmetal@gmail.com>`; **zero** `Co-Authored-By` / "Generated with" /
   agent-or-tool attribution in any message body. CONFIRMED.

Consumer-skill edits (sdd-explore Step 3d, sdd-tasks two-site read-only note, chained-pr
advisory line, embed_test) all match the design and reinforce the R-13/R-14 constraints.

## Issues

- **CRITICAL:** none.
- **WARNING:** none blocking. Two carry-forward notes for the PR (below).
- **SUGGESTION:** display.go's enabled-branch field labels have cosmetic mixed alignment
  (`Version:    ` vs `Output Dir: ` vs `Semantic:   `). This is verbatim from design §1 and
  matches the Impeccable idiom; the test asserts substring presence, not alignment. No action
  required — noted only for polish.

## Warnings to Carry Into the PR

1. **R-07..R-16 have no automated coverage** — agent-executed skill behaviour with no Go exec
   wrapper to mock in this repo (same as `npx impeccable`). Correctness rests on the SKILL.md
   prose reviewed in T-19 plus the T-20 dry-run. This is the approved, inherited coverage
   model; the PR reviewer should read `skills/graphify/SKILL.md` directly rather than expect
   unit tests for these rows.
2. **templates.go archive-before-PR drift** — templates.go remains behind the root docs (root
   CLAUDE.md rule 10 archive-before-PR content, and the AGENTS.md 10 vs CLAUDE.md 11 rule-count
   asymmetry, both come from that drift). Bringing templates.go current is an explicitly
   scoped-out follow-up recorded in tasks.md — call it out in the PR body so it is not read as
   an inconsistency.

## What the Judge Should Look At Hardest

- **SKILL.md as the primary deliverable** — R-07..R-16 live entirely in prose. Re-read §§1–10,
  especially the §5 sdd-tasks read-only guarantee (structural, two-site) and the §4 staleness
  algorithm (mtime `<` HEAD `%ct`), since these are the most easily eroded on a future edit.
- **The narrowed Phase-Models guard** — confirm the reviewer agrees scoping to
  `phaseModelsBlock` did not create a blind spot; the Phase Models section must never be
  describable as "advisory."
- **Rule-count asymmetry across the three files** (templates 10 / CLAUDE 11 / AGENTS 10) — an
  automated diff reviewer may flag it; it is intended and documented.

## Verdict

**PASS.** All machine-testable requirements pass at runtime with no CRITICAL or blocking
WARNING issues; all agent-executed requirements are satisfied by the T-19 review and T-20
dry-run; every scope and hygiene guard held. Ready to advance to the judge phase.
