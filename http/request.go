package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
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

var actionFields = []string{"method", "path", "baseURL", "headers", "query", "context", "body"}

var mutatingMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

func parseRequestInput(input []any) (RequestInput, error) {
	if len(input) != 1 {
		return RequestInput{}, fmt.Errorf("expected input object")
	}
	raw, ok := input[0].(map[string]any)
	if !ok {
		return RequestInput{}, fmt.Errorf("input must be an object")
	}
	ri, err := parseRequestFields(raw)
	if err != nil {
		return RequestInput{}, err
	}
	if ri.Method != http.MethodGet && ri.Method != http.MethodPost {
		return RequestInput{}, fmt.Errorf("method must be GET or POST")
	}
	readonly, _ := raw["readonly"].(bool)
	if ri.Method == http.MethodPost && !readonly {
		return RequestInput{}, fmt.Errorf("POST requires readonly: true")
	}
	if ri.Body != nil && ri.Method != http.MethodPost {
		return RequestInput{}, fmt.Errorf("body is only allowed with POST")
	}
	if ri.Method == http.MethodPost && ri.Body == nil {
		return RequestInput{}, fmt.Errorf("POST requires a body")
	}
	return ri, nil
}

func parseActionInput(data map[string]any) (RequestInput, error) {
	for key := range data {
		if !slices.Contains(actionFields, key) {
			return RequestInput{}, fmt.Errorf("unknown field %q", key)
		}
	}
	ri, err := parseRequestFields(data)
	if err != nil {
		return RequestInput{}, err
	}
	if !slices.Contains(mutatingMethods, ri.Method) {
		return RequestInput{}, fmt.Errorf("method must be POST, PUT, PATCH, or DELETE")
	}
	return ri, nil
}

func parseRequestFields(raw map[string]any) (RequestInput, error) {
	method, ok := raw["method"].(string)
	if !ok || method == "" {
		return RequestInput{}, fmt.Errorf("method must be a non-empty string")
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
		Body:    raw["body"],
	}
	if raw["baseURL"] != nil {
		s, ok := raw["baseURL"].(string)
		if !ok || s == "" {
			return RequestInput{}, fmt.Errorf("baseURL must be a non-empty string")
		}
		ri.BaseURL = s
	}
	err := parseStringMap(raw, "headers", "header", ri.Headers)
	if err != nil {
		return RequestInput{}, err
	}
	err = parseStringMap(raw, "query", "query", ri.Query)
	if err != nil {
		return RequestInput{}, err
	}
	err = parseStringMap(raw, "context", "context", ri.Context)
	if err != nil {
		return RequestInput{}, err
	}
	return ri, nil
}

func parseStringMap(raw map[string]any, field string, label string, out map[string]string) error {
	if raw[field] == nil {
		return nil
	}
	m, ok := raw[field].(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be an object", field)
	}
	for k, v := range m {
		if v == nil {
			continue
		}
		s, err := stringFromAny(v)
		if err != nil {
			return fmt.Errorf("%s %q: %w", label, k, err)
		}
		out[k] = s
	}
	return nil
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

func runRequest(ctx context.Context, cfg *Config, ri RequestInput) (map[string]any, error) {
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
