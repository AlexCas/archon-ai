## Exploration: Opencode "archon-leader" custom mode + configurable leader model (Slice 2)

### Project Type
**Web testing**: not-web

This is a Go CLI/TUI (`archon`) built on Cobra-style command wiring plus a Bubbletea TUI (`internal/tui`). No web framework, no `package.json`, no browser surface, no E2E tooling. Playwright must stay disabled.

### Current State

**What `archon init` writes for the opencode agent today (CONFIRMED from real code):**
- `internal/initcmd/init.go` `Run()` does exactly five things: guard/abort if the orchestrator file exists, `ensureAgentDir`, `refreshSkills`, `buildConfig` + `cfg.Save()`, write the rollback manifest, `writeTemplate`, `createOpenSpecDir`.
- For opencode the orchestrator file is `AGENTS.md` (`templateFilePath`, init.go:161-166 — only `claude` gets `CLAUDE.md`).
- `refreshSkills` (refresh.go) extracts embedded skills to the machine-wide `~/.config/opencode/skills/` and symlinks them into the project `.opencode/skills/` (`resolveProjectSkillsDir`, init.go:168-181).
- **CONFIRMED: archon writes ONLY the orchestrator markdown (`AGENTS.md`) + skill links + `.archon/config.yaml` + `.archon/rollback.json` + the `openspec/` scaffold. It NEVER writes or merges any opencode config** — there is no reference to `opencode.json`, `~/.config/opencode/opencode.json`, custom modes, or a `"mode"`/`"model"`/`"prompt"` JSON shape anywhere in `internal/`. The only `~/.config/opencode/` touch is the skills subdir. The agent detector (`internal/agent/detect.go`) only *reads* `.opencode/` to detect the agent; it does not parse its contents.

**Where a new "write/merge opencode mode" step would slot in:** inside `Run()` in init.go, after `writeTemplate` (line 96-98) and before/after `createOpenSpecDir`. It must be gated on `agentName == "opencode"` (a no-op for claude/agents/codex). It should add its created/modified paths to `buildRollbackManifest` (init.go:225-239) so rollback stays truthful, and — critically — it must be additive/merge-safe rather than a clobbering write (see Risks).

**Config model (CONFIRMED):** `internal/config/config.go` `Config.Models` is a `ModelConfig` = `{Default string; Phases map[string]string}` (`internal/config/model.go:9-12`), serialized under `models:` in `.archon/config.yaml`. There is no "leader" concept anywhere. `Config.Clone()` (config.go:86-104) hand-copies every field — a `TestConfig_CloneRoundtrip` test (config_test.go:181) fails loudly via `reflect.DeepEqual` if any new field is added to the struct but not to `Clone`. Any new leader field MUST be added to `Clone` and to that test's fixture.

**Note on Slice 1 / PR #40:** The prompt said the multi-provider `NormalizeModel` (PR #40) "may or may not be present." It IS present and merged: `internal/config/model.go` already has the cross-provider `providerFamilies` table, `NormalizeModel`, `StaticModels()` (Claude → Gemini → OpenAI → Opencode Go), and `Validate`. So the leader model value can and should reuse `config.NormalizeModel`/`config.Validate` exactly like phase models do — no dependency risk remains.

**TUI (CONFIRMED):** `internal/tui/model.go` wires five tabs (`ModelsTab`, `JudgeTab`, `MutationTab`, `PlaywrightTab`, `AgentTab`) via the `Tab` iota + `tabCount`. Each tab is a `*TabState` with the same contract: `newXTabState(cfg)`, `update(msg)`, `view(w,h)`, `applyToConfig(cfg)`, `setWidth(w)`. Tabs are registered in five places: the `Tab` const block, the `Model` struct fields, `NewModel`, the `WindowSizeMsg`/`Update` switch, `agentInitDoneMsg` rebuild, `renderTabs` (the `tabs []string` slice + headers), and `renderTabContent`. `saveConfig` (model.go:304) clones the config, calls every tab's `applyToConfig`, saves, then `regenerateTemplate` re-renders `AGENTS.md`/`CLAUDE.md` from `config.ResolvePhaseModels`. The models tab (`internal/tui/models_tab.go`) is a vertical list of `textinput.Model`s: index 0 = default, indices 1..8 = phases, with `ctrl+n/ctrl+p` cycling `config.StaticModels()` and live `config.Validate` warnings.

