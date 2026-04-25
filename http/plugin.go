package http

import (
	"net/http"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/marcbran/jpoet/pkg/jpoet"
)

const defaultRequestTimeout = 30 * time.Second

type Config struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string
}

type Option func(*Config)

func WithBaseURL(u string) Option {
	return func(c *Config) {
		c.BaseURL = u
	}
}

func WithHeaders(h map[string]string) Option {
	return func(c *Config) {
		if h == nil {
			c.Headers = map[string]string{}
			return
		}
		c.Headers = make(map[string]string, len(h))
		for k, v := range h {
			c.Headers[k] = v
		}
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		if client == nil {
			return
		}
		c.Client = client
	}
}

func Plugin(name string, opts ...Option) *jpoet.Plugin {
	cfg := &Config{
		Client: &http.Client{
			Timeout: defaultRequestTimeout,
		},
		Headers: map[string]string{},
	}
	for _, o := range opts {
		o(cfg)
	}
	return jpoet.NewPlugin(name, []jsonnet.NativeFunction{
		Request(cfg),
	})
}
