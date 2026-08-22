package route

import "regexp"

// judgeVerbs and verifyVerbs are the single source of truth shared by BOTH
// the D3 dual-action ambiguity rule and the keyword table's judge/verify
// rows (see keywordTable below). They MUST NOT be duplicated elsewhere in
// this package — any code that needs a judge or verify verb set references
// these same symbols so the D3 rule and the keyword table can never drift
// apart. See design.md decision A2.
var judgeVerbs = []string{"revisa", "review", "juzga", "dictamen"}
var verifyVerbs = []string{"prueba", "pruebas", "test", "valida", "verify"}

// conjunctions are the coordinating conjunctions the D3 rule looks for
// alongside a judge-verb and a verify-verb.
var conjunctions = []string{"y", "and", "e"}

// controlWords trigger the "control" rule (resume / next / next-nochange).
var controlWords = []string{"siguiente", "continuemos", "adelante", "sigamos", "continua"}

// implicitVerbs trigger the "implicit" start/resume rule. Deliberately
// narrow: conjugated phase verbs (e.g. "disenemos") are NOT implicit-start
// verbs, they fall through to the keyword table instead (fixture #11).
var implicitVerbs = []string{"trabajemos", "empecemos", "comencemos", "hagamos", "armemos", "pongamonos", "arranquemos"}

// judgeKeywordExtras and verifyKeywordExtras extend judgeVerbs/verifyVerbs
// with additional synonyms for the keyword table ONLY — the D3 rule never
// sees these, keeping the ambiguity check narrow per spec D3 rationale.
var judgeKeywordExtras = []string{"revisar", "revision", "revisa el codigo", "dual review", "judge"}
var verifyKeywordExtras = []string{"verifica", "corre las pruebas", "run tests", "validate"}

// keywordTable maps each SDD phase to its Spanish + English keyword set
// (ROUTER.md). The judge and verify rows are built from judgeVerbs and
// verifyVerbs (concat, above) so they cannot diverge from the D3 rule's verb
// sets. Every phase carries at least two Spanish and two English keywords.
//
// Note: "tareas" is intentionally NOT a bare tasks keyword — "Implementa las
// tareas" must resolve to apply (via "implementa"), not collide with tasks
// (fixture #12). The more specific "desglose"/"plan de tareas" phrasing
// still covers the tasks-phase fixtures without that collision.
var keywordTable = map[string][]string{
	"explore": {"explora", "exploremos", "investiga", "entender", "analiza el codigo", "explore", "investigate", "understand", "research"},
	"propose": {"propon", "propuesta", "propongamos", "idea", "enfoque", "propose", "proposal", "approach", "suggest"},
	"spec":    {"especificacion", "spec", "requisitos", "gherkin", "escenarios", "specification", "requirements"},
	"design":  {"diseno", "disenemos", "arquitectura", "plan tecnico", "design", "architecture", "technical plan"},
	"tasks":   {"desglose", "checklist", "plan de tareas", "to-do", "tasks", "breakdown", "work items"},
	"apply":   {"implementa", "codifica", "aplica", "escribe el codigo", "construye", "apply", "implement", "code", "build"},
	"verify":  concat(verifyVerbs, verifyKeywordExtras),
	"judge":   concat(judgeVerbs, judgeKeywordExtras),
	"archive": {"archiva", "finaliza", "cierra", "completa el cambio", "archive", "finalize", "close", "complete"},
}

// concat returns a fresh slice containing a followed by b. It never mutates
// either input's backing array (unlike append(a, b...) on a shared slice).
func concat(a, b []string) []string {
	out := make([]string, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// phaseAliases maps a navigation-position noun to its target phase, for the
// explicit-agent rule. Canonical phase names map to themselves; a small set
// of derived nouns (e.g. "exploracion") map to their phase.
var phaseAliases = map[string]string{
	"explore":     "explore",
	"propose":     "propose",
	"spec":        "spec",
	"design":      "design",
	"tasks":       "tasks",
	"apply":       "apply",
	"verify":      "verify",
	"judge":       "judge",
	"archive":     "archive",
	"exploracion": "explore",
}

// archonAgentRe matches a literal archon-<phase> token, e.g. "archon-design".
var archonAgentRe = regexp.MustCompile(`archon-(explore|propose|spec|design|tasks|apply|verify|judge|archive)`)

// navigationRe matches a navigation-marker phrase followed by a phase noun,
// e.g. "lanza el agente de exploracion", "corre el apply", "volvamos al spec",
// "regresa a design", "salta a tasks", "vamos a la fase spec".
var navigationRe = regexp.MustCompile(`(?:lanza(?: el agente de| el)?|agente de|corre el|vamos a(?: la fase)?|volvamos al?(?: la fase)?|regresa al?(?: la fase)?|salta al?(?: la fase)?)\s+([a-z]+)`)

// matchAny reports whether msg contains any of words as a whole-word (or
// whole-phrase, for multi-word entries) match.
func matchAny(msg string, words []string) bool {
	for _, w := range words {
		if wordBoundaryMatch(msg, w) {
			return true
		}
	}
	return false
}

func wordBoundaryMatch(msg, phrase string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(phrase) + `\b`)
	return re.MatchString(msg)
}
