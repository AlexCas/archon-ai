## Exploration: local-model-provider

Enable archon to point SDD phases at local models served by Ollama and LocalAI
(both expose OpenAI-compatible HTTP endpoints), while archon keeps its role of
config generator + orchestrator. Archon must NOT become a gateway/proxy: routing
and hosting stay with Ollama/LocalAI and the underlying agent (OpenCode).

### Project Type
**Web testing**: not-web

archon is a Go CLI + Bubbletea TUI. No browser surface, no web framework, no E2E
tooling for a web app. Playwright and Impeccable stay disabled (matches the cached
preflight: Playwright No, Impeccable No). No Impeccable recommendation applies.

### Current State (code-grounded)

**1. ModelRef is the single unit of provider/model assignment.**
`internal/config/model.go:14-18`:
```go
type ModelRef struct {
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	Effort   string `yaml:"effort,omitempty"`
}
```
`Provider` is purely a string prefix. `FullID()` (`model.go:24-32`) concatenates
to `"<provider>/<model>"` (or the bare model if provider is empty, or the model
verbatim if it already contains a `/`). Provider drives NO routing, auth, or
endpoint — it is only a label that ends up in the generated agent config string.
There is NO `baseURL`/`endpoint` field anywhere on ModelRef today.

**2. There is no explicit "provider catalog."** The shaped-problem note that
`model.go:89-129` holds a provider catalog is inaccurate — those are *model*
catalogs (`ClaudeModels`, `GeminiModels`, `OpenAIModels`, `OpencodeModels`). The
closest thing to a provider list is the agent-CLI detector
`internal/models/detect.go:16-24`, which probes PATH for `{"opencode","claude",
"codex","gemini"}` — these are *agent CLIs*, not model backends. Available
model *providers* are discovered dynamically from the OpenCode cache
`~/.cache/opencode/models.json` via `internal/opencode/models.go`
(`DetectAvailableProviders`, lines 115-124); no hard-coded provider allowlist and
no provider-level validation exist. Ollama/LocalAI are model servers, not agent
CLIs, so they would never (and should never) appear in the PATH detector.

**3. Config load / save / clone.** `internal/config/config.go`. `ModelConfig`
(`model.go:80-84`) has `Default`, `Leader`, `Phases map[string]ModelRef`.
`Clone()` (`config.go:136-156`) is a hand-rolled deep copy that value-copies
`Default`/`Leader` and copies each `Phases` entry by value — so adding a scalar
field like `BaseURL` to ModelRef is transparently preserved by Clone (no per-field
edit needed there). `TestConfig_CloneRoundtrip` guards omissions.

**4. Config CLI.** `cmd/archon/config.go`. `set`/`get`/`list` only understand
`models.default` and `models.phases.<phase>`, each parsed via
`config.ParseModelRef(value)` (`config.go:155,225`) — which splits only on the
first `/`. A baseURL cannot be expressed through the current `provider/model`
scalar; it needs either a new key form or a structured parse.

**5. TUI Models tab.** `internal/tui/models_tab.go` (658 lines). State machine
`newModelsTabState`→`update` with modes rowNav / providerSelect / modelSelect /
effortSelect / freeForm. Providers/models are picked from the OpenCode cache, with
a free-form `provider/model` text entry (mode `freeForm`, `openFreeForm` at :197,
committed via `config.ParseModelRef` at :422). Rows persist back via
`applyToConfig` (:634). A baseURL would need a new sub-mode/prompt (or an extended
free-form parse) plus a render change in `renderRow` (:494).

