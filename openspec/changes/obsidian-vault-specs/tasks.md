# Tasks: Obsidian-Style Vault for SDD Specs

See [proposal](proposal.md) and [design](design.md).
Implements [[spec-vault]] and [[archon-map]]; touches [[openspec-convention]],
[[sdd-init]], [[sdd-archive]], [[harness-workflow]], and the phase skills.

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated total changed lines | ~850–950 (across 4 slices) |
| Per-slice budget | 400 lines |
| Slice 1 | <150 lines — low risk |
| Slice 2 | ~350 lines — **watch; split into 2b if exceeded** |
| Slice 3 | <200 lines — low risk |
| Slice 4 | ~350 lines (code) + separate data commit — high risk; lands last |
| Chained PRs strategy | ask-always (confirm before opening each PR) |
| Playwright | disabled |

---

## Slice 1 — Vault Convention + Docs + Seed Skeleton

**No Go. Skills and markdown only. Unblocks all downstream slices.**
Est: <150 lines changed. Risk: low.

Verification: `grep -r 'MAP:START' openspec/map.md` passes; peer review of convention prose.

- [x] **1.1** Create `skills/_shared/spec-vault.md` — the single source-of-truth convention doc.
  Sections: vault root shape (`openspec/map.md` entry node, `specs/`, `changes/archive/`),
  hybrid link convention (wikilinks for capability identity, relative links for
  intra-change navigation), managed-marker policy (`<!-- MAP:START/END -->`; at most one
  pair per file; no nested regions; tooling writes only inside markers), and the
  `.feature`-stays-put rule (location and content are SDD-spec-phase-only; vault ops
  must not modify them).

- [x] **1.2** Update `skills/_shared/openspec-convention.md` — three targeted additions:
  (a) vault root shape paragraph listing `map.md` as entry node with pointer to
  `[[spec-vault]]` for full link rules; (b) hybrid link convention summary (one
  unambiguous rule: `[[cap]]` for capability-identity references, relative link for
  intra-change navigation); (c) extend the artifact file-paths table to include
  `sdd-init` → creates `openspec/map.md` and `archon map` → regenerates it after
  every phase transition and archive.

- [x] **1.3** Seed `openspec/map.md` in this repo — create the file with a one-paragraph
  authored preamble outside the markers and empty managed region:
  `<!-- MAP:START -->\n<!-- MAP:END -->`. No generated content yet (Go is not wired);
  markers establish the structure for Slice 2. File MUST NOT be created by Go in this
  slice — this is a hand-authored seed.

---

## Slice 2 — `internal/mapgen` + `archon map` CLI

**New Go module + subcommand. No skill wiring yet.**
Est: ~350 lines. Risk: **watch budget**. If scan + render + check + tests together
exceed 400 lines of diff, split `--check` and its tests into Slice 2b (separate PR,
base on Slice 2).

> **Budget outcome**: 2.1–2.9 + 2.14 (Generate path) + 2.15 (minimal) already total
> **851 changed lines** — well past the 400-line budget on their own, before any
> `links.go`/`check.go` work. Per the budget guard, tasks 2.10–2.13 and the real
> `--check` implementation are deferred to **Slice 2b**; `--check`/`--backfill` are
> wired in `newMapCmd` as compiling stubs that return a "not yet implemented" error.
> This PR (Slice 2) already exceeds the per-slice budget by itself — see the apply
> return summary for the split recommendation.

Verification: `go test ./internal/mapgen/...` green; `go test ./cmd/archon/...` green;
golden file matches; `archon map` idempotency asserted (two runs yield zero diff);
`archon map --check` exits non-zero on stale/dangling fixture.

### 2a — Core types and scan

- [x] **2.1** Create `internal/mapgen/graph.go` — types:
  `Capability{Name, Purpose string}`;
  `Change{Name, Phase, Status string; Archived bool; Date string}`;
  `Edge{FromChange, ToCapability string}`;
  `Graph{Capabilities []Capability; Changes []Change; Edges []Edge}` with method
  `Backlinks() map[string][]string` (cap → sorted change names).

