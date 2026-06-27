## Exploration: judge-model-under-verify

### Project Type
**Web testing**: not-web

This is a Go CLI tool (no `package.json`, no web framework, no browser-facing routes, no
Playwright/Cypress/Selenium signals in the codebase). Playwright is explicitly disabled in
`.archon/config.yaml` (`playwright.enabled: false`).

---

### Current State

#### Model resolution mechanism (confirmed)

Every delegated SDD phase (explore through archive) gets its model from **two coupled sources**
that must stay in sync:

1. `.archon/config.yaml` → `models.phases.<phase>` — stores the configured `ModelRef`
   (provider + model + effort). Serialised as a YAML mapping with an effort field.
2. `.claude/agents/archon-<phase>.md` frontmatter `model:` — the **hard gate** that Claude
   Code actually enforces. This is a bare model id (`claude-opus-4-8`, not a provider-prefixed
   string). The harness generates this from the config.

The two are kept in sync by `internal/initcmd/claude_mode.go: writeClaudeAgents()`, which
calls `config.ResolvePhaseModels()` and writes one `archon-<phase>.md` per entry in
`config.PhaseOrder`. Both `archon init` and the TUI save path invoke this writer.

#### The judge gap (confirmed)

Judge has **neither** an `archon-judge.md` subagent nor a `models.phases.judge` entry.

Evidence:
- `.claude/agents/` contains exactly 8 files: archon-{apply,archive,design,explore,
  propose,spec,tasks,verify}.md. No `archon-judge.md`.
- `.archon/config.yaml` `models.phases` has 8 keys: apply, archive, design, explore,
  propose, spec, tasks, verify. No `judge` key.
- `internal/config/model.go` line 167: `PhaseOrder = []string{"explore", "propose", "spec",
  "design", "tasks", "apply", "verify", "archive"}` — judge is explicitly excluded with
  the comment "Excludes judge, which is not delegated to an sdd-* sub-agent."
- `internal/config/model.go` lines 153-163: `ValidPhases` map has the same 8 phases, no judge.

#### How harness-judge / judgment-day run today (no model pin)

`harness-judge` is a **skill** (loaded by the orchestrator inline, not a delegated subagent).
It is invoked via `Skill("harness-judge")` which loads `skills/harness-judge/SKILL.md`.
Because it runs inline on the orchestrator's model context, it inherits whatever model the
orchestrator is on — there is no model pin.

`judgment-day` is also a **skill**, invoked from within `harness-judge` via delegation to
the `judgment-day` skill (`Skill("judgment-day")`). `judgment-day` in turn launches two
blind judge sub-agents concurrently (Step 3 of `skills/judgment-day/SKILL.md`) — those
sub-agents are ad-hoc, not named `archon-judge`, and carry no model frontmatter from the
harness.

Neither `skills/harness-judge/SKILL.md` nor `skills/judgment-day/SKILL.md` contain any
reference to a model selector, `--model` flag, or `archon-judge` subagent invocation.
The dual review work runs on whatever model the skill executor inherits from the calling
context.

**A model pin for judge would need to be injected at the point `judgment-day` launches its
two judge sub-agents** (Step 3 of `judgment-day` SKILL.md), OR by having `harness-judge`
delegate to an `archon-judge` subagent whose frontmatter carries `model: claude-opus-4-8`.

---

### Affected Areas

- `.archon/config.yaml` — `models.phases` is missing a `judge` key. Any approach that adds
  `models.phases.judge` touches this file (generated, not hand-authored after init).
- `.claude/agents/` — has no `archon-judge.md`. Creating one is possible; the generator
  must know to emit it.
- `internal/config/model.go` — defines `PhaseOrder` (line 167) and `ValidPhases` (lines
  153-163). Both exclude `judge`. Any approach that adds `judge` to the config/agent system
  must update these two variables.
- `internal/initcmd/claude_mode.go` — `writeClaudeAgents()` iterates `PhaseOrder` via
  `config.ResolvePhaseModels()`. If `judge` is added to `PhaseOrder` it will automatically
  emit `archon-judge.md`. No structural change needed to this file; only the data it iterates.
- `internal/initcmd/templates.go` — `phaseModelsClaude` template iterates `.PhaseModels`
  (which comes from `ResolvePhaseModels()`). If judge is included in `PhaseOrder`, it will
  appear in the "Phase Models" section of CLAUDE.md automatically. The `orchestratorRulesClaude`
  constant (line 160) currently reads "invoke its `archon-<phase>` subagent" — if
  `archon-judge.md` exists it would be implied but the text would need no change.
- `internal/initcmd/init.go` — `buildConfig()` accepts `modelPhases map[string]string` from
  the caller. If the CLI adds a `--model-judge` flag, it flows through here unchanged.
- `cmd/archon/main.go` — lines 192-200 define per-phase `--model-*` init flags. No
  `--model-judge` flag exists. If a judge model is user-configurable at init time, a new
  flag must be added here.
- `cmd/archon/config.go` — `setConfigValue()` (line 180) and `getConfigValue()` (line 208)
  both guard against unknown phases using `config.ValidPhases`. The error message (lines 181,
  209) explicitly lists 8 phases, no judge. Any `archon config set models.phases.judge ...`
  command would fail with "unknown phase" until `ValidPhases` is updated.
- `skills/harness-judge/SKILL.md` — Step 2 calls `judgment-day` skill inline. If option (a)
  is chosen (archon-judge subagent), this skill would need to invoke the subagent instead.
  If option (b) or (c), this skill would need to read the model from config and pass it to
  judgment-day's sub-agents.
- `skills/judgment-day/SKILL.md` — Step 3 launches two blind judge sub-agents concurrently.
  These are currently generic (no model pin). A model pin would be added here or at the
  harness-judge delegation point.
