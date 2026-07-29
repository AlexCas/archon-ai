# Local Model Provider Specification

<!-- proposal: ../../proposal.md | exploration: ../../exploration.md -->

## Purpose

Define how archon exposes OpenAI-compatible local model endpoints (Ollama, LocalAI)
through a `ModelRef.BaseURL` field, emits the required OpenCode V1 `provider` block,
and guards the Claude path where no local endpoint can be honored.

This is a wholly new capability with no existing `openspec/specs/` file to delta
against; this is therefore a full spec.

PR mapping: **PR-A** covers REQ-1 through REQ-5; **PR-B** covers REQ-6 and REQ-7.

---

## Requirements

### Requirement: REQ-1 — ModelRef.BaseURL Field

`ModelRef` MUST carry an optional `BaseURL string` field (YAML key `base_url`,
`omitempty`). A ref with `BaseURL == ""` MUST continue to round-trip its existing
on-disk form byte-identically (scalar for refs without `Effort`, mapping for refs
that already carry `Effort`). A ref with `BaseURL != ""` MUST marshal as a YAML
mapping regardless of whether `Effort` is set.

`UnmarshalYAML` MUST continue to accept both the legacy scalar form and the mapping
form; the scalar form MUST always decode to `BaseURL == ""` (no inference from
Provider or Model strings).

`Clone()` value-copy already covers scalar fields; no structural change is needed,
but `TestConfig_CloneRoundtrip` MUST include a fixture with a non-empty `BaseURL`.

**PR-A**

#### Scenario: Scalar ref round-trips byte-identically

```gherkin
Scenario: Scalar ref round-trips byte-identically
  Given a config YAML with a phase ref set to the scalar "ollama/llama3"
  When the config is loaded and saved without modification
  Then the emitted YAML is byte-identical to the input
```

#### Scenario: BaseURL ref marshals as mapping

```gherkin
Scenario: BaseURL ref marshals as mapping
  Given a ModelRef with Provider "ollama", Model "llama3", and BaseURL "http://localhost:11434/v1"
  When the ref is marshaled to YAML
  Then the output is a mapping with keys provider, model, and base_url
  And the scalar form is NOT emitted
```

#### Scenario: Scalar input decodes with empty BaseURL

```gherkin
Scenario: Scalar input decodes with empty BaseURL
  Given a YAML scalar value "anthropic/claude-sonnet-4-6"
  When the value is unmarshaled into a ModelRef
  Then Provider is "anthropic", Model is "claude-sonnet-4-6", and BaseURL is ""
```

---

### Requirement: REQ-2 — CLI set/get/list for BaseURL

The `archon config set` command MUST accept the key
`models.phases.<phase>.base_url` and `models.default.base_url` to write the
`BaseURL` field of the corresponding `ModelRef`. `archon config get` on the same
keys MUST return the stored value (empty string if unset). `archon config list`
MUST include a `base_url = <value>` line for every ref whose `BaseURL != ""`,
grouped with its sibling `models.phases.<phase>` or `models.default` line.

Setting a `base_url` key MUST NOT alter the model/provider fields of the same ref.

**PR-A**

#### Scenario: Set and get base_url for a phase

```gherkin
Scenario: Set and get base_url for a phase
  Given an archon project with no base_url configured
  When the user runs: archon config set models.phases.apply.base_url http://localhost:11434/v1
  Then archon config get models.phases.apply.base_url prints "http://localhost:11434/v1"
  And the provider and model fields of the apply ref are unchanged
```

#### Scenario: list shows base_url lines

```gherkin
Scenario: list shows base_url lines
  Given a config where models.phases.apply has BaseURL "http://localhost:11434/v1"
  When the user runs: archon config list
  Then the output includes a line "models.phases.apply.base_url = http://localhost:11434/v1"
```

#### Scenario: Get base_url when unset returns empty

```gherkin
Scenario: Get base_url when unset returns empty
  Given a config where models.default has no BaseURL
  When the user runs: archon config get models.default.base_url
  Then the command exits 0 and prints nothing
```

---

### Requirement: REQ-3 — Advisory Validation of BaseURL

When a `BaseURL` is present on a ref, the system SHOULD validate it at set-time
(CLI) and at config-load time (init/TUI) and emit a WARNING to stderr. The system
MUST NOT reject or fail on an invalid BaseURL — validation is advisory only.

