package ask_test

import (
	"context"
	"testing"

	"sshbot/internal/ai"
	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	askplugin "sshbot/internal/plugins/ask"
)

type fakeProvider struct {
	lastRequest ai.AskRequest
}

func (p *fakeProvider) Name() string { return "openai" }

func (p *fakeProvider) Ask(_ context.Context, request ai.AskRequest) (ai.AskResponse, error) {
	p.lastRequest = request
	return ai.AskResponse{
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Answer:   "hello from ai",
	}, nil
}

type fakeStreamingProvider struct {
	fakeProvider
	chunks []ai.AskStreamChunk
}

func (p *fakeStreamingProvider) AskStream(_ context.Context, request ai.AskRequest, callback ai.StreamCallback) (ai.AskResponse, error) {
	p.lastRequest = request
	for _, chunk := range p.chunks {
		if err := callback(chunk); err != nil {
			return ai.AskResponse{}, err
		}
	}
	return ai.AskResponse{
		Provider:   "openai",
		Model:      "gpt-4.1-mini",
		Answer:     "hello from ai",
		ResponseID: "resp-stream",
	}, nil
}

type noopStore struct{}

func (noopStore) EnsureSchema(buildprofile.Profile) error        { return nil }
func (noopStore) UpsertPrincipal(core.Principal) error           { return nil }
func (noopStore) SaveSession(core.SessionState) error            { return nil }
func (noopStore) RecordAudit(core.AuditEntry) error              { return nil }
func (noopStore) ListRecentAudit(int) ([]core.AuditEntry, error) { return nil, nil }
func (noopStore) Close() error                                   { return nil }

type collectingStreamWriter struct {
	chunks []core.StreamChunk
	result core.CommandResult
}

func (w *collectingStreamWriter) WriteChunk(chunk core.StreamChunk) error {
	w.chunks = append(w.chunks, chunk)
	return nil
}

func (w *collectingStreamWriter) SetResult(result core.CommandResult) {
	w.result = result
}

func (w *collectingStreamWriter) Flush() error {
	return nil
}

func TestAskPluginReturnsAnswerInData(t *testing.T) {
	provider := &fakeProvider{}
	plugin := askplugin.New(provider)

	result, err := plugin.Execute(context.Background(), core.CommandContext{
		Profile: buildprofile.Current(),
		Store:   noopStore{},
	}, core.CommandEnvelope{
		Name:    "ask",
		RawText: "ask hello",
		Args: map[string]string{
			"prompt":           "hello",
			"proxy_session_id": "proxy-123",
		},
		Transport: "http_chat",
		Principal: core.Principal{ID: "operator-local"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Message != "ai answer generated" {
		t.Fatalf("unexpected audit-safe message %q", result.Message)
	}
	if got := result.Data["answer"]; got != "hello from ai" {
		t.Fatalf("unexpected answer %v", got)
	}
	if provider.lastRequest.ProxySessionID != "proxy-123" {
		t.Fatalf("expected proxy session id to flow into provider, got %q", provider.lastRequest.ProxySessionID)
	}
}

func TestAskPluginExecuteStreamDeliversChunks(t *testing.T) {
	provider := &fakeStreamingProvider{
		chunks: []ai.AskStreamChunk{
			{Delta: "hello"},
			{Delta: " world", FinishReason: "stop"},
		},
	}
	plugin := askplugin.New(provider)
	writer := &collectingStreamWriter{}

	err := plugin.ExecuteStream(context.Background(), core.CommandContext{
		Profile: buildprofile.Current(),
		Store:   noopStore{},
	}, core.CommandEnvelope{
		Name:      "ask",
		Transport: "http_chat",
		Principal: core.Principal{ID: "operator-local"},
		Args: map[string]string{
			"prompt": "hello",
		},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if len(writer.chunks) != 2 {
		t.Fatalf("expected two chunks, got %d", len(writer.chunks))
	}
	if writer.result.Message != "ai answer generated" {
		t.Fatalf("unexpected result message %q", writer.result.Message)
	}
	if got := writer.result.Data["response_id"]; got != "resp-stream" {
		t.Fatalf("unexpected response id %v", got)
	}
	if provider.lastRequest.SessionID != "http_chat:operator-local" {
		t.Fatalf("unexpected session id %q", provider.lastRequest.SessionID)
	}
}

func TestAskPluginExecuteStreamFallsBackToAsk(t *testing.T) {
	provider := &fakeProvider{}
	plugin := askplugin.New(provider)
	writer := &collectingStreamWriter{}

	err := plugin.ExecuteStream(context.Background(), core.CommandContext{
		Profile: buildprofile.Current(),
		Store:   noopStore{},
	}, core.CommandEnvelope{
		Name:      "ask",
		Transport: "http_chat",
		Principal: core.Principal{ID: "operator-local"},
		Args: map[string]string{
			"prompt":           "hello",
			"proxy_session_id": "proxy-123",
		},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if len(writer.chunks) != 1 {
		t.Fatalf("expected one fallback chunk, got %d", len(writer.chunks))
	}
	if writer.chunks[0].Delta != "hello from ai" {
		t.Fatalf("unexpected chunk delta %q", writer.chunks[0].Delta)
	}
	if got := writer.result.Data["provider"]; got != "openai" {
		t.Fatalf("unexpected provider %v", got)
	}
}
