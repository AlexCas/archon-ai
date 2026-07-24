# Verification Report — obsidian-vault-specs

Links: [proposal](proposal.md) · [design](design.md) · [tasks](tasks.md) · specs under [specs/](specs/).
Capabilities verified: [[spec-vault]], [[archon-map]], [[openspec-convention]],
[[sdd-archive]], [[sdd-init]], [[sdd-phase-skills]], [[harness-workflow]].

- **Change**: obsidian-vault-specs
- **Branch**: feat/obsidian-vault-specs-s5-checkfix (not switched)
- **Persistence mode**: openspec (interactive · ask-always · 400-line budget · Playwright disabled)
- **Verdict**: **PASS WITH WARNINGS**

## Completeness

| Dimension | Result |
|-----------|--------|
| Tasks checked | 35 / 35 (0 unchecked implementation tasks) |
| Proposal / design / specs / tasks present | Yes (full artifact set) |
| Playwright | Disabled per preflight — skipped, no `@web` scenarios expected |
| Security | `security.enabled` not set for this change — skipped |

## Build / Test Evidence

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `go test ./...` | exit 0 — all packages ok (mapgen, cmd/archon, initcmd, tui, etc.) |

mapgen has 27 test functions across scan/render/region/links/check/archive/backfill
plus 3 CLI integration tests (`TestMapCommand_GeneratesManagedRegion`,
`TestMapCommand_Check`, `TestMapCommand_Backfill`) and `TestRun_SeedsMapMD` in initcmd.
All green.

## CLI Behavior (built `/tmp/archon`, run from repo root)

| Check | Observed | Result |
|-------|----------|--------|
| `archon map` regenerates managed region | exit 0, "regenerated" | PASS |
| Idempotent (run twice, 2nd run zero diff) | byte-identical | PASS |
| `--check` passes when map.md is fresh | exit 0, "up to date" | PASS |
| `--check` fails on hand-mutated managed region | exit 1, reports `stale: managed region does not match a fresh regeneration` | PASS |
| `--check` fails on dangling relative link | exit 1, reports `dangling: link "…" does not resolve` + source file | PASS |
| `--check` ignores link syntax in code fences / inline code / double-backtick spans (slice-5 fix) | exit 0 — fenced + inline examples not flagged; a real link in the same file still flagged | PASS |
| `--backfill` idempotent (run twice, 2nd run zero diff) | only map.md touched; identical output | PASS |
| `--backfill` touches no `.feature` file | 47 `.feature` sha256 unchanged before/after | PASS |

Note on `--check` against the *committed* tree: it currently exits 1 because the
committed `map.md` was last regenerated at phase `apply`, while live `state.yaml`
now reads `verify`. This is `--check` working correctly (the managed region is
genuinely stale vs. live state); it is a natural consequence of verify running as a
later phase, not a defect. A fresh `archon map` (which the transition hook fires
automatically) reconciles it.

## Managed-Marker & Prose Safety

- Generation writes only between `<!-- MAP:START -->` / `<!-- MAP:END -->`
  (`Splice` in region.go; `TestSplice_PreservesProseOutsideMarkers`,
  `TestGenerate_PreservesProseOutsideMarkers`). Authored preamble in
  `openspec/map.md` is preserved byte-for-byte across regen. PASS
- Wikilinks are never rewritten on archive (`Rewrite` only edits
  `[text](target)` spans; `TestRewrite_LeavesWikilinksByteIdentical`). PASS
- `.feature` files never modified: `RewriteMove`/`Backfill` skip non-`.md` files;
  confirmed by sha256 diff over all 47 `.feature` files. PASS

## Spec Coverage Matrix

Legend: MET / PARTIAL / UNMET. Evidence = code path, test name, or CLI observation.