Validation rules:
- Provider id MUST be non-empty when BaseURL is set; violation message:
  `warning: base_url is set but provider is empty — provider id required for local model routing`.
- BaseURL MUST parse as an `http` or `https` URL with a non-empty host; violation
  message: `warning: base_url "<value>" is not a valid http/https URL`.

**PR-A**

#### Scenario: Valid BaseURL produces no warning

```gherkin
Scenario: Valid BaseURL produces no warning
  Given a ModelRef with Provider "ollama" and BaseURL "http://localhost:11434/v1"
  When advisory validation runs
  Then no warning is emitted
```

#### Scenario: Non-http BaseURL triggers a warning

```gherkin
Scenario: Non-http BaseURL triggers a warning
  Given a ModelRef with Provider "ollama" and BaseURL "ftp://localhost/v1"
  When advisory validation runs
  Then stderr contains: warning: base_url "ftp://localhost/v1" is not a valid http/https URL
```

#### Scenario: BaseURL set but provider empty triggers a warning

```gherkin
Scenario: BaseURL set but provider empty triggers a warning
  Given a ModelRef with empty Provider and BaseURL "http://localhost:11434/v1"
  When advisory validation runs
  Then stderr contains: warning: base_url is set but provider is empty — provider id required for local model routing
```

---

### Requirement: REQ-4 — OpenCode provider block emission

When `mergeOpencodeAgent` is called and one or more resolved `ModelRef` values have
`BaseURL != ""`, the system MUST emit a top-level `"provider"` key in `opencode.json`
alongside the existing `"agent"` key.

**Block shape (OpenCode V1 schema):**

```json
"provider": {
  "<provider-id>": {
    "npm": "@ai-sdk/openai-compatible",
    "options": { "baseURL": "<BaseURL>" },
    "models": { "<model>": { "name": "<model>" } }
  }
}
```

- `"npm"` MUST be `"@ai-sdk/openai-compatible"` for every local provider block.
- `"options"` MUST contain only `"baseURL"` — no `"apiKey"` key is emitted (keyless
  server contract; resolves API-key decision).
- `"models"` MUST contain one entry per distinct model routed through this provider
  id, with the model id as key and `{"name": "<model-id>"}` as value.
- The agent entries in `"agent"` MUST continue to reference `"<provider-id>/<model-id>"`
  (i.e., `FullID()` is unchanged).

**Coalescing rule:** Multiple refs sharing the same `Provider` id (and therefore the
same OpenCode provider id) MUST be merged into ONE provider block. The `models` map
is the union of all model ids from those refs. If two refs share the same `Provider`
but carry **different** `BaseURL` values the system MUST emit a WARNING to stderr
(message: `warning: provider "<id>" declared with conflicting baseURLs — using first
occurrence "<url>"`) and MUST use the BaseURL of the **first encountered ref** (in
`PhaseOrder` traversal order, default ref last).

**Deterministic ordering:** Provider ids in the `"provider"` map MUST be sorted
lexicographically. Model ids within each provider's `"models"` map MUST be sorted
lexicographically. The resulting JSON MUST be byte-identical across re-runs with the
same config.

**Additive merge:** Existing non-archon keys in the top-level `"provider"` map MUST
be preserved. If the existing file has no `"provider"` key, one is created. If the
existing file has a `"provider"` key with user-defined entries, those entries MUST
NOT be deleted or modified.

**PR-A**

#### Scenario: Ollama happy path — single phase

```gherkin
Scenario: Ollama happy path — single phase
  Given a config with models.phases.apply set to Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
  When archon init runs with agent "opencode"
  Then opencode.json contains a top-level "provider" key
  And provider.ollama.npm equals "@ai-sdk/openai-compatible"
  And provider.ollama.options.baseURL equals "http://localhost:11434/v1"
  And provider.ollama.models contains key "llama3" with value {"name":"llama3"}
  And agent.archon-apply.model equals "ollama/llama3"
  And the "options" object does NOT contain an "apiKey" key
```

#### Scenario: LocalAI happy path — single phase

```gherkin
Scenario: LocalAI happy path — single phase
  Given a config with models.phases.spec set to Provider "localai", Model "gpt-4-vision", BaseURL "http://localhost:8080/v1"
  When archon init runs with agent "opencode"
  Then provider.localai.npm equals "@ai-sdk/openai-compatible"
  And provider.localai.options.baseURL equals "http://localhost:8080/v1"
  And provider.localai.models contains key "gpt-4-vision"
  And agent.archon-spec.model equals "localai/gpt-4-vision"
```

