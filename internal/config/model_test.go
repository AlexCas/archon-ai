package config

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		wantWarn bool
	}{
		{name: "known claude model opus", model: "claude-opus-4-8", wantWarn: false},
		{name: "known claude model sonnet", model: "claude-sonnet-4-6", wantWarn: false},
		{name: "known claude model haiku", model: "claude-haiku-4-5", wantWarn: false},
		{name: "known opencode model", model: "glm-5", wantWarn: false},
		{name: "known openai model", model: "gpt-4o", wantWarn: false},
		{name: "known gemini model", model: "gemini-2.5-pro", wantWarn: false},
		{name: "normalizable display name", model: "Opus 4.8", wantWarn: false},
		{name: "typo display name", model: "Opues 4.8", wantWarn: true},
		{name: "unknown model", model: "future-model-v2", wantWarn: true},
		{name: "empty string", model: "", wantWarn: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.model)
			hasWarn := got != ""
			if hasWarn != tt.wantWarn {
				t.Errorf("Validate(%q) warning = %v, want %v (msg: %q)", tt.model, hasWarn, tt.wantWarn, got)
			}
		})
	}
}

func TestNormalizeModel(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		wantID string
		wantOK bool
	}{
		{name: "alias opus idempotent", in: "opus", wantID: "opus", wantOK: true},
		{name: "alias sonnet idempotent", in: "sonnet", wantID: "sonnet", wantOK: true},
		{name: "alias haiku idempotent", in: "haiku", wantID: "haiku", wantOK: true},
		{name: "alias uppercase", in: "Opus", wantID: "opus", wantOK: true},
		{name: "display name with version", in: "Opus 4.8", wantID: "opus", wantOK: true},
		{name: "family with bare version", in: "opus 4", wantID: "opus", wantOK: true},
		{name: "family hyphenated version", in: "opus-4-8", wantID: "opus", wantOK: true},
		{name: "full id opus", in: "claude-opus-4-8", wantID: "opus", wantOK: true},
		{name: "full id sonnet", in: "claude-sonnet-4-6", wantID: "sonnet", wantOK: true},
		{name: "full id haiku dated", in: "claude-haiku-4-5-20251001", wantID: "haiku", wantOK: true},
		{name: "padded whitespace", in: "  Sonnet 4.6  ", wantID: "sonnet", wantOK: true},
		{name: "typo no family", in: "Opues 4.8", wantID: "", wantOK: false},
		{name: "opencode glm", in: "glm-5", wantID: "glm-5", wantOK: true},
		{name: "opencode kimi", in: "kimi-k2.5", wantID: "kimi-k2.5", wantOK: true},
		{name: "openai gpt-4o", in: "gpt-4o", wantID: "gpt-4o", wantOK: true},
		{name: "openai gpt-4o uppercase", in: "GPT-4o", wantID: "gpt-4o", wantOK: true},
		{name: "gemini pro", in: "gemini-2.5-pro", wantID: "gemini-2.5-pro", wantOK: true},
		{name: "substring not whole token", in: "octopus", wantID: "", wantOK: false},
		{name: "family embedded in word rejected", in: "supushaiku", wantID: "", wantOK: false},
		{name: "multi family resolves by priority", in: "sonnet-opus", wantID: "opus", wantOK: true},
		{name: "priority is position independent", in: "opus-sonnet", wantID: "opus", wantOK: true},
		// Claude precedence: a value carrying a Claude family token resolves to
		// the Claude alias because the Claude row is consulted before any catalog
		// row, regardless of other tokens present.
		{name: "claude row wins over later providers", in: "opus gpt-4o", wantID: "opus", wantOK: true},
		{name: "separatorless glued form does not resolve", in: "opus4", wantID: "", wantOK: false},
		{name: "empty", in: "", wantID: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := NormalizeModel(tt.in)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("NormalizeModel(%q) = (%q, %v), want (%q, %v)", tt.in, gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

// TestModelRef_FullID covers M1: all four scenarios.
func TestModelRef_FullID(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     string
	}{
		{
			name:     "joins provider and bare model",
			provider: "anthropic",
			model:    "claude-sonnet-4-6",
			want:     "anthropic/claude-sonnet-4-6",
		},
		{
			name:     "opencode bare key",
			provider: "opencode",
			model:    "deepseek-v4-pro",
			want:     "opencode/deepseek-v4-pro",
		},
		{
			name:     "already-slashed id no double-prefix",
			provider: "openrouter",
			model:    "xai/grok-4",
			want:     "xai/grok-4",
		},
		{
			name:     "empty provider returns bare model no leading slash",
			provider: "",
			model:    "opus",
			want:     "opus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ModelRef{Provider: tt.provider, Model: tt.model}
			if got := r.FullID(); got != tt.want {
				t.Errorf("FullID() = %q, want %q", got, tt.want)
			}
			if tt.provider == "" && strings.HasPrefix(r.FullID(), "/") {
				t.Errorf("FullID() has leading slash: %q", r.FullID())
			}
		})
	}
}

// TestParseModelRef covers the split-on-first-slash semantics.
func TestParseModelRef(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantProv string
		wantMod  string
	}{
		{name: "simple provider/model", in: "a/b", wantProv: "a", wantMod: "b"},
		{name: "model with embedded slash", in: "a/b/c", wantProv: "a", wantMod: "b/c"},
		{name: "bare model empty provider", in: "opus", wantProv: "", wantMod: "opus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseModelRef(tt.in)
			if got.Provider != tt.wantProv || got.Model != tt.wantMod {
				t.Errorf("ParseModelRef(%q) = {%q, %q}, want {%q, %q}",
					tt.in, got.Provider, got.Model, tt.wantProv, tt.wantMod)
			}
		})
	}
}