### [[spec-vault]] (4 requirements)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Vault Root Shape | MET | `skills/_shared/spec-vault.md` documents map.md + specs/ + changes/archive/; init seeds the tree (`createOpenSpecDir`) |
| Hybrid Link Convention | MET | spec-vault.md + openspec-convention.md specify `[[cap]]` vs relative; change artifacts use both |
| Managed-Marker Policy | MET | region.go single-pair `Splice`, `ErrNestedRegion`; `TestSplice_NestedRegion`; prose-preservation tests |
| Feature Files Stay Put | MET | RewriteMove skips non-`.md`; sha256 diff shows 0 `.feature` changes |

### [[archon-map]] (5 requirements)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| map.md Generation (index + backlinks, deterministic) | **PARTIAL** | render.go emits all sections deterministically; `TestRender_Golden`, `TestGenerate_Idempotent`, CLI idempotency PASS. **BUT** the Backlinks section is polluted by non-capability tokens (see WARNING W1) |
| Managed-Region Integrity | MET | Splice + create-when-absent (`defaultMap`); `TestMapCommand_GeneratesManagedRegion`, prose-preservation tests |
| --check Mode | MET | check.go stale + dangling detection; CLI observations above; `TestCheck_*` |
| Archive Link Rewriting | MET | archive.go `RewriteMove` shifts `../` depth, leaves wikilinks; `TestRewriteMove_FixtureArchiveShift` |
| --backfill Mode (idempotent, forward-only default) | MET | mapgen.go `Backfill`; `hasDanglingRelLink` idempotency guard; `TestBackfill_Idempotent`; CLI double-run zero diff |

### [[openspec-convention]] (3 requirements)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Vault Root Shape Documentation | MET | openspec-convention.md lines ~7,30 document map.md entry node + `[[spec-vault]]` pointer |
| Hybrid Link Convention in Convention Doc | MET | openspec-convention.md line ~37 states the unambiguous rule |
| Artifact File Paths Table Updated | MET | table rows for `sdd-init` creating map.md and `archon map` regenerating it |

### [[sdd-archive]] (2 requirements)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Archive-Triggered Link Rewrite | MET | SKILL.md Step 3b invokes `archon map --backfill` then regen (see risk R1 on backfill-vs-targeted) |
| Post-Archive --check Guard | MET | SKILL.md Step 3b.3 + Step 4 checklist: STOP on non-zero `--check`, do not mark complete |

### [[sdd-init]] (1 requirement)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Seed map.md on Init | MET | initcmd/init.go `seedMapMD` writes preamble + empty managed region, never overwrites; `TestRun_SeedsMapMD` |

### [[sdd-phase-skills]] (2 requirements)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Capability Wikilink Emission | MET | link-convention comment block in sdd-propose/spec/design/tasks/verify; change's own proposal/design/tasks use `[[…]]` (3/14/9 tokens) |
| Intra-Change Relative Navigation | MET | same convention blocks specify relative links by depth; this report and tasks.md use them |

### [[harness-workflow]] (1 requirement)
| Requirement | Status | Evidence |
|-------------|--------|----------|
| Automatic map.md Regen on Phase Transition (best-effort, non-blocking) | MET | SKILL.md Step 3 instructs shell-out to `archon map` after state.yaml write, regen failure is a warning that MUST NOT roll back |

**Tally**: 18 requirements — **17 MET, 1 PARTIAL, 0 UNMET.**

## Constraints

| Constraint | Result |
|------------|--------|
| Go never edits authored prose outside markers | CONFIRMED (Splice-only writes; prose-preservation tests + live regen) |
| No `.feature` content changed anywhere | CONFIRMED (sha256 over 47 files unchanged through regen + double backfill) |
| Backfill produced only map.md changes in this repo | CONFIRMED (`git status` after double backfill shows only map.md; and that diff is only the live-state phase delta) |

## Issues

### WARNING — W1: Backlinks section pollution (spec-relevant, cosmetic)
The rendered `## Backlinks` in `openspec/map.md` contains three tokens that are NOT
capabilities: `[[spec.md]]`, `[[specs/harness-workflow/spec]]`, and `[[wikilinks]]`.

