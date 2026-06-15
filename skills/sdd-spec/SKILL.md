---
name: sdd-spec
description: "Write SDD delta specs with requirements and scenarios. Trigger: orchestrator launches spec work for a change."
disable-model-invocation: true
user-invocable: false
license: MIT
metadata:
  
  version: "2.0"
  delegate_only: true
---

> **ORCHESTRATOR GATE**: If you loaded this skill via the `skill()` tool, you are
> the ORCHESTRATOR — STOP. Do NOT execute these instructions inline. Delegate to
> the dedicated `sdd-spec` sub-agent using your platform's delegation primitive
> (e.g., `task(...)`, sub-agent invocation, etc.). This skill is for EXECUTORS
> only.

## Executor Override

If you ARE the `sdd-spec` sub-agent (NOT the orchestrator), the gate above does NOT apply to you. Continue with the phase work below. Do NOT delegate. Do NOT call the Skill tool. You are the executor — execute.


## Language Domain Contract

Generated technical artifacts default to English. Do not inherit the user's conversational language or the active persona's regional voice for SDD artifacts unless the user explicitly requests that artifact language or the project convention requires it.

If Spanish technical artifacts are explicitly requested, use neutral/professional Spanish unless the user explicitly asks for a regional variant.

Public/contextual comments follow the target context language by default. Explicit user language or tone overrides win; Spanish comments default to neutral/professional Spanish unless the user or target context clearly calls for regional tone.

## Purpose

You are a sub-agent responsible for writing SPECIFICATIONS. You take the proposal and produce delta specs — structured requirements and scenarios that describe what's being ADDED, MODIFIED, REMOVED, or RENAMED from the system's behavior.

## What You Receive

From the orchestrator:
- Change name
- Artifact store mode (`engram | openspec | hybrid | none`)

## Execution and Persistence Contract

> Follow **Section B** (retrieval) and **Section C** (persistence) from `skills/_shared/sdd-phase-common.md`.

- **engram**: Read `sdd/{change-name}/proposal` (required). If specs span multiple domains, concatenate into a single artifact with domain headers. Save as `sdd/{change-name}/spec`.
- **openspec**: Read and follow `skills/_shared/openspec-convention.md`.
- **hybrid**: Follow BOTH conventions — persist to Engram (single concatenated artifact) AND write domain files to filesystem.
- **none**: Return result only. Never create or modify project files.

## What to Do

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Identify Affected Domains

Read the proposal's **Capabilities section** — this is your primary contract:

```
FOR EACH entry under "New Capabilities":
├── This becomes a NEW full spec: openspec/specs/<capability-name>/spec.md
└── Write a complete spec (not a delta) — no existing behavior to reference

FOR EACH entry under "Modified Capabilities":
├── This becomes a DELTA spec: openspec/changes/{change-name}/specs/<capability-name>/spec.md
└── Read existing openspec/specs/<capability-name>/spec.md first — your delta modifies it
```

If the proposal has no Capabilities section (older format), fall back to inferring from "Affected Areas". But always prefer the explicit Capabilities mapping when present.

### Step 3: Read Existing Specs

**IF mode is `openspec` or `hybrid`:** If `openspec/specs/{domain}/spec.md` exists, read it to understand CURRENT behavior. Your delta specs describe CHANGES to this behavior.

**IF mode is `engram`:** Existing specs were already retrieved from Engram in the Persistence Contract. Skip filesystem reads.

**IF mode is `none`:** Skip — no existing specs to read.

### Step 4: Write Delta Specs

**IF mode is `openspec` or `hybrid`:** Create specs inside the change folder:

```
openspec/changes/{change-name}/
├── proposal.md              ← (already exists)
└── specs/
    └── {domain}/
        ├── spec.md          ← Delta spec (requirements + Gherkin scenarios)
        └── {domain}.feature ← Formal Gherkin feature file (executable use cases)
```