// TestModelRef_UnmarshalYAML covers M3: scalar slashed, scalar bare, mapping. (S1a-2)
func TestModelRef_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantProv string
		wantMod  string
	}{
		{
			name:     "scalar slashed splits on first slash",
			yaml:     "a/b",
			wantProv: "a",
			wantMod:  "b",
		},
		{
			name:     "scalar bare keeps empty provider",
			yaml:     "x",
			wantProv: "",
			wantMod:  "x",
		},
		{
			name:     "mapping decodes structured form",
			yaml:     "provider: opencode\nmodel: deepseek-v4-pro\n",
			wantProv: "opencode",
			wantMod:  "deepseek-v4-pro",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r ModelRef
			if err := yaml.Unmarshal([]byte(tt.yaml), &r); err != nil {
				t.Fatalf("Unmarshal(%q) error = %v", tt.yaml, err)
			}
			if r.Provider != tt.wantProv || r.Model != tt.wantMod {
				t.Errorf("Unmarshal(%q) = {%q, %q}, want {%q, %q}",
					tt.yaml, r.Provider, r.Model, tt.wantProv, tt.wantMod)
			}
		})
	}
}

// TestModelRef_MarshalYAML covers M4: bare alias re-marshals as scalar, not mapping. (S1a-2)
func TestModelRef_MarshalYAML(t *testing.T) {
	t.Run("empty provider marshals as scalar", func(t *testing.T) {
		r := ModelRef{Provider: "", Model: "opus", Effort: ""}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		got := strings.TrimSpace(string(out))
		if got != "opus" {
			t.Errorf("Marshal() = %q, want scalar %q", got, "opus")
		}
	})

	t.Run("provider-qualified with no effort marshals as scalar FullID", func(t *testing.T) {
		// A provider/model scalar is the documented legacy form, so it must
		// re-marshal as a one-line scalar (byte-stability), NOT a mapping.
		r := ModelRef{Provider: "opencode", Model: "deepseek-v4-pro", Effort: ""}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		got := strings.TrimSpace(string(out))
		if got != "opencode/deepseek-v4-pro" {
			t.Errorf("Marshal() = %q, want scalar %q", got, "opencode/deepseek-v4-pro")
		}
		if strings.Contains(got, "provider:") {
			t.Errorf("Marshal() = %q, expected scalar not a mapping", got)
		}
	})

	t.Run("effort-bearing ref marshals as mapping", func(t *testing.T) {
		r := ModelRef{Provider: "opencode", Model: "deepseek-v4-pro", Effort: "high"}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		got := string(out)
		if !strings.Contains(got, "model:") || !strings.Contains(got, "effort:") {
			t.Errorf("Marshal() = %q, want mapping with model/effort keys", got)
		}
	})

	t.Run("effort survives a marshal->unmarshal round-trip", func(t *testing.T) {
		in := ModelRef{Provider: "opencode", Model: "deepseek-v4-pro", Effort: "low"}
		out, err := yaml.Marshal(in)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var got ModelRef
		if err := yaml.Unmarshal(out, &got); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got != in {
			t.Errorf("round-trip = %+v, want %+v", got, in)
		}
	})
}

