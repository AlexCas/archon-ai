# Design: Weave Security Baseline into the SDD Flow

## Technical Approach

Realize the 6 capabilities by mirroring the tested `playwright` opt-in pattern across
every layer: Go config struct + `Clone()` → CLI get/set → `--security` init flag →
TUI tab → a new `_shared` reference module → conditional `security.enabled` hooks in
five phase skills, plus an end-to-end `@security` Gherkin tag (authored in spec,
checked in verify, gated in judge). Default OFF; the gate's absence yields zero
behavior change. Profiles are restricted to `cli` and `web` (`llm`/`agentic` deferred).

## Architecture Decisions

| Decision | Choice | Rationale / Alternative rejected |
|---|---|---|
| Activation surface | Full Playwright parity (Approach B) | `setConfigValue` errors on unknown keys (`config.go:189`), so skill-only (A/C) is functionally blocked without the struct anyway. |
| Profile validation site | Validate in `setConfigValue` (`config.go`) + TUI selector constrained to `cli`/`web`; NOT in loader | Matches existing pattern — `Playwright` has no loader validation. Loader stays lenient (forward-compat); the CLI is the typed entry point. Init only ever writes valid values. |
| Default | No pre-seed; zero value `Enabled:false`, `Profile:""` | Opposite of `Judge` (which pre-seeds `true` at `config.go:69`). Opt-in semantics. |
| `web` profile word budget | Reference OWASP Top 10 categories by name compactly (one line each), not expanded prose | Resolves the spec's "module bloat" risk while satisfying `security-baseline-module` web coverage. |
| Embed | No `embed.go` change | `//go:embed ... all:_shared` (`skills/embed.go:5`) already ships every `_shared` file. |

## Data Shapes

```go
// internal/config/config.go — sibling of Playwright (config.go:31-35)
type Security struct {
    Enabled bool   `yaml:"enabled"`
    Profile string `yaml:"profile,omitempty"` // "cli" | "web"
}
// Config struct (config.go:43): add field after Playwright (line 50):
//   Security Security `yaml:"security"`
// Clone() (config.go:86): add `Security: c.Security,` (value type — plain copy).
```

Profile validation error contract (`setConfigValue`):
`invalid profile %q for security.profile (supported: cli, web)`.

## File Changes

| File | Action | Change (anchor) |
|---|---|---|
| `internal/config/config.go` | Modify | Add `Security` struct (~:36); `Config.Security` field (:50); `Clone()` copy (:95). |
| `internal/config/config_test.go` | Modify | Extend `TestConfig_CloneRoundtrip` (:222) with populated `Security{Enabled:true,Profile:"web"}`; add load-default + Clone assertions. |
| `cmd/archon/config.go` | Modify | `setConfigValue` (:157-168): `security.enabled` (parseBool), `security.profile` (validate cli/web). `getConfigValue` (:197-204): both. Update both key-list strings (:189,:213). |
| `cmd/archon/main.go` | Modify | `securityFlag bool` (:80); `opts.Security=securityFlag` (:166); register `--security` flag (~:197). |
| `internal/initcmd/init.go` | Modify | `Options.Security bool` (:27); `buildConfig` param + `Security: config.Security{Enabled: security}` (:240-242); callsite (:85). |
| `internal/tui/security_tab.go` | Create | Mirror `playwright_tab.go`: `enabled` toggle + `profile` selector (cli/web cycle, not free text); `view`, `update`, `applyToConfig`, `setWidth`. |
| `internal/tui/model.go` | Modify | `SecurityTab` iota after `PlaywrightTab` (:26); `securityTab` field (:47); init (:106), setWidth (:125), update (:165), view (:290), apply (:328), reload (:195); add `"Security"` to tab names (:263). |
| `skills/_shared/security-baseline.md` | Create | Profile-scaled OWASP checklist (see below). |
| `skills/sdd-propose/SKILL.md` | Modify | Rules (:200): security-risk row hook. |
| `skills/sdd-spec/SKILL.md` | Modify | Tag guidance (:110) + Rules tag list (:293): `@security` abuse cases. |
| `skills/sdd-tasks/SKILL.md` | Modify | Phase 4 template (:196) + Rules (:254): scanner task. |
| `skills/sdd-verify/SKILL.md` | Modify | Hard Rules (:43): `@security` coverage check. |
| `skills/harness-judge/SKILL.md` | Modify | Opt-in gates list (:28): `@security` gate. |

## security-baseline.md Structure

Non-invocable (no frontmatter `name`, referenced by path). Sections:
`## Profile: cli` (argument/path injection, secret handling, dependency integrity) and
`## Profile: web` (= cli + a compact "OWASP Top 10" list naming each category:
broken access control, cryptographic failures, injection, insecure design, security
misconfiguration, vulnerable components, auth failures, integrity failures, logging
failures, SSRF — one control line each). No `llm`/`agentic`. Risk-taxonomy table for
sdd-propose. Target ~120 lines.

## Skill Hook Wiring (`@security` end-to-end)

Each phase edit uses the conditional phrasing pattern that mirrors `playwright.enabled`:
- **sdd-propose** (:200): "If `security.enabled`, load `skills/_shared/security-baseline.md` and add a mandatory Security Risk row to the Risks table."
- **sdd-spec** (:110,:293): add `@security` to the tag list; "If `security.enabled`, derive ≥1 `@security` abuse-case scenario per MUST requirement; prohibitions use RFC 2119 `MUST NOT`." (authors the tag)
- **sdd-tasks** (:196,:254): "If `security.enabled`, emit a tool-agnostic `@security` task running SAST + secret + dependency scans; fail CI on HIGH/CRITICAL."
- **sdd-verify** (:43): "If `security.enabled`, confirm each `@security` scenario maps to a covering test/scanner; report gaps as CRITICAL." (checks the tag)
- **harness-judge** (:28): "Security is OPT-IN: read `security.enabled`; treat unresolved `@security` CRITICAL gaps as a failing gate." (gates the tag)

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | `Clone()` roundtrip | Extend `TestConfig_CloneRoundtrip` with `Security` fields. |
| Unit | Default-off load | New: config with no `security:` block → `Enabled:false`, `Profile:""`. |
| Unit | Profile validation | `setConfigValue` accepts `cli`/`web`; rejects `llm`/`agentic`/garbage with the error contract. |
| Unit | CLI get/set | set then get roundtrip for `security.enabled` + `security.profile`. |
| Unit | Init flag | `--security` → emitted config `security.enabled:true`; without → `false`. |
| TUI | Tab toggle/select | Optional teatest of `applyToConfig` writing `enabled`/`profile`. |

## Slice Boundaries (chained PRs)

- **Slice 1 — Go foundations (~100 ln):** `config.go`, `config_test.go`, `cmd/archon/config.go`, `main.go`, `init.go`. Ships the config gate + CLI + init flag end to end.
- **Slice 2 — Skill layer (~200 ln):** `security-baseline.md` + the five SKILL.md hooks. Inert until Slice 1's gate is on; delivers the `@security` flow.
- **Slice 3 — TUI (~130 ln):** `security_tab.go` + `model.go` wiring. Pure UX, depends on Slice 1's struct.

## Migration / Rollout

No migration. Revert PR(s) to roll back; existing projects unaffected (default false).

## Open Questions

- [ ] TUI profile control: cycle-selector (recommended, matches restricted enum) vs free-text input — defaulting to cycle to enforce cli/web at the UI layer.