### Affected Areas
- `internal/config/config.go` — add the leader model field to `Config` (or to `ModelConfig`) and MUST add it to `Clone()`.
- `internal/config/model.go` — natural home for a `LeaderModel` field on `ModelConfig` and any leader-normalization helper; already exposes `NormalizeModel`/`Validate`/`StaticModels`.
- `internal/config/config_test.go` — extend the `TestConfig_CloneRoundtrip` fixture (and roundtrip test) for the new field.
- `internal/initcmd/init.go` — add an opencode-only `writeOpencodeMode` step in `Run()` after `writeTemplate`; gate on `agentName == "opencode"`; thread the leader model from `cfg`; register any created/merged paths in `buildRollbackManifest`. `buildConfig` may need a new `ModelLeader` field on `Options`.
- `internal/initcmd/templates.go` (or a new `internal/initcmd/opencode_mode.go`) — JSON shape/merge logic for the `archon-leader` mode block (kept out of init.go for clarity).
- `internal/tui/models_tab.go` — add a "Leader model" input (opencode only) following the existing `textinput` + `Validate` pattern; OR a dedicated control elsewhere (see Approaches).
- `internal/tui/model.go` — if the leader field lives outside the models tab, register a tab/control in all the wiring sites; if it lives in the models tab, no model.go change beyond the existing `applyToConfig` call. Also: `saveConfig`/`regenerateTemplate` must (for opencode) re-write the opencode mode file so the TUI stays in sync with init.
- `internal/initcmd/refresh.go` / `update.go` — `archon update` deliberately preserves user config and never rewrites the orchestrator. Decide whether `update` should also (re)write the opencode mode; current bias is that it should NOT (update is skill-only). Note explicitly so design doesn't regress this invariant.
- `README.md` — the model docs and TUI tab list will need a leader-mode note (docs only).

### Approaches

#### A. Where the leader model lives in config

1. **New field on `ModelConfig` (e.g. `Leader string` → `models.leader`)** — reuse the existing models section/struct.
   - Pros: cohesive with `Default`/`Phases`; one `models:` block; trivial TUI placement (next to default); reuses `NormalizeModel`/`Validate`; smallest serialization surface.
   - Cons: `ModelConfig` starts carrying an opencode-only concern; slightly muddies "models for SDD phases" vs "model for the leader mode."
   - Effort: Low.

2. **New top-level config section (e.g. `opencode: { leader_model: ... }` or `leader: { model: ... }`)** — separate the opencode-only concern.
   - Pros: clearly scopes the field as opencode-only / leader-only; room to grow (e.g. mode name, prompt path) without overloading `models`.
   - Cons: more YAML surface, another struct + `Clone` entry + roundtrip coverage; a second place the TUI must read/write.
   - Effort: Medium.

#### B. Global vs project opencode config

1. **Project-level (`<project>/.opencode/opencode.json` or `<project>/opencode.json`)** — write the mode into the project the user just ran init in.
   - Pros: matches archon's per-project model (`.opencode/skills/`, `AGENTS.md`, `.archon/config.yaml` are all project-scoped); rollback/`.archon/rollback.json` can track it; the AGENTS.md prompt file the mode references lives in the same project; no surprise edits to the user's global machine config.
   - Cons: must confirm Opencode actually honors a project-level custom-mode file and its exact filename/location (OPEN QUESTION).
   - Effort: Medium (mostly the merge logic + schema confirmation).

2. **Global (`~/.config/opencode/opencode.json`)** — write the mode into the user's machine-wide opencode config.
   - Pros: skills already live globally under `~/.config/opencode/skills/`, so there is precedent for a global touch.
   - Cons: edits a shared machine file that may contain unrelated user modes/providers → much higher merge/clobber risk; not per-project; a relative `AGENTS.md` prompt reference would be ambiguous from a global file (OPEN QUESTION on how a global mode references a project prompt).
   - Effort: Medium-High (merge safety is harder on a shared file).