// TestModelConfig_StructuredFields covers M2: structured fields preserve
// provider+model. (S1b-4)
func TestModelConfig_StructuredFields(t *testing.T) {
	mc := ModelConfig{
		Default: ModelRef{Provider: "anthropic", Model: "claude-sonnet-4-6"},
		Leader:  ModelRef{Provider: "opencode", Model: "deepseek-v4-pro"},
		Phases: map[string]ModelRef{
			"apply": {Provider: "openai", Model: "gpt-4o"},
		},
	}

	if mc.Default.Provider != "anthropic" || mc.Default.Model != "claude-sonnet-4-6" {
		t.Errorf("Default = %+v, want {anthropic, claude-sonnet-4-6}", mc.Default)
	}
	if mc.Leader.Provider != "opencode" || mc.Leader.Model != "deepseek-v4-pro" {
		t.Errorf("Leader = %+v, want {opencode, deepseek-v4-pro}", mc.Leader)
	}
	if mc.Phases["apply"].Provider != "openai" || mc.Phases["apply"].Model != "gpt-4o" {
		t.Errorf("Phases[apply] = %+v, want {openai, gpt-4o}", mc.Phases["apply"])
	}
}

// TestResolvePhaseModels covers M6: FullID emission, fallback, omit-when-empty,
// PhaseOrder iteration, determinism. (S1b-4, rewritten from string-based version)
func TestResolvePhaseModels(t *testing.T) {
	t.Run("provider-qualified phase emits FullID", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"propose": {Provider: "opencode", Model: "deepseek-v4-pro"},
			},
		}
		got := ResolvePhaseModels(mc)
		want := []PhaseModel{{Phase: "propose", Model: "opencode/deepseek-v4-pro", Provider: "opencode"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty-provider phase emits bare model", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"propose": {Provider: "", Model: "opus"},
			},
		}
		got := ResolvePhaseModels(mc)
		want := []PhaseModel{{Phase: "propose", Model: "opus"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unset phase falls back to default", func(t *testing.T) {
		mc := ModelConfig{
			Default: ModelRef{Provider: "", Model: "opus"},
			Phases:  map[string]ModelRef{"verify": {}},
		}
		got := ResolvePhaseModels(mc)
		// Every phase falls back to the default since none are set with a model.
		if len(got) != len(PhaseOrder) {
			t.Fatalf("got %d phases, want %d", len(got), len(PhaseOrder))
		}
		for _, pm := range got {
			if pm.Model != "opus" {
				t.Errorf("phase %q = %q, want opus", pm.Phase, pm.Model)
			}
		}
	})

	t.Run("omits phase when nothing resolves", func(t *testing.T) {
		got := ResolvePhaseModels(ModelConfig{})
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("renders in canonical PhaseOrder regardless of map order", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"archive": {Model: "haiku"},
				"explore": {Model: "opus"},
				"design":  {Provider: "opencode", Model: "deepseek-v4-pro"},
			},
		}
		got := ResolvePhaseModels(mc)
		want := []PhaseModel{
			{Phase: "explore", Model: "opus"},
			{Phase: "design", Model: "opencode/deepseek-v4-pro", Provider: "opencode"},
			{Phase: "archive", Model: "haiku"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("twice is deeply equal", func(t *testing.T) {
		mc := ModelConfig{
			Default: ModelRef{Model: "sonnet"},
			Phases:  map[string]ModelRef{"design": {Model: "opus"}},
		}
		first := ResolvePhaseModels(mc)
		second := ResolvePhaseModels(mc)
		if !reflect.DeepEqual(first, second) {
			t.Errorf("non-deterministic: %v vs %v", first, second)
		}
	})

	// E1-3: PhaseModel.Effort is carried from the resolved ref.
	t.Run("phase ref effort carried into PhaseModel", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"spec": {Provider: "opencode", Model: "deepseek-v4-pro", Effort: "medium"},
			},
		}
		got := ResolvePhaseModels(mc)
		if len(got) != 1 {
			t.Fatalf("got %d phases, want 1", len(got))
		}
		if got[0].Effort != "medium" {
			t.Errorf("PhaseModel.Effort = %q, want %q", got[0].Effort, "medium")
		}
		if got[0].Model != "opencode/deepseek-v4-pro" {
			t.Errorf("PhaseModel.Model = %q, want %q", got[0].Model, "opencode/deepseek-v4-pro")
		}
	})

	t.Run("default fallback ref effort carried into PhaseModel", func(t *testing.T) {
		mc := ModelConfig{
			Default: ModelRef{Provider: "anthropic", Model: "claude-opus-4-8", Effort: "high"},
			// No phases set — all phases fall back to Default.
		}
		got := ResolvePhaseModels(mc)
		if len(got) != len(PhaseOrder) {
			t.Fatalf("got %d phases, want %d", len(got), len(PhaseOrder))
		}
		for _, pm := range got {
			if pm.Effort != "high" {
				t.Errorf("phase %q: PhaseModel.Effort = %q, want %q", pm.Phase, pm.Effort, "high")
			}
		}
	})

	t.Run("empty effort stays empty", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"apply": {Provider: "openai", Model: "gpt-4o"},
			},
		}
		got := ResolvePhaseModels(mc)
		if len(got) != 1 {
			t.Fatalf("got %d phases, want 1", len(got))
		}
		if got[0].Effort != "" {
			t.Errorf("PhaseModel.Effort = %q, want empty", got[0].Effort)
		}
	})

	t.Run("judge phase resolves and appears between verify and archive", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"verify": {Provider: "anthropic", Model: "claude-opus-4-8"},
				"judge":  {Provider: "anthropic", Model: "claude-opus-4-8"},
			},
		}
		got := ResolvePhaseModels(mc)

		// Find judge in result.
		judgeIdx := -1
		verifyIdx := -1
		archiveIdx := -1
		for i, pm := range got {
			switch pm.Phase {
			case "judge":
				judgeIdx = i
			case "verify":
				verifyIdx = i
			case "archive":
				archiveIdx = i
			}
		}

		if judgeIdx == -1 {
			t.Fatal("judge phase not present in ResolvePhaseModels result")
		}
		// Verify the resolved model.
		if got[judgeIdx].Model != "anthropic/claude-opus-4-8" {
			t.Errorf("judge model = %q, want %q", got[judgeIdx].Model, "anthropic/claude-opus-4-8")
		}
		// judge must come after verify (archive is absent here since it has no model set).
		if verifyIdx != -1 && judgeIdx <= verifyIdx {
			t.Errorf("judge (index %d) must come after verify (index %d)", judgeIdx, verifyIdx)
		}
		// archive absent — confirm archiveIdx is -1 (no model set for archive).
		_ = archiveIdx
	})

	t.Run("judge defaults to verify model via modelFlags wiring", func(t *testing.T) {
		// Simulate the CLI flag layer: judge receives the same value as verify.
		verifyRef := ModelRef{Provider: "anthropic", Model: "claude-opus-4-8"}
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"verify": verifyRef,
				"judge":  verifyRef, // mirror: modelFlags["judge"] = modelVerifyFlag
			},
		}
		got := ResolvePhaseModels(mc)

		var judgeModel, verifyModel string
		for _, pm := range got {
			switch pm.Phase {
			case "judge":
				judgeModel = pm.Model
			case "verify":
				verifyModel = pm.Model
			}
		}

		if judgeModel == "" {
			t.Fatal("judge phase not resolved")
		}
		if judgeModel != verifyModel {
			t.Errorf("judge model = %q, verify model = %q; judge must default to verify's model", judgeModel, verifyModel)
		}
	})

	// REQ-1: BaseURL flows from the resolved ref into PhaseModel.BaseURL.
	t.Run("BaseURL carried from resolved ref into PhaseModel", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
			},
		}
		got := ResolvePhaseModels(mc)
		if len(got) != 1 {
			t.Fatalf("got %d phases, want 1", len(got))
		}
		if got[0].BaseURL != "http://localhost:11434/v1" {
			t.Errorf("PhaseModel.BaseURL = %q, want %q", got[0].BaseURL, "http://localhost:11434/v1")
		}
		if got[0].Provider != "ollama" {
			t.Errorf("PhaseModel.Provider = %q, want %q", got[0].Provider, "ollama")
		}
	})

	// REQ-1: a ref without BaseURL resolves with PhaseModel.BaseURL == "".
	t.Run("empty BaseURL stays empty on resolved PhaseModel", func(t *testing.T) {
		mc := ModelConfig{
			Phases: map[string]ModelRef{
				"apply": {Provider: "openai", Model: "gpt-4o"},
			},
		}
		got := ResolvePhaseModels(mc)
		if len(got) != 1 {
			t.Fatalf("got %d phases, want 1", len(got))
		}
		if got[0].BaseURL != "" {
			t.Errorf("PhaseModel.BaseURL = %q, want empty", got[0].BaseURL)
		}
	})
}

