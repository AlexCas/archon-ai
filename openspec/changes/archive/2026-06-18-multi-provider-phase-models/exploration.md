## Exploration: Multi-provider per-phase models (Slice 1 MVP)

### Project Type
**Web testing**: not-web

This is a Go CLI/TUI project (`internal/config`, `internal/initcmd`, `internal/tui`
built on bubbletea/lipgloss). No web framework, no `package.json`, no dev server, no
browser-facing routes, no Playwright/Cypress tooling. `playwright.enabled` MUST stay
`false`; no Playwright generation applies to this change.

### Current State

Per-phase model config (`ModelConfig{Default, Phases}`, `internal/config/model.go:9-12`)
is fully plumbed end-to-end as of the prior change `phase-model-propagation`:

- `ResolvePhaseModels(cfg.Models)` (`model.go:119-132`) walks the canonical
  `PhaseOrder` and, per phase, normalizes the explicit `Phases[p]` value, falling back
  to `Default`, omitting the line when neither resolves.
- Both render paths feed the same resolver: `init.writeTemplate` (`init.go:96`) and
  `tui.regenerateTemplate` (`tui/model.go:338`). The template block
  (`templates.go:162-171`) is guarded by `{{if .PhaseModels}}` and simply ranges over
  the resolved slice — it is **already provider-agnostic**. Nothing in the template or
  the plumbing is Claude-specific.

**The single point that gates non-Claude providers is `NormalizeModel`**
(`model.go:97-113`). It only matches the three `claudeFamilies` tokens
(`opus|sonnet|haiku`, `model.go:78`) and returns `("", false)` for everything else —
including every Opencode/Gemini/OpenAI value (verified by tests `model_test.go:54-56`:
`glm-5`, `kimi-k2.5`, `gpt-4` all → `ok=false`). Because `ResolvePhaseModels` drops any
phase where both the phase value and the default fail to normalize, a project configured
entirely with non-Claude models resolves to an **empty** slice, so `{{if .PhaseModels}}`
omits the whole `## Phase Models` block. That is exactly the reported bug.

Catalogs today: `ClaudeModels` (`model.go:17-21`) and `OpencodeModels`
(`model.go:27-36`); `StaticModels()` concatenates them (Claude first), `KnownModels`
(`model.go:51-57`) is the advisory validation set built from `StaticModels()`. The TUI
(`models_tab.go:129,186`) cycles and lists `StaticModels()`. `Validate` (`model.go:134-145`)
is advisory: it returns a warning string and NEVER rejects (covered by `model_test.go:8-33`).

Important: Opencode models are already "known" (they validate clean, no warning) yet
`NormalizeModel` returns `ok=false` for them. So today they pass validation but never
reach the rendered block. The user has confirmed the Opencode delegation runtime DOES
accept these provider model IDs, so emitting them is useful, not cosmetic.

### Affected Areas

- `internal/config/model.go` — core change. Replace the Claude-only `NormalizeModel`
  with per-provider normalization; add curated `GeminiModels` and `OpenAIModels`
  catalogs; fold them into `StaticModels()`/`KnownModels`. `ResolvePhaseModels`,
  `Validate`, `PhaseOrder`, `PhaseModel` stay structurally as-is (they call through
  `NormalizeModel`, so they inherit multi-provider support for free).
- `internal/config/model_test.go` — update/extend. Existing rows assert
  `glm-5`/`kimi-k2.5`/`gpt-4` → `ok=false` (`model_test.go:54-56`); these expectations
  INVERT under the new behavior and MUST be rewritten. Add provider tables for Gemini,
  OpenAI, Opencode; keep the typo `Opues 4.8`→`ok=false` and `octopus` whole-token
  cases.
- `internal/tui/models_tab.go` — no logic change required; it already renders whatever
  `StaticModels()` returns, so new catalogs surface automatically in the cycle list and
  the "Static:" hint. Worth a visual check that the longer list still fits the hint line.