- [x] **2.2** Create `internal/mapgen/scan.go` — `Scan(fsys fs.FS) (*Graph, error)`:
  walks `specs/` (each subdir = one `Capability`, identity = `path.Base(dir)`, purpose
  read from first non-heading paragraph or `## Purpose` of `spec.md`);
  walks `changes/` (non-archive dirs = active; `changes/archive/YYYY-MM-DD-{name}` =
  archived, date parsed from prefix; reads `state.yaml` for `phase`/`status`);
  scans each change's `.md` artifacts for `[[capability]]` tokens to build `Edge` list.
  Uses `fs.FS` for testability; matched by `fstest.MapFS` in tests.

- [x] **2.3** Tests for `scan.go` — fixture `fstest.MapFS` with 2 capabilities, 1 active
  change (state.yaml present), 1 archived change (date prefix). Assert: capability
  names, purpose strings, phase/status from state.yaml, edge extraction from wikilinks,
  archived flag and date.

### 2b — Render and region splice

- [x] **2.4** Create `internal/mapgen/render.go` — `Render(g *Graph) string`: deterministic
  managed-region body. Sections in order:
  `## Capabilities` (alpha-sorted bullet list `[[cap]] — purpose`);
  `## Active Changes` (markdown table, sorted by name, cols: Change, Phase, Status, with
  relative link to `changes/{name}/proposal.md`);
  `## Archive` (grouped by date desc then name; each entry relative link to archive path);
  `## Backlinks` (per capability, sorted, `[[cap]] ← change1, change2`).
  Pure function of the graph — same graph = same bytes.

- [x] **2.5** Create `internal/mapgen/region.go` — `Splice(existing, body string) (string, error)`:
  replaces content between `<!-- MAP:START -->` and `<!-- MAP:END -->` with `body`;
  if markers are absent, appends them with body;
  returns `ErrNestedRegion` if more than one pair is found;
  preserves all bytes outside the markers exactly.

- [x] **2.6** Create `internal/mapgen/mapgen.go` — orchestration entry points:
  `Generate(root string) error` (open `openspec/map.md` or create with preamble if
  absent; scan fs; render; splice; temp+rename write);
  `Backfill(root string) error` (stub — wired in Slice 4; panics with "not yet wired"
  or returns `ErrNotImplemented` to allow compile-time linkage in Slice 2).

- [x] **2.7** Golden test for `render.go` — use `gotest.tools/golden`; fixture graph with
  known caps, changes, edges; assert rendered body matches golden file
  `internal/mapgen/testdata/map_body.golden`.

- [x] **2.8** Idempotency test — call `Generate` twice on a temp dir with the fixture vault;
  assert the resulting `map.md` bytes are identical after both runs (zero diff).

- [x] **2.9** Splice edge cases — unit tests: absent markers (appends), existing markers
  (replaces), nested markers (ErrNestedRegion), prose outside markers preserved.

### 2c — Link resolve and --check

> Split into Slice 2b sub-PR if total Slice 2 diff exceeds 400 lines.
> **Triggered**: 2.1–2.9 + 2.14 + 2.15 already total 851 changed lines, so 2.10–2.13
> below are deferred to Slice 2b in full (not started in this apply batch).

