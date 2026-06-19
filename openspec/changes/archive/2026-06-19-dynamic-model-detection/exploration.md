# Exploration — dynamic-model-detection (Slice 3)

## Project Type

Not-web. Go CLI/TUI (Cobra-style `cmd/`, Bubbletea TUI in `internal/tui/`). No web frontend; group E (Playwright) does not apply.

## Goal (restated)

Make archon's model choices reflect what is actually available on the user's machine, instead of the hardcoded static catalogs in `internal/config/model.go`. The TUI Models tab (ctrl+n/p cycling, the "Static: ..." hint) and `KnownModels` advisory validation both consume `StaticModels()`. Free-form entry MUST keep working.

## Current State (real code, file:line)

### Static catalogs — `internal/config/model.go`
- `ClaudeModels` (`model.go:18-22`): `claude-opus-4-8`, `claude-sonnet-4-6`, `claude-haiku-4-5`.
- `OpencodeModels` (`model.go:28-37`): `deepseek-v4-flash`, `deepseek-v4-pro`, `glm-5`, `glm-5.1`, `kimi-k2.5`, `kimi-k2.6`, `qwen3.6-plus`, `qwen3.7-plus`.
- `StaticModels()` (`model.go:42-47`): concatenates `ClaudeModels` then `OpencodeModels`. **Confirmed single seam.**
- `KnownModels` (`model.go:52-58`): built by iterating `StaticModels()` — so it derives from the same seam.
- `Validate` (`model.go:135-146`): advisory only; unknown models accepted with a warning. Returns "" for any `KnownModels` hit OR any value `NormalizeModel` resolves to a Claude family.
- `NormalizeModel` (`model.go:98-114`): token-based Claude-family matcher (opus/sonnet/haiku). Independent of `StaticModels()`. **Untouched by this slice — the Slice 1–2 seam holds.**

> Note: the slice brief mentions `GeminiModels`/`OpenAIModels` and an `OpencodeModels`-via-`StaticModels()` shape. In the current tree those Gemini/OpenAI vars **do not exist** — the catalog was already reduced to Claude + Opencode. The single-seam intent is intact: only `StaticModels()` feeds the TUI and `KnownModels`.

### Static catalog vs. reality (verified on this machine)
`opencode models opencode-go` on this machine returns:
`deepseek-v4-flash, deepseek-v4-pro, glm-5.1, glm-5.2, kimi-k2.6, kimi-k2.7-code, mimo-v2.5, mimo-v2.5-pro, minimax-m2.7, minimax-m3, qwen3.6-plus, qwen3.7-max, qwen3.7-plus`.
The hardcoded `OpencodeModels` already **drifts**: it lists `glm-5`, `kimi-k2.5`, `qwen3.7-plus` while the live catalog has `glm-5.2`, `kimi-k2.7-code`, `minimax-m3`, etc. This is concrete evidence the static list goes stale — the core motivation for this slice.

### Agent detection — `internal/agent/` + `internal/initcmd/init.go`
- `agent.Detect(fsys fs.FS)` (`internal/agent/detect.go:23-40`): probes **project directories only** — `.opencode`, `.claude`, `.agents`, `.codex` (in that priority order, `detect.go:8-16`). It does **NOT** probe PATH. Returns first match as `Agent`, all matches as `Dirs`.
- `agent.Resolver.Resolve` (`internal/agent/resolve.go:15-42`): picks the agent given a `--agent` flag or prompts on ambiguity.
- `initcmd.detectAgent` (`init.go:127-144`): wraps `agent.Detect` over `os.DirFS(projectDir)`; explicit `--agent` flag wins.
- The TUI Agent tab uses a **fixed list** `availableAgents = {opencode, claude, codex, agents}` (`internal/tui/agent_tab.go:11`) and lets the user pick + run init; it does not detect installed CLIs either.

**Conclusion: today archon detects agents by PROJECT FILES, never by PATH. No `os/exec` / `exec.LookPath` exists anywhere in non-test code (verified by grep).**