- `internal/initcmd/templates.go`, `internal/initcmd/init.go`, `internal/tui/model.go`
  — NO code change. The block, `TemplateData.PhaseModels`, and both render paths are
  already provider-agnostic; they begin rendering non-Claude phases the moment
  `NormalizeModel` returns `ok=true` for them. (Confirm with a render test, but expect
  zero edits.)
- `internal/initcmd/templates_test.go` — add a render assertion proving a non-Claude
  config (e.g. default `glm-5`) now emits the `## Phase Models` block. Existing golden
  checks stay green (they use Claude or empty data).

### Approaches

All three approaches concern ONLY how `NormalizeModel` decides a value's normalized
identifier per provider. The contract stays `(id string, ok bool)`, advisory, pure,
idempotent, deterministic.

1. **Provider-family table (ordered list of `{provider, family-tokens, canonical}` entries)**
   — generalize the existing `claudeFamilies` approach into a single ordered table
   covering Claude, Gemini, OpenAI, Opencode. Tokenize the input (reuse the existing
   non-alphanumeric `FieldsFunc` split + whole-token match) and walk the table in a
   fixed order; first family whose token appears wins.
   - Pros: minimal departure from the current proven algorithm and its whole-token
     semantics (octopus-safe); one data structure to read in review; deterministic by
     table order; trivial to extend later (new provider = new rows); keeps everything in
     one function.
   - Cons: must define what each provider's "canonical output" is (alias vs. full ID) —
     a real design decision per the user's "normalize to the identifier the delegation
     tool accepts for that provider"; a single flat table mixes providers, so the
     priority ordering across providers needs care to stay deterministic.
   - Effort: Low

2. **Per-provider normalizer functions (dispatch)** — one small function per provider
   (`normalizeClaude`, `normalizeGemini`, `normalizeOpenAI`, `normalizeOpencode`);
   `NormalizeModel` tries each in a fixed order and returns the first `ok`.
   - Pros: each provider's rules (and its canonical output form) are isolated and
     independently testable; cleanest place to encode provider-specific quirks (e.g.
     Gemini `gemini-2.5-pro` vs. display `Gemini 2.5 Pro`, OpenAI `gpt-4o` family
     handling); easy to reason about which provider matched.
   - Cons: more surface area than the table; provider-precedence still must be fixed and
     documented; mild duplication of the tokenize/match boilerplate unless shared.
   - Effort: Low–Medium

3. **Ordered matcher list (predicate + canonicalizer closures)** — a `[]matcher` where
   each entry is `{match func(tokens) bool, canonical func(tokens) string}`.
   - Pros: most flexible; supports models that need a computed canonical (e.g. preserve a
     point-release) without special-casing.
   - Cons: closures are harder to scan in review than data; flexibility we do not need
     for Slice 1 (curated catalogs are small and known); easiest to make accidentally
     non-deterministic.
   - Effort: Medium

### Recommendation

**Approach 1 (provider-family table), with the per-provider canonical output captured as
a column.** It is the smallest, most reviewable step from the existing, well-tested
algorithm: it preserves the whole-token / `octopus`-safe / priority-ordering semantics
the prior change deliberately established, and it keeps all normalization logic in one
place sharing the existing tokenizer. The only genuinely new design decisions it forces
are the right ones to surface in the spec/design phase:

- **Canonical output per provider.** The user explicitly chose per-provider
  normalization over pass-through and confirmed Opencode accepts provider model IDs.
  Most provider catalogs (Gemini, OpenAI, Opencode) have no short family-alias system
  comparable to Claude's `opus|sonnet|haiku`; their accepted IDs ARE the catalog
  strings (`glm-5`, `gpt-4o`, `gemini-2.5-pro`). So the natural canonical form for
  non-Claude providers is the matched catalog ID itself (lowercased, normalized
  separators), while Claude keeps emitting its short alias for the reasons the prior
  design recorded (stable across dated point-releases, clean in CLAUDE.md). Design must
  state this split explicitly.