- `CLAUDE.md` — "Phase Models" section lists 8 phases, no judge. Generated by the template
  in `templates.go`; will auto-update if `PhaseOrder` includes judge. But the "Phase Models"
  intro text says "Each phase runs in its named `archon-<phase>` subagent" — judge currently
  does NOT run as a named subagent, so the intro text may need a clarifying note.
- `internal/tui/models_tab.go` — line 104 iterates `config.PhaseOrder` to build the model
  configuration rows. If judge is added to `PhaseOrder`, a judge model row will appear in the
  TUI automatically.

---

### Approaches

1. **Create `archon-judge.md` subagent + have harness-judge delegate to it**
   The harness-judge skill, instead of invoking judgment-day inline, delegates to a new
   `archon-judge` subagent whose frontmatter pins `model: claude-opus-4-8`. That subagent
   internally invokes judgment-day. The Go codegen path must add `judge` to `PhaseOrder`
   and `ValidPhases` so the agent file is emitted at init time and `archon config set
   models.phases.judge` is accepted.
   - Pros: model is a hard gate (Claude Code enforces frontmatter), consistent with all other
     phases, fully re-init-safe, TUI and config set/get work transparently.
   - Cons: requires a non-trivial skill change (harness-judge becomes a thin delegator to
     archon-judge, which then hosts the judgment-day logic); highest conceptual complexity;
     archon-judge.md body would need to replicate or reference the harness-judge logic.
   - Effort: Medium-High

2. **Add `models.phases.judge` to config; harness-judge reads it and passes model to judgment-day**
   Add `judge` to `PhaseOrder` and `ValidPhases`. At init/re-init, config and CLAUDE.md
   get a `judge: claude-opus-4-8` entry. harness-judge reads `.archon/config.yaml →
   models.phases.judge` (or falls back to `models.phases.verify`) and passes the resolved
   model id as a parameter when it delegates the `judgment-day` skill's judge sub-agents.
   No `archon-judge.md` subagent is created; model governance stays in config, not a
   frontmatter gate.
   - Pros: minimal skill change (harness-judge gains one config read + one parameter pass);
     backward-compatible (falls back to verify's model when unset); config key is visible in
     TUI and `archon config set/get`.
   - Cons: the model is NOT a hard gate via frontmatter (Claude Code cannot enforce it; it is
     advisory), consistent with how the orchestrator runs today but weaker than the subagent
     pattern; judgment-day sub-agents need to accept a model parameter (may require skill edit).
   - Effort: Low-Medium

3. **Mirror/alias: harness-judge always reads verify's model (no new config key)**
   harness-judge reads `models.phases.verify` from `.archon/config.yaml` and injects that
   value when delegating to judgment-day. No new `ValidPhases` entry, no new config key, no
   new subagent file. The "judge runs on verify's model" invariant is encoded in the skill,
   not the config.
   - Pros: zero Go codegen changes; zero config schema changes; zero TUI changes; minimal
     blast radius; implements the stated goal exactly ("inherit/mirror verify's model").
   - Cons: judge model is not independently configurable (they are permanently coupled);
     not visible in `archon config list` or the TUI; if verify's model changes, judge
     silently follows (may or may not be desired); judgment-day sub-agents still need to
     accept a model parameter.
   - Effort: Low

---

### Recommendation

Option 3 (mirror verify's model from skill, no new config key) is the minimal-change path
that exactly satisfies the stated goal. If independent configurability of judge's model is
desired in the future, option 2 is the next step up. Option 1 is the full-harness solution
but carries the most complexity and skill restructuring.

This is for propose to decide — flagging both option 2 and option 3 as strong candidates.

---

### Risks

- Judgment-day launches sub-agents "concurrently via delegation" (judgment-day SKILL.md Step
  3). These sub-agents are generic (no frontmatter). If the delegation primitive used by
  judgment-day does not support a `model:` parameter, options 2 and 3 may require a
  judgment-day skill edit in addition to harness-judge.
- Adding `judge` to `PhaseOrder` (option 1 or 2) is a behavioral change to `ResolvePhaseModels`,
  which is called by `writeClaudeAgents` and `writeTemplate`. This must not break existing
  projects that have no `models.phases.judge` in their config — `ResolvePhaseModels` already
  handles missing phases gracefully (falls back to Default, omits if neither yields a model),
  so the risk is low but must be tested.
- The `Clone()` method in `internal/config/config.go` (line 86) has a comment: "every new
  Config field MUST be added here" — if a new struct field is added for judge model, it must
  be cloned or the round-trip test will fail loudly.
- CLAUDE.md "Phase Models" section intro text ("Each phase runs in its named `archon-<phase>`
  subagent") would be misleading for judge if option 2 or 3 is chosen (no subagent created).
  Template text may need a guard or footnote.

---

### Open Questions for Propose

1. Should the judge model be **independently configurable** (option 2), or always mirror
   verify (option 3)? The change goal says "inherit/mirror verify's" — option 3 reads cleaner
   against that goal.
2. Does judgment-day's delegation of judge sub-agents currently support a `model:` parameter?
   If not, is it acceptable to add it to `skills/judgment-day/SKILL.md`, or is that out of
   scope for this change?
3. If option 1 or 2 is chosen, should a `--model-judge` CLI flag be added to `archon init`
   for consistency with all other phase flags?
4. Should `archon config set models.phases.judge` be accepted (requires `ValidPhases` update)
   or should judge model always be derived from config (option 3)?

### Ready for Proposal
Yes — ground truth is sufficient. Propose should choose between option 2 (independent config
key) and option 3 (mirror verify inline), with the open questions above resolved.