### Opencode config / auth (verified on disk)
- `~/.config/opencode/opencode.json` here: just `{"$schema": ..., "autoupdate": false}` — no `provider`/`model` keys present. So a user config does NOT reliably reveal models.
- `~/.local/share/opencode/auth.json` here: `{"opencode-go": {"type": "api", "key": "..."}}` — reveals **authenticated providers** by key. `opencode auth list` confirms "OpenCode Go".
- `opencode models` (no arg) returned both `opencode/*` (free, unauthenticated provider) and `opencode-go/*` lines — so `opencode models` enumerates the **models.dev catalog** for known providers, NOT strictly the authenticated set. Auth.json is a better signal for "which providers can I actually call".

### The crux — what is realistically detectable
- **Agent CLIs ARE detectable via PATH.** Verified: `opencode`, `claude`, `codex`, `gemini` all resolve via `command -v` here. `exec.LookPath` is the idiomatic Go equivalent.
- **`opencode models [provider]` is a REAL enumeration subcommand** (verified via `opencode models --help` and live run). It prints `provider/model` lines, supports `[provider]` filtering and `--refresh` (pulls from models.dev). This is a genuine, machine-local source of the real Opencode catalog.
- **Other CLIs do NOT enumerate models.** `claude --help`, `codex --help`, `gemini --help` only expose `--model <model>` for *setting* a model (aliases like `sonnet`/`opus`). No list subcommand. Claude/Gemini/OpenAI models are remote API/subscription models — there is **nothing "installed" to detect**; the curated Claude family list must stay static.

## Affected Areas

- `internal/config/model.go` — where a detector would replace/augment `StaticModels()` (or feed it). Keep `NormalizeModel`, `Validate`, `KnownModels` derivation intact.
- `internal/tui/models_tab.go` — `cycleStaticModel` (`models_tab.go:152-170`) and the `view` "Static:" hint (`models_tab.go:210`) both call `config.StaticModels()`. If detection is async/expensive, these need a resolved list passed in (constructor already takes `*config.Config`, `models_tab.go:39`).
- `internal/agent/` — natural home for a new PATH-based CLI detector (e.g. `agent.DetectCLIs()`), parallel to the existing project-file `Detect`.
- Possibly a new small `internal/models/` (or function in `config`) that shells out to `opencode models` and parses lines.

## Approaches (honestly scoped)

### A — Detect installed agent CLIs via PATH; gate/order the curated catalogs
Add `exec.LookPath` checks for `opencode`/`claude`/`codex`/`gemini`. Keep the curated lists, but **filter/reorder** so present agents' models surface first (e.g. hide Opencode models if `opencode` isn't on PATH).
- Pros: trivial, deterministic-ish, no external coupling to subcommand output, fully offline.
- Cons: still ships stale curated lists (the drift problem remains for Opencode). PATH presence ≠ provider authenticated.

### B — Read opencode config/auth to surface configured providers
Parse `~/.local/share/opencode/auth.json` (authenticated providers) and/or `~/.config/opencode/opencode.json`.
- Pros: tells us which providers the user can actually call; small file reads, no subprocess.
- Cons: gives *providers*, not *models* (the config here has no `model` keys). Paths are platform/install-specific (`~/.local/share` vs XDG vs macOS). Best as a **signal**, not the model source.

### C — Query `opencode models [provider]` for the live model list
Shell out to `opencode models` (optionally `opencode models opencode-go`), parse `provider/model` lines, strip the `provider/` prefix for display, dedupe.
- Pros: **the only source of the REAL, current Opencode catalog**; fixes drift directly. Verified to exist and work. Supports per-provider filtering.
- Cons: introduces the repo's FIRST `os/exec` dependency; output format is an external contract that could change; latency + needs network for `--refresh`; non-deterministic in tests (must be injectable/mockable). Only covers Opencode — Claude stays static.

### D — Hybrid (RECOMMENDED): detect agents via PATH → for Opencode, enumerate via `opencode models`; Claude stays curated; static lists as fallback; free-form always
1. `exec.LookPath` to learn which agent CLIs are present (cheap, offline).
2. If `opencode` is present, call `opencode models opencode-go` (timeout-bounded) to get the live Opencode catalog; on any error/timeout fall back to the curated `OpencodeModels`.
3. Claude family list stays the curated `ClaudeModels` (nothing to enumerate).
4. Compose into the same ordered shape `StaticModels()` returns, so the TUI and `KnownModels` consume it unchanged.
5. Free-form entry untouched.

