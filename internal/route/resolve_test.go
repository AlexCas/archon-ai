package route

import "testing"

// TestResolve ports the 18 fixtures from prototype/sdd-router/fixtures.md
// and openspec/changes/local-model-router/specs/local-model-router/
// local-model-router.feature against the pure Resolve function — no
// filesystem, no subprocess. Fixture #16 asserts the CLASSIFY fallthrough
// (model-path signal); #16b (model classifier returning ASK) is model-path
// and is intentionally NOT unit-tested here, per design's Testing Strategy.
func TestResolve(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want Result
	}{
		{
			name: "fixture 1: implicit start, no active change (trabajemos)",
			in:   Input{Message: "Trabajemos en esta especificacion", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "implicit-start", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 2: implicit resume, spec in_progress (trabajemos)",
			in:   Input{Message: "Trabajemos en esta especificacion", Phase: "spec", Status: "in_progress", ActiveChange: "test-change"},
			want: Result{Phase: "spec", Rule: "implicit-resume", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 3: implicit start, no active change (empecemos)",
			in:   Input{Message: "Empecemos con esto", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "implicit-start", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 4: implicit start, no active change (hagamos)",
			in:   Input{Message: "Hagamos esta feature", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "implicit-start", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 5: explicit agent wins over co-present implicit verb",
			in:   Input{Message: "Hagamos esta especificacion. Lanza el agente de exploracion", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "explicit-agent", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 6: control next, propose/completed -> spec",
			in:   Input{Message: "Continuemos", Phase: "propose", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "spec", Rule: "next", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 7: control next, apply/completed -> verify",
			in:   Input{Message: "Siguiente", Phase: "apply", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "verify", Rule: "next", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 8: control resume, design/in_progress",
			in:   Input{Message: "Continuemos", Phase: "design", Status: "in_progress", ActiveChange: "test-change"},
			want: Result{Phase: "design", Rule: "resume", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 9: control next-nochange, no active change (adelante)",
			in:   Input{Message: "Adelante", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "next-nochange", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 10: keyword explora -> explore",
			in:   Input{Message: "Explora el codigo de billing", Phase: "", Status: "", ActiveChange: "none"},
			want: Result{Phase: "explore", Rule: "keyword", Path: "code", ActiveChange: "none"},
		},
		{
			name: "fixture 11: keyword disenemos -> design (not implicit)",
			in:   Input{Message: "Disenemos la arquitectura del API", Phase: "spec", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "design", Rule: "keyword", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 12: keyword implementa -> apply, not tasks",
			in:   Input{Message: "Implementa las tareas", Phase: "tasks", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "apply", Rule: "keyword", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 13: keyword corre las pruebas -> verify",
			in:   Input{Message: "Corre las pruebas", Phase: "apply", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "verify", Rule: "keyword", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 14: keyword archiva -> archive",
			in:   Input{Message: "Archiva el cambio", Phase: "judge", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "archive", Rule: "keyword", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 15: D3 dual-action judge+verify+conjunction -> ASK",
			in:   Input{Message: "Revisa y prueba esto", Phase: "verify", Status: "in_progress", ActiveChange: "test-change"},
			want: Result{Phase: "ASK", Rule: "ambiguous", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 16: unclassifiable falls through to CLASSIFY (model path)",
			in:   Input{Message: "Que opinas del clima?", Phase: "spec", Status: "in_progress", ActiveChange: "test-change"},
			want: Result{Phase: "CLASSIFY", Rule: "classify", Path: "model", ActiveChange: "test-change"},
		},
		{
			name: "fixture 17: backward nav resolves via explicit-agent (blocking is harness-workflow's job)",
			in:   Input{Message: "Volvamos al spec", Phase: "design", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "spec", Rule: "explicit-agent", Path: "code", ActiveChange: "test-change"},
		},
		{
			name: "fixture 18: literal phase token in imperative position -> explicit-agent",
			in:   Input{Message: "corre el apply", Phase: "tasks", Status: "completed", ActiveChange: "test-change"},
			want: Result{Phase: "apply", Rule: "explicit-agent", Path: "code", ActiveChange: "test-change"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.in)
			if got.Phase != tt.want.Phase || got.Rule != tt.want.Rule || got.Path != tt.want.Path {
				t.Errorf("Resolve(%+v) = %+v, want %+v", tt.in, got, tt.want)
			}
			if got.ActiveChange != tt.want.ActiveChange {
				t.Errorf("Resolve(%+v).ActiveChange = %q, want %q", tt.in, got.ActiveChange, tt.want.ActiveChange)
			}
		})
	}
}

// TestResolveKeywordOutline covers the feature file's "Keyword table: all
// nine SDD phases covered" Scenario Outline.
func TestResolveKeywordOutline(t *testing.T) {
	tests := []struct {
		message string
		phase   string
	}{
		{"Explora el codigo de billing", "explore"},
		{"Propongamos el enfoque", "propose"},
		{"Escribe los requisitos gherkin", "spec"},
		{"Disenemos la arquitectura", "design"},
		{"Haz el desglose de tareas", "tasks"},
		{"Implementa las tareas", "apply"},
		{"Corre las pruebas", "verify"},
		{"Juzga el codigo", "judge"},
		{"Archiva el cambio", "archive"},
	}

	for _, tt := range tests {
		t.Run(tt.phase, func(t *testing.T) {
			got := Resolve(Input{Message: tt.message, Phase: "spec", Status: "completed", ActiveChange: "test-change"})
			if got.Phase != tt.phase || got.Rule != "keyword" {
				t.Errorf("Resolve(%q) = {phase:%q rule:%q}, want {phase:%q rule:keyword}", tt.message, got.Phase, got.Rule, tt.phase)
			}
		})
	}
}

// TestNormalize covers spec §Text Normalization: lowercase + diacritic
// strip, applied before any rule evaluation.
func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"accented word strips to plain ascii", "especificación", "especificacion"},
		{"uppercase lowers", "TRABAJEMOS", "trabajemos"},
		{"combined accent + uppercase", "Diseño", "diseno"},
		{"full accented sentence", "Trabajemos en esta especificación", "trabajemos en esta especificacion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestImplicitAboveKeyword covers spec §Implicit-above-keyword Precedence:
// a start verb with a loose phase noun must resolve to implicit-start, not
// keyword.
func TestImplicitAboveKeyword(t *testing.T) {
	got := Resolve(Input{Message: "Trabajemos en esta especificacion", Phase: "", Status: "", ActiveChange: "none"})
	if got.Rule != "implicit-start" {
		t.Errorf("Rule = %q, want implicit-start", got.Rule)
	}
	if got.Rule == "keyword" {
		t.Errorf("Rule must NOT be keyword")
	}
}
