# Design: Obsidian-Style Vault for SDD Specs

Implements [[spec-vault]] and [[archon-map]]; touches [[openspec-convention]],
[[sdd-init]], [[sdd-archive]], [[harness-workflow]], and the phase skills
([[sdd-propose]] / [[sdd-spec]] / [[sdd-design]] / [[sdd-tasks]] / [[sdd-verify]]).
See [proposal](proposal.md) and [exploration](exploration.md).

## Technical Approach

A deterministic Go module (`internal/mapgen`) walks `openspec/{specs,changes}`, builds
an in-memory graph, and renders it into the managed region of `openspec/map.md`. Go only
writes markdown inside `<!-- MAP:START/END -->` markers or rewrites recognized relative
link patterns; it never edits authored prose or `.feature` files. `state.yaml` and phase
transitions remain LLM-owned (skills write them via temp+rename); Go is invoked as a
subprocess after those writes. All decisions in [proposal](proposal.md) are LOCKED.

## Architecture Decisions

| Decision | Choice | Why / rejected |
|----------|--------|----------------|
| Capability identity | Folder name of `specs/{cap}/` (never filename) | Files are literally `spec.md`; folder is the unique, stable id |
| Graph model | `Graph{Capabilities, Changes, Edges}` built in memory per run | Rebuild-from-scan is simplest deterministic source of truth; no incremental cache to invalidate |
| Managed region | Single `<!-- MAP:START/END -->` block, full replace | Preserves authored prose byte-for-byte; idempotent; no diff-merge logic |
| Ordering | caps A→Z, active by name, archive by date desc then name | Total order ⇒ regen twice = zero diff |
| Link rewrite scope | Regex over known relative-link patterns only | Never touch prose; wikilinks resolve by name, need no rewrite |
| Regen trigger | `archon map` CLI + skill-invoked subprocess after `state.yaml` write | State stays LLM-owned; Go stays a pure function of the tree |
| Dangling-link fs | `os.DirFS(openspec)` for scan; real FS for writes | Testable with `fstest.MapFS`, matches existing `internal/` pattern |

## Go Package Layout (`internal/mapgen`)

```
internal/mapgen/
  scan.go      Scan(fsys fs.FS) (*Graph, error) — walk specs/ + changes/
  graph.go     types: Capability{Name,Purpose}; Change{Name,Phase,Status,Archived,Date}
               Edge{FromChange,ToCapability}; Graph aggregates + Backlinks() map[cap][]change
  render.go    Render(g *Graph) string — deterministic managed-region body
  region.go    Splice(existing, body string) (string, error) — replace between markers,
               create/append markers if absent; ErrNestedRegion on >1 pair
  links.go     FindRelLinks(md string) []Link; Rewrite(md, oldDir, newDir string) string;
               Resolve(srcPath, link string) (target string, ok bool)
  check.go     Check(fsys) (report []Issue, err) — stale regions + dangling rel links
  archive.go   RewriteMove(root, oldRel, newRel string) error — boundary-edge rewrite
  mapgen.go    Generate(root string) error; Backfill(root string) error (orchestration)
```

`Scan` derives capability identity from `path.Base(dir)` under `specs/`; reads the
one-line purpose from the first non-heading paragraph or `## Purpose` of each `spec.md`.
Change nodes read `phase`/`status` from `state.yaml`; archived changes are those under
`changes/archive/`, date parsed from the `YYYY-MM-DD-` folder prefix. Edges come from
scanning each change's artifacts for `[[capability]]` wikilink tokens.

## map.md Format (rendered example)

Authored preamble lives outside the markers and is never touched.

```markdown
# openspec Map

<!-- MAP:START -->
## Capabilities
- [[archon-map]] — deterministic map.md generation + link rewrite
- [[harness-workflow]] — SDD phase state machine gate

## Active Changes
| Change | Phase | Status |
|--------|-------|--------|
| [obsidian-vault-specs](changes/obsidian-vault-specs/proposal.md) | design | in_progress |

## Archive
### 2026-07-22
- [ai-orchestration-harness](changes/archive/2026-07-22-ai-orchestration-harness/proposal.md)

## Backlinks
- [[archon-map]] ← obsidian-vault-specs
- [[harness-workflow]] ← obsidian-vault-specs
<!-- MAP:END -->
```

Determinism: every list is fully sorted; the managed body is a pure function of the
graph, so `Generate` → `Generate` yields byte-identical output (idempotency test).

## Link Model

- **Wikilinks** `[[cap]]` are emitted verbatim by skills for capability identity;
  mapgen only reads them (edge extraction) and never rewrites them.
- **Relative intra-change links** (`[design](design.md)`, `[proposal](../../proposal.md)`)
  are computed by skills from the artifact's own nesting depth.
- `map.md` → change and change → `specs/` are the only relative edges that cross the
  archive boundary; mapgen emits/rewrites these.
