package http

import (
	"context"
	"encoding/json"
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
		if err != nil {
			return "", err
		}
		status, _ := out["status"].(float64)
		if status < 200 || status > 299 {
			return "", fmt.Errorf("%s %s: %d %s", ri.Method, ri.Path, int(status), failureMessage(out))
		}
		raw, err := json.Marshal(out)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}
}

func failureMessage(out map[string]any) string {
	body, _ := out["body"].(map[string]any)
	message, _ := body["message"].(string)
	return message
}
