## Exploration: Ajustar la personalidad del Leader (Issue #16)

### Current State

The archon-ai harness generates `AGENTS.md` or `CLAUDE.md` in target projects via `internal/initcmd/templates.go`. These generated files act as the "Leader" orchestrator instructions that the AI agent reads when running the SDD workflow.

**Critical finding**: The generated templates contain **only SDD workflow instructions** (phase order, preflight gates, review gates) and **zero persona, language, tone, or behavioral instructions**. The harness assumes the AI agent's global configuration will provide personality, but when tested on other teams, the agent defaults to its base behavior — causing the reported issues:

- Responds in English even when user writes in Spanish
- Not warm enough in tone
- Makes evasive comments
- Uses voseo (Argentine Spanish)

### Affected Areas

- `internal/initcmd/templates.go` — The single source of truth for generated orchestrator instructions (`AGENTS.md` / `CLAUDE.md`). This is where the persona section must be added. **This is the primary file to change.**
- `internal/initcmd/templates_test.go` — Tests verify the generated content. Must be updated to assert the new persona section is present.
- `internal/tui/model.go` — The TUI regenerates templates via `regenerateTemplate()` which calls `initcmd.RenderClaudeMD()` / `RenderAgentsMD()`. No code change needed here (it reuses the same templates), but the resulting files will include the new persona after the template change.
- `openspec/changes/issue-16-leader-personality/exploration.md` — The previous exploration incorrectly identified the Leader as `gentle-orchestrator` in `~/.config/opencode/AGENTS.md`. This file should be replaced with the correct findings.

### Key Files

| File | Role | What it controls |
|------|------|-----------------|
| `internal/initcmd/templates.go` | Template engine | Generates `AGENTS.md` / `CLAUDE.md` — currently only workflow rules, no persona |
| `internal/initcmd/templates_test.go` | Test coverage | Verifies generated content has required sections |
| `skills/*/SKILL.md` | SDD phase skills | Contain "Language Domain Contract" for **technical artifacts** (sub-agents), NOT for orchestrator chat |
| `CLAUDE.md` (project root) | Project's own orchestrator | Same template output — also lacks persona instructions |

### Root Causes

1. **Missing persona in templates**: `internal/initcmd/templates.go` has no `## Personality` or `## Persona` section. The orchestrator instructions only cover workflow mechanics.
2. **Missing language directive**: There is no instruction telling the orchestrator to match the user's language in chat replies.
3. **Missing tone guidance**: No warmth, directness, or anti-evasiveness rules are embedded in the generated instructions.
4. **Missing regional language rule**: No instruction to use neutral/professional Spanish instead of voseo.
5. **Previous exploration was wrong**: The old `exploration.md` pointed to `~/.config/opencode/AGENTS.md` — that is the **global** agent config, not the **archon-ai harness** definition. The harness must define its own orchestrator persona so it works consistently across teams.

### Approaches

1. **Add persona section to templates** (Recommended)
   - Insert a `## Leader Persona` block into `orchestratorSections` in `internal/initcmd/templates.go`.
   - Include language matching, tone, warmth, and anti-evasiveness rules.
   - Update tests in `templates_test.go` to verify the new section.
   - **Pros**: Single source of truth, affects all generated projects, no config changes needed, fixes the root cause.
   - **Cons**: Hard-coded persona in Go code — requires a new CLI release to update persona text. Updating persona requires editing source code and recompiling.
   - **Effort**: Low (template edit + test update)

2. **Add persona config to `.archon/config.yaml` + template parameter**
   - Add a `persona` struct to `internal/config/model.go` and render it in the template.
   - **Pros**: Configurable per-project without recompiling the CLI.
   - **Cons**: Requires config schema changes, TUI updates, more complexity. Overkill for a persona that should be consistent across all archon projects.
   - **Effort**: Medium

3. **External persona file (e.g., `.archon/persona.md`)**
   - Read an external markdown file and append it to the generated orchestrator instructions.
   - **Pros**: Maximum flexibility, teams can customize.
   - **Cons**: Requires new file handling, CLI flag, documentation. The default would still be missing unless the file is present.
   - **Effort**: High

### Recommendation

**Approach 1**: Add a hard-coded persona section to `internal/initcmd/templates.go`.

The archon-ai harness should embed its own orchestrator persona so that every project initialized with `archon init` gets consistent, correct behavior. The persona should be a markdown block inserted at the top of the generated file, before the workflow instructions, so the AI agent sees it first.

Suggested content for the persona block:
- **Language**: "ALWAYS reply in the user's current language. If the user writes in Spanish, you MUST reply in Spanish. Do not default to English for chat replies."
- **Tone**: "Warm and direct, but from a place of CARING. Use gentle emphasis, avoid CAPS."
- **Spanish variant**: "When replying to the user in Spanish, use neutral/professional Spanish. Do NOT use voseo or regional slang."
- **Behavior**: "Seek clarification and ask for context when the user asks for code without sufficient understanding. Guide them rather than pushing back or making evasive comments."
- **Scope**: "These persona rules apply ONLY to your chat replies to the user. Technical artifacts (code, comments, docs, specs) default to English unless the user explicitly requests otherwise."

### Risks

- **Template regeneration**: The TUI's `regenerateTemplate` will overwrite existing `AGENTS.md`/`CLAUDE.md` with the new persona. Teams that have customized their orchestrator file will lose their customizations. Mitigation: Add a backup/merge warning or make the persona additive.
- **Agent override**: Some AI agents may still ignore embedded persona instructions if their global system prompt is stronger. The persona block should be prominent and use strong imperative language.
- **Testing**: The existing tests in `templates_test.go` are strict about content (e.g., "exactly 5 rules"). Adding a persona section requires careful test updates to avoid breaking existing assertions.
- **Localization**: The preflight gate prompt is already in Spanish. The persona should also be in Spanish or bilingual, matching the expected user base. Since the current preflight is Spanish, the persona block should be in Spanish.

### Ready for Proposal

**Yes.** The scope is clear: add a persona section to `internal/initcmd/templates.go`, update `templates_test.go`, and correct the previous `exploration.md`. The root cause is the absence of persona instructions in the generated orchestrator file, and the fix is straightforward.

The orchestrator should show the user:
- The exact template file and lines to modify
- The proposed persona block text
- The test updates needed
- Ask: "¿Querés ajustar algo en esta exploración antes de pasar a la propuesta?"