**6. Two divergent generation backends — this is the crux.** `init` dispatches by
`agentName` (`internal/initcmd/init.go:105,118`):
   - **OpenCode path** — `internal/initcmd/opencode_mode.go`. `mergeOpencodeAgent`
     writes ONLY the `agent` map into `opencode.json`: each `archon-<phase>` and
     `archon-leader` gets `{mode, hidden, model, variant, description, prompt}`
     where `model` = `FullID()` (e.g. `opencode-go/deepseek-v4-pro`, see the
     repo's own `opencode.json`). It does NOT emit a top-level `provider` block.
   - **Claude path** — `internal/initcmd/claude_mode.go`. Writes
     `.claude/agents/archon-<phase>.md` with YAML frontmatter `{name, description,
     model}`. Critically, `claudeFrontmatterModel` (`claude_mode.go:87-92`)
     **strips the provider prefix**, emitting only the bare model id, because
     Claude Code's `model:` frontmatter rejects a provider-qualified id and has NO
     field for a baseURL/endpoint.

### The Key Constraint (item #3 in the brief)

**Local providers are OpenCode-only.** Claude Code subagent frontmatter accepts a
bare model id and nothing else — no provider prefix, no baseURL, no endpoint
(`claude_mode.go:81-92`). There is no seam to point a Claude agent at a local
HTTP server. Therefore a local/OpenAI-compatible provider can only be honored on
the OpenCode generation path. The design must either (a) scope local providers to
OpenCode projects and warn/reject when `agent == "claude"`, or (b) accept that the
Claude path silently drops the baseURL and routes to Claude Code's own resolution
(almost certainly wrong for a local model). Option (a) is the honest choice.

### How OpenCode expects a custom provider (OPEN QUESTION — flagged, not guessed)

archon currently emits ONLY the `agent` block of `opencode.json`; it never writes
a top-level `provider` block. There is **no code, test, doc, or committed example
in this repo** describing how OpenCode declares a custom / OpenAI-compatible
provider with a baseURL (grep for `baseURL|base_url|npm|options|@ai-sdk|
openai-compatible` across `internal/`, `docs/`, `README`, `openspec/` returns
nothing model-related — the only `base_url` is Playwright's). OpenCode's public
config schema (`https://opencode.ai/config.json`, referenced by the repo's own
`opencode.json`) does support a top-level `provider` key for custom providers
(typically `provider.<id>.npm` = an ai-sdk package such as
`@ai-sdk/openai-compatible`, plus `provider.<id>.options.baseURL` and a
`models` map, with agents referencing `<id>/<model>`), but the exact required
field shape MUST be verified against OpenCode docs during propose/design rather
than assumed here. **This is the primary open question the propose phase must
resolve** — it determines the entire opencode_mode.go write shape.

### Affected Areas
- `internal/config/model.go` — add `BaseURL` (endpoint) to `ModelRef`; decide
  scalar-vs-mapping marshal impact (a baseURL forces the mapping form, like
  Effort does today, breaking the byte-identical scalar for these refs only).
- `internal/config/config.go` — `Clone()` value-copy already covers a scalar
  field; extend `TestConfig_CloneRoundtrip` fixtures.
- `internal/initcmd/opencode_mode.go` — emit a top-level `provider` block for
  local providers so OpenCode can route to the baseURL; agents keep referencing
  `<provider>/<model>`. This is the largest and riskiest change.
- `internal/initcmd/claude_mode.go` — surface that local providers are
  unsupported on the Claude path (warn/skip/reject); no silent drop.
- `internal/tui/models_tab.go` — UI plumbing for a baseURL field (new sub-mode +
  render).
- `cmd/archon/config.go` — new set/get/list handling for the baseURL (either a
  `models.default.base_url` style key or a `models.providers.<id>` block).
- `internal/config/model.go` `Validate` / TUI validation — advisory validation
  for a local/openai-compatible provider id.
- Tests: `opencode_mode_test.go` (golden output for the new provider block),
  `claude_mode_test.go` (local-provider handling), `config_test.go`,
  `models_tab_test.go`, `cmd/archon/config` tests.

### Approaches

1. **Per-ref baseURL on ModelRef, OpenCode-only routing** — add
   `ModelRef.BaseURL`; when set (and agent==opencode) synthesize a top-level
   `provider.<id>` block in opencode.json and point the agent at `<id>/<model>`.
   Claude path warns/skips.
   - Pros: minimal schema surface; fits the existing ModelRef flow; scalar configs
     without a baseURL stay byte-identical.
   - Cons: a baseURL on a per-phase ref means the same provider could be declared
     from several refs — needs dedup/coalescing into one provider block; forces
     mapping-form marshal for local refs; must confirm OpenCode's exact provider
     schema first.
   - Effort: Medium.

2. **Dedicated `models.providers` map in ModelConfig** — declare local providers
   once (`{id, base_url, api model list}`) and let refs reference them by id.
   - Pros: clean one-place provider declaration; natural mapping to OpenCode's
     top-level `provider` block; no per-ref duplication.
   - Cons: larger config schema + CLI + TUI surface; more new code and tests;
     likely over the 400-line budget on its own.
   - Effort: High.

3. **Doc-only / passthrough** — document that users hand-edit `opencode.json`'s
   `provider` block themselves; archon only preserves it (it already preserves
   unknown top-level keys).
   - Pros: near-zero code; honest about archon-not-a-gateway.
   - Cons: no TUI/CLI support, defeats the "config generator" value; users still
     must know OpenCode's schema. Weak.
   - Effort: Low.

### Recommendation

Lean toward **Approach 1** (per-ref `BaseURL`, OpenCode-only) as the MVP first
slice, but the choice hinges on the OpenCode custom-provider schema open question
— resolve that in propose before committing. Approach 2 is the cleaner end state
if multiple local providers/models are expected; it can be a follow-up slice.
Whichever wins, the Claude-path constraint (local = OpenCode-only) is
non-negotiable and must be stated as an explicit non-goal/guard.

### Size Estimate vs 400-line Budget

Rough, code + tests:
- ModelRef.BaseURL + marshal/parse + Clone/roundtrip test: ~60-90 lines.
- opencode_mode.go provider-block emission + coalescing + golden tests: ~120-180
  lines (the golden-file tests dominate).
- claude_mode.go guard + test: ~30-50 lines.
- CLI set/get/list + tests: ~60-90 lines.
- TUI sub-mode + render + tests: ~120-180 lines.

Total ~390-590 lines. **Likely at or over the 400-line budget** — recommend
**splitting into chained PRs**, e.g.:
- PR-A: config core (ModelRef.BaseURL, marshal, Clone, CLI set/get/list) +
  opencode_mode provider-block emission — the functional spine.
- PR-B: TUI Models-tab baseURL editing + Claude-path guard/warning.
Per the cached PR strategy (ask-always at 400), surface this to the user at the
propose/tasks gate.

### Risks
- **OpenCode custom-provider schema is unverified in-repo.** Wrong field shape =
  a generated opencode.json OpenCode silently ignores or rejects. Must confirm
  against OpenCode docs before design. (Primary risk.)
- **Byte-identical marshal invariant.** ModelRef currently round-trips scalar
  configs byte-for-byte; a baseURL forces mapping form for those refs. Must scope
  this to only refs that carry a baseURL so existing configs stay stable.
- **Provider-block coalescing.** If baseURL lives per-ref, multiple refs pointing
  at the same local provider must merge into ONE top-level provider entry
  deterministically (sorted keys) to preserve idempotent, byte-identical output.
- **Claude-path silent-wrong-routing.** Without a guard, a local ref on the Claude
  path drops the baseURL and mis-routes. Must warn/skip explicitly.
- **Scope creep into a gateway.** Keep the non-goal firm: archon writes config,
  never proxies traffic. Embeddings stay out (no consumer in archon today).

### Open Questions (for propose)
1. **OpenCode custom-provider schema**: exact required shape of the top-level
   `provider.<id>` block (npm package id, `options.baseURL`, per-model map, api
   key handling for keyless local servers). Confirm against OpenCode docs.
2. **Config surface**: per-ref `ModelRef.BaseURL` (Approach 1) vs a
   `models.providers` declaration block (Approach 2)?
3. **Provider id / naming**: reuse a single `local` id, or user-named ids
   (`ollama`, `localai`, arbitrary)? How do refs reference them?
4. **Claude path**: hard-reject a local ref at init when `agent==claude`, or
   allow-with-warning-and-skip?
5. **API key for keyless local servers**: does OpenCode require a placeholder
   `apiKey`/env for Ollama/LocalAI, or can it be omitted?
6. **Validation**: what advisory validation (if any) for a local provider id /
   baseURL format?
7. **PR split**: confirm chained-PR split with the user given the >400-line
   estimate.

### Ready for Proposal
Yes — with the caveat that the OpenCode custom-provider schema (Open Question 1)
should be verified early in propose, as it gates the whole opencode_mode.go
design. Orchestrator should present the two config-surface approaches (per-ref
BaseURL vs providers block) and the likely chained-PR split to the user at the
propose gate.
