package hubrelay

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	transport CommandTransport
}

type ResultStream interface {
	Next() bool
	Chunk() StreamChunk
	Result() (CommandResult, error)
	Close() error
}

type CommandTransport interface {
	Discover(ctx context.Context) (DiscoveryResponse, error)
	Health(ctx context.Context) (HealthResponse, error)
	Capabilities(ctx context.Context, principal Principal) (CapabilitiesResponse, error)
	Execute(ctx context.Context, req CommandRequest) (CommandResult, error)
	ExecuteStream(ctx context.Context, req CommandRequest) (ResultStream, error)
	EgressStatus(ctx context.Context, principal Principal) (EgressStatusResponse, error)
	Close() error
}

type rawTransport interface {
	Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error)
	Close() error
}

func NewHTTPClient(baseURL string, opts ...Option) *Client {
	cfg := applyOptions(opts...)
	return &Client{
		transport: newCommandTransport(newHTTPTransport(strings.TrimSpace(baseURL), cfg), cfg.principal),
	}
}

func NewUnixClient(socketPath string, opts ...Option) *Client {
	cfg := applyOptions(opts...)
	return &Client{
		transport: newCommandTransport(newUnixTransport(strings.TrimSpace(socketPath), cfg), cfg.principal),
	}
}

func NewClient(kind string, opts ...Option) (*Client, error) {
	cfg := applyOptions(opts...)
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "http":
		return NewHTTPClient(cfg.baseURL, opts...), nil
	case "unix":
		return NewUnixClient(cfg.socketPath, opts...), nil
	case "grpc":
		return nil, ErrTransportNotImplemented
	default:
		return nil, ErrUnsupportedTransport
	}
}

func (c *Client) Close() error {
	if c == nil || c.transport == nil {
		return nil
	}
	return c.transport.Close()
}

func (c *Client) Discover(ctx context.Context) (DiscoveryResponse, error) {
	return c.transport.Discover(ctx)
}

func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	return c.transport.Health(ctx)
}

func (c *Client) Capabilities(ctx context.Context, principal Principal) (CapabilitiesResponse, error) {
	return c.transport.Capabilities(ctx, principal)
}

func (c *Client) Execute(ctx context.Context, req CommandRequest) (CommandResult, error) {
	return c.transport.Execute(ctx, req)
}

func (c *Client) ExecuteStream(ctx context.Context, req CommandRequest) (*Stream, error) {
	stream, err := c.transport.ExecuteStream(ctx, req)
	if err != nil {
		return nil, err
	}
	typed, ok := stream.(*Stream)
	if !ok {
		return nil, ErrUnexpectedStreamType
	}
	return typed, nil
}

func (c *Client) EgressStatus(ctx context.Context, principal Principal) (EgressStatusResponse, error) {
	return c.transport.EgressStatus(ctx, principal)
}
