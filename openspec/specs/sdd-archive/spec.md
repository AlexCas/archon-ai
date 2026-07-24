# Delta for sdd-archive

## ADDED Requirements

### Requirement: Archive-Triggered Link Rewrite

When `sdd-archive` moves `changes/{c}/` to `changes/archive/YYYY-MM-DD-{c}/`, it MUST
invoke the `archon map` Go step to deterministically rewrite all boundary-crossing
relative links (map→change edges and change→specs edges that shift depth) and regenerate
`map.md`. This MUST happen as part of the same atomic archive operation, before the
move is considered complete.

#### Scenario: Archive operation rewrites boundary links

```gherkin
@happy
Scenario: Archive operation rewrites boundary links
  Given an active change my-feature with relative links into openspec/specs/
  When sdd-archive archives my-feature
  Then archon map is invoked to rewrite boundary-crossing relative links
  And map.md is regenerated to reflect the new archive path
  And archon map --check passes after the operation
```

#### Scenario: Wikilinks survive archive unchanged

```gherkin
@happy
Scenario: Wikilinks survive archive unchanged
  Given a change artifact containing [[harness-workflow]] and [[spec-vault]]
  When sdd-archive archives the change
  Then both wikilinks are byte-identical after the archive move
```

### Requirement: Post-Archive --check Guard

After the archive move and link rewrite, `sdd-archive` MUST run `archon map --check`
to verify no dangling relative links remain. If `--check` exits non-zero, `sdd-archive`
MUST surface the failure to the orchestrator and MUST NOT mark the archive as complete.

#### Scenario: Archive aborts if --check fails

```gherkin
@error
Scenario: Archive aborts if --check fails
  Given an archive operation that produces a dangling relative link
  When archon map --check runs after the move
  Then the command exits non-zero
  And sdd-archive surfaces the failure to the orchestrator
  And the archive is not marked complete
```
