#!/usr/bin/env bash
# Run the SDD router fixtures against a local ollama model and score them.
# Usage: ./run_local.sh [model]   (default: qwen3-orch:latest)
set -uo pipefail

MODEL="${1:-qwen3-orch:latest}"
API="http://localhost:11434/api/generate"

read -r -d '' SYS <<'EOF'
You are a SIMPLE, LITERAL routing model. Follow the ruleset EXACTLY, top-to-bottom,
first match wins. Do not use outside intuition. Do NOT execute anything. Do NOT
explain. Output EXACTLY ONE line and nothing else, in this format:
  -> Router: archon-<phase>  (rule: <rule-id>)
or
  -> Router: ASK  (rule: ask)

Phase order: explore -> propose -> spec -> design -> tasks -> apply -> verify -> judge -> archive.
Check the rules STRICTLY in this order. As soon as a rule matches, STOP and output it.

Rule 1 explicit-agent: if MSG names an agent or a phase token directly as the thing to
run OR as a navigation target (e.g. "lanza explore", "agente de exploracion",
"corre el apply", "vamos a la fase spec", "volvamos al spec", "regresa a design",
"salta a tasks", "archon-design") -> resolve to THAT phase. Always wins. A bare phase
name after "a/al/a la fase/volvamos a/regresa a/salta a" counts here, NOT as keyword.

Rule 2 control: else if MSG contains any of: siguiente|continuemos|continua|adelante|
sigamos|next|procede|avanza:
- STATE none -> explore (id next-nochange)
- status in_progress -> current_phase (id resume)
- status completed -> look up current_phase in this SUCCESSOR TABLE and output the value (id next):
    explore->propose, propose->spec, spec->design, design->tasks, tasks->apply,
    apply->verify, verify->judge, judge->archive
  (Do NOT compute; just read the pair whose left side equals current_phase.)

Rule 3 implicit: else if MSG contains any of: trabajemos|empecemos|comencemos|hagamos|
armemos|pongamonos|arranquemos -> IGNORE every other word in the sentence (even if it
looks like a phase name such as "especificacion", "diseno", "tareas") and output:
- STATE none -> explore (id implicit-start)
- else -> current_phase (id implicit-resume)
NOTE volvamos/regresemos/cambiemos are NOT in this list.

Rule 4 keyword: else scan MSG against this table. FIRST check if words from TWO
DIFFERENT rows appear; if yes, output ASK (id ambiguous). Otherwise the single
matching row is the phase (id keyword). Table:
  explore: explora, exploremos, investiga, entender, analiza el codigo
  propose: propon, propuesta, propongamos, idea, enfoque
  spec: especificacion, spec, requisitos, gherkin, escenarios
  design: diseno, disenemos, arquitectura, plan tecnico
  tasks: tareas, desglose, checklist, plan de tareas, to-do
  apply: implementa, codifica, aplica, escribe el codigo, construye
  verify: verifica, prueba, pruebas, corre las pruebas, valida
  judge: juzga, revisa, revisar, revision, revisa el codigo, dictamen
  archive: archiva, finaliza, cierra, completa el cambio
  -> exactly one phase matches -> that phase (id keyword); two+ -> ASK (id ambiguous).

Rule ask: else ASK (id ask).
EOF

# fixture: MSG|||STATE|||expected_phase   (expected "ASK" means any ASK is correct)
FIXTURES=(
"Trabajemos en esta especificacion|||none|||explore"
"Trabajemos en esta especificacion|||spec/in_progress|||spec"
"Empecemos con esto|||none|||explore"
"Hagamos esta feature|||none|||explore"
"Hagamos esta especificacion. Lanza el agente de exploracion|||none|||explore"
"Continuemos|||propose/completed|||spec"
"Siguiente|||apply/completed|||verify"
"Continuemos|||design/in_progress|||design"
"Adelante|||none|||explore"
"Explora el codigo de billing|||none|||explore"
"Disenemos la arquitectura del API|||spec/completed|||design"
"Implementa las tareas|||tasks/completed|||apply"
"Corre las pruebas|||apply/completed|||verify"
"Archiva el cambio|||judge/completed|||archive"
"Revisa y prueba esto|||verify/in_progress|||ASK"
"Que opinas del clima?|||spec/in_progress|||ASK"
"Volvamos al spec|||design/completed|||spec"
"corre el apply|||tasks/completed|||apply"
)

pass=0; fail=0; n=0
printf '%-3s %-52s %-20s %-10s %-10s %s\n' "#" "MSG" "STATE" "EXPECT" "GOT" "OK"
printf '%.0s-' {1..110}; echo
for fx in "${FIXTURES[@]}"; do
  n=$((n+1))
  MSG="${fx%%|||*}"; rest="${fx#*|||}"; STATE="${rest%%|||*}"; EXP="${rest##*|||}"
  USER="MSG=\"$MSG\" | STATE=$STATE"
  # think:false disables qwen3 reasoning; keep temperature 0 for determinism.
  REQ=$(jq -n --arg m "$MODEL" --arg s "$SYS" --arg p "$USER" \
    '{model:$m, system:$s, prompt:$p, stream:false, think:false, options:{temperature:0}}')
  RESP=$(curl -s "$API" -d "$REQ")
  RAW=$(echo "$RESP" | jq -r '.response // empty')
  # strip any <think>...</think> block, then pull the phase token after "Router:"
  # (accept the phase with or without the archon- prefix; the model often drops it).
  CLEAN=$(echo "$RAW" | sed -E 's#<think>.*</think>##g')
  GOT=$(echo "$CLEAN" | grep -ioE 'Router:[[:space:]]*(archon-)?(explore|propose|spec|design|tasks|apply|verify|judge|archive|ASK)' \
        | head -1 | grep -ioE '(explore|propose|spec|design|tasks|apply|verify|judge|archive|ASK)' | head -1)
  # normalize ASK casing
  echo "$GOT" | grep -qi '^ask$' && GOT="ASK"
  GOT="${GOT:-<none>}"
  if [ "$GOT" = "$EXP" ]; then OK="PASS"; pass=$((pass+1)); else OK="FAIL"; fail=$((fail+1)); fi
  printf '%-3s %-52s %-20s %-10s %-10s %s\n' "$n" "${MSG:0:52}" "$STATE" "$EXP" "$GOT" "$OK"
done
printf '%.0s-' {1..110}; echo
echo "RESULT: $pass passed, $fail failed, of $n"
