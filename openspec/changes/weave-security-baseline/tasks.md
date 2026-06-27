# Tasks: Weave Security Baseline into the SDD Flow

## Review Workload Forecast

| Dimension | Value |
|---|---|
| Estimated changed lines | ~330–440 |
| 400-line budget risk | High |
| Decision needed before apply | Yes |
| Chained PRs recommended | Yes |
| Chain strategy | feature-branch-chain |
| Slices | 3 (Go foundations / skill layer / TUI) |

> **Chain strategy: feature-branch-chain.** A tracker branch
> `feature/weave-security-baseline` accumulates the integration and is the only
> branch that merges to main. PR 1 base = tracker; PR 2 base = PR 1 branch;
> PR 3 base = PR 2 branch. Apply starts with Slice 1 (PR 1) only.

---

## Suggested Work Units

| PR | Slice | Files | Est. lines |
|---|---|---|---|
| PR 1 | Go foundations | `config.go`, `config_test.go`, `cmd/archon/config.go`, `main.go`, `init.go` | ~100 |
| PR 2 | Skill layer | `security-baseline.md`, 5 × SKILL.md hooks | ~200 |
| PR 3 | TUI | `security_tab.go`, `model.go` | ~130 |

---

## Phase 1 — Slice 1: Go Foundations (PR 1)

- [x] **[S1-1] Add `Security` struct to `internal/config/config.go`** — define `Security{Enabled bool, Profile string}` as a sibling of `Playwright` (~:36); add `Security Security` field to `Config` struct after `Playwright` (~:50). Covers scenario: *Absent block defaults to disabled*.
- [x] **[S1-2] Extend `Clone()` in `internal/config/config.go`** — add `Security: c.Security` value copy (~:95). Covers scenario: *Clone preserves security fields*.
- [x] **[S1-3] Add `security.*` cases to `cmd/archon/config.go`** — in `setConfigValue` (:157–168): `security.enabled` (parseBool) and `security.profile` (validate `cli`/`web`; reject all else with `invalid profile %q for security.profile (supported: cli, web)`); in `getConfigValue` (:197–204): both keys; update both key-list strings (:189, :213). Covers scenarios: *Setting a valid profile succeeds*, *Setting an invalid profile is rejected*.
- [x] **[S1-4] Wire `--security` flag in `cmd/archon/main.go`** — declare `securityFlag bool` (~:80); pass `opts.Security=securityFlag` (~:166); register `--security` flag (~:197).
- [x] **[S1-5] Add `Security bool` to `Options` and `buildConfig` in `internal/initcmd/init.go`** — `Options.Security bool` (:27); propagate into `buildConfig` param and emit `Security: config.Security{Enabled: security}` (:240–242); update callsite (:85). Covers scenarios: *Init flag enables the gate*, *Init without flag leaves security off*.
- [x] **[S1-6] Extend `TestConfig_CloneRoundtrip` in `internal/config/config_test.go`** — populate `Security{Enabled:true, Profile:"web"}` in the roundtrip fixture (:222); add a separate test asserting load with no `security:` block yields `Enabled:false, Profile:""`. Covers scenarios: *Clone preserves security fields*, *Absent block defaults to disabled*.
- [x] **[S1-7] Add CLI get/set unit tests in `internal/config/config_test.go` (or `cmd/archon/`)** — set then get `security.enabled` and `security.profile`; assert `llm`/`agentic`/garbage are rejected. Covers scenarios: *Setting a valid profile succeeds*, *Setting an invalid profile is rejected*.

---

## Phase 2 — Slice 2: Skill Layer (PR 2, depends on PR 1 merged)

- [x] **[S2-1] Create `skills/_shared/security-baseline.md`** — non-invocable; `## Profile: cli` (argument/path injection, secret handling, dependency integrity); `## Profile: web` (cli controls + compact OWASP Top 10 list, one line per category: broken access control, cryptographic failures, injection, insecure design, security misconfiguration, vulnerable components, auth failures, integrity failures, logging failures, SSRF); risk-taxonomy table for sdd-propose. Target ~120 lines. No `llm`/`agentic` sections. Covers scenarios: *CLI profile surfaces CLI-relevant controls*, *Web profile adds web Top 10 controls*, *Module is reference-only*.
- [x] **[S2-2] Hook `sdd-propose/SKILL.md` (:200)** — add rule: "If `security.enabled`, load `skills/_shared/security-baseline.md` and add a mandatory Security Risk row to the Risks table." Covers scenarios: *Security risk row is emitted when enabled*, *Missing security row is treated as incomplete*, *No security row when disabled*.
- [x] **[S2-3] Hook `sdd-spec/SKILL.md` (:110, :293)** — add `@security` to the tag list; add rule: "If `security.enabled`, derive ≥1 `@security` abuse-case scenario per MUST requirement using RFC 2119 `MUST NOT` in prohibitions." Covers scenarios: *Each MUST requirement gets an abuse case*, *Abuse case describes the prohibited behavior*, *No security tag when disabled*.
- [x] **[S2-4] Hook `sdd-tasks/SKILL.md` (:196, :254)** — add rule: "If `security.enabled`, emit a tool-agnostic `@security` CI task: run SAST, secret detection, and dependency vulnerability scans; fail CI on any HIGH or CRITICAL finding; do not name a specific vendor tool." Covers scenarios: *Scanning task is emitted when enabled*, *Scanning task names no specific tool*, *No scanning task when disabled*.
- [x] **[S2-5] Hook `sdd-verify/SKILL.md` (:43)** — add hard rule: "If `security.enabled`, confirm each `@security` scenario maps to a covering test or scanner; report any gap as CRITICAL." Covers scenarios: *Full coverage passes verify*, *Uncovered abuse case fails verify*.
- [x] **[S2-6] Hook `harness-judge/SKILL.md` (:28)** — add to opt-in gates list: "Security is OPT-IN: read `security.enabled`; treat unresolved `@security` CRITICAL coverage gaps as a failing gate." Covers scenarios: *Judge blocks on uncovered security scenario*, *No security gate when disabled*.

---

## Phase 3 — Slice 3: TUI (PR 3, depends on PR 1 merged)

- [ ] **[S3-1] Create `internal/tui/security_tab.go`** — mirror `playwright_tab.go`; `enabled` bool toggle; `profile` cycle-selector cycling `cli` → `web` → `cli` (not free text, enforces the two-value enum at the UI layer); implement `view`, `update`, `applyToConfig`, `setWidth`.
- [ ] **[S3-2] Wire Security tab into `internal/tui/model.go`** — add `SecurityTab` iota after `PlaywrightTab` (:26); add `securityTab` field (:47); update `init` (:106), `setWidth` (:125), `update` (:165), `view` (:290), `apply` (:328), `reload` (:195); add `"Security"` to tab names (:263).
- [ ] **[S3-3] Optional: add teatest for `security_tab.go` `applyToConfig`** — assert toggle + cycle sets `enabled`/`profile` correctly; assert invalid profile cannot be selected via the cycle.

---

## Cross-Slice: CI Scanning (tool-agnostic, all PRs)

- [x] **[CI-1] @security CI task in task lists** — for any apply that touches security-enabled paths, verify the task list carries a tool-agnostic scan step: run SAST, secret detection, and dependency vuln checks; gate on HIGH/CRITICAL; no vendor name hard-coded. (This is the runtime enforcement of the `tasks-security-scanner` spec across all three PRs.)
