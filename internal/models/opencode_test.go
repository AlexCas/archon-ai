package models

import (
	"reflect"
	"testing"
)

func TestParseModels(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "strips provider prefix",
			in:   "opencode-go/glm-5.2\nopencode-go/kimi-k2.6",
			want: []string{"glm-5.2", "kimi-k2.6"},
		},
		{
			name: "skips blank and whitespace-only lines",
			in:   "opencode-go/glm-5.2\n\n   \nopencode-go/qwen3.7-plus\n",
			want: []string{"glm-5.2", "qwen3.7-plus"},
		},
		{
			name: "keeps no-slash lines as-is",
			in:   "glm-5.2\nkimi-k2.6",
			want: []string{"glm-5.2", "kimi-k2.6"},
		},
		{
			name: "trims surrounding whitespace",
			in:   "  opencode-go/glm-5.2  \n\tplain-model\t",
			want: []string{"glm-5.2", "plain-model"},
		},
		{
			name: "drops bare provider/ with empty model name",
			in:   "opencode-go/\nopencode-go/glm-5.2\n/",
			want: []string{"glm-5.2"},
		},
		{
			name: "strips trailing CR from CRLF output",
			in:   "opencode-go/glm-5.2\r\nopencode-go/kimi-k2.6\r\n",
			want: []string{"glm-5.2", "kimi-k2.6"},
		},
		{
			name: "strips only the first slash (namespaced ids keep the rest)",
			in:   "provider/family/model",
			want: []string{"family/model"},
		},
		{
			name: "empty input yields nil",
			in:   "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseModels([]byte(tt.in))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseModels(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
