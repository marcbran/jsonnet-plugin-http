package http

import (
	"context"
	"fmt"
	"maps"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

type inputModifier func(ri RequestInput, input map[string]any) (map[string]any, error)

func withInputModifier(modify inputModifier) Option {
	return func(c *Config) {
		c.pluginOpts = append(c.pluginOpts,
			jpoet.WithHook(modifierInvokeHook(modify)),
			jpoet.WithActionHook(modifierActionHook(modify)),
		)
	}
}

func modifierInvokeHook(modify inputModifier) jpoet.InvokeHook {
	return func(next jpoet.Invoker, funcName string, args []any) (any, error) {
		if funcName != "request" {
			return next.Invoke(funcName, args)
		}
		ri, err := parseRequestInput(args)
		if err != nil {
			return next.Invoke(funcName, args)
		}
		input, err := modify(ri, args[0].(map[string]any))
		if err != nil {
			return nil, err
		}
		mergedArgs := make([]any, len(args))
		copy(mergedArgs, args)
		mergedArgs[0] = input
		return next.Invoke(funcName, mergedArgs)
	}
}

func modifierActionHook(modify inputModifier) jpoet.ActionHook {
	return func(ctx context.Context, next jpoet.Executor, action string, data map[string]any) (string, error) {
		if action != "request" {
			return next.Exec(ctx, action, data)
		}
		ri, err := parseActionInput(data)
		if err != nil {
			return next.Exec(ctx, action, data)
		}
		input, err := modify(ri, data)
		if err != nil {
			return "", err
		}
		return next.Exec(ctx, action, input)
	}
}

func WithBaseURLByRequest(
	baseURLFor func(RequestInput) (string, error),
) Option {
	if baseURLFor == nil {
		return func(*Config) {}
	}
	return withInputModifier(func(ri RequestInput, input map[string]any) (map[string]any, error) {
		baseURL, err := baseURLFor(ri)
		if err != nil {
			return nil, fmt.Errorf("base url by request: %w", err)
		}
		if baseURL == "" {
			return input, nil
		}
		return injectBaseURL(input, baseURL), nil
	})
}

func injectBaseURL(input map[string]any, baseURL string) map[string]any {
	merged := maps.Clone(input)
	merged["baseURL"] = baseURL
	return merged
}

func WithBodyByRequest(
	bodyFor func(RequestInput) (any, error),
) Option {
	if bodyFor == nil {
		return func(*Config) {}
	}
	return withInputModifier(func(ri RequestInput, input map[string]any) (map[string]any, error) {
		body, err := bodyFor(ri)
		if err != nil {
			return nil, fmt.Errorf("body by request: %w", err)
		}
		if body == nil {
			return input, nil
		}
		return injectBody(input, body), nil
	})
}

func injectBody(input map[string]any, body any) map[string]any {
	merged := maps.Clone(input)
	merged["body"] = body
	return merged
}

func WithHeadersByRequest(
	headersFor func(RequestInput) (map[string]string, error),
	override bool,
) Option {
	if headersFor == nil {
		return func(*Config) {}
	}
	return withInputModifier(func(ri RequestInput, input map[string]any) (map[string]any, error) {
		headers, err := headersFor(ri)
		if err != nil {
			return nil, fmt.Errorf("headers by request: %w", err)
		}
		if len(headers) == 0 {
			return input, nil
		}
		return injectHeaders(input, headers, override), nil
	})
}

func injectHeaders(input map[string]any, headers map[string]string, override bool) map[string]any {
	merged := maps.Clone(input)
	mergedHeaders := map[string]any{}
	if !override {
		for k, v := range headers {
			mergedHeaders[k] = v
		}
	}
	if existing, ok := input["headers"].(map[string]any); ok {
		maps.Copy(mergedHeaders, existing)
	}
	if override {
		for k, v := range headers {
			mergedHeaders[k] = v
		}
	}
	merged["headers"] = mergedHeaders
	return merged
}
