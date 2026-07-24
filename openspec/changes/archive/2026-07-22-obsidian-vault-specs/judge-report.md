# Judge Report — obsidian-vault-specs

- **Change**: obsidian-vault-specs
- **Branch**: feat/obsidian-vault-specs-s6-w1fix
- **Judge phase**: Round 1 (post-verify, post-W1-fix)
- **Verdict**: **FAIL**
- **Date**: 2026-07-22

Links: [proposal](proposal.md) · [design](design.md) · [tasks](tasks.md) · [verify-report](verify-report.md)

---

## Build / Test Baseline

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `go test ./...` | exit 0 — all packages ok |

Mutation testing: not configured for this repo. Relied on dual adversarial review + probe tests.

---

## Judgment Method

Two blind judges (A and B) reviewed independently with identical adversarial criteria covering all
7 areas in the judge charter. Findings were cross-classified (confirmed / suspect / theoretical / INFO).
All CONFIRMED WARNING (real) findings were independently verified with probe tests executed
against the live codebase before writing this report.

---

## Verdict Table

| # | Finding | Judge A | Judge B | Classification | Blocking |
|---|---------|---------|---------|----------------|----------|
| F1 | `Rewrite()` rewrites links inside fenced code blocks (no `maskCodeRegions`) | WARNING (real) | WARNING (real) | **CONFIRMED** | Yes |
| F2 | `Splice()` orphan single marker corrupts file irreversibly | WARNING (real) | WARNING (theoretical) | **CONFIRMED** | Yes |
| F3 | `--check` silently skips stale-check when `map.md` has no markers | WARNING (real) | (mechanism noted, not classified as WARNING) | **CONFIRMED** | Yes |
| F4 | `!capNames[cap]` filter in `scanEdges` has no mutation-catching test | (not flagged) | WARNING (real) | **SUSPECT** | No (test gap, not a runtime defect) |
| F5 | `hasDanglingRelLink` idempotency guard: path-collision false-negative | WARNING (theoretical) | WARNING (theoretical) | Theoretical | No |
| F6 | `renderBacklinks` sort removal not deterministically caught | INFO | WARNING (theoretical) | Theoretical | No |

---

## Confirmed Blocking Defects

### F1 — `Rewrite()` mutates links inside fenced code blocks

**File**: `internal/mapgen/links.go:196`, `internal/mapgen/archive.go:52`

`Rewrite()` calls `relLinkRe.ReplaceAllStringFunc` directly on the raw markdown without first
calling `maskCodeRegions`. When `RewriteMove` processes a file that has a real dangling relative
link (triggering `hasDanglingRelLink → true`) **and** also has relative link syntax inside a
fenced code block (e.g. a documentation example), `Rewrite()` depth-shifts the code-block link
too. The result is mathematically valid (resolves to the same absolute target), but the authored
example text in the code block now shows a different relative path than the author wrote. Authored
documentation content is silently mutated.

**Probe evidence** (live execution):
```
Input:   ```\nExample: [link](../../specs/foo/spec.md)\n```
Output:  ```\nExample: [link](../../../specs/foo/spec.md)\n```
```

**Fix**: apply `maskCodeRegions` before `relLinkRe.ReplaceAllStringFunc` in `Rewrite()`, preserving
the same masking approach used by `FindRelLinks`.

---

### F2 — `Splice()` orphan single marker causes irreversible corruption

**File**: `internal/mapgen/region.go:28-33`

When `map.md` contains only one of the two markers (e.g., only `<!-- MAP:START -->`, no
`<!-- MAP:END -->`), `Splice` takes the "absent markers" branch (`endIdx == -1`) and appends a
fresh `MAP:START…MAP:END` block. The resulting file has `MAP:START` count = 2, `MAP:END` count = 1.
Every subsequent `Generate()` and `Check()` call immediately returns `ErrNestedRegion`, permanently
breaking the vault's map generation until the file is manually repaired.

Precondition: a user hand-edits `map.md` and accidentally removes exactly one marker. This is
normal intended use of the file (the preamble above the markers is explicitly authored prose).

**Probe evidence** (live execution):
```
Splice(orphan START only): err=<nil>
result contains MAP:START count=2, MAP:END count=1
```

