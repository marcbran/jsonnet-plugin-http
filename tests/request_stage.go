//go:build e2e

package tests

import (
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marcbran/jpoet/pkg/jpoet"
	pluginhttp "github.com/marcbran/jsonnet-plugin-http/http"
	"github.com/stretchr/testify/require"
)

type Stage struct {
	t      require.TestingT
	srv    *httptest.Server
	plugin *jpoet.Plugin
	out    map[string]any
	err    error
}

func scenario(t *testing.T) (*Stage, *Stage, *Stage) {
	s := &Stage{t: t}
	return s, s, s
}

func (s *Stage) and() *Stage {
	return s
}

func (s *Stage) a_server_returning_empty_body() *Stage {
	s.srv = httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
	}))
	s.t.(*testing.T).Cleanup(s.srv.Close)
	s.plugin = pluginhttp.Plugin("http", pluginhttp.WithBaseURL(s.srv.URL), pluginhttp.WithHTTPClient(s.srv.Client()))
	return s
}

func (s *Stage) a_server_returning_not_found(body string) *Stage {
	s.srv = httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNotFound)
		_, err := w.Write([]byte(body))
		require.NoError(s.t, err)
	}))
	s.t.(*testing.T).Cleanup(s.srv.Close)
	s.plugin = pluginhttp.Plugin("http", pluginhttp.WithBaseURL(s.srv.URL), pluginhttp.WithHTTPClient(s.srv.Client()))
	return s
}

func (s *Stage) a_slow_server_and_short_timeout() *Stage {
	s.srv = httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"ok":true}`))
		require.NoError(s.t, err)
	}))
	s.t.(*testing.T).Cleanup(s.srv.Close)
	client := s.srv.Client()
	client.Timeout = 10 * time.Millisecond
	s.plugin = pluginhttp.Plugin(
		"http",
		pluginhttp.WithBaseURL(s.srv.URL),
		pluginhttp.WithHTTPClient(client),
	)
	return s
}

func (s *Stage) a_server_echoing_query_and_header() *Stage {
	s.srv = httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(
			w,
			`{"query":"%s","header":"%s"}`,
			r.URL.Query().Get("query"),
			r.Header.Get("X-Test"),
		)
		require.NoError(s.t, err)
	}))
	s.t.(*testing.T).Cleanup(s.srv.Close)
	s.plugin = pluginhttp.Plugin("http", pluginhttp.WithBaseURL(s.srv.URL), pluginhttp.WithHTTPClient(s.srv.Client()))
	return s
}

func (s *Stage) a_jsonnet_request_is_evaluated(method string, path string) *Stage {
	snippet := fmt.Sprintf(`std.native('invoke:http')('request', [{method: '%s', path: '%s'}])`, method, path)
	s.err = jpoet.Eval(
		jpoet.WithPlugin(s.plugin),
		jpoet.SnippetInput("test.jsonnet", snippet),
		jpoet.ValueOutput(&s.out),
		jpoet.Serialize(false),
	)
	return s
}

func (s *Stage) a_jsonnet_request_with_query_and_headers_is_evaluated(method string, path string, query string, header string) *Stage {
	snippet := fmt.Sprintf(
		`std.native('invoke:http')('request', [{method: '%s', path: '%s', query: {query: '%s'}, headers: {'X-Test': '%s'}}])`,
		method,
		path,
		query,
		header,
	)
	s.err = jpoet.Eval(
		jpoet.WithPlugin(s.plugin),
		jpoet.SnippetInput("test.jsonnet", snippet),
		jpoet.ValueOutput(&s.out),
		jpoet.Serialize(false),
	)
	return s
}

func (s *Stage) the_eval_has_no_error() *Stage {
	require.NoError(s.t, s.err)
	return s
}

func (s *Stage) the_result_is_an_empty_map() *Stage {
	require.Len(s.t, s.out, 0)
	return s
}

func (s *Stage) the_result_has_kind(expected string) *Stage {
	require.Equal(s.t, expected, s.out["kind"])
	return s
}

func (s *Stage) the_result_has_code(expected float64) *Stage {
	require.Equal(s.t, expected, s.out["code"])
	return s
}

func (s *Stage) the_result_has_message(expected string) *Stage {
	require.Equal(s.t, expected, s.out["message"])
	return s
}

func (s *Stage) the_result_message_contains(expected string) *Stage {
	msg, ok := s.out["message"].(string)
	require.True(s.t, ok)
	require.True(s.t, strings.Contains(msg, expected))
	return s
}

func (s *Stage) the_result_has_field(field string, expected any) *Stage {
	require.Equal(s.t, expected, s.out[field])
	return s
}