## Recommendation

**Approach D (hybrid), built behind the existing `StaticModels()` seam**, with C's enumeration as the value-add and A/curated as the always-available fallback.

Concretely:
- Add a PATH detector in `internal/agent/` (e.g. `DetectCLIs() map[string]bool` via `exec.LookPath`) — the project's first, narrowly-scoped `os/exec` use.
- Add a resolver (e.g. `config.ResolveModels(detector)` or a new `internal/models` package) that: starts from curated `ClaudeModels`; if `opencode` is on PATH, replaces curated `OpencodeModels` with parsed `opencode models opencode-go` output, else keeps curated. Keep the pure `StaticModels()` as the offline fallback and as the seed for `KnownModels`.
- The TUI passes the resolved list into `newModelsTabState` (it already takes `*config.Config`); `cycleStaticModel` and the hint line read the resolved slice instead of calling `config.StaticModels()` directly.
- Detection runs ONCE at TUI open (and/or init), result cached on the model — NOT per keystroke.

This is genuinely feasible, low-risk (curated fallback guarantees the TUI never breaks offline), preserves free-form entry, and keeps `NormalizeModel`/`Validate` untouched. It does NOT over-promise enumerating Claude/Gemini/OpenAI models, which are not locally enumerable.

## Risks

- **Over-promising "model detection":** only Opencode is truly enumerable; Claude/Gemini/OpenAI stay curated. Naming/UX must not imply full auto-detection. Mitigate with a clear "detected vs. curated" label.
- **First `os/exec` dependency:** new attack/maintenance surface. Must bound with `context` timeout, ignore stderr noise, never fail the TUI on subprocess error.
- **External CLI output contract:** `opencode models` line format (`provider/model`) could change across opencode versions. Parse defensively; fall back to curated on unexpected output.
- **Cross-platform PATH:** `exec.LookPath` handles `.exe` on Windows, but install locations differ; keep it advisory.
- **Test determinism:** detection is env-dependent. Must inject the detector/lister (interface or func field) so tests use a fake; never shell out in unit tests. Matches existing `Detect(fsys fs.FS)` injection style and `Prompter` interface convention.
- **Auth vs. catalog mismatch:** `opencode models` lists catalog models even for unauthenticated providers (saw `opencode/*` free lines). If we want "callable" models, cross-reference `auth.json` — adds complexity; likely out of scope for v1.
- **Latency/network:** `--refresh` hits models.dev; default (cached) read is fast. Use cached read, never `--refresh` synchronously on the UI thread.

## Open Questions / Product Decisions (for the user)

1. **Scope of detection:** agents-only (Approach A — just filter/order curated lists by which CLIs exist), OR also enumerate Opencode models live via `opencode models` (Approach C/D)? D is recommended but is more work and adds `os/exec`.
2. **Filter vs. reorder:** when an agent's CLI is absent, HIDE its models, or just deprioritize them (still selectable/free-form)? Hiding is cleaner; reordering is safer if detection is wrong.
3. **Where does detection run:** at `init` (bake resolved list into config?), at TUI open (live each session), or both? Recommendation: TUI open (live), since installed CLIs change over time; do NOT freeze into config.
4. **Authenticated-only?** Should we cross-reference `auth.json` to show only providers the user can actually call, or show the full catalog `opencode models` returns? Recommend: full catalog for v1 (simpler), revisit later.
5. **Claude models:** confirm Claude stays a curated static list (no local enumeration exists). Recommend yes.
6. **Failure UX:** when `opencode` is absent or the subcommand fails, silently fall back to curated lists (recommended), or surface a hint that detection failed?

## Readiness

**Blocked on product decisions** — primarily Q1 (agents-only vs. live model enumeration) and Q2 (filter vs. reorder), which determine the size and shape of the change. The technical seam is confirmed (`StaticModels()` is the single consumer; `opencode models` is a verified live source; `exec.LookPath` is viable). Once Q1/Q2 (and ideally Q3) are answered, this is **Ready for Proposal**.
