package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

type RequestInput struct {
	BaseURL string
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Context map[string]string
	Body    any
}

func Request(cfg *Config) jsonnet.NativeFunction {
	return jsonnet.NativeFunction{
		Name:   "request",
		Params: ast.Identifiers{"input"},
		Func: func(input []any) (any, error) {
			ri, err := parseRequestInput(input)
			if err != nil {
				return clientFailureStatus(400, err.Error()), nil
			}
			out, err := runRequest(context.Background(), cfg, ri)
			if err != nil {
				return nil, err
			}
			return out, nil
		},
	}
}

func parseRequestInput(input []any) (RequestInput, error) {
	if len(input) != 1 {
		return RequestInput{}, fmt.Errorf("expected input object")
	}
	raw, ok := input[0].(map[string]any)
	if !ok {
		return RequestInput{}, fmt.Errorf("input must be an object")
	}
	method, ok := raw["method"].(string)
	if !ok || method == "" {
		return RequestInput{}, fmt.Errorf("method must be a non-empty string")
	}
	if method != http.MethodGet && method != http.MethodPost {
		return RequestInput{}, fmt.Errorf("method must be GET or POST")
	}
	readonly, _ := raw["readonly"].(bool)
	if method == http.MethodPost && !readonly {
		return RequestInput{}, fmt.Errorf("POST requires readonly: true")
	}
	path, ok := raw["path"].(string)
	if !ok || path == "" {
		return RequestInput{}, fmt.Errorf("path must be a non-empty string")
	}
	ri := RequestInput{
		Method:  method,
		Path:    path,
		Headers: map[string]string{},
		Query:   map[string]string{},
		Context: map[string]string{},
	}
	if raw["baseURL"] != nil {
		s, ok := raw["baseURL"].(string)
		if !ok || s == "" {
			return RequestInput{}, fmt.Errorf("baseURL must be a non-empty string")
		}
		ri.BaseURL = s
	}
	if raw["headers"] != nil {
		hm, ok := raw["headers"].(map[string]any)
		if !ok {
			return RequestInput{}, fmt.Errorf("headers must be an object")
		}
		for k, v := range hm {
			if v == nil {
				continue
			}
			s, err := stringFromAny(v)
			if err != nil {
				return RequestInput{}, fmt.Errorf("header %q: %w", k, err)
			}
			ri.Headers[k] = s
		}
	}
	if raw["query"] != nil {
		qm, ok := raw["query"].(map[string]any)
		if !ok {
			return RequestInput{}, fmt.Errorf("query must be an object")
		}
		for k, v := range qm {
			if v == nil {
				continue
			}
			s, err := stringFromAny(v)
			if err != nil {
				return RequestInput{}, fmt.Errorf("query %q: %w", k, err)
			}
			ri.Query[k] = s
		}
	}
	if raw["context"] != nil {
		cm, ok := raw["context"].(map[string]any)
		if !ok {
			return RequestInput{}, fmt.Errorf("context must be an object")
		}
		for k, v := range cm {
			if v == nil {
				continue
			}
			s, err := stringFromAny(v)
			if err != nil {
				return RequestInput{}, fmt.Errorf("context %q: %w", k, err)
			}
			ri.Context[k] = s
		}
	}
	if raw["body"] != nil {
		if method != http.MethodPost {
			return RequestInput{}, fmt.Errorf("body is only allowed with POST")
		}
		ri.Body = raw["body"]
	}
	if method == http.MethodPost && ri.Body == nil {
		return RequestInput{}, fmt.Errorf("POST requires a body")
	}
	return ri, nil
}

func stringFromAny(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(t), nil
	default:
		return "", fmt.Errorf("must be string, number, or bool")
	}
}

func runRequest(ctx context.Context, cfg *Config, ri RequestInput) (any, error) {
	baseURL := ri.BaseURL
	if baseURL == "" {
		baseURL = cfg.BaseURL
	}
	if baseURL == "" {
		return clientFailureEnvelope(400, "base url not configured"), nil
	}
	baseStr := strings.TrimRight(baseURL, "/")
	pathStr := ri.Path
	if !strings.HasPrefix(pathStr, "/") {
		pathStr = "/" + pathStr
	}
	fullRaw := baseStr + pathStr
	u, err := url.Parse(fullRaw)
	if err != nil {
		return clientFailureEnvelope(400, "invalid url"), nil
	}
	q := u.Query()
	for k, v := range ri.Query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	var bodyReader *bytes.Reader
	if ri.Body != nil {
		bodyBytes, err := json.Marshal(ri.Body)
		if err != nil {
			return clientFailureEnvelope(400, err.Error()), nil
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	var reqBody io.Reader
	if bodyReader != nil {
		reqBody = bodyReader
	}
	req, err := http.NewRequestWithContext(ctx, ri.Method, u.String(), reqBody)
	if err != nil {
		return clientFailureEnvelope(400, err.Error()), nil
	}
	if ri.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range ri.Headers {
		req.Header.Set(k, v)
	}
	resp, err := cfg.Client.Do(req)
	if err != nil {
		return clientFailureEnvelope(500, err.Error()), nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return clientFailureEnvelope(500, err.Error()), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return clientFailureEnvelope(500, err.Error()), nil
	}
	headers := headersToMap(resp.Header)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return envelope(resp.StatusCode, headers, clientFailureStatus(int32(resp.StatusCode), msg)), nil
	}
	if len(body) == 0 {
		return envelope(resp.StatusCode, headers, map[string]any{}), nil
	}
	var out any
	err = json.Unmarshal(body, &out)
	if err != nil {
		return envelope(resp.StatusCode, headers, clientFailureStatus(500, err.Error())), nil
	}
	return envelope(resp.StatusCode, headers, out), nil
}

func headersToMap(h http.Header) map[string]any {
	m := make(map[string]any, len(h))
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func envelope(status int, headers map[string]any, body any) map[string]any {
	return map[string]any{
		"status":  float64(status),
		"headers": headers,
		"body":    body,
	}
}

func clientFailureEnvelope(code int32, msg string) map[string]any {
	return envelope(int(code), map[string]any{}, clientFailureStatus(code, msg))
}

func clientFailureStatus(code int32, msg string) map[string]any {
	reason := http.StatusText(int(code))
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Status",
		"status":     "Failure",
		"code":       float64(code),
		"message":    msg,
		"reason":     reason,
	}
}