#### Scenario: Multiple phases same provider are coalesced

```gherkin
Scenario: Multiple phases same provider are coalesced
  Given phases apply and verify both use Provider "ollama" and BaseURL "http://localhost:11434/v1" with models "llama3" and "mistral" respectively
  When archon init runs with agent "opencode"
  Then the "provider" object contains exactly ONE "ollama" key
  And provider.ollama.models contains both "llama3" and "mistral"
```

#### Scenario: Mixed local and remote phases

```gherkin
Scenario: Mixed local and remote phases
  Given phase apply uses Provider "ollama" with BaseURL "http://localhost:11434/v1" and model "llama3"
  And phase spec uses Provider "anthropic" with no BaseURL and model "claude-sonnet-4-6"
  When archon init runs with agent "opencode"
  Then opencode.json contains a "provider" block for "ollama"
  And opencode.json does NOT contain a "provider" block for "anthropic"
  And agent.archon-spec.model equals "anthropic/claude-sonnet-4-6"
```

#### Scenario: Conflicting BaseURLs for same provider id

```gherkin
Scenario: Conflicting BaseURLs for same provider id
  Given phase apply uses Provider "ollama", BaseURL "http://localhost:11434/v1"
  And phase spec uses Provider "ollama", BaseURL "http://remote-ollama:11434/v1"
  When archon init runs with agent "opencode"
  Then stderr contains: warning: provider "ollama" declared with conflicting baseURLs — using first occurrence "http://localhost:11434/v1"
  And provider.ollama.options.baseURL equals "http://localhost:11434/v1"
```

#### Scenario: No BaseURL refs — no provider block emitted

```gherkin
Scenario: No BaseURL refs — no provider block emitted
  Given a config where no ModelRef has a BaseURL
  When archon init runs with agent "opencode"
  Then opencode.json does NOT contain a top-level "provider" key
  And the output is byte-identical to previous runs with the same config
```

#### Scenario: Existing user-defined provider entries are preserved

```gherkin
Scenario: Existing user-defined provider entries are preserved
  Given opencode.json already has provider.myprovider with custom user data
  And a config with models.phases.apply using Provider "ollama" with BaseURL
  When archon init runs with agent "opencode"
  Then provider.myprovider is unchanged
  And provider.ollama is added alongside it
```

---

### Requirement: REQ-5 — Deterministic and idempotent opencode.json output

The emitted `opencode.json` MUST be byte-identical across repeated invocations of
`mergeOpencodeAgent` with the same `ModelConfig`. This extends the existing
idempotency contract to include the `"provider"` block.