**Fix**: detect the partial-marker case separately (exactly one of the two markers present)
and return a specific error (e.g., `ErrPartialMarker`) rather than silently appending.

---

### F3 — `--check` silently skips stale-check when `map.md` has no markers

**File**: `internal/mapgen/check.go:74-87`, `internal/mapgen/region.go:102-116`

`extractManagedBody` returns `ok=false` when `map.md` has zero `MAP:START`/`MAP:END` markers.
The stale-check branch is silently skipped — `Check` reports no issues. A `map.md` that was
created manually without a managed region (or whose markers were fully stripped) passes
`archon map --check` with exit 0. This is a false negative on the staleness check.

**Probe evidence** (live execution):
```
extractManagedBody("# Vault Map\n\nHand written.\n") → ok=false
Stale check branch: skipped with no issue reported.
```

**Fix**: when `extractManagedBody` returns `ok=false` and the file exists, distinguish the
"no markers present" case from the "nested markers" case and emit an `IssueMissingRegion` issue
(or equivalent) rather than silently passing.

---

## Suspect Findings (not blocking, no auto-fix)

### F4 — `!capNames[cap]` capability filter in `scanEdges` has no effective test

**File**: `internal/mapgen/scan_test.go:76-99`

`TestScan_WikilinkMaskingAndCapabilityFilter` places non-capability wikilinks (`[[not-a-cap]]`)
only inside masked regions (inline code span + fenced code block). After `maskCodeRegions`, those
tokens are blanked before `wikilinkRe` runs. Removing the `!capNames[cap]` guard at `scan.go:216`
produces no observable change to the test result — the mutation survives the full test suite.

This is a test gap, not a runtime defect: the filter itself is present and correct. A future
fixture placing a non-capability wikilink in plain prose would close the gap.

---

## Confirmed Non-Blocking / Theoretical

- **F5** (hasDanglingRelLink false-negative on coincidental path collision): requires an existing
  file at exactly the wrong relative path inside `openspec/`. Practically impossible in a
  controlled vault. Theoretical only.
- **F6** (`renderBacklinks` sort removal not deterministically caught): `sort.Strings(caps)` is
  present and correct; the mutation-catch weakness is a test quality observation. The golden test
  provides a backstop under most runtime orderings.

---

## Hard Constraints Status

| Constraint | Status |
|------------|--------|
| Go never writes outside MAP:START/END | SATISFIED — `Splice` preserves all bytes before/after markers; `Generate` writes only via `tmp+rename` |
| `.feature` content never modified | SATISFIED — `RewriteMove` skips non-`.md`; `Scan`/`Check` skip non-`.md` |
| harness-workflow auto-regen hook is warning-only | SATISFIED — `harness-workflow/SKILL.md` Step 3: regen failure MUST NOT roll back state or block transition |
| sdd-archive aborts on `--check` failure | SATISFIED — `sdd-archive/SKILL.md` Step 3b: STOP on non-zero `--check`; do not mark complete |

---

## Skill Resolution

Matched from registry: `go-testing`, `harness-workflow`, `sdd-archive`. Injected into both judge
prompts. No registry file was missing — registry dir exists but has no indexed `.md`; judges
proceeded with the explicit skill paths provided.

---

## Final Verdict

**JUDGMENT: FAIL**

Three confirmed blocking defects (F1, F2, F3). All are independently verified by probe tests
against the live codebase. The defects are in `links.go:Rewrite`, `region.go:Splice`, and
`check.go:Check`. No CRITICAL issues (no data loss, no invariant corruption of the vault's
authoritative graph). All three are fixable with targeted, low-risk changes.

Recommended re-apply scope:
1. `links.go`: add `maskCodeRegions` call in `Rewrite()` before `ReplaceAllStringFunc`.
2. `region.go`: detect partial-marker case and return `ErrPartialMarker` instead of appending.
3. `check.go`: emit `IssueMissingRegion` when `map.md` exists but has no markers.
4. Add a plain-prose non-capability wikilink fixture to `TestScan_WikilinkMaskingAndCapabilityFilter` (closes F4 test gap).

---

---

# Judge Report — Round 2

- **Change**: obsidian-vault-specs
- **Branch**: feat/obsidian-vault-specs-s7-judgefix
- **Fix commit**: `0182075`
- **Judge phase**: Round 2 (re-judge after re-apply of F1/F2/F3)
- **Verdict**: **PASS**
- **Date**: 2026-07-22