- **Deterministic cross-provider precedence.** Define one fixed provider order (e.g.
  Claude → Gemini → OpenAI → Opencode) and document it, mirroring the existing intra-
  Claude `opus→sonnet→haiku` priority rule, so a value that somehow matches two
  providers resolves deterministically.

If, during design, per-provider quirks turn out to need isolated handling, Approach 1
refactors cleanly into Approach 2 — the table entries become per-provider functions
without changing the `NormalizeModel` contract or any caller. That keeps the door open
without paying for it now.

Keep `Validate` advisory and keep free-form entry working (both are explicit constraints
and already true; the new catalogs just widen `KnownModels`, so more values validate
clean). No template, init, or TUI-render code changes are expected — the fix is almost
entirely inside `NormalizeModel` plus the catalogs.

### Risks

- **Behavior inversion breaks existing tests.** `model_test.go:54-56` currently assert
  non-Claude values do NOT normalize; those rows MUST be rewritten, not just added to.
  This is intended behavior change, but it will (correctly) fail the prior golden
  expectations until updated.
- **Catalog accuracy / staleness.** Curated Gemini/OpenAI/Opencode lists can drift from
  real provider catalogs. Mitigate by keeping lists small, current, and clearly "edit me
  as catalogs change," and by keeping free-form entry as the always-available escape
  hatch (the existing pattern).
- **Canonical-ID acceptance assumption.** Emitting `glm-5`/`gpt-4o`/`gemini-2.5-pro`
  into the block assumes the Opencode delegation runtime accepts those exact IDs. The
  user confirmed this for Opencode; design should state it as the documented assumption
  and note that, like the prior Claude-alias risk, the canonical output form lives in
  one function and can be swapped if a platform needs a different form.
- **Cross-provider collision / precedence.** A future ambiguous value could match more
  than one provider; the fixed, documented provider order prevents non-determinism.
- **Whole-token semantics for non-Claude families.** Some provider IDs glue family +
  version without separators (e.g. `gpt4o`, `glm5`). The existing tokenizer splits on
  non-alphanumeric boundaries and matches whole tokens, so a glued form may not match
  (cf. `model_test.go:61` "separatorless glued form does not resolve"). Design must
  decide each catalog's expected input forms and test them so a real configured value
  like `gpt-4o` resolves while typos do not.

### Deferred (Slice 1 must not foreclose; do NOT design now)

- **(a) Opencode "archon-leader" mode + a leader-model field in the TUI** (writing a
  leader mode into the opencode config). Out of scope. Slice 1 stays inside
  normalization + catalogs + the existing advisory block; it adds no new config schema,
  no opencode-config writer, and no TUI leader field — so adding those later is purely
  additive.
- **(b) Dynamic detection of installed agents/models.** Out of scope. Slice 1 keeps the
  curated STATIC catalogs as Go vars (`ClaudeModels`/`OpencodeModels`/new
  `GeminiModels`/`OpenAIModels`). Because `StaticModels()` is already the single source
  consumed by the TUI and `KnownModels`, a future dynamic detector can replace/augment
  that one function without touching `NormalizeModel`, the template, or callers.

### Ready for Proposal

**Yes.** The problem is precisely located (the Claude-only `NormalizeModel`), the fix is
small and mostly confined to `internal/config/model.go` plus its tests, the template and
both render paths already work for any provider, and the two related ideas are clearly
deferred without being foreclosed. The one decision the proposal/design must nail down is
the per-provider canonical output form (Claude alias vs. catalog-ID for the others) and
the fixed cross-provider precedence — both flagged above.

Orchestrator: tell the user this is ready to propose. Confirm with them, during
propose/spec, the exact contents of the curated `GeminiModels` and `OpenAIModels`
catalogs and the canonical-ID form to emit for each non-Claude provider (we recommend the
catalog ID itself for non-Claude, short alias for Claude).
