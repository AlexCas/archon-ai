#!/usr/bin/env bash
# Hybrid router: deterministic CODE handles state-dependent transitions
# (control words + successor lookup + resume). The MODEL only does fuzzy
# classification (explicit-agent / implicit-start / single keyword) where it is
# reliable. This plays to the 5B model's strength and removes its failure modes.
# Usage: ./run_hybrid.sh [model]
set -uo pipefail

MODEL="${1:-qwen3-orch:latest}"
API="http://localhost:11434/api/generate"

# ---- deterministic successor table (code, not model) ----
declare -A NEXT=( [explore]=propose [propose]=spec [spec]=design [design]=tasks
                  [tasks]=apply [apply]=verify [verify]=judge [judge]=archive )

# control words -> resolved in code from STATE, zero model involvement
CONTROL='siguiente|continuemos|continua|continúa|adelante|sigamos|next|procede|avanza'

read -r -d '' SYS <<'EOF'
You are a SIMPLE, LITERAL classifier. Output EXACTLY ONE line, nothing else:
  PHASE=<explore|propose|spec|design|tasks|apply|verify|judge|archive|ASK>
No explanation. Do not execute anything. Rules, checked in order, first match wins:

1. If MSG names a phase/agent directly to run or as a destination (e.g. "lanza
   explore", "agente de exploracion", "corre el apply", "volvamos al spec",
   "regresa a design", "salta a tasks", "vamos a la fase spec") -> that phase.

2. If MSG contains a START verb (trabajemos, empecemos, comencemos, hagamos, armemos,
   pongamonos, arranquemos): IGNORE every other word (even if it looks like a phase
   name) and output: STATE_NONE=true -> explore ; otherwise -> RESUME (the caller
   handles resume, so just output explore here and the caller overrides). Simplest:
   output PHASE=explore for a start verb.

3. Otherwise pick the SINGLE phase whose topic the sentence is about. An ACTION VERB
   beats an object noun: "implementa las tareas" is about implementing -> apply (not
   tasks). Topic map:
   explore: explorar/investigar/entender el codigo
   propose: proponer/propuesta/enfoque/idea
   spec: especificacion/requisitos/gherkin/escenarios
   design: disenar/arquitectura/plan tecnico
   tasks: desglosar/lista de tareas/checklist
   apply: implementar/codificar/aplicar/construir
   verify: verificar/probar/correr pruebas/validar
   judge: juzgar/revisar codigo/dictamen
   archive: archivar/finalizar/cerrar

4. If the sentence asks for TWO clearly different actions from different phases
   (e.g. "revisa y prueba" = judge review AND verify testing) -> ASK.

5. If none apply -> ASK.
EOF

# fixture: MSG|||STATE|||expected
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

classify() { # $1=MSG  -> echoes PHASE token or ASK (model call)
  local msg="$1"
  local req; req=$(jq -n --arg m "$MODEL" --arg s "$SYS" --arg p "MSG=\"$msg\"" \
    '{model:$m,system:$s,prompt:$p,stream:false,think:false,options:{temperature:0}}')
  local raw; raw=$(curl -s "$API" -d "$req" | jq -r '.response // empty' | sed -E 's#<think>.*</think>##g')
  echo "$raw" | grep -ioE 'PHASE=[A-Za-z]+' | head -1 | sed -E 's/PHASE=//I' \
    | grep -ioE 'explore|propose|spec|design|tasks|apply|verify|judge|archive|ASK' | head -1
}

pass=0; fail=0; n=0
printf '%-3s %-52s %-20s %-8s %-8s %-6s %s\n' "#" "MSG" "STATE" "EXPECT" "GOT" "PATH" "OK"
printf '%.0s-' {1..112}; echo
for fx in "${FIXTURES[@]}"; do
  n=$((n+1))
  MSG="${fx%%|||*}"; rest="${fx#*|||}"; STATE="${rest%%|||*}"; EXP="${rest##*|||}"
  CUR="${STATE%%/*}"; ST="${STATE##*/}"; [ "$STATE" = "none" ] && CUR="" && ST=""
  low=$(echo "$MSG" | tr '[:upper:]' '[:lower:]')

  START='trabajemos|empecemos|comencemos|hagamos|armemos|pongamonos|arranquemos'
  # NOTE: explicit-agent must beat the start-verb code path, so #5
  # ("hagamos ... lanza el agente de exploracion") still resolves via the model.
  # Here both give explore, but the guard documents the intended precedence.
  HAS_EXPLICIT=$(echo "$low" | grep -qE 'lanza|agente de|archon-|corre el|vamos a la fase|volvamos a|regresa a|salta a' && echo 1 || echo "")

  if echo "$low" | grep -qE "\b($CONTROL)\b"; then
    # ---- deterministic path: control words resolved in CODE ----
    PATHK="code"
    if [ -z "$CUR" ]; then GOT="explore"
    elif [ "$ST" = "in_progress" ]; then GOT="$CUR"
    else GOT="${NEXT[$CUR]:-ASK}"; fi
  elif echo "$low" | grep -qE "\b($START)\b" && [ -z "$HAS_EXPLICIT" ]; then
    # ---- deterministic path: start verbs resolved in CODE ----
    PATHK="code"
    if [ -z "$CUR" ]; then GOT="explore"; else GOT="$CUR"; fi
  else
    # ---- model path: fuzzy classification only ----
    PATHK="model"
    GOT=$(classify "$MSG")
    echo "$GOT" | grep -qi '^ask$' && GOT="ASK"
    # unclassifiable -> ASK (safe default; never guess a phase)
    [ -z "$GOT" ] && GOT="ASK"
  fi

  if [ "$GOT" = "$EXP" ]; then OK="PASS"; pass=$((pass+1)); else OK="FAIL"; fail=$((fail+1)); fi
  printf '%-3s %-52s %-20s %-8s %-8s %-6s %s\n' "$n" "${MSG:0:52}" "$STATE" "$EXP" "$GOT" "$PATHK" "$OK"
done
printf '%.0s-' {1..112}; echo
echo "RESULT: $pass passed, $fail failed, of $n   (code path = deterministic, model path = qwen3)"
