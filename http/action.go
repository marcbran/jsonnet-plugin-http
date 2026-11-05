package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/marcbran/jpoet/pkg/jpoet"
)

func RequestAction(cfg *Config) jpoet.ActionFunc {
	return func(ctx context.Context, data map[string]any) (string, error) {
		ri, err := parseActionInput(data)
		if err != nil {
			return "", err
		}
		out, err := runRequest(ctx, cfg, ri)
		var failure *jpoet.Failure
		if errors.As(err, &failure) {
			return "", fmt.Errorf("%s %s: %d %s", ri.Method, ri.Path, failure.Code, failure.Message)
		}
		if err != nil {
			return "", err
		}
		raw, err := json.Marshal(out)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
}
