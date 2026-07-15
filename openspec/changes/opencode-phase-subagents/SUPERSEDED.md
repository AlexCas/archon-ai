# SUPERSEDED

This change (`opencode-phase-subagents`) targeted a VERBATIM model-writing approach
for per-phase opencode subagents.

After reviewing how gentle-ai resolves models (provider captured at selection time,
reading opencode's structured cache `~/.cache/opencode/models.json`), the user chose
to pursue the robust path: a structured provider+model data model for archon.

This change is SUPERSEDED by `../structured-model-resolution`, a broader multi-PR
initiative. The exploration/proposal/spec/design here remain for traceability; the
per-phase subagent writing becomes Slice 2 of the new initiative.

Date: 2026-06-19
