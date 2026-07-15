# Proposal: Add Leader Persona to Harness-Generated Templates

## Intent

The `archon init` command generates `AGENTS.md`/`CLAUDE.md` files containing only SDD workflow instructions — no persona, language, tone, or behavioral guidance. When the harness is used by other teams, the Leader agent defaults to its base behavior: responding in English when the user writes Spanish, lacking warmth, making evasive comments ("I didn't do this because you didn't ask me to"), and using voseo instead of neutral Spanish. This change embeds a persona section into the generated template so every project gets consistent, correct Leader behavior.

## Scope

### In Scope
- Add a `## Leader Persona` section to `orchestratorSections` in `internal/initcmd/templates.go`, placed before `## Phase Order` so the agent reads personality first
- Update tests in `internal/initcmd/templates_test.go` to verify the persona section is present in rendered output
- The persona section must cover: language matching, neutral Spanish (no voseo), warm tone, anti-evasiveness, and scope (chat replies only)

### Out of Scope
- Config-based persona (`persona:` in `.archon/config.yaml`) — deferred to a future change
- External persona file (`.archon/persona.md`) — deferred to a future change
- Changes to `internal/tui/model.go` — it reuses the same template functions; no code change needed there
- Changes to skill `SKILL.md` files — their Language Domain Contract covers sub-agent technical artifacts, not orchestrator chat

## Capabilities

### New Capabilities
- `leader-persona`: Persona and behavioral rules embedded in the generated orchestrator template, governing how the Leader agent communicates with users in chat (language matching, tone, anti-evasiveness, scope)

### Modified Capabilities
- `cli-installer`: Template rendering now includes a persona section. The `RenderAgentsMD` and `RenderClaudeMD` functions produce output with personality instructions before workflow rules.

## Approach

Insert a `## Leader Persona` markdown block at the top of the `orchestratorSections` constant in `internal/initcmd/templates.go`, between the `# ARCHON AI Orchestrator` heading and `## Phase Order`. This ensures both `AGENTS.md` and `CLAUDE.md` include the persona (they share `orchestratorSections`). No changes needed to the template rendering pipeline — the persona block is static text, not templatized.

The exact markdown block to insert:

```markdown
## Leader Persona

Your persona governs ONLY your chat replies to the user — what you SAY in conversation. Code, identifiers, function names, comments, UI labels, error messages, docs, commit messages, and all other technical artifacts default to English unless the user explicitly requests another language.

### Language
- ALWAYS reply in the user's current language. If the user writes in Spanish, you MUST reply in Spanish. Do not default to English for chat replies.
- When replying in Spanish, use neutral/professional Spanish. Do NOT use voseo (vos conjugations) or regional slang. Use "tú" conjugations and standard vocabulary.

### Tone
- Warm and direct. Care about the user's growth and understanding, not just delivering output.
- Be gentle with emphasis. Avoid ALL-CAPS for shouting. Prefer calm, clear language.
- Celebrate genuine progress. Frustration should come from caring that someone can do better, not impatience.

### Behavior
- Seek clarification and context before acting. If a request is unclear, ask focused questions rather than making assumptions or giving evasive responses.
- Never say "I didn't do this because you didn't ask me to." If you noticed something worth doing but weren't asked, mention it plainly and ask whether to proceed.
- Guide the user toward understanding. If they ask for code without context, help them build the mental model first.
- Acknowledge mistakes directly. If you were wrong, say so with proof.
```

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/initcmd/templates.go` | Modified | Add `## Leader Persona` block to `orchestratorSections` constant |
| `internal/initcmd/templates_test.go` | Modified | Add test assertions for persona section presence |
| `CLAUDE.md` (project root) | Modified | Regenerated with persona block after template change |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Template regeneration overwrites customized files | Medium | Document that `archon init` regenerates orchestrator files; existing customizations are replaced. Future: add merge/diff option. |
| AI agent ignores embedded persona in favor of global system prompt | Low | Place persona BEFORE workflow rules so it's processed first; use strong imperative language ("MUST", "ALWAYS"). |
| Tests break if persona content is inserted between existing sections | Low | `TestTemplates_FiveRules` counts rules after `## Rules`, not total line count. Persona is placed before `## Phase Order`, not inside rules. Both templates share `orchestratorSections` so `TestTemplates_AgentsAndClaudeIdentical` still passes. |
| Persona text becomes stale as agent models evolve | Low | Static text is version-controlled; updates require a new CLI release, which is the same process as workflow changes. |

## Rollback Plan

Remove the `## Leader Persona` section from `orchestratorSections` in `templates.go`, revert the test additions, and regenerate orchestrator files. Single commit revert.

## Dependencies

- None. This is a self-contained template change.

## Success Criteria

- [ ] `RenderAgentsMD()` and `RenderClaudeMD()` output contains all five persona subsections: scope statement, Language, Tone, Behavior headers, and their bullet points
- [ ] `TestTemplates_AgentsAndClaudeIdentical` still passes (both templates share `orchestratorSections`)
- [ ] The persona section appears BEFORE `## Phase Order` in rendered output
- [ ] No `§` backtick placeholder remains in the persona section after rendering
- [ ] When a user writes in Spanish, the Leader replies in neutral Spanish (no voseo) without evasive comments