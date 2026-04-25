package main

import (
	"os"
	"strings"

	"github.com/marcbran/jsonnet-plugin-http/http"
)

func main() {
	name := os.Getenv("HTTP_PLUGIN_NAME")
	if strings.TrimSpace(name) == "" {
		name = "http"
	}
	var opts []http.Option
	base := os.Getenv("HTTP_BASE_URL")
	if strings.TrimSpace(base) != "" {
		opts = append(opts, http.WithBaseURL(base))
	}
	http.Plugin(name, opts...).Serve()
}
