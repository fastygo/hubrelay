package hubrelay

import (
	"context"
	"io"
	"net/http"
	"strings"
)

type httpTransport struct {
	baseURL string
	client  *http.Client
}

func newHTTPTransport(baseURL string, cfg clientConfig) *httpTransport {
	client := cfg.httpClient
	if client == nil {
		client = &http.Client{Timeout: cfg.timeout}
	}
	return &httpTransport{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

func (t *httpTransport) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	return t.client.Do(request)
}

func (t *httpTransport) Close() error {
	if transport, ok := t.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}