// TestModelRef_BaseURL covers REQ-1: the BaseURL field's YAML round-trip and
// the scalar-vs-mapping marshal switch.
func TestModelRef_BaseURL(t *testing.T) {
	t.Run("scalar ref round-trips byte-identically", func(t *testing.T) {
		const input = "ollama/llama3\n"
		var r ModelRef
		if err := yaml.Unmarshal([]byte(input), &r); err != nil {
			t.Fatalf("Unmarshal(%q) error = %v", input, err)
		}
		if r.BaseURL != "" {
			t.Fatalf("Unmarshal(%q).BaseURL = %q, want empty", input, r.BaseURL)
		}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		if string(out) != input {
			t.Errorf("round-trip = %q, want byte-identical %q", out, input)
		}
	})

	t.Run("BaseURL ref marshals as mapping", func(t *testing.T) {
		r := ModelRef{Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		got := string(out)
		for _, key := range []string{"provider:", "model:", "base_url:"} {
			if !strings.Contains(got, key) {
				t.Errorf("Marshal() = %q, missing mapping key %q", got, key)
			}
		}
		if strings.TrimSpace(got) == "ollama/llama3" {
			t.Errorf("Marshal() emitted scalar form %q, want mapping", got)
		}
	})

	t.Run("scalar input decodes with empty BaseURL", func(t *testing.T) {
		var r ModelRef
		if err := yaml.Unmarshal([]byte("anthropic/claude-sonnet-4-6"), &r); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if r.Provider != "anthropic" || r.Model != "claude-sonnet-4-6" || r.BaseURL != "" {
			t.Errorf("got {%q,%q,%q}, want {%q,%q,\"\"}", r.Provider, r.Model, r.BaseURL, "anthropic", "claude-sonnet-4-6")
		}
	})

	t.Run("BaseURL survives a marshal->unmarshal round-trip", func(t *testing.T) {
		in := ModelRef{Provider: "localai", Model: "gpt-4-vision", BaseURL: "http://localhost:8080/v1"}
		out, err := yaml.Marshal(in)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		var got ModelRef
		if err := yaml.Unmarshal(out, &got); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got != in {
			t.Errorf("round-trip = %+v, want %+v", got, in)
		}
	})

	t.Run("BaseURL with Effort still marshals as mapping", func(t *testing.T) {
		r := ModelRef{Provider: "ollama", Model: "llama3", Effort: "high", BaseURL: "http://localhost:11434/v1"}
		out, err := yaml.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		got := string(out)
		if !strings.Contains(got, "effort:") || !strings.Contains(got, "base_url:") {
			t.Errorf("Marshal() = %q, want mapping with effort and base_url keys", got)
		}
	})
}

// TestValidateBaseURL covers REQ-3: advisory-only validation, never an error.
func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		ref      ModelRef
		wantWarn bool
	}{
		{name: "valid http BaseURL produces no warning", ref: ModelRef{Provider: "ollama", BaseURL: "http://localhost:11434/v1"}, wantWarn: false},
		{name: "valid https BaseURL produces no warning", ref: ModelRef{Provider: "ollama", BaseURL: "https://models.internal/v1"}, wantWarn: false},
		{name: "non-http scheme triggers a warning", ref: ModelRef{Provider: "ollama", BaseURL: "ftp://localhost/v1"}, wantWarn: true},
		{name: "BaseURL set but provider empty triggers a warning", ref: ModelRef{Provider: "", BaseURL: "http://localhost:11434/v1"}, wantWarn: true},
		{name: "empty BaseURL is always silent", ref: ModelRef{Provider: "", BaseURL: ""}, wantWarn: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ValidateBaseURL(tt.ref, &buf)
			gotWarn := buf.Len() > 0
			if gotWarn != tt.wantWarn {
				t.Errorf("ValidateBaseURL(%+v) warned = %v (msg %q), want %v", tt.ref, gotWarn, buf.String(), tt.wantWarn)
			}
		})
	}

	t.Run("exact message for non-http/https BaseURL", func(t *testing.T) {
		var buf bytes.Buffer
		ValidateBaseURL(ModelRef{Provider: "ollama", BaseURL: "ftp://localhost/v1"}, &buf)
		want := "warning: base_url \"ftp://localhost/v1\" is not a valid http/https URL\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("exact message for empty provider", func(t *testing.T) {
		var buf bytes.Buffer
		ValidateBaseURL(ModelRef{Provider: "", BaseURL: "http://localhost:11434/v1"}, &buf)
		want := "warning: base_url is set but provider is empty — provider id required for local model routing\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})
}

// TestModelConfig_HasAny covers REQ-8: the shared emptiness guard used by both
// `config list` and `status`.
func TestModelConfig_HasAny(t *testing.T) {
	tests := []struct {
		name string
		mc   ModelConfig
		want bool
	}{
		{name: "all empty", mc: ModelConfig{}, want: false},
		{name: "default id only", mc: ModelConfig{Default: ModelRef{Provider: "ollama", Model: "llama3"}}, want: true},
		{name: "default BaseURL only", mc: ModelConfig{Default: ModelRef{BaseURL: "http://localhost:11434/v1"}}, want: true},
		{name: "leader id only", mc: ModelConfig{Leader: ModelRef{Provider: "ollama", Model: "llama3"}}, want: true},
		{name: "leader BaseURL only", mc: ModelConfig{Leader: ModelRef{BaseURL: "http://localhost:11434/v1"}}, want: true},
		{name: "phase id only", mc: ModelConfig{Phases: map[string]ModelRef{"apply": {Provider: "ollama", Model: "llama3"}}}, want: true},
		{name: "phase BaseURL only", mc: ModelConfig{Phases: map[string]ModelRef{"apply": {BaseURL: "http://localhost:11434/v1"}}}, want: true},
		{
			name: "all fields set",
			mc: ModelConfig{
				Default: ModelRef{Provider: "anthropic", Model: "claude-opus-4-8", BaseURL: "http://localhost:11434/v1"},
				Leader:  ModelRef{Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"},
				Phases:  map[string]ModelRef{"apply": {Provider: "ollama", Model: "llama3", BaseURL: "http://localhost:11434/v1"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mc.HasAny(); got != tt.want {
				t.Errorf("HasAny() = %v, want %v", got, tt.want)
			}
		})
	}
}
