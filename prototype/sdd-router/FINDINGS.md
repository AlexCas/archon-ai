# SDD Router — empirical findings (local model test)

Target model: **qwen3-orch:latest** (Qwen3 ~5B, via ollama, `think:false`, temp 0).
Fixtures: 18 (see `fixtures.md`). Harnesses: `run_local.sh` (model-only),
`run_hybrid.sh` (code + model).

## Scores

| Design                       | Score  | Notes |
|------------------------------|--------|-------|
| Model-only, base prompt      | 15/18  | after fixing the scoring parser (model drops the `archon-` prefix) |
| Model-only, hardened prompt  | 14/18  | fixed the headline implicit case (#1) but destabilized arithmetic/counting |
| **Hybrid (code + model)**    | **17/18** | stable across 3 runs; only #15 (dual-action ambiguity) fails |

## What the weak model is RELIABLE at

Fuzzy text→category classification:
- explicit-agent ("corre el apply", "volvamos al spec") ✓
- implicit start/resume once told to IGNORE competing words ✓
- single-keyword topic ("disenemos la arquitectura" → design) ✓
- action-verb-beats-object-noun ("implementa las tareas" → apply, not tasks) ✓
- unclassifiable → ASK ✓

## What the weak model is UNRELIABLE at (must be CODE, not model)

1. **State arithmetic.** "the phase after `propose`" — even given an explicit
   successor table in the prompt, the model returned the current phase or ASK.
   Fix: resolve control words (`siguiente|continuemos|adelante|...`) in code from
   `state.yaml` + successor map. 100% reliable, zero model calls.
2. **Resume vs start.** "trabajemos ..." with a competing phase keyword
   ("especificacion") derailed the model to the keyword. Fix: detect start verbs
   in code; `none → explore`, else `→ current_phase`.
3. **Multi-match counting.** "revisa y prueba" (two phases) → the model picks one
   instead of ASK. Counting keyword-rows is unreliable in-model (and semantically
   wrong for "implementa las tareas", where the noun is an object, not a phase).

## Architectural conclusion

**Do not make a 5B model do what deterministic code already does.** The Go harness
holds `state.yaml` + `PhaseOrder`; state-dependent transitions are a pure function
of that. Split responsibilities:

```
user message
   │
   ├─ CODE pre-router (deterministic, no model):
   │     • control words  → state.yaml + successor table → target
   │     • start verbs     → none? explore : current_phase (resume)
   │     • explicit "archon-<phase>" / "fase <x>" literal → target
   │     └─ if none of the above, fall through ↓
   │
   └─ MODEL classifier (fuzzy, single responsibility):
         "which phase family is this sentence about?" → phase | ASK
   │
   ▼
 harness-workflow gate (validates legality; blocks illegal jumps)
   ▼
 archon-<phase> subagent
```

This is why the reported bug ("trabajemos en esta especificacion" doesn't launch
explore) happens: the current design leans entirely on the model inferring intent.
Moving the deterministic slice to code fixes it structurally, and the model only
handles what it is actually good at.

## Residual (#15)

"revisa y prueba esto" → model returns `judge` instead of `ASK`. Options:
- Accept it: `harness-workflow` still gates the (legal-or-not) transition, and the
  Human Review Gate shows the user before execution.
- Or add a targeted code rule: message contains a judge-verb AND a verify-verb with
  a coordinating "y"/"and" → ASK. (Narrow; avoid growing a whack-a-mole list.)

## How to reproduce

```
./run_local.sh          # model-only baseline
./run_hybrid.sh         # code + model (recommended design)
```
Both print a per-fixture table and a PASS/FAIL total. `PATH=code` rows are
deterministic; `PATH=model` rows call qwen3.
