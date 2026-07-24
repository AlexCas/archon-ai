# Spec Vault Convention (shared across all SDD skills)

Single source of truth for the Obsidian-style vault layout under `openspec/`: the
root shape, the hybrid link convention, the managed-marker policy, and the rule
that `.feature` files stay put. Every phase skill inherits this convention rather
than restating it.

## Vault Root Shape

`openspec/` has a fixed root shape:

```
openspec/
├── map.md          <- Entry node: generated overview + backlinks (see below)
├── specs/           <- Capability source-of-truth tree, {capability}/spec.md
└── changes/         <- Active changes, plus changes/archive/ for completed ones
```

- `map.md` is the entry node for the vault — the first file a human or agent opens
  to see what capabilities exist and what changes touch them.
- `specs/{capability}/` is the capability source-of-truth tree; the folder name is
  the capability's stable identity (never the filename — files are literally
  `spec.md`).
- `changes/{change-name}/` holds active changes; `changes/archive/YYYY-MM-DD-{name}/`
  holds completed ones. See `openspec-convention.md` for the full change-folder
  layout.

No other top-level file MAY be introduced under `openspec/` without updating this
module first.

## Hybrid Link Convention

Two link forms are permitted, each tied to exactly one reference category:

| Reference category | Form | Example |
|---------------------|------|---------|
| Capability identity (change → capability it touches; `map.md` → capability) | `[[capability]]` wikilink | `Implements [[archon-map]]` |
| Intra-change navigation (proposal ↔ spec ↔ design ↔ tasks within the same change folder) | Relative markdown link | `[design](design.md)`, `[proposal](../../proposal.md)` |

Rule of thumb: if the target is "a capability" (identity, resolved by name,
survives moves unmodified), use a wikilink. If the target is "a sibling artifact
in this same change" (resolved by path, depth-sensitive), use a relative link. No
other link form is permitted for these two categories.

Wikilinks are stable across the archive move — `[[capability]]` resolves by name,
not by path, so it is never rewritten. Relative links that cross a folder boundary
(e.g. a change artifact linking out to `specs/`) DO need rewriting when the change
moves one level deeper into `changes/archive/`; that rewrite is `archon map`'s job,
not a skill's.

## Managed-Marker Policy

Generated content inside vault files (currently only `map.md`) lives inside a
managed region delimited by:

```
<!-- MAP:START -->
<!-- MAP:END -->
```

Rules:

- Tooling (`archon map`) writes ONLY inside this marker pair. It never touches
  authored prose outside the markers.
- A document MAY contain **at most one** managed region pair.
- Nested managed regions are PROHIBITED — a second `MAP:START` before the matching
  `MAP:END` is an error, not a nesting feature.
- Authored preambles (title, human notes) live outside the markers and are
  preserved byte-for-byte across every regeneration.

## `.feature` Files Stay Put

`.feature` files beside `spec.md` are owned exclusively by the SDD spec phase:

- Their **location** (always `specs/{capability}/{capability}.feature`) and their
  **content** (Gherkin scenarios) are SDD-spec-phase-only.
- No vault operation — link rewrite, archive move, map regeneration — may move,
  rename, or modify a `.feature` file's content. When a change archives, the
  `.feature` file moves alongside `spec.md` as an inert byte-identical payload.
- Markdown artifacts MAY reference a `.feature` file with a relative link, but that
  reference is read-only; it never causes the `.feature` file itself to change.
