# Track B — Skill Ecosystem Hardening (Future)

> **Status**: Backlog / not scheduled. Documented as a future, separate track.
> **Scope**: Hardens ARCHON *itself* against the [OWASP Agentic Skills Top 10](https://owasp.org/www-project-agentic-skills-top-10/).
> **Out of scope here**: Secure-by-design woven into the SDD flow — that is Track A
> (`skills/_shared/security-baseline.md` + propose/spec/tasks/verify), which protects the
> software ARCHON *builds*.

## Why this is separate from the SDD flow

Track B does **not** live in the `propose → spec → tasks` phases. Its controls are
properties of how ARCHON is packaged and distributed: they belong to `archon init`, the
skill-embedding pipeline, and CI. They are transparent to the end user running the SDD
flow and protect the *tool*, not the *product the tool builds*.

| | Track A (chosen, in flight) | Track B (this doc) |
|---|---|---|
| Protects | Software the user builds | ARCHON's own skill supply chain |
| Lives in | propose/spec/tasks/verify | `archon init` / build / CI |
| Part of SDD flow? | Yes | No — binary/infra concern |

## OWASP Agentic Skills Top 10 → ARCHON hardening backlog

ARCHON embeds and distributes 24 skills via `archon init`. The following items apply
directly to that distribution machinery.

| AST | Risk | Proposed ARCHON control | Where |
|-----|------|-------------------------|-------|
| **AST01** Malicious Skills | Harmful code embedded in a skill | Merkle-root signing of the embedded skill bundle; verify signature at `init` time | embed pipeline + `archon init` |
| **AST02** Supply Chain Compromise | Trusted skill source becomes an attack vector | Provenance metadata per skill (source, commit, author); record in a manifest checked into the binary | build + manifest |
| **AST03** Over-Privileged Skills | Skills request more than they need | Per-skill permission manifest (declared tools/scopes); validate against a least-privilege schema at load | skill frontmatter + loader |
| **AST04** Insecure Metadata | Dangerous parsing of skill config | Safe YAML parser, strict schema validation, reject unknown/dangerous keys in frontmatter | config/frontmatter loader |
| **AST05** Untrusted External Instructions | Skills pull instructions from unverified sources at runtime | Inventory of any external fetches; pin + checksum referenced content; rescan | skill content audit |
| **AST06** Weak Isolation | Insufficient sandboxing → host access | Document/enforce the sandbox the harness runs phases under; bound the blast radius of apply | runtime/harness |
| **AST07** Update Drift | Uncontrolled skill updates open exploit windows | Version + hash pinning of the embedded skill set; verify on upgrade | release/versioning |
| **AST08** Poor Scanning | Pattern-matching misses sophisticated threats | Multi-tool scan (semantic + secret + behavioral) of skills in CI before release | CI |
| **AST09** No Governance | No inventory / audit visibility | Skill inventory + audit log of which skills ran in a session; emit on demand | telemetry/governance |
| **AST10** Cross-Platform Reuse | Skills ported without security metadata | Carry security metadata in the canonical skill format used across agents (claude/opencode) | skill format |

## Suggested phasing (if/when scheduled)

1. **Inventory + provenance manifest** (AST02, AST09) — cheapest, unlocks the rest.
2. **Permission manifests + schema validation** (AST03, AST04) — least-privilege baseline.
3. **Signing + hash pinning** (AST01, AST07) — integrity of the embedded bundle.
4. **CI scanning pipeline** (AST08) — gate releases.
5. **Isolation hardening + external-content audit** (AST05, AST06) — deeper, runtime-level.

AST10 is mostly satisfied once the canonical skill format carries the metadata produced
by steps 1–3.
