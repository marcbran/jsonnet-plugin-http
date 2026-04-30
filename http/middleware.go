package http

import (
	"fmt"
	"maps"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

func HeadersByRequest(
	headersFor func(RequestInput) (map[string]string, error),
	override bool,
) jpoet.Middleware {
	return jpoet.HookMiddleware(func(next jpoet.Invoker, funcName string, args []any) (any, error) {
		if funcName != "request" || headersFor == nil {
			return next.Invoke(funcName, args)
		}
		ri, err := parseRequestInput(args)
		if err != nil {
			return next.Invoke(funcName, args)
		}
		headers, err := headersFor(ri)
		if err != nil {
			return nil, fmt.Errorf("headers by request: %w", err)
		}
		if len(headers) == 0 {
			return next.Invoke(funcName, args)
		}
		return next.Invoke(funcName, injectHeaders(args, headers, override))
	})
}

func injectHeaders(args []any, headers map[string]string, override bool) []any {
	if len(args) == 0 {
		return args
	}
	input, ok := args[0].(map[string]any)
	if !ok {
		return args
	}
	mergedInput := make(map[string]any, len(input)+1)
	maps.Copy(mergedInput, input)
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
	mergedInput["headers"] = mergedHeaders
	mergedArgs := make([]any, len(args))
	copy(mergedArgs, args)
	mergedArgs[0] = mergedInput
	return mergedArgs
}