Provider id keys and model id keys within `"provider"` MUST be sorted as part of
the `json.MarshalIndent` pass (Go's `map[string]any` sorts keys when marshaled).

**PR-A**

#### Scenario: Re-run produces identical output

```gherkin
Scenario: Re-run produces identical output
  Given a config with two local providers "aaa" and "zzz" each with one model
  When archon init runs twice with the same config
  Then both runs produce byte-identical opencode.json content
  And provider keys appear in lexicographic order: "aaa" before "zzz"
```

---

### Requirement: REQ-6 — Claude path warn-and-skip guard

When `writeClaudeAgents` processes a phase ref that has `BaseURL != ""`, the system
MUST emit a visible warning to stderr before writing the agent file, then write the
agent file with the bare model id as if `BaseURL` were empty. The system MUST NOT
silently ignore the BaseURL and MUST NOT hard-reject or abort writing the file.

Warning message (exact): `warning: phase "<phase>" has base_url set but agent is
"claude" — local endpoint ignored; claude agents do not support custom baseURLs`.

The `model:` frontmatter in the emitted `.claude/agents/archon-<phase>.md` MUST be
the bare model id (current `claudeFrontmatterModel` behavior, unchanged).

**PR-B**

#### Scenario: Local ref on Claude path triggers warn-and-skip

```gherkin
Scenario: Local ref on Claude path triggers warn-and-skip
  Given a config with models.phases.apply set to Provider "ollama", Model "llama3", BaseURL "http://localhost:11434/v1"
  When archon init runs with agent "claude"
  Then stderr contains: warning: phase "apply" has base_url set but agent is "claude" — local endpoint ignored; claude agents do not support custom baseURLs
  And the file .claude/agents/archon-apply.md is written
  And its model: frontmatter equals "llama3"
```

#### Scenario: Remote ref on Claude path has no warning

```gherkin
Scenario: Remote ref on Claude path has no warning
  Given a config with models.phases.apply set to Provider "anthropic", Model "claude-sonnet-4-6", and no BaseURL
  When archon init runs with agent "claude"
  Then stderr contains no base_url warning
  And .claude/agents/archon-apply.md is written normally
```

#### Scenario: Multiple local phases each emit a warning

```gherkin
Scenario: Multiple local phases each emit a warning
  Given phases apply and verify both have BaseURL set and agent is "claude"
  When archon init runs with agent "claude"
  Then stderr contains one warning per local phase
  And both agent files are written with bare model ids
```

---

### Requirement: REQ-7 — TUI Models-tab BaseURL editing

The TUI Models tab MUST provide a way for the user to view and edit the `BaseURL`
of any model row. When a row has `BaseURL != ""`, the row's rendered display MUST
show the BaseURL value alongside the provider/model assignment. The tab MUST offer
a sub-mode to type or clear the BaseURL for the focused row.

The BaseURL sub-mode MUST use the existing `textinput` component pattern (matching
`freeForm` mode) and MUST commit the typed value to the row's `ModelRef.BaseURL` on
Enter. Escape MUST cancel without committing. Empty input on commit MUST clear the
BaseURL (setting it to `""`).

`applyToConfig` MUST persist the edited `BaseURL` back to `ModelConfig` alongside
the provider/model fields.

**PR-B**

#### Scenario: Row with BaseURL shows endpoint in display

```gherkin
Scenario: Row with BaseURL shows endpoint in display
  Given the TUI Models tab is open
  And the apply row has BaseURL "http://localhost:11434/v1"
  When the tab renders
  Then the apply row shows "http://localhost:11434/v1" in its display text
```

#### Scenario: User can set BaseURL via TUI sub-mode

```gherkin
Scenario: User can set BaseURL via TUI sub-mode
  Given the TUI Models tab is open and the apply row is focused
  When the user activates the BaseURL sub-mode and types "http://localhost:11434/v1" and presses Enter
  Then the apply row's BaseURL is "http://localhost:11434/v1"
  And saving the config persists the BaseURL to .archon/config.yaml
```

#### Scenario: User can clear BaseURL via TUI sub-mode

```gherkin
Scenario: User can clear BaseURL via TUI sub-mode
  Given the apply row has BaseURL "http://localhost:11434/v1"
  When the user activates the BaseURL sub-mode, clears the input, and presses Enter
  Then the apply row's BaseURL is ""
  And the config YAML for that ref reverts to scalar form
```

#### Scenario: Escape cancels BaseURL edit

```gherkin
Scenario: Escape cancels BaseURL edit
  Given the apply row has BaseURL "http://localhost:11434/v1"
  When the user activates the BaseURL sub-mode, types a new value, and presses Escape
  Then the apply row's BaseURL is still "http://localhost:11434/v1"
```

---

## PR Mapping Summary

| REQ | Description | PR |
|-----|-------------|-----|
| REQ-1 | ModelRef.BaseURL — YAML marshal/unmarshal + round-trip | PR-A |
| REQ-2 | CLI set/get/list for base_url key | PR-A |
| REQ-3 | Advisory validation — warnings, never hard-fail | PR-A |
| REQ-4 | opencode.json provider block — coalesce, shape, preserve | PR-A |
| REQ-5 | Deterministic/idempotent output | PR-A |
| REQ-6 | Claude path warn-and-skip guard | PR-B |
| REQ-7 | TUI Models-tab BaseURL sub-mode | PR-B |

## Key Invariants

1. **Byte-identical round-trip**: refs without BaseURL MUST emit the exact same
   YAML bytes they were loaded from. No behavioural change for existing configs.
2. **Coalescing**: multiple refs with the same provider id yield exactly ONE
   `provider.<id>` block; union of models; first-seen BaseURL wins on conflict.
3. **Deterministic ordering**: provider ids and model ids sorted lexicographically;
   output is byte-identical across runs.
4. **Keyless API**: no `apiKey` field emitted in the provider block.
5. **Warn-never-fail validation**: advisory warnings go to stderr; the operation
   always completes.
6. **OpenCode V1 target**: `npm = "@ai-sdk/openai-compatible"`, `options.baseURL`.
   Document V2 divergence risk in code comments; do not abstract away from it.
