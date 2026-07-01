# Proposal: Weave Security Baseline into the SDD Flow

## Intent

ARCHON has no proactive security guidance in the SDD flow — only a reactive hint in judgment-day. Teams get no spec-level abuse cases, no risk prompt, no CI scanning task. This change weaves OWASP-derived secure-by-design guidance into the EXISTING phases, gated by a new opt-in `security` config block mirroring `playwright`.

## Scope

### In Scope
- `Security{Enabled, Profile}` config block with `cli` + `web` profiles.
- `--security` init flag, `archon config set/get security.*`, TUI Security tab.
- `skills/_shared/security-baseline.md`: profile-scaled OWASP checklist.
- Conditional security hooks in sdd-propose, sdd-spec, sdd-tasks, sdd-verify, harness-judge.

### Out of Scope
- `llm` and `agentic` profiles (deferred to future work).
- Track B — hardening ARCHON's own skill supply chain (see `docs/security/skill-ecosystem-hardening.md`).
- Prescribing specific scanner tools; the CI task stays tool-agnostic.
- Behavior change for existing projects (`security.enabled=false` by default).

## Capabilities

### New Capabilities
- `security-config`: Go `Security` block + CLI get/set + `--security` init flag + TUI tab.
- `security-baseline-module`: `skills/_shared/security-baseline.md`, profile-scaled (`cli`, `web`).
- `propose-security-risk`: mandatory security-risk row in sdd-propose when enabled.
- `spec-security-scenarios`: `@security`-tagged abuse-case Gherkin in sdd-spec.
- `tasks-security-scanner`: tool-agnostic CI scanning task (SAST, secrets, dependency vulns) in sdd-tasks.
- `verify-security-gate`: `@security` coverage check in sdd-verify plus harness-judge gate.

### Modified Capabilities
- None (no existing capability spec covers these skills).

## Approach

Full Playwright parity. The `playwright.enabled` flow is a tested template for every layer (struct → `Clone()` → CLI → init flag → TUI → skill conditional). `setConfigValue` rejects unknown keys, so the Go struct is required — no skills-only shortcut. The new module embeds automatically via `all:_shared`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/config/config.go` | Modified | `Security` struct, `Clone()` copy |
| `cmd/archon/config.go` | Modified | set/get cases + key lists |
| `cmd/archon/main.go`, `internal/initcmd/init.go` | Modified | `--security` flag + `buildConfig` |
| `internal/tui/` | New/Modified | `security_tab.go`, tab wiring |
| `skills/_shared/security-baseline.md` | New | OWASP checklist module |
| `sdd-propose/spec/tasks/verify`, `harness-judge` | Modified | gated security hooks |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Profile values misaligned with OWASP | Med | Validate `cli` against ASVS L1 in sdd-spec |
| Module word bloat | Med | Scope to profile-relevant controls only |
| Change exceeds 400-line budget (~330–440 est.) | High | Chained PRs: Go foundations / skill layer / TUI |
| Missed `Clone()` field | Low | `TestConfig_CloneRoundtrip` fails loudly |

## Rollback Plan

Revert the PR(s). Existing projects are unaffected since `security.enabled` defaults to `false`; no config migration. The embedded module is inert unless the gate is on.

## Dependencies

- None external. Reuses the existing config, init, TUI, and skill-embed infrastructure.

## Success Criteria

- [ ] `archon init --security` and `archon config set security.profile web` work end to end.
- [ ] `Clone()` roundtrip test passes with the new fields.
- [ ] When `security.enabled`, the four phases emit security risk row, `@security` scenarios, scanner task, and coverage gate.
- [ ] Existing projects (gate off) show zero behavior change.