---

## Build / Test Baseline

| Command | Result |
|---------|--------|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `go test ./...` | exit 0 — all 12 packages pass |

Mutation testing: not configured for this repo. Relied on dual adversarial review + probe tests (Judge A: 31 probe cases; Judge B: 16 probe cases + caller audit).

---

## Judgment Method

Two blind judges (A and B) reviewed independently with adversarial criteria focused on F1/F2/F3
re-probing and fresh regression discovery. Judges used independent probe sets and angles. Findings
were cross-classified per the judgment-day skill contract.

---

## Verdict Table — Round 2

| # | Finding | Judge A | Judge B | Classification | Blocking |
|---|---------|---------|---------|----------------|----------|
| F1 fix | `Rewrite()` span-skipping via masked-copy comparison | CONFIRMED FIXED | CONFIRMED FIXED | Fixed | — |
| F2 fix | `Splice()` `ErrPartialMarker` + no caller breakage | CONFIRMED FIXED | CONFIRMED FIXED | Fixed | — |
| F3 fix | `Check()` `IssueMissingRegion` on markerless/partial `map.md` | CONFIRMED FIXED | CONFIRMED FIXED | Fixed | — |
| R2-S1 | `ErrNestedRegion` on `map.md` silently swallowed by `Check()` | INFO | SUSPECT (non-blocking) | Pre-existing intentional design; explicitly documented in code | No |
| R2-S2 | `maskInlineCode` does not mask multi-line inline code | INFO | INFO | Correct per CommonMark; documented | No |

No new blocking defects found. No regressions introduced by commit `0182075`.

---

## F1 Fix Confirmation — `Rewrite()` code-region span skipping

**Mechanism verified** (`internal/mapgen/links.go:197–241`): `Rewrite()` calls `maskCodeRegions(md)`
to produce a masked copy, then for each regex match position in the *original* `md` compares
`masked[start:end] != md[start:end]`. If the span was blanked by masking, it falls within a code
region and is emitted verbatim. The same `maskCodeRegions` function is used by `FindRelLinks` and
`Rewrite`, guaranteeing consistent behavior — a link found by `FindRelLinks` is treated identically
in `Rewrite`.