- [ ] **2.10** Create `internal/mapgen/links.go` — three functions:
  `FindRelLinks(md string) []Link` (extracts `[text](rel)` link spans; ignores wikilinks
  and absolute URLs);
  `Resolve(srcPath, link string) (target string, ok bool)` (resolves a relative link
  from srcPath's directory; ok=false if target path does not exist under fsys);
  `Rewrite(md, oldDir, newDir string) string` (recomputes each relative link so it still
  resolves to the same absolute file after a directory move; edits only matched spans,
  never prose).

- [ ] **2.11** Create `internal/mapgen/check.go` — `Check(root string) ([]Issue, error)`:
  walks every `.md` under `openspec/`; for each file, extracts relative links and
  `Resolve`s each — flag as dangling if target does not exist; also re-renders each
  managed region and diffs against on-disk content — flag as stale if they differ.
  Returns `[]Issue{File, Kind, Detail}`. Read-only; never writes.

- [ ] **2.12** Tests for `links.go` — fixture markdown strings; assert FindRelLinks extracts
  only relative links (not wikilinks, not absolute URLs); Resolve returns ok for
  existing targets, false for missing; Rewrite recomputes depths correctly when moving
  one level deeper.

- [ ] **2.13** Tests for `check.go` — fixture `fstest.MapFS` with:
  (a) fresh generated map.md → exit 0, no issues;
  (b) stale managed region (hand-edited) → stale issue reported;
  (c) dangling relative link in a change artifact → dangling issue reported;
  assert read-only (no writes to fixture).

### 2d — CLI wiring

- [x] **2.14** Add `newMapCmd` in `cmd/archon/main.go` — mirrors the existing subcommand
  pattern (`newInitCmd`, `newStatusCmd`). Flags: `--check` (bool), `--backfill` (bool).
  Behavior: no flags → `mapgen.Generate(cwd)`; `--check` → `mapgen.Check(cwd)`, emit
  per-issue report, exit non-zero on any issue; `--backfill` → `mapgen.Backfill(cwd)`
  (Slice 4 fully implements; Slice 2 wires the flag, returns stub error if called).
  Register in `newRootCmd`.
  **Deviation**: `--check` is a compiling stub in this slice (returns a "not yet
  implemented" error) rather than calling `mapgen.Check` — `Check` itself is deferred
  to Slice 2b per the budget guard. `--backfill` behaves as designed (calls the
  `mapgen.Backfill` stub from 2.6).

- [x] **2.15** Integration test for `archon map` in `cmd/archon/main_test.go` — seed a
  `t.TempDir()` with fixture `openspec/`; run `archon map`; assert `map.md` contains
  the expected managed region; run `archon map --check`; assert exit 0; mutate the
  managed region; assert `archon map --check` exits non-zero.
  **Deviation (minimal, per budget guard)**: implemented `TestMapCommand_GeneratesManagedRegion`
  (Generate path: managed region + capability entry present) and
  `TestMapCommand_CheckIsStubbed` (asserts `--check` compiles and returns the stub
  "not yet implemented" error) instead of the full stale/exit-0/exit-non-zero
  `--check` assertions — those require the real `Check` implementation, deferred to
  Slice 2b along with 2.10–2.13.

---

## Slice 3 — Auto-Regen Hook + Link Emission

**Skill/doc edits only. No new Go files; one small addition to `internal/initcmd/init.go`.**
Est: <200 lines. Risk: low.

Verification: manual `archon init` in a tmpdir creates `openspec/map.md` with markers;
peer review of skill SKILL.md diffs confirms the convention pointer is minimal and correct.

- [ ] **3.1** Update `skills/harness-workflow/SKILL.md` — add one instruction block in Step 3
  (after the `state.yaml` temp+rename write): shell out to `archon map`; surface failure
  as a warning to the orchestrator; MUST NOT roll back or block the recorded transition.
  Instruction MUST specify execution order: state.yaml written first, then `archon map`.

- [ ] **3.2** Update `skills/sdd-init/SKILL.md` — add one instruction: when scaffolding
  `openspec/`, call `archon init` (which seeds `map.md`) or note that `createOpenSpecDir`
  is responsible; clarify the skill does not need to create `map.md` directly — the Go
  init step does it.

- [ ] **3.3** Update `internal/initcmd/init.go` `createOpenSpecDir` — after creating
  `openspec/changes/`, also write `openspec/map.md` with a minimal preamble +
  `<!-- MAP:START -->\n<!-- MAP:END -->` if the file does not already exist. Use
  temp+rename write. MUST NOT overwrite if file already exists.

- [ ] **3.4** Test `createOpenSpecDir` — extend `internal/initcmd/init_test.go`: assert
  that after `archon init` in a clean `t.TempDir()`, `openspec/map.md` exists and
  contains both marker strings; assert re-running `archon init` leaves existing
  `map.md` unchanged (idempotent).

- [ ] **3.5** Update `skills/sdd-propose/SKILL.md` — add one-line convention pointer:
  emit `[[capability]]` for any capability referenced by name; emit relative links for
  intra-change artifact navigation. Reference `[[spec-vault]]` for the full rule. Include
  a one-line example in the artifact template showing both forms.

- [ ] **3.6** Update `skills/sdd-spec/SKILL.md` — same one-line convention pointer +
  example as 3.5; relative link depth from `changes/{c}/specs/{cap}/spec.md` to
  `changes/{c}/proposal.md` is `../../proposal.md`.

- [ ] **3.7** Update `skills/sdd-design/SKILL.md` — same convention pointer + example;
  artifact lives at `changes/{c}/design.md` (depth = same dir as proposal).

- [ ] **3.8** Update `skills/sdd-tasks/SKILL.md` — same convention pointer + example;
  artifact lives at `changes/{c}/tasks.md`.

- [ ] **3.9** Update `skills/sdd-verify/SKILL.md` — same convention pointer + example;
  artifact lives at `changes/{c}/verify-report.md`.

---

## Slice 4 — Archive Link Rewrite + Backfill

**Riskiest slice. Lands last. Ships with fixture-move tests. The 20-change backfill
is a separate data-only commit reviewable apart from the code changes.**
Est: ~350 lines (code PR) + backfill diff (data commit). Risk: high.

Verification: `go test ./internal/mapgen/...` green including archive + backfill tests;
`archon map --check` exits 0 after a fixture archive move; `archon map --backfill` in a
tmpdir with the 20 archived changes exits 0 and `--check` passes; second `--backfill`
run produces zero diff.

### 4a — RewriteMove (boundary-edge rewrite)

- [ ] **4.1** Implement `internal/mapgen/archive.go` — `RewriteMove(root, oldRel, newRel string) error`:
  for every `.md` file under the moved folder's new location, call `links.Rewrite` to
  recompute each relative link so it still resolves to the same absolute target after the
  directory shift (one level deeper). Wikilinks are never touched. Writes via temp+rename.

- [ ] **4.2** Fixture-move tests for `archive.go` — `fstest.MapFS` with a change containing
  relative `../../../specs/foo/spec.md` links; simulate move one level deeper; assert:
  (a) rewritten link resolves to the same absolute file; (b) wikilinks in the same file
  are byte-identical after rewrite; (c) `.feature` file content is untouched.

### 4b — Implement Backfill

- [ ] **4.3** Implement `mapgen.Backfill(root string) error` in `internal/mapgen/mapgen.go` —
  walks all `openspec/changes/archive/YYYY-MM-DD-{name}/` dirs; for each, calls
  `RewriteMove` treating the archive path as the current location and the expected
  pre-archive path as `oldRel`; then calls `Generate`. Must be idempotent: a second run
  on already-correct files produces no diff.

- [ ] **4.4** Backfill idempotency test — run `Backfill` twice on a fixture with 2 archived
  changes; capture all file bytes after each run; assert zero diff between run 1 and
  run 2 outputs.

- [ ] **4.5** Wire `archon map --backfill` fully in `cmd/archon/main.go` `newMapCmd` —
  remove the Slice-2 stub; call `mapgen.Backfill(cwd)`; emit per-change progress lines
  to stdout; exit non-zero on any error.

### 4c — sdd-archive skill wiring

- [ ] **4.6** Update `skills/sdd-archive/SKILL.md` — add three steps after the folder-move
  instruction: (1) invoke `archon map` (rewrite + regen as a combined step via
  `mapgen.Generate` which calls `RewriteMove` internally for the moved path); (2) run
  `archon map --check`; (3) if `--check` exits non-zero, surface failure to orchestrator
  and abort — do NOT mark archive complete. Wikilinks require no rewrite (document
  explicitly to avoid confusion).

### 4d — One-shot backfill of this repo's 20 archived changes (data-only commit)

- [ ] **4.7** Run `archon map --backfill` against this repo's `openspec/changes/archive/`
  (the 20 existing archived changes) — execute in a clean working tree; commit the
  resulting `.md` diffs as a standalone data-only commit with message
  `chore(openspec): backfill vault links in 20 archived changes`. This commit MUST be
  separate from any code change commit and must be the last commit of the slice.

- [ ] **4.8** Verify backfill result — run `archon map --check` after the backfill commit;
  assert exit 0 and zero issues reported.

---

## Cross-Slice Constraints

- `.feature` files are never modified by any task in any slice. Go writes only inside
  `<!-- MAP:START/END -->` markers and known relative link patterns.
- All Go writes use temp+rename (atomic, matching existing `internal/` conventions).
- All Go packages use `fs.FS` injection for scan paths to support `fstest.MapFS` in tests.
- Tests are co-located with the code they cover (work-unit-commits discipline).
- Slice 4 code PR and the backfill data commit are separate reviewable units.
- Slices 2b (`--check` split) and 4 (code vs. data) each have a defined split boundary
  to stay under the 400-line per-PR budget.
