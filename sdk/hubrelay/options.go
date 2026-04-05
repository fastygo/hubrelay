package hubrelay

import (
	"net/http"
	"time"
)

type clientConfig struct {
	timeout    time.Duration
	httpClient *http.Client
	principal  Principal
	baseURL    string
	socketPath string
	grpcTarget string
}

type Option func(*clientConfig)

func WithTimeout(d time.Duration) Option {
	return func(cfg *clientConfig) {
		if d > 0 {
			cfg.timeout = d
		}
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		if client != nil {
			cfg.httpClient = client
		}
	}
}

func WithPrincipal(principal Principal) Option {
	return func(cfg *clientConfig) {
		cfg.principal = principal
	}
}

func WithBaseURL(baseURL string) Option {
	return func(cfg *clientConfig) {
		cfg.baseURL = baseURL
	}
}

func WithSocketPath(socketPath string) Option {
	return func(cfg *clientConfig) {
		cfg.socketPath = socketPath
	}
}

func WithGRPCTarget(target string) Option {
	return func(cfg *clientConfig) {
		cfg.grpcTarget = target
	}
}

func applyOptions(opts ...Option) clientConfig {
	cfg := clientConfig{
		timeout: 30 * time.Second,
		principal: Principal{
			ID:      "sdk-client",
			Display: "sdk-client",
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}