**Probe results** (combined 25 cases across both judges):
- Prose link adjacent to fenced code → depth-shifted (correct).
- Link inside fenced code (```` ``` ````, `~~~go`) → byte-identical (correct).
- Link inside inline-code span → byte-identical (correct).
- Link immediately after closing backtick → depth-shifted, no off-by-one (correct).
- Empty fence immediately before link → link shifted (correct).
- Link with backtick in text field, not code-span opener → shifted (correct).
- Unclosed backtick (CommonMark: not a code span) → link inside accessible, shifted (correct).
- Link at byte-offset 0 → shifted (correct).
- Multiple consecutive code spans before a real link → only prose link shifted (correct).

**Off-by-one assessment**: `maskInlineCode` sets `i = closeEnd` after blanking; the byte at `closeEnd`
is the first byte after the closing backtick run. A link starting at that position is NOT masked —
correct. Adjacent-to-code-span probe passes.

**F1: CONFIRMED FIXED.**

---

## F2 Fix Confirmation — `Splice()` `ErrPartialMarker`

**Mechanism verified** (`internal/mapgen/region.go:29–71`): `markerSpan` counts `startCount` and
`endCount` separately. If they differ (exactly one marker present), `ErrPartialMarker` is returned
before any write path is reached.

**Caller audit**: sole production caller is `Generate()` at `mapgen.go:48–51`. It wraps all errors
with `fmt.Errorf("splice map.md: %w", err)` and returns — it does NOT branch on specific sentinel
values, so adding `ErrPartialMarker` breaks no callers. `ErrPartialMarker` is declared via
`errors.New` and is natively `errors.Is`-compatible.

**Probe results** (7 cases):
- START-only → `ErrPartialMarker`, no output written.
- END-only → `ErrPartialMarker`, no output written.
- Both absent → appends normally; output contains exactly one START and one END.
- Both present → normal splice; prose preamble and epilogue preserved.
- Nested (two full pairs) → `ErrNestedRegion`.
- Reversed markers (END before START) → `ErrNestedRegion`.
- `errors.Is(ErrPartialMarker)` and `errors.Is(ErrNestedRegion)` are distinct.

Atomicity: `Generate()` only reaches `os.WriteFile(tmp, ...)` after `Splice()` returns `nil`; on
error, `map.md` is never overwritten.

**F2: CONFIRMED FIXED.**

---

## F3 Fix Confirmation — `Check()` `IssueMissingRegion` for `map.md`

**Mechanism verified** (`internal/mapgen/check.go:87–119`): when `p == mapPath` (exact absolute
path `filepath.Join(openspecDir, "map.md")`), a switch on `extractErr` handles:
- `errNoManagedRegion` → `IssueMissingRegion` with "no MAP:START/MAP:END" detail.
- `ErrPartialMarker` → `IssueMissingRegion` with "partial region" detail.
- `ErrNestedRegion` → intentionally swallowed (pre-existing design, code comment documents this).
- `nil` → stale check (existing behavior).

**Scoping**: `p == mapPath` is an absolute-path comparison. Files named `vault-map.md`,
`roadmap.md`, or `map.md` in any subdirectory do NOT match.

**Probe results** (8 cases across both judges):
- `map.md` with no markers → `IssueMissingRegion` reported.
- `map.md` with START-only (partial) → `IssueMissingRegion` reported, no crash.
- Zero-byte `map.md` → `IssueMissingRegion` reported.
- `map.md` with swapped markers (END before START) → `ErrNestedRegion`, intentionally swallowed.
- `map.md` with dangling link AND no markers → both `IssueDangling` and `IssueMissingRegion` reported.
- `vault-map.md` with no markers → NOT flagged.
- `openspec/changes/foo/map.md` with no markers → NOT flagged.
- Ordinary `proposal.md` with no markers → NOT flagged.

**F3: CONFIRMED FIXED.**

---

## Non-Blocking / INFO Findings

### R2-S1 — `ErrNestedRegion` silently swallowed in `Check()` for `map.md`

**File**: `internal/mapgen/check.go:114–118`

A `map.md` with two full `MAP:START/MAP:END` pairs (doubly-keyed) passes `archon map --check` with
zero issues. In contrast, `archon map` (Generate) returns `ErrNestedRegion` and fails loudly on
the same file. This is an observable inconsistency between `--check` and `map` behavior. However,
the code comment explicitly documents this as matching prior behavior, no data-corruption path
exists (Generate refuses to write), and this is a pre-existing design choice not introduced by the
current fix.

Classification: SUSPECT (non-blocking). No fix required for this round.

### R2-S2 — `maskInlineCode` does not mask multi-line inline code

**File**: `internal/mapgen/links.go:81–123`

`maskInlineCode` operates line-by-line; a backtick span crossing a newline is not masked. This is
correct per CommonMark (inline code is single-line). Documented. No fix required.

---

## Hard Constraints Status — Round 2

| Constraint | Status |
|------------|--------|
| Go never writes outside MAP:START/END | SATISFIED — `Splice` preserves all bytes before/after markers; `Generate` writes only via `tmp+rename`; `Splice` returns error before write on partial/nested/absent markers |
| `.feature` content never modified | SATISFIED — `RewriteMove` filters on `.md` suffix; `Scan`/`Check`/`Generate` only touch `.md` |
| harness-workflow auto-regen hook is warning-only | SATISFIED — `harness-workflow/SKILL.md`: regen failure MUST NOT roll back state or block transition |
| sdd-archive aborts on `--check` failure | SATISFIED — `sdd-archive/SKILL.md`: STOP on non-zero `--check`; `cmd/archon/main.go:450–463` returns non-nil error on issues |

---

## Skill Resolution

Matched from registry: `go-testing`, `harness-workflow`, `sdd-archive`. Injected into both judge
prompts.

---

## Final Verdict — Round 2

**JUDGMENT: APPROVED**

All three Round-1 confirmed blocking defects (F1, F2, F3) are correctly and completely fixed.
Fixes are well-scoped, consistent, and introduce no regressions. Combined 41+ adversarial probe
cases across both judges — all pass. No new blocking or non-blocking defects found. All hard
constraints hold. The two non-blocking findings (R2-S1, R2-S2) are pre-existing intentional design
decisions explicitly documented in the code.