#### C. Merge strategy into an existing opencode config

1. **Read-modify-write JSON merge, additive on the `mode`/`modes` key only** — load existing JSON (if any) into a generic `map[string]any`, ensure the modes container exists, set ONLY the `archon-leader` key, re-marshal, atomic temp+rename (the pattern `config.Save`/`writeTemplate` already use).
   - Pros: idempotent; preserves unknown user keys and other modes; safe to re-run on `archon init`.
   - Cons: needs the exact JSON shape (OPEN QUESTION); must preserve formatting/ordering reasonably; must handle "file absent," "file present but no modes," "archon-leader already present" (overwrite our own block only).
   - Effort: Medium.

2. **Write a fresh `opencode.json` only when absent; otherwise skip/warn** — never touch an existing file.
   - Pros: zero clobber risk.
   - Cons: users who already have an `opencode.json` (likely) never get the leader mode → defeats the feature for the common case.
   - Effort: Low but inadequate.

### Recommendation

- **Config (A1):** add `Leader string` to `ModelConfig` serialized as `models.leader`. Lowest-friction, reuses the existing `NormalizeModel`/`Validate`/`StaticModels` machinery and the models tab. Remember to extend `Config.Clone()` and the `TestConfig_CloneRoundtrip` fixture (hard invariant). If design later wants mode-name/prompt-path knobs, promote to a small section then.
- **Location (B1):** write the mode into a **project-level** opencode config so it matches archon's per-project footprint, lets rollback track it, and keeps the `AGENTS.md` prompt reference local. Pending confirmation of Opencode's actual project-config filename/path.
- **Merge (C1):** additive read-modify-write that only sets the `archon-leader` mode key, atomic temp+rename, idempotent, opencode-only. Add the written path to `buildRollbackManifest`.
- **TUI:** add a single "Leader model" `textinput` to the **models tab**, shown/active only for the opencode agent, reusing `config.Validate` warnings and `ctrl+n/ctrl+p` cycling. On save, for opencode, `regenerateTemplate` (or a sibling) must also re-run the additive mode merge so the TUI and init stay consistent. A whole new tab is overkill for one field.
- **Claude/agents/codex:** explicitly NO leader mode. Opencode is the only agent with a mode concept; for claude/agents/codex the orchestration stays driven by `CLAUDE.md`/`AGENTS.md`. The leader field/UI should be inert (and ideally hidden) for non-opencode agents.
- **`archon update`:** keep it skill-only; do NOT have update rewrite the opencode mode (preserves the existing "update never rewrites orchestrator/user config" invariant). Surface this decision in the proposal.

### Risks
- **Opencode external schema uncertainty (HIGH).** Opencode is an external tool; its custom-mode config format (file location, JSON shape, how a mode declares `model`/`prompt`, how a mode references a prompt file like `AGENTS.md`) is NOT in this repo and cannot be verified from code. Designing the writer/merger before confirming this risks shipping a mode Opencode ignores. Must be confirmed against Opencode's actual docs before design.
- **Merge safety / clobber (HIGH).** Writing into an existing user `opencode.json` can destroy unrelated user modes, providers, or keys if done as a blind overwrite. The writer must be a strictly additive, idempotent read-modify-write touching only the `archon-leader` key, with atomic temp+rename and rollback registration.
- **Opencode-only scope (MEDIUM).** The feature applies ONLY to the opencode agent. Every new code path (init step, TUI field, save regen) must be a clean no-op for claude/agents/codex, and tests must cover the non-opencode path doing nothing.
- **Clone/roundtrip invariant (MEDIUM).** A new config field not added to `Config.Clone()` is silently dropped on every `archon update`. The roundtrip test guards this but only if the fixture is updated.
- **TUI ↔ init drift (MEDIUM).** If init writes the mode but the TUI save path does not (or vice versa), the leader model in `.archon/config.yaml` and the written opencode mode can diverge. Both write paths must run the same merge logic.
- **Prompt-path coupling (LOW-MEDIUM).** The mode must point at the orchestrator prompt (`AGENTS.md`). Whether Opencode resolves that path relative to the config file, the project root, or an absolute path is an OPEN QUESTION that affects whether a global vs project config even works.