- **Root cause**: `internal/mapgen/scan.go` `scanEdges` extracts EVERY `[[…]]`
  token as a capability edge. It does not (a) mask code fences / inline-code spans —
  the very fix slice 5 applied to relative links in `FindRelLinks` was never applied
  to wikilink edge extraction — nor (b) validate each token against the real
  capability set (`specs/{name}/` folders from `scanCapabilities`).
- **Where from**: `exploration.md` prose examples — `` ### A. Obsidian `[[wikilinks]]` ``,
  `` `[[spec.md]]` `` and `` `[[specs/harness-workflow/spec|harness-workflow]]` `` — all
  illustrative wikilink syntax inside inline-code spans/headings, not real references.
- **Severity**: WARNING, not CRITICAL. It is deterministic and idempotent, so
  `--check` still passes and no link is broken; build/tests are green. But the
  archon-map "map.md Generation" requirement says the backlink section associates
  "each capability with changes that reference it" — emitting non-capability tokens
  violates the spirit of that requirement and adds navigation noise to the entry node.
- **Suggested fix (do not apply in verify)**: in `scanEdges`, reuse
  `maskCodeRegions` before matching wikilinks, and/or filter edges to
  `ToCapability ∈ capabilitySet` in `Render`/`Backlinks`.

### SUGGESTION — R1 (risk item a): archive uses `--backfill` (loops ALL archived), not a targeted rewrite
`sdd-archive` Step 3b calls `archon map --backfill`, which iterates every archived
change on each archive. Assessment: **correct and safe, but O(all-archived) per
archive.** The `hasDanglingRelLink` guard in `RewriteMove` makes already-correct
changes no-op, so idempotency holds and no already-archived file is disturbed (verified:
double backfill = zero diff). The freshly-moved change is rewritten identically to a
targeted call because `Backfill` treats `changes/{name}/`→`changes/archive/{dir}/` for
every entry. Acceptable; the only cost is redundant work at scale. A targeted
single-change rewrite would be a future optimization, not a correctness fix.

### SUGGESTION — R2 (risk item b): Resolve/Check operate on the real filesystem, not fs.FS
`links.go Resolve` and `check.go Check` use `os.Stat` / `filepath.WalkDir` on the real
FS, while `Scan` uses `os.DirFS`. Assessment: **acceptable and consistent with the
design table** ("`os.DirFS(openspec)` for scan; real FS for writes"). `--check` is
exercised end-to-end by the CLI integration test in a `t.TempDir()`, so it is testable;
the split just means `Resolve`/`Check` can't be unit-tested with a pure `fstest.MapFS`.
Low risk; no change required.

### INFO — R3 (risk item c): backlinks are capability→change only
Confirmed intentional and matches proposal/design open questions (change→artifact depth
explicitly out of scope). Not a gap.

## Design Coherence

Implementation matches design.md: package layout (scan/graph/render/region/links/check/
archive/mapgen) is exactly as specified; managed-region full-replace, A→Z / date-desc
ordering, `os.DirFS` scan + real-FS writes, CLI `map` / `--check` / `--backfill`
surface, init seed, and the harness-workflow best-effort hook are all present. The only
divergence from the design's intent is W1 (edge extraction does not validate/mask), which
the design's Link Model section implies ("mapgen only reads [wikilinks] for edge
extraction") but does not explicitly require validation for — so W1 is a latent gap
rather than a stated-design deviation.

## Final Verdict

**PASS WITH WARNINGS.** Build, vet, and the full test suite are green; all 8 CLI
behaviors (idempotency, `--check` pass/stale/dangling, slice-5 code-fence exemption,
backfill idempotency, `.feature` safety) behave as specified; 17/18 spec requirements
are MET with concrete evidence. The single PARTIAL (archon-map map.md Generation) is due
to W1 — non-capability tokens leaking into the Backlinks section from prose examples in
`exploration.md`. W1 is cosmetic (deterministic, `--check`-clean, no broken links) but
does technically violate the backlink requirement, so it is surfaced for the orchestrator/
user to decide whether to fix before archive. Risk items (a) and (b) were assessed and
found acceptable; (c) is intentional scope.
