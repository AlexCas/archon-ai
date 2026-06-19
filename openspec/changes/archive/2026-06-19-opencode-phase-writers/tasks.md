# Tasks — opencode-phase-writers (Slice 2)

Ordered, dependency-aware. Signature change ripples into call sites + tests; apply them
together (do not expect a green build mid-change). Tag = spec requirement.

## W1 — Writer (internal/initcmd/opencode_mode.go)
- [x] W1-1 Add import `internal/config`; add `phaseAgentName(phase)` helper and `archonPhaseAgent`
      struct (Mode, Hidden, Model, Description, Prompt; field order = JSON order; Hidden no omitempty).
- [x] W1-2 Change `MergeOpencodeAgent`/`mergeOpencodeAgent` signature to take `config.ModelConfig`.
      Update the exported delegate.
- [x] W1-3 Implement resolution + no-op: `leaderFull := models.Leader.FullID()`;
      `phases := config.ResolvePhaseModels(models)`; no-op when `leaderFull=="" && len(phases)==0`.
- [x] W1-4 Gate leader write on `leaderFull != ""`; loop `phases` writing
      `agents[phaseAgentName(pm.Phase)] = archonPhaseAgent{...}` with `Model: pm.Model`. Keep
      additive merge / MarshalIndent / trailing newline / atomic rename unchanged. Update doc comment.

## W2 — Call sites (apply together with W1)
- [x] W2-1 `internal/initcmd/init.go:102` → `mergeOpencodeAgent(opts.ProjectDir, cfg.Models)`.
- [x] W2-2 `internal/tui/model.go:334` → `initcmd.MergeOpencodeAgent(m.projectDir, cfg.Models)`.

## W3 — Tests (internal/initcmd/opencode_mode_test.go)
- [x] W3-1 Migrate the 6 writer call sites to `config.ModelConfig{Leader: config.ParseModelRef(testLeaderModel)}`.
- [x] W3-2 Add `phaseAgentFrom(t, doc, phase)` helper.
- [x] W3-3 Rename/rework `_EmptyLeaderWritesNothing` → `_NothingConfiguredWritesNothing` (empty ModelConfig{}). [no-op req]
- [x] W3-4 `_WritesSubagentPerResolvablePhase` (+ no archon-judge). [per-phase emission]
- [x] W3-5 `_PhaseModelMatchesResolvedFullID`. [resolved FullID]
- [x] W3-6 `_PhaseFallsBackToDefault`. [resolved FullID/edge]
- [x] W3-7 `_SubagentFixedFields` (mode/hidden/description/prompt). [fixed shape]
- [x] W3-8 `_PhasesSetEmptyLeaderWritesSubagentsNoLeader`. [no-op edge]
- [x] W3-9 Extend `_PreservesExisting` + `_Idempotent` to the leader+subagent set (byte-identical re-run, user agent preserved). [determinism/preservation]

## W4 — TUI test
- [x] W4-1 `internal/tui/model_test.go:675` reference merge: pass `config.ModelConfig` (new signature);
      keep the byte-identical TUI-vs-init cross-path assertion.

## W5 — Gates
- [x] W5-1 `gofmt -l` on changed files empty.
- [x] W5-2 `go build ./...` clean.
- [x] W5-3 `go vet ./...` clean.
- [x] W5-4 `go test ./...` all green.

## Definition of Done
- opencode.json gets `archon-<phase>` subagents for every ResolvePhaseModels phase, each model ==
  the advisory FullID; judge never emitted.
- Leader unchanged shape; written only when its FullID is non-empty; no-op when nothing configured.
- Re-run byte-identical; existing top-level keys + user agents preserved.
- build/vet/test green; templates untouched (already emit FullID); rollback unchanged.