### Open Questions (must be answered before propose/design)
1. **Opencode custom-mode config location:** Does Opencode read custom modes from a project file (`<project>/.opencode/opencode.json`? `<project>/opencode.json`?) or only from the global `~/.config/opencode/opencode.json`? (Drives Approach B.)
2. **Exact JSON shape of a custom mode:** Is it `{"mode": {"archon-leader": {"model": "...", "prompt": "..."}}}` or `{"modes": [...]}` or another shape? What are the exact key names for the model and the prompt? (Drives the writer/merger.)
3. **How a mode references a prompt file:** Does the `prompt` value take a file path (relative/absolute) pointing at `AGENTS.md`, or an inline string? How is a relative path resolved? (Drives B1 vs B2 viability.)
4. **Model identifier format expected by Opencode modes:** Does Opencode expect a provider-prefixed id (e.g. `anthropic/claude-...`, `opencode/glm-5`) or a bare id? Should the leader value be passed through `config.NormalizeModel` and/or re-prefixed for the mode file? (Affects what we write vs what we store.)
5. **Should `archon update` also (re)write the opencode mode,** or stay strictly skill-only? (Recommendation: stay skill-only — confirm with the user.)
6. **TUI placement preference:** leader-model input inside the existing Models tab (recommended) vs a dedicated control/tab. (Product/UX choice for the user.)

### Ready for Proposal
**No — blocked on open questions 1–4** (Opencode's external mode schema/location/prompt-reference/model-format). These cannot be answered from the repo and must come from Opencode's docs or the user before design can specify the writer/merger. Questions 5–6 are product decisions the user can settle quickly. The archon-side design (config field A1, additive merge C1, opencode-only gating, TUI field, rollback registration) is well understood and ready as soon as the Opencode schema is confirmed.

---

## Orchestrator Addendum — External Open Questions RESOLVED (from Opencode docs, 2026-06-18)

Source: https://opencode.ai/docs/agents/ and https://opencode.ai/docs/config/ (the older `/docs/modes/` now 404s — "modes" were renamed to "agents").

1. **Config location & precedence (Q1):** Opencode merges config from multiple sources; the relevant writable ones are the **global** `~/.config/opencode/opencode.json` and the **project** `opencode.json` (project root). Project overrides global. Configs are MERGED (non-conflicting keys preserved) — which supports an additive write. `.opencode/` subdirs also contribute (agents/commands/plugins).
2. **Exact JSON shape (Q2):** Custom modes are custom **agents** under the top-level `agent` key:
   ```json
   {
     "agent": {
       "archon-leader": {
         "mode": "primary",
         "description": "...",
         "model": "provider/model-id",
         "prompt": "{file:./AGENTS.md}"
       }
     }
   }
   ```
   `mode` ∈ `primary | subagent | all`. For the leader we want `primary`. Built-in primary agents are **Build** and **Plan**; custom primary agents **auto-appear in the Tab switcher** alongside them — exactly the requested UX.
3. **Prompt reference (Q3):** `prompt` accepts `"{file:./relative/path}"`, resolved **relative to the config file location**. Pointing it at the archon-written `AGENTS.md` works. NOTE: relative resolution means a **project** `opencode.json` next to `AGENTS.md` is the simplest viable placement (favors project-level write).
4. **Model id format (Q4):** Opencode requires **`provider/model-id`** (e.g. `anthropic/claude-sonnet-4-20250514`, and provider-prefixed forms for google/openai/opencode). NOT a bare alias. DESIGN IMPLICATION: the configured leader model must be mapped/stored to a `provider/model-id`. Decide whether `.archon/config.yaml` stores the full `provider/model-id` directly, or stores a catalog id (reusing Slice 1 catalogs) plus a provider→prefix mapping applied when writing the agent. This is the main new design decision for propose/spec.

Remaining for the user (product decisions): Q5 (`archon update` stays skill-only — recommended) and Q6 (TUI placement — Models tab recommended), plus confirming the Q4 mapping approach.

**Revised readiness:** Ready for Proposal — external schema confirmed; only product decisions (Q4 mapping shape, Q5, Q6) remain, to settle at the review gate.