**IF mode is `engram` or `none`:** Do NOT create any `openspec/` directories or files. Compose the spec content (including the Gherkin feature) in memory — you will persist it in Step 5.

#### Gherkin Feature Files (MANDATORY)

Every domain MUST also produce a formal Gherkin `.feature` file. Test/use cases
are authored in **formal Gherkin**, not loose bullet lists — this is the contract
the verify phase tests against, and (for web projects) the source the harness uses
to generate Playwright specs.

Rules for `.feature` files:
- One `.feature` per domain, named `{domain}.feature`, placed beside `spec.md`.
- Start with a `Feature:` line and a short description. Use `Background:` for shared
  preconditions, `Scenario:` for examples, and `Scenario Outline:` + `Examples:` for
  data-driven cases.
- Steps use the real Gherkin keywords `Given`, `When`, `Then`, `And`, `But` —
  capitalized as keywords, NOT in all-caps prose, NOT as Markdown bullets.
- Tag scenarios for traceability and selective execution, e.g. `@happy`, `@edge`,
  `@error`, and `@web` for browser-facing flows that Playwright will exercise.
- Every requirement in `spec.md` MUST map to at least one scenario in the
  `.feature` file. Keep the scenario names identical in both files.

Example `{domain}.feature`:

```gherkin
Feature: Session expiration
  Inactive sessions expire so that stale credentials cannot be reused.

  Background:
    Given an authenticated user with an active session

  @happy
  Scenario: Session stays valid within the timeout window
    When the user makes a request 5 minutes after login
    Then the request succeeds

  @edge @error
  Scenario: Session expires after the inactivity timeout
    When the user makes a request 31 minutes after login
    Then the response is 401 Unauthorized
    And the session token is invalidated
```

#### MODIFIED Requirements Workflow (CRITICAL — read before writing deltas)

When writing a `## MODIFIED Requirements` section, follow this exact workflow:

```
1. Locate the requirement in openspec/specs/{domain}/spec.md
2. COPY the ENTIRE requirement block — from `### Requirement:` through ALL its scenarios
3. PASTE it under `## MODIFIED Requirements`
4. EDIT the copy to reflect the new behavior
5. Add "(Previously: {one-line summary of what changed})" under the requirement text

Why copy-full-then-edit?
→ The archive step REPLACES the requirement in main specs with your MODIFIED block
→ If your block is partial, the archive will lose scenarios you didn't copy
→ Common pitfall: only writing the changed scenario and losing the rest
→ If adding NEW behavior WITHOUT changing existing behavior, use ADDED instead
```

#### Delta Spec Format

```markdown
# Delta for {Domain}

## ADDED Requirements

### Requirement: {Requirement Name}

{Description using RFC 2119 keywords: MUST, SHALL, SHOULD, MAY}

The system {MUST/SHALL/SHOULD} {do something specific}.

#### Scenario: {Happy path scenario}

```gherkin
Scenario: {Happy path scenario}
  Given {precondition}
  When {action}
  Then {expected outcome}
  And {additional outcome, if any}
```

#### Scenario: {Edge case scenario}

```gherkin
Scenario: {Edge case scenario}
  Given {precondition}
  When {action}
  Then {expected outcome}
```

## MODIFIED Requirements

### Requirement: {Existing Requirement Name}

{Full updated requirement text — replaces the existing one entirely}
(Previously: {what it was before, in one line})

#### Scenario: {Unchanged scenario — keep if still valid}

```gherkin
Scenario: {Unchanged scenario — keep if still valid}
  Given {precondition}
  When {action}
  Then {outcome}
```

#### Scenario: {Updated or new scenario}

```gherkin
Scenario: {Updated or new scenario}
  Given {updated precondition}
  When {updated action}
  Then {updated outcome}
```

## REMOVED Requirements

### Requirement: {Requirement Being Removed}

(Reason: {why this requirement is being deprecated/removed})
(Migration: {what replaces it, or "None" if no migration is needed})

## RENAMED Requirements

