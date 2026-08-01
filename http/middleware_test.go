package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectBaseURL(t *testing.T) {
	tests := []struct {
		name             string
		jsonnetInput     map[string]any
		injectedBaseURL  string
		wantOriginalBase any
	}{
		{
			name: "sets baseURL when none is present",
			jsonnetInput: map[string]any{
				"method": "GET",
				"path":   "/x",
			},
			injectedBaseURL:  "https://eu-west-1.example.com",
			wantOriginalBase: nil,
		},
		{
			name: "overrides an existing baseURL",
			jsonnetInput: map[string]any{
				"method":  "GET",
				"path":    "/x",
				"baseURL": "https://static.example.com",
			},
			injectedBaseURL:  "https://eu-west-1.example.com",
			wantOriginalBase: "https://static.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []any{tt.jsonnetInput}

			got := injectBaseURL(args, tt.injectedBaseURL)

			require.Equal(t, tt.injectedBaseURL, got[0].(map[string]any)["baseURL"])
			require.Equal(t, tt.wantOriginalBase, args[0].(map[string]any)["baseURL"])
		})
	}
}

func TestInjectBody(t *testing.T) {
	tests := []struct {
		name             string
		jsonnetInput     map[string]any
		injectedBody     any
		wantOriginalBody any
	}{
		{
			name: "sets body when none is present",
			jsonnetInput: map[string]any{
				"method":   "POST",
				"path":     "/x",
				"readonly": true,
			},
			injectedBody: map[string]any{
				"name": "dynamic",
			},
			wantOriginalBody: nil,
		},
		{
			name: "overrides an existing body",
			jsonnetInput: map[string]any{
				"method":   "POST",
				"path":     "/x",
				"readonly": true,
				"body": map[string]any{
					"name": "jsonnet",
				},
			},
			injectedBody: map[string]any{
				"name": "dynamic",
			},
			wantOriginalBody: map[string]any{
				"name": "jsonnet",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []any{tt.jsonnetInput}

			got := injectBody(args, tt.injectedBody)

			require.Equal(t, tt.injectedBody, got[0].(map[string]any)["body"])
			require.Equal(t, tt.wantOriginalBody, args[0].(map[string]any)["body"])
		})
	}
}

func TestInjectHeaders(t *testing.T) {
	tests := []struct {
		name            string
		jsonnetHeaders  map[string]any
		injectedHeaders map[string]string
		override        bool
		want            map[string]any
	}{
		{
			name: "keeps jsonnet headers when override is false",
			jsonnetHeaders: map[string]any{
				"Authorization": "Bearer jsonnet",
				"X-Request":     "jsonnet",
			},
			injectedHeaders: map[string]string{
				"Authorization": "Bearer dynamic",
				"X-Dynamic":     "dynamic",
			},
			override: false,
			want: map[string]any{
				"Authorization": "Bearer jsonnet",
				"X-Request":     "jsonnet",
				"X-Dynamic":     "dynamic",
			},
		},
		{
			name: "overrides jsonnet headers when override is true",
			jsonnetHeaders: map[string]any{
				"Authorization": "Bearer jsonnet",
				"X-Request":     "jsonnet",
			},
			injectedHeaders: map[string]string{
				"Authorization": "Bearer dynamic",
				"X-Dynamic":     "dynamic",
			},
			override: true,
			want: map[string]any{
				"Authorization": "Bearer dynamic",
				"X-Request":     "jsonnet",
				"X-Dynamic":     "dynamic",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []any{map[string]any{
				"method":  "GET",
				"path":    "/x",
				"headers": tt.jsonnetHeaders,
			}}

			got := injectHeaders(args, tt.injectedHeaders, tt.override)

			headers := got[0].(map[string]any)["headers"]
			require.Equal(t, tt.want, headers)
			require.Equal(t, "Bearer jsonnet", args[0].(map[string]any)["headers"].(map[string]any)["Authorization"])
		})
	}
}