- `--check` walks every `.md` under `openspec/`, extracts relative links, `Resolve`s each
  against the source file's dir, and flags misses; it also re-renders each managed region
  and diffs it against on-disk content to flag stale regions. Read-only; exits non-zero on
  any issue with a per-issue report (file + link/region).

## Archive Link Rewriting

On `changes/{c}/` → `changes/archive/YYYY-MM-DD-{c}/` the change descends one level.

1. Wikilink identity edges: unchanged (resolve by name).
2. `map.md` edges pointing at `changes/{c}/...`: rewritten to the new archive path
   (`region.go` re-renders from the post-move graph, so this falls out of a plain regen).
3. Relative links **inside** moved files pointing out to `specs/`: each gains one `../`
   segment (`../../../specs/x` → `../../../../specs/x`). `Rewrite(md, oldRel, newRel)`
   recomputes each relative target so it still resolves to the same absolute file, editing
   only the matched link spans.
4. `.feature` files move with `spec.md` untouched.

Coordination: [[sdd-archive]] performs the folder move, then calls `archon map` (rewrite +
regen), then `archon map --check`; a non-zero check aborts marking the archive complete.

## CLI + Integration Surface

`archon map` added to `newRootCmd` in `cmd/archon/main.go` via `newMapCmd`, mirroring the
existing subcommand pattern (resolves `openspec/` under cwd):

| Invocation | Behavior |
|------------|----------|
| `archon map` | Regenerate `map.md` managed region (forward-only; ignores archive rewrite) |
| `archon map --check` | Read-only; non-zero on stale region or dangling link |
| `archon map --backfill` | One-shot rewrite of all `changes/archive/*` boundary edges + regen; idempotent |

**Init seed**: `internal/initcmd/init.go` `createOpenSpecDir` also writes
`openspec/map.md` (preamble + empty managed region) when absent; never overwrites.

**Auto-regen hook**: added as an instruction in [[harness-workflow]] Step 3 — after the
`state.yaml` temp+rename write, shell out to `archon map`. Failure is surfaced as a
warning and MUST NOT roll back or block the recorded transition (state is LLM-owned; Go
regen is best-effort). No new Go transition code.

## Phase-Skill Link Emission

A single rule added to [[openspec-convention]] and inherited by each phase skill: emit
`[[capability]]` for any capability reference (canonical = `specs/{cap}/` folder name) and
relative links for intra-change navigation, computed from the artifact's own depth. Skills
do NOT validate or repair links — integrity is Go's job (`--check`). This keeps skill
edits to a one-line convention pointer plus example, not a maintenance burden.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | scan, render, region splice, link resolve/rewrite | `fstest.MapFS` fixture vault (2 caps, 1 active, 1 archived) |
| Unit | idempotency | `Generate` twice ⇒ byte-identical (assert equal) |
| Unit | `--check` failures | stale region + dangling link fixtures ⇒ non-zero + report |
| Unit | archive move | fixture move one level deeper ⇒ rel links still resolve; wikilinks unchanged; `.feature` untouched |
| Unit | backfill idempotency | run twice ⇒ zero diff |
| Golden | rendered `map.md` body | `gotest.tools/golden` on the fixture graph |
| Integration | `archon map` in `t.TempDir()` | seed → regen → `--check` passes |

Follow existing `internal/` conventions: inject FS/roots, table-driven tests, temp+rename
writes.

## Slice Alignment (chained PRs, 400-line budget)

| Slice | Content | Est. | Risk |
|-------|---------|------|------|
| 1 | `spec-vault.md` + `openspec-convention.md` edits; seed repo `map.md` skeleton (no Go) | <150 | low |
| 2 | `internal/mapgen` (scan/graph/render/region/check) + `archon map`/`--check` + tests + golden | ~350 | **watch** |
| 3 | init seed (Go, small); harness-workflow auto-regen hook; phase-skill link-emission convention | <200 | low |
| 4 | `links.go` rewrite + `archive.go` + `--backfill` + [[sdd-archive]] wiring + fixture-move tests; run backfill on 20 archived changes | ~350 (+ backfill diff) | high; lands last |

Slice 2 is the budget risk: if scan+render+check+tests approach 400, split `--check`
(+ its tests) into a 2b sub-slice. The 20-change backfill in slice 4 is a data-only
commit reviewable separately from the code.

## Migration / Rollout

Additive. Rollback = delete `map.md`, remove `internal/mapgen` + `archon map`, revert the
hook and skill convention edits; managed-region removal leaves authored prose intact.
Backfill is a discrete git-revertable commit.

## Open Questions

- [ ] Backlink granularity is capability→change only (change→artifact out of scope) —
      confirmed by proposal; no deeper depth planned.
- [ ] `--check` is dev-loop-only this change (no CI wiring) — confirmed; a future change
      may add the CI hook.