### Requirement: {Old Requirement Name} → {New Requirement Name}

(Reason: {why the requirement is being renamed})
(Migration: {how references/tests/docs should update, or "None" if no migration is needed})
```

#### For NEW Specs (No Existing Spec)

If this is a completely new domain, create a FULL spec (not a delta):

```markdown
# {Domain} Specification

## Purpose

{High-level description of this spec's domain.}

## Requirements

### Requirement: {Name}

The system {MUST/SHALL/SHOULD} {behavior}.

#### Scenario: {Name}

```gherkin
Scenario: {Name}
  Given {precondition}
  When {action}
  Then {outcome}
```
```

### Step 5: Persist Artifact

**This step is MANDATORY — do NOT skip it.**

Follow **Section C** from `skills/_shared/sdd-phase-common.md`.
- artifact: `spec`
- topic_key: `sdd/{change-name}/spec`
- type: `architecture`

The persisted `spec` artifact MUST include the Gherkin feature content. In
`openspec`/`hybrid` mode also write the `{domain}.feature` files to disk (Step 4).
In `engram`/`none` mode, embed each domain's feature block inside the artifact under
a clearly labelled `## Feature ({domain})` section.

### Step 6: Return Summary

Return to the orchestrator:

```markdown
## Specs Created

**Change**: {change-name}

### Specs Written
| Domain | Type | Requirements | Scenarios | Feature file |
|--------|------|-------------|-----------|--------------|
| {domain} | Delta/New | {N added, M modified, K removed} | {total scenarios} | {domain}.feature |

### Coverage
- Happy paths: {covered/missing}
- Edge cases: {covered/missing}
- Error states: {covered/missing}

### Next Step
Ready for design (sdd-design). If design already exists, ready for tasks (sdd-tasks).
```

## Rules

- ALWAYS author scenarios in FORMAL Gherkin (`Feature:`, `Scenario:`, `Given/When/Then/And/But`, optional `Background:`, `Scenario Outline:` + `Examples:`) — not all-caps prose or Markdown bullets
- ALWAYS produce a `{domain}.feature` file per domain alongside `spec.md`, with every requirement mapped to at least one scenario and matching scenario names
- Tag scenarios (`@happy`, `@edge`, `@error`, `@web`) so verify and Playwright generation can select them
- ALWAYS use RFC 2119 keywords (MUST, SHALL, SHOULD, MAY) for requirement strength
- Read the proposal's **Capabilities section** first — it tells you exactly which spec files to create
- If existing specs exist, write DELTA specs (ADDED/MODIFIED/REMOVED sections)
- If NO existing specs exist for the domain, write a FULL spec
- Every requirement MUST have at least ONE scenario
- Include both happy path AND edge case scenarios
- Keep scenarios TESTABLE — someone should be able to write an automated test from each one
- DO NOT include implementation details in specs — specs describe WHAT, not HOW
- **MODIFIED requirements MUST be the FULL block** — copy entire requirement + all scenarios from main spec, then edit. Partial MODIFIED blocks lose content at archive time.
- If adding new behavior without changing existing behavior → use ADDED, not MODIFIED
- REMOVED requirements MUST include Reason and SHOULD include Migration when consumers, persisted behavior, docs, or tests are affected
- RENAMED requirements MUST state both old and new names explicitly and SHOULD include Migration guidance for references/tests/docs
- Apply any `rules.specs` from `openspec/config.yaml`
- **Size budget**: Spec artifact MUST be under 650 words. Prefer requirement tables over narrative descriptions. Each scenario: 3-5 lines max.
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.

## RFC 2119 Keywords Quick Reference

| Keyword | Meaning |
|---------|---------|
| **MUST / SHALL** | Absolute requirement |
| **MUST NOT / SHALL NOT** | Absolute prohibition |
| **SHOULD** | Recommended, but exceptions may exist with justification |
| **SHOULD NOT** | Not recommended, but may be acceptable with justification |
| **MAY** | Optional |
