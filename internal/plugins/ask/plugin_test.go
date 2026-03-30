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

type noopStore struct{}

func (noopStore) EnsureSchema(buildprofile.Profile) error        { return nil }
func (noopStore) UpsertPrincipal(core.Principal) error           { return nil }
func (noopStore) SaveSession(core.SessionState) error            { return nil }
func (noopStore) RecordAudit(core.AuditEntry) error              { return nil }
func (noopStore) ListRecentAudit(int) ([]core.AuditEntry, error) { return nil, nil }
func (noopStore) Close() error                                   { return nil }

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
