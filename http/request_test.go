package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRequestInput(t *testing.T) {
	tests := []struct {
		name      string
		input     []any
		want      RequestInput
		wantError string
	}{
		{
			name: "minimal get",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/api/v1/query",
			}},
			want: RequestInput{
				Method:  "GET",
				Path:    "/api/v1/query",
				Headers: map[string]string{},
				Query:   map[string]string{},
			},
		},
		{
			name: "with query and headers",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/x",
				"query":  map[string]any{"q": "up", "n": float64(1)},
				"headers": map[string]any{
					"X-Foo": "bar",
					"X-B":   true,
				},
			}},
			want: RequestInput{
				Method: "GET",
				Path:   "/x",
				Headers: map[string]string{
					"X-Foo": "bar",
					"X-B":   "true",
				},
				Query: map[string]string{
					"q": "up",
					"n": "1",
				},
			},
		},
		{
			name: "skip nil query value",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/y",
				"query":  map[string]any{"a": nil, "b": "c"},
			}},
			want: RequestInput{
				Method:  "GET",
				Path:    "/y",
				Headers: map[string]string{},
				Query:   map[string]string{"b": "c"},
			},
		},
		{
			name: "skip nil header",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/y",
				"headers": map[string]any{
					"X-Optional": nil,
					"X-Set":      "v",
				},
			}},
			want: RequestInput{
				Method: "GET",
				Path:   "/y",
				Headers: map[string]string{
					"X-Set": "v",
				},
				Query: map[string]string{},
			},
		},
		{
			name:      "wrong arity",
			input:     []any{},
			wantError: "expected input object",
		},
		{
			name:      "input not object",
			input:     []any{"x"},
			wantError: "input must be an object",
		},
		{
			name: "missing method",
			input: []any{map[string]any{
				"path": "/",
			}},
			wantError: "method must be a non-empty string",
		},
		{
			name: "method not get",
			input: []any{map[string]any{
				"method": "POST",
				"path":   "/",
			}},
			wantError: "method must be GET",
		},
		{
			name: "missing path",
			input: []any{map[string]any{
				"method": "GET",
			}},
			wantError: "path must be a non-empty string",
		},
		{
			name: "headers not object",
			input: []any{map[string]any{
				"method":  "GET",
				"path":    "/",
				"headers": "bad",
			}},
			wantError: "headers must be an object",
		},
		{
			name: "query not object",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/",
				"query":  []any{},
			}},
			wantError: "query must be an object",
		},
		{
			name: "bad header value",
			input: []any{map[string]any{
				"method": "GET",
				"path":   "/",
				"headers": map[string]any{
					"X": map[string]any{},
				},
			}},
			wantError: `header "X"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRequestInput(tt.input)
			if tt.wantError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
