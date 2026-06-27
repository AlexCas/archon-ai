# Design: Leader Persona Template Section

## Technical Approach

Insert a `## Leader Persona` markdown block at line 17 of `internal/initcmd/templates.go`, between the `# ARCHON AI Orchestrator` heading and `## Phase Order`. Both `agentsTemplate` and `claudeTemplate` inherit from `orchestratorSections`, so the persona appears in every generated orchestrator file with zero rendering-pipeline changes.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Persona storage | Inline in `orchestratorSections` constant | Separate `personaSection` const, external file | Inline follows the existing pattern (all sections are inline raw strings). A separate const adds indirection with no reuse benefit. External file is out of scope per proposal. |
| Backtick escaping | No § placeholders needed | Pre-escape with § | Persona text contains zero backtick characters. The existing rendering pipeline (`strings.ReplaceAll("§", "\`")`) operates on `orchestratorSections` unchanged. |
| Insertion order | After H1, before `## Phase Order` | Before H1, at document end | Spec requires persona before workflow rules. Placing immediately after the title ensures the LLM processes personality before procedural rules. |

## Data Flow

```
orchestratorSections (const, now includes persona)
    │
    ▼
agentsTemplate = orchestratorSections + "## Rules\n..."
claudeTemplate  = orchestratorSections + "## Rules\n..."
    │
    ▼
renderTemplate() → strings.ReplaceAll("§"→"`") → text/template.Execute()
    │
    ▼
AGENTS.md / CLAUDE.md (identical shared-core output)
```

No new branches. Persona is static markdown with zero template variable interpolation.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/initcmd/templates.go` | Modify | Insert `## Leader Persona` block at line 17 of `orchestratorSections` (after `# ARCHON AI Orchestrator\n`, before `\n## Phase Order`). ~40 lines of markdown. |
| `internal/initcmd/templates_test.go` | Modify | Add `TestTemplates_LeaderPersona` function (~35 lines). No existing tests modified. |

## Persona Block Content

Exact markdown to insert (from proposal, zero backticks — no § escaping required):

```
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

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Persona present in both templates | `TestTemplates_LeaderPersona` — renders both `RenderAgentsMD` and `RenderClaudeMD`, asserts `strings.Contains` for `## Leader Persona`. |
| Unit | Ordering: persona before Phase Order | `strings.Index("## Leader Persona") < strings.Index("## Phase Order")` on rendered output of both templates. |
| Unit | Four domains covered | Asserts: "governs ONLY your chat replies" (scope), "ALWAYS reply in the user's current language" (language), "Warm and direct" (tone), "Never say" (behavior). |
| Regression | Both templates identical | Existing `TestTemplates_AgentsAndClaudeIdentical` — unchanged, passes because both share `orchestratorSections`. |
| Regression | Exactly five rules, no § artifacts | Existing `TestTemplates_FiveRules` and `TestTemplates_BacktickRendering` — unchanged, continue passing. |

New test function pseudo-code:

```go
func TestTemplates_LeaderPersona(t *testing.T) {
    data := TemplateData{Agent: "test", HarnessVersion: "1.0.0", SkillCount: 10}
    cases := []struct{
        name   string
        render func(TemplateData) (string, error)
    }{
        {"AGENTS.md", RenderAgentsMD},
        {"CLAUDE.md", RenderClaudeMD},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            content, _ := c.render(data)
            // 1. Header present
            // 2. Ordering: personaIdx < phaseOrderIdx
            // 3. Four domain checks via Contains
        })
    }
}
```

## Risks

| Risk | Mitigation |
|------|------------|
| Persona accidentally contains backtick needing § | Verified: zero backticks in the block. `TestTemplates_BacktickRendering` catches regression. |
| `strings.Index` returns -1 breaking ordering assertion | Guard: run header-exists checks first. Ordering assertion only executes if both sections confirmed present. |
| Test fragility on exact strings | Uses key invariant phrases ("governs ONLY", "ALWAYS reply", "Warm and direct", "Never say") — stable against minor rewording. |

## Open Questions

None. The insertion point, exact content, test assertions, and risk mitigations are fully specified.

---

### Next Step
Ready for tasks (`sdd-tasks`).
