---
description: "Security baseline reference: profile-scaled controls for cli and web projects. Non-invocable; referenced by path from phase skills."
disable-model-invocation: true
user-invocable: false
---

# Security Baseline Reference

This module is **not invocable**. It is loaded by path from phase skills when
`security.enabled` is true. Do not invoke it as a standalone skill.

## Default Profile

When `security.enabled` is true and `security.profile` is empty or not set,
skills MUST default to the `cli` profile (minimal baseline). The `web` profile
is only active when `security.profile` is explicitly set to `"web"`.

## Profile: cli

Controls that apply to all CLI and non-web workloads.

| Control | Description |
|---|---|
| Argument injection | All CLI arguments MUST be validated or sanitized before use in shell commands, file paths, or subprocess calls. |
| Path traversal | File path inputs MUST be resolved and constrained to expected directories; `..` sequences MUST be rejected. |
| Secret handling | Secrets (tokens, keys, passwords) MUST NOT be logged, printed to stdout, or stored in plaintext config files. Prefer environment variables or a secrets backend. |
| Dependency integrity | Third-party dependencies MUST be pinned to exact versions (go.sum, package-lock.json, etc.) and verified at build time. |
| Least privilege | Processes MUST request only the permissions required for their task. |
| Error disclosure | Internal error details (stack traces, internal paths) MUST NOT be surfaced to untrusted callers. |

## Profile: web

Includes all `cli` controls above, plus the OWASP Top 10 (2021) web controls:

| # | Category | Minimum Control |
|---|---|---|
| A01 | Broken access control | Enforce authorization on every endpoint; deny by default. |
| A02 | Cryptographic failures | Use TLS for data in transit; use strong algorithms (AES-256, SHA-256+) for data at rest. |
| A03 | Injection | Parameterize all queries and commands; never interpolate user input into SQL, shell, or HTML. |
| A04 | Insecure design | Threat-model new features; include abuse-case scenarios in every spec with security requirements. |
| A05 | Security misconfiguration | Harden defaults; remove unused routes, headers, and services; review config before deploy. |
| A06 | Vulnerable & outdated components | Scan dependencies for known CVEs; fail CI on HIGH/CRITICAL findings. |
| A07 | Identification & auth failures | Enforce MFA where applicable; rotate session tokens on privilege change; limit login attempts. |
| A08 | Software & data integrity failures | Verify supply-chain integrity (signed packages, SBOM); validate deserialized data. |
| A09 | Security logging & monitoring failures | Log authentication events, access control failures, and input validation errors; alert on anomalies. |
| A10 | SSRF | Validate and allowlist all server-side URL destinations; block requests to internal metadata endpoints. |

## Risk Taxonomy (for sdd-propose)

Use this table to fill the mandatory Security Risk row in the proposal Risks table
when `security.enabled` is true.

| Category | Example Risk | Default Likelihood | Suggested Mitigation |
|---|---|---|---|
| Input / injection | Malicious argument or path injected into a shell call | Med | Validate and sanitize all inputs; use parameterized APIs. |
| Secret leakage | API key logged or committed | Med | Audit log calls; use `.gitignore` + secrets scanning in CI. |
| Dependency vulnerability | CVE in a pinned dependency | Med | Dependency scan in CI; pin + regularly upgrade. |
| Auth / access control (web) | Unauthenticated access to protected resource | High | Deny-by-default middleware; integration test for each protected route. |
| Cryptographic failure (web) | Sensitive data transmitted over plain HTTP | High | Enforce HTTPS; HSTS header; test TLS configuration. |
| SSRF (web) | Server fetches attacker-controlled URL | Med | Allowlist outbound destinations; block metadata endpoints. |

When writing the Security Risk row, select the category most relevant to the
change being proposed, or list multiple rows if the change crosses categories.
The row format for the proposal Risks table is:

```markdown
| Security — {category} | {Likelihood} | {Mitigation from taxonomy or custom} |
```
