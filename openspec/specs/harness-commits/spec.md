# harness-commits Specification

## Purpose

Ensure commits created through the harness are attributed only to the user.

## Requirements

### Requirement: User-only commit authorship

When the harness or any sub-agent creates a commit on the user's behalf, the commit
MUST be authored solely by the user's git account. The harness MUST NOT add
`Co-Authored-By` trailers, "Generated with" lines, agent/assistant names, or any other
co-author or tool attribution to commit messages or PR bodies. Conventional-commit
format applies to the subject; the body describes the change, not the tool.

#### Scenario: Commit created without co-author trailers

```gherkin
Scenario: Commit created without co-author trailers
  Given the harness is asked to commit a completed work unit
  When the commit is created
  Then the author is the user's git account
  And the commit message contains no "Co-Authored-By" trailer
  And the commit message contains no agent or tool attribution
```
