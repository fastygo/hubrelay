package hubrelay

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	transport        Transport
	defaultPrincipal Principal
}

type Transport interface {
	Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error)
	Close() error
}

func NewHTTPClient(baseURL string, opts ...Option) *Client {
	cfg := applyOptions(opts...)
	return &Client{
		transport:        newHTTPTransport(strings.TrimSpace(baseURL), cfg),
		defaultPrincipal: cfg.principal,
	}
}

func NewUnixClient(socketPath string, opts ...Option) *Client {
	cfg := applyOptions(opts...)
	return &Client{
		transport:        newUnixTransport(strings.TrimSpace(socketPath), cfg),
		defaultPrincipal: cfg.principal,
	}
}

func (c *Client) Close() error {
	if c == nil || c.transport == nil {
		return nil
	}
	return c.transport.Close()
}

func (c *Client) principalOrDefault(principal Principal) Principal {
	if strings.TrimSpace(principal.ID) == "" {
		principal = c.defaultPrincipal
	}
	if strings.TrimSpace(principal.Display) == "" {
		principal.Display = principal.ID
	}
	return principal
}
