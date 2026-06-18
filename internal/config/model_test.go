package config

import (
	"reflect"
	"testing"
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
		{name: "opencode glm", in: "glm-5", wantID: "", wantOK: false},
		{name: "opencode kimi", in: "kimi-k2.5", wantID: "", wantOK: false},
		{name: "gpt", in: "gpt-4", wantID: "", wantOK: false},
		{name: "substring not whole token", in: "octopus", wantID: "", wantOK: false},
		{name: "family embedded in word rejected", in: "supushaiku", wantID: "", wantOK: false},
		{name: "multi family resolves by priority", in: "sonnet-opus", wantID: "opus", wantOK: true},
		{name: "priority is position independent", in: "opus-sonnet", wantID: "opus", wantOK: true},
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

func TestResolvePhaseModels(t *testing.T) {
	t.Run("explicit phase resolves to its alias", func(t *testing.T) {
		got := ResolvePhaseModels(ModelConfig{Phases: map[string]string{"propose": "Sonnet 4.6"}})
		want := []PhaseModel{{Phase: "propose", Model: "sonnet"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unset phase falls back to default", func(t *testing.T) {
		got := ResolvePhaseModels(ModelConfig{Default: "Opus 4.8", Phases: map[string]string{"verify": ""}})
		// Every phase falls back to the default since none are set.
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

	t.Run("renders in canonical order regardless of map order", func(t *testing.T) {
		mc := ModelConfig{Phases: map[string]string{
			"archive": "haiku",
			"explore": "opus",
			"design":  "Opus 4.8",
		}}
		got := ResolvePhaseModels(mc)
		want := []PhaseModel{
			{Phase: "explore", Model: "opus"},
			{Phase: "design", Model: "opus"},
			{Phase: "archive", Model: "haiku"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("twice is deeply equal", func(t *testing.T) {
		mc := ModelConfig{Default: "sonnet", Phases: map[string]string{"design": "opus"}}
		first := ResolvePhaseModels(mc)
		second := ResolvePhaseModels(mc)
		if !reflect.DeepEqual(first, second) {
			t.Errorf("non-deterministic: %v vs %v", first, second)
		}
	})

	t.Run("unresolvable phase value falls through to default", func(t *testing.T) {
		mc := ModelConfig{Default: "haiku", Phases: map[string]string{"propose": "Opues 4.8"}}
		got := ResolvePhaseModels(mc)
		for _, pm := range got {
			if pm.Phase == "propose" && pm.Model != "haiku" {
				t.Errorf("propose = %q, want haiku (fell back to default)", pm.Model)
			}
		}
	})
}
