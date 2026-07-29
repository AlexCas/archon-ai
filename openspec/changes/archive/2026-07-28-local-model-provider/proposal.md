# Proposal: Local Model Provider (Ollama / LocalAI via OpenCode)

## Intent

Users running local models (Ollama, LocalAI — both OpenAI-compatible HTTP
servers) cannot point archon-generated SDD phases at them today. `ModelRef`
carries only `provider/model` labels with no endpoint, and archon emits only the
`agent` block of `opencode.json`, never a top-level `provider` block. Archon stays
a config generator + orchestrator: it writes the provider declaration OpenCode
needs; it never proxies traffic. Embeddings are out of scope.

## Scope

### In Scope
- Add `ModelRef.BaseURL` (endpoint) scalar field.
- Emit a coalesced top-level `provider.<id>` block in `opencode.json` for refs
  that carry a BaseURL (OpenCode path only).
- CLI `set/get/list` support for the BaseURL.
- TUI Models-tab BaseURL editing (PR-B).
- Claude-path guard: explicit warn-and-skip when a local ref meets `agent==claude`
  (PR-B). No silent drop.
- Advisory validation of provider id + BaseURL format.

### Out of Scope
- A gateway/proxy — archon never routes model traffic.
- Embeddings (no consumer in archon today).
- A dedicated `models.providers` block (Approach 2) — deferred follow-up.
- Auto-detection of running Ollama/LocalAI servers.

## Capabilities

### New Capabilities
- `local-model-provider`: declaring an OpenAI-compatible local endpoint on a
  model ref and emitting the corresponding OpenCode custom-provider block.

### Modified Capabilities
- None. (No existing spec files under `openspec/specs/`; behavior is additive.)

## OpenCode Custom-Provider Schema — VERIFIED

Confirmed against OpenCode official docs (`opencode.ai/docs/providers`). The repo
targets the V1 schema (`$schema: https://opencode.ai/config.json`). Shape:

```json
"provider": {
  "<id>": {
    "npm": "@ai-sdk/openai-compatible",
    "name": "<display>",
    "options": { "baseURL": "http://localhost:11434/v1" },
    "models": { "<model-id>": { "name": "<display>" } }
  }
}
```
- `npm`: `@ai-sdk/openai-compatible` for `/v1/chat/completions` (Ollama, LocalAI).
- `options.baseURL`: the local endpoint. `apiKey` OMITTED for keyless servers
  (resolves Open-Q 5). Agents reference `"<id>/<model>"` — matches `FullID()`.
- **Risk**: OpenCode V2 renames `options.baseURL`→`settings.baseURL` and prefixes
  npm with `aisdk:`. We target V1 (repo's schema URL); revisit if archon migrates.

## Approach

**Approach 1 — per-ref `ModelRef.BaseURL`, OpenCode-only** (recommended MVP).
When a resolvable ref has `BaseURL != ""` and `agent==opencode`, coalesce all such
refs by provider id into ONE top-level `provider.<id>` block (npm =
`@ai-sdk/openai-compatible`, `options.baseURL`, `models` = union of that provider's
models), additively merged into `opencode.json` like the existing `agent` block.
Agents keep referencing `<provider>/<model>`.

**Resolved decisions:**
1. Config surface: **per-ref `BaseURL`** — leaner MVP, fits ModelRef flow, scalar
   refs without a BaseURL stay byte-identical. `models.providers` deferred.
2. Naming: **user-named provider ids** (`ollama`, `localai`, arbitrary) — the ref's
   `Provider` IS the OpenCode provider id; no forced `local` alias.
3. Claude path: **warn-and-skip** (not hard-reject) — emit the Claude agent with a
   visible warning that the local endpoint is dropped; never silent.
4. API key: **omitted** for keyless local servers (verified).
5. Validation: **advisory** — non-empty provider id, BaseURL parses as http(s)
   URL; warn (don't fail) otherwise.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/model.go` | Modified | `ModelRef.BaseURL`; mapping-form marshal only when set; advisory `Validate` |
| `internal/config/config.go` | Modified | `Clone` scalar copy already covers it; extend roundtrip test |
| `internal/initcmd/opencode_mode.go` | Modified | Emit + coalesce `provider.<id>` block (largest change) |
| `internal/initcmd/claude_mode.go` | Modified | Warn-and-skip guard for local refs |
| `cmd/archon/config.go` | Modified | `set/get/list` BaseURL key handling |
| `internal/tui/models_tab.go` | Modified | BaseURL sub-mode + render |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| OpenCode V2 schema divergence (`settings.baseURL`, `aisdk:` prefix) | Med | Target V1 (repo's `$schema`); document; revisit on migration |
| Byte-identical marshal broken for existing configs | Med | Mapping form ONLY for refs with a BaseURL; golden tests |
| Provider-block coalescing non-deterministic | Med | Sort ids + model keys; idempotent golden test |
| Claude-path silent mis-route | Med | Explicit warn-and-skip guard + test |
| Scope creep into a gateway | Low | Firm non-goal; config-only, no traffic |

## Rollback Plan

Additive and behind a set BaseURL. Revert the PRs (or unset the BaseURL); refs
without a BaseURL emit byte-identical `opencode.json` as before. No migration.

## Dependencies

- OpenCode consuming the V1 custom-provider schema (verified).
- A running local server (Ollama/LocalAI) at the configured BaseURL (user-owned).

## Success Criteria

- [ ] A ref with `Provider=ollama`, `BaseURL=http://localhost:11434/v1` yields a
      valid, idempotent `provider.ollama` block plus `ollama/<model>` agent refs.
- [ ] LocalAI happy path works identically with a different id/BaseURL.
- [ ] Configs without a BaseURL stay byte-identical.
- [ ] Claude path warns and skips the local endpoint — no silent drop.
- [ ] CLI + TUI can set and read the BaseURL.

## PR Split (>400-line estimate — chained)

- **PR-A** (config core + emission): `ModelRef.BaseURL`, marshal, Clone/roundtrip,
  CLI set/get/list, advisory Validate, `opencode_mode.go` provider-block emission +
  coalescing + golden tests. Est. ~240-320 lines — near/over budget on its own.
- **PR-B** (surfaces + guard): TUI BaseURL sub-mode + render + tests,
  `claude_mode.go` warn-and-skip guard + test. Est. ~150-230 lines.

Combined ~390-550 lines exceeds the 400-line budget → chained split recommended
per the ask-always strategy. Confirm at the gate.
