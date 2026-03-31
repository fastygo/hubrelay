package hubrelay

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
)

type unixTransport struct {
	socketPath string
	client     *http.Client
}

func newUnixTransport(socketPath string, cfg clientConfig) *unixTransport {
	client := cfg.httpClient
	if client == nil {
		client = &http.Client{
			Timeout: cfg.timeout,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var dialer net.Dialer
					return dialer.DialContext(ctx, "unix", socketPath)
				},
			},
		}
	}
	return &unixTransport{
		socketPath: strings.TrimSpace(socketPath),
		client:     client,
	}
}

func (t *unixTransport) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, "http://unix"+path, body)
	if err != nil {
		return nil, err
	}
	return t.client.Do(request)
}

func (t *unixTransport) Close() error {
	if transport, ok := t.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}
