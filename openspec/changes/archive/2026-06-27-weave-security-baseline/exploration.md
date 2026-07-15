# Exploration: weave-security-baseline

## Project Type

**Web testing**: not-web — Go CLI with no browser surface, no package.json, no web framework detected.

---

## Current State

ARCHON has no security guidance woven into the SDD flow. The judgment-day skill lists
"Security: injection, secrets, auth boundaries" as a generic review criterion
(`skills/judgment-day/references/prompts-and-formats.md:19`), but this is reactive
(post-code review), not proactive (spec-level requirements). No `@security` tag, no
security checklist module, no `security` block in config.

Track B (hardening ARCHON's own skill supply chain) is already documented and
explicitly separate — see `docs/security/skill-ecosystem-hardening.md`.

---

## Affected Areas (file:line anchors)

### 1. Config plumbing (Go)

**Config struct** — `internal/config/config.go`

- Struct: `Config` at line 43; all feature flags follow the same `struct + yaml tag` pattern.
- `Playwright` (lines 31–35) is the canonical model to follow: a struct with `Enabled bool` plus optional string fields.
- Default handling: `Judge.Enabled` is pre-seeded to `true` before `yaml.Unmarshal` (line 69) so an absent section means "enabled". Security default must be `false` (opt-in) — no pre-seed needed; absent section → zero value → `Enabled: false`.
- `Clone()` at line 86: **every new Config field MUST be added here** (comment at line 83 is explicit; `TestConfig_CloneRoundtrip` fails loudly if a field is missed). The new `Security` struct must be copied.
- `Save()` at line 106: struct-level yaml marshal; no changes needed there.

**Config CLI set/get** — `cmd/archon/config.go`

- `setConfigValue` at line 152: switch on dot-delimited keys. New cases `security.enabled` and `security.profile` follow the `playwright.*` pattern (lines 157–168).
- `getConfigValue` at line 193: mirror of set. Same cases needed.
- Error message at line 189 (set) and line 213 (get) lists supported keys — both must be updated.

**Init** — `internal/initcmd/init.go`

- `buildConfig` at line 218: constructs `config.Config` from `Options`; `Playwright` is set at line 241 from `opts.Playwright bool`. The `Security` struct must be added here with `Enabled: false` default (the field is already zero-value false, so no extra Options flag is needed unless the user wants an `--security` init flag — likely yes for discoverability, mirroring `--playwright`).
- `Options` struct at line 17: `Playwright bool` at line 27 — `Security bool` field goes here.

**CLI init command** — `cmd/archon/main.go`

- `newInitCmd` at line 75: `playwrightFlag bool` at line 81; `--playwright` flag registered at line 197. New `--security` flag mirrors this. `opts.Playwright = playwrightFlag` at line 165; same for `opts.Security`.

**TUI** — `internal/tui/`

- `internal/tui/model.go:25`: `PlaywrightTab Tab = iota` const; a new `SecurityTab` would be added after `PlaywrightTab` and before `AgentTab`, bumping `tabCount`.
- `internal/tui/playwright_tab.go` is the full reference implementation for a TUI tab (toggle + optional fields). A `security_tab.go` follows the same shape: `enabled` toggle + `profile` string selector.
- `internal/tui/model.go:263`: `tabs := []string{"Models", "Judge", "Mutation Testing", "Playwright", "Agent"}` — add `"Security"`.

### 2. Config emission at init

When `archon init` writes `.archon/config.yaml`, the `security` block appears if and only if `buildConfig` sets it (like `playwright`). The YAML marshaler emits the block from the struct tags. No template changes needed for CLAUDE.md/AGENTS.md — the security block is config-only (parallel to `playwright` which also has no CLAUDE.md mention beyond "When playwright.enabled…").

The `sdd-init` skill (`skills/sdd-init/`) does not directly write `.archon/config.yaml` — it delegates to the Go harness via `archon init`. The init skill would need a prompt update to mention `--security` / `security.enabled` as a preflight group (similar to group E for Playwright).

### 3. Shared-module pattern

All SDD phase skills load shared modules by path. Convention:
- Files live in `skills/_shared/`.
- Skills reference them by explicit path: `skills/_shared/sdd-phase-common.md`, `skills/_shared/openspec-convention.md`, etc.
- The `_shared/SKILL.md` has `disable-model-invocation: true` and `user-invocable: false` — it is a pure reference package.
- Modules are referenced in phase skill rules sections: "Follow **Section B** from `skills/_shared/sdd-phase-common.md`".

`skills/_shared/security-baseline.md` follows this exact convention. It is not invocable. Phase skills reference it conditionally: "If `security.enabled`, load `skills/_shared/security-baseline.md` and apply the checklist for the configured `profile`."

### 4. Gherkin tag + verify selection — the `@web` template to mirror for `@security`

| Step | Where | Detail |
|------|-------|--------|
| Tag authored | `sdd-spec/SKILL.md:110` | "Tag scenarios…`@web` for browser-facing flows that Playwright will exercise." The equivalent instruction for `@security` goes in the same Rule block. |
| Tag authored | `sdd-spec/SKILL.md:293` | Rules list: "Tag scenarios (`@happy`, `@edge`, `@error`, `@web`)" — add `@security` here. |
| Config gate (apply) | `sdd-apply/SKILL.md:164` | "Only if `playwright.enabled: true`…" — new block: "Only if `security.enabled: true`" for emitting an SBOM / running a scanner. |
| Config gate (verify) | `sdd-verify/SKILL.md:43` | "For web projects with `playwright.enabled: true`: confirm Playwright specs were generated for `@web` scenarios." Mirror: "If `security.enabled: true`: confirm each `@security` scenario has a covering test or CI scanner invocation." |
| Config gate (harness-judge) | `harness-judge/SKILL.md:28` | Playwright gate pattern. A `security` gate could follow: "If `security.enabled: true`: confirm the CI scanner ran and passed." |
| Tasks | `sdd-tasks/SKILL.md:196–197` | "Web + playwright.enabled: a task to generate Playwright specs from `@web` Gherkin scenarios." Mirror: "If `security.enabled`: emit a CI scanning task tagged `@security`." |

### 5. Phase hook points

| Phase | Injection point | Current mechanism | Security hook |
|-------|----------------|-------------------|---------------|
| `sdd-propose` | Rules section (line 200) | `Apply any rules.proposal from openspec/config.yaml` | Add: "If `security.enabled`, add a **Security Risk** row to the Risks table (mandatory). Load `skills/_shared/security-baseline.md` for risk taxonomy." |
| `sdd-spec` | Rules section (line 306) | `Apply any rules.specs from openspec/config.yaml` | Add: "If `security.enabled`, derive `@security`-tagged Gherkin abuse-case scenarios for each MUST requirement. Use RFC 2119 `MUST NOT` for prohibitions." |
| `sdd-tasks` | Phase 4 template + Rules (line 250, 254) | `rules.tasks from openspec/config.yaml`; playwright task pattern | Add: "If `security.enabled`, emit a Phase 4 task: `Run {scanner} against changed files; fail CI if any HIGH/CRITICAL finding`." |
| `sdd-verify` | Hard Rules (line 43) | playwright.enabled gate | Add: "If `security.enabled: true`: check that `@security` scenarios map to covering tests. Report missing coverage as CRITICAL." |
| `harness-judge` | Step 3b pattern | Playwright gate after judgment-day | Optional: "If `security.enabled` and a scanner is configured: confirm scanner passed as a gate." |

**`openspec/config.yaml` rules mechanism**: the `rules.*` keys (`rules.proposal`, `rules.specs`, `rules.tasks`) are freeform YAML lists read by each phase skill. Adding security guidance to these lists is a per-project opt-in that works today, but it is not config-gated by `security.enabled`. The new design adds `security.enabled` as the gate, mirroring `playwright.enabled` in `.archon/config.yaml`.

### 6. Size/word budgets

| Phase artifact | Budget |
|---------------|--------|
| `proposal.md` | 450 words (`sdd-propose/SKILL.md:205`) |
| `spec.md` | 650 words (`sdd-spec/SKILL.md:307`) |
| `tasks.md` | 530 words (`sdd-tasks/SKILL.md:256`) |
| `exploration.md` | (no stated budget in sdd-explore SKILL.md) |

Security content added per artifact:
- `proposal.md`: +1 table row to Risks table — trivial, well inside 450 words.
- `spec.md`: +N `@security` scenarios. Each Gherkin scenario is 3–5 lines. For a CLI tool (profile `cli`) expect 2–4 abuse cases per change — adds ~20–40 words, within 650-word budget.
- `tasks.md`: +1–2 tasks in Phase 4 for scanner invocation — adds ~20–30 words, within 530-word budget.
- `security-baseline.md`: standalone shared module, not counted in per-artifact budgets.

---

## Approaches

### Approach A — Pure skill injection (no Go changes)

Add `security-baseline.md` to `skills/_shared/`. Modify the four phase skills to check `security.enabled` from `.archon/config.yaml` (which phases already read for playwright). No new Go struct, no new CLI flag, no TUI tab. Users activate via `archon config set security.enabled true` with the existing config CLI.

- Pros: zero Go changes; ships faster; no new CLI surface.
- Cons: no `--security` init flag (discovery); no TUI tab; `archon config set` must be documented; `Clone()` not affected but the YAML block must still be handled gracefully by the existing YAML decoder (unknown keys are silently dropped by `gopkg.in/yaml.v3` if no struct field exists — this means the config CLI `set` call would need to handle the unknown key; OR the struct field must be added anyway). This approach is actually blocked unless the Go struct is extended, because `setConfigValue` returns an error for unknown keys (line 189).
- Effort: Medium (struct still needed)

### Approach B — Full parity with Playwright (recommended)

Add `Security` struct to `internal/config/config.go`. Wire `--security` init flag and `archon config set security.*` support. Add TUI tab. Add `security-baseline.md` shared module. Add conditional injection to the four phase skills. This is the exact pattern used for Playwright.

- Pros: full UX parity; discoverable; TUI toggle; correct error handling; `archon status` can surface it.
- Cons: more files touched (Go struct + test + CLI + TUI + 5 skill files + 1 new module).
- Effort: Medium-High — estimated 5–7 Go files changed + 5–6 skill files changed + 1 new skill module.

### Approach C — Skill-only, no struct, just `openspec/config.yaml` rules

Only add the `security-baseline.md` module and update `openspec/config.yaml` `rules.*` sections. No `.archon/config.yaml` key; no Go changes. Activation by adding security rules to `openspec/config.yaml` manually.

- Pros: no Go changes.
- Cons: no config gate (security content always injected if rules are present, not gated by `enabled` flag); no `archon config set`; no TUI; inconsistent with playwright pattern; harder to disable per-project.
- Effort: Low — but incomplete design (misses the config-gated activation requirement).

---

## Recommendation

**Approach B**, full Playwright parity.

The `setConfigValue` error on unknown keys makes Approach A functionally blocked without the struct anyway. The Playwright implementation is a complete, tested reference for every layer (struct → Clone → CLI set/get → init flag → TUI tab → skill conditional). Reusing that pattern minimizes design decisions and produces a consistent, discoverable UX.

The `profile` field (`web | llm | agentic | cli`) controls which OWASP-derived checklist section `security-baseline.md` surfaces. For this project (Go CLI) the default profile is `cli`.

---

## Capability Decomposition (suggested for proposal)

| Capability | Description | New spec? |
|-----------|-------------|-----------|
| `security-config` | Go struct + CLI set/get + init flag + TUI tab | New |
| `security-baseline-module` | `skills/_shared/security-baseline.md` shared module (OWASP-derived, profile-scaled) | New |
| `propose-security-risk` | Mandatory security risk row in `sdd-propose` when enabled | Modified |
| `spec-security-scenarios` | `@security`-tagged Gherkin abuse-case scenarios in `sdd-spec` | Modified |
| `tasks-security-scanner` | CI scanning task in `sdd-tasks` when enabled | Modified |
| `verify-security-gate` | `@security` scenario coverage check in `sdd-verify` | Modified |

---

## Risks

| Risk | Likelihood | Notes |
|------|------------|-------|
| `security-baseline.md` word bloat exceeds spec budget when many scenarios | Medium | Mitigate by scoping the module to emit only profile-relevant controls; keep per-scenario content to 3–5 lines |
| TDD coverage gap: new `Security` struct fields + `Clone()` | Low | `TestConfig_CloneRoundtrip` already enforces this loudly; easy to catch |
| Phase skill word budget overflow from security content | Low | Estimated additions are well within existing budgets (see size analysis above) |
| Misalignment between profile values and real OWASP categories | Medium | Validate `cli` profile controls against OWASP ASVS Level 1 before writing the module |
| `archon update` skipping `security-baseline.md` if embed list is hardcoded | Unknown | Need to verify how `skills.FS` embed glob works — if it embeds all files under `skills/`, the new file is picked up automatically |

---

## Open Question (Resolved)

`skills/embed.go:5`: `//go:embed */SKILL.md all:_shared`

The `all:_shared` directive embeds ALL files under `skills/_shared/`, including dotfiles. `skills/_shared/security-baseline.md` is therefore picked up automatically at compile time — no changes to `embed.go` needed.

---

## Size Estimate / Chained PR forecast

| Layer | Files changed | Estimated line delta |
|-------|--------------|----------------------|
| Go: Config struct + Clone + test | `internal/config/config.go`, `config_test.go` | ~30–40 |
| Go: CLI set/get | `cmd/archon/config.go` | ~25–30 |
| Go: Init flag | `cmd/archon/main.go`, `internal/initcmd/init.go` | ~15–20 |
| Go: TUI tab | `internal/tui/security_tab.go` (new), `model.go` | ~120–150 |
| Skills: phase injections | `sdd-propose`, `sdd-spec`, `sdd-tasks`, `sdd-verify`, `harness-judge` | ~60–80 |
| Skills: new shared module | `skills/_shared/security-baseline.md` | ~80–120 |
| **Total** | 10–11 files | **~330–440 lines** |

**400-line budget risk: Medium-High.** The change is borderline. A TUI-less slice (Approach A foundations) could stay under 400 lines; the full TUI tab alone is ~130 lines. Chained PRs are likely warranted:

- **Slice 1**: Go config struct + CLI + init flag (no TUI). ~100 lines.
- **Slice 2**: `security-baseline.md` + phase skill injections. ~200 lines.
- **Slice 3**: TUI security tab. ~130 lines.

---

## Ready for Proposal

Yes. The design direction from the goal statement is validated. All hook points are confirmed with file:line anchors. The capability decomposition and PR forecast are ready to hand to `sdd-propose`.
