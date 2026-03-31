package core_test

import (
	"context"
	"testing"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
)

type memoryStore struct {
	audit []core.AuditEntry
}

func (m *memoryStore) EnsureSchema(buildprofile.Profile) error { return nil }
func (m *memoryStore) UpsertPrincipal(core.Principal) error    { return nil }
func (m *memoryStore) SaveSession(core.SessionState) error     { return nil }
func (m *memoryStore) Close() error                            { return nil }
func (m *memoryStore) RecordAudit(entry core.AuditEntry) error {
	m.audit = append(m.audit, entry)
	return nil
}
func (m *memoryStore) ListRecentAudit(limit int) ([]core.AuditEntry, error) {
	return m.audit, nil
}

type stubPlugin struct{}

func (stubPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name: "capabilities",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityPluginCapabilities,
		},
	}
}

func (stubPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{
		Status:  "ok",
		Message: "stub executed",
	}, nil
}

type askStubPlugin struct{}

func (askStubPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name: "ask",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}
}

func (askStubPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{
		Status:  "ok",
		Message: "ai answer generated",
	}, nil
}

type answerOnlyPlugin struct{}

func (answerOnlyPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name: "ask",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}
}

func (answerOnlyPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{
		Status:  "ok",
		Message: "ai answer generated",
		Data: map[string]any{
			"answer": "fallback answer",
		},
	}, nil
}

type streamStubPlugin struct{}

func (streamStubPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name: "ask",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}
}

func (streamStubPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{
		Status:  "ok",
		Message: "fallback",
		Data: map[string]any{
			"answer": "fallback answer",
		},
	}, nil
}

func (streamStubPlugin) ExecuteStream(_ context.Context, _ core.CommandContext, _ core.CommandEnvelope, writer core.StreamWriter) error {
	if err := writer.WriteChunk(core.StreamChunk{Delta: "hello"}); err != nil {
		return err
	}
	if err := writer.WriteChunk(core.StreamChunk{Delta: " world"}); err != nil {
		return err
	}
	writer.SetResult(core.CommandResult{
		Message: "ai answer generated",
		Data: map[string]any{
			"provider": "openai",
		},
	})
	return nil
}

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

func TestServiceExecutesKnownCommand(t *testing.T) {
	store := &memoryStore{}
	service, err := core.NewService(buildprofile.Current(), store, []core.Plugin{stubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Execute(context.Background(), core.CommandEnvelope{
		ID:          "test-1",
		Transport:   "http_chat",
		Name:        "capabilities",
		Principal:   core.Principal{ID: "operator", Transport: "http_chat"},
		RequestedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %q", result.Status)
	}
	if len(store.audit) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(store.audit))
	}
}

func TestServiceRejectsUnknownCommand(t *testing.T) {
	store := &memoryStore{}
	service, err := core.NewService(buildprofile.Current(), store, []core.Plugin{stubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Execute(context.Background(), core.CommandEnvelope{
		ID:        "test-2",
		Transport: "http_chat",
		Name:      "missing",
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	})
	if err == nil {
		t.Fatalf("expected error for unknown command")
	}
	if result.Status != "error" {
		t.Fatalf("expected error status, got %q", result.Status)
	}
}

func TestServiceBlocksSensitiveAskPrompt(t *testing.T) {
	store := &memoryStore{}
	profile := buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}

	service, err := core.NewService(profile, store, []core.Plugin{askStubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	result, err := service.Execute(context.Background(), core.CommandEnvelope{
		ID:        "test-ask",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "please analyze this token=supersecret123 and alice@example.com",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result.Status != "error" {
		t.Fatalf("expected error status, got %q", result.Status)
	}
	if result.Message != "request blocked by sensitive data policy" {
		t.Fatalf("unexpected message %q", result.Message)
	}
	if len(store.audit) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(store.audit))
	}
	if store.audit[0].Message != "request blocked by sensitive data policy" {
		t.Fatalf("unexpected audit message %q", store.audit[0].Message)
	}
	if got := store.audit[0].Metadata["roles"]; got != "" {
		t.Fatalf("expected empty roles metadata, got %q", got)
	}
}

func TestServiceExecuteStreamUsesStreamingPlugin(t *testing.T) {
	store := &memoryStore{}
	profile := buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}

	service, err := core.NewService(profile, store, []core.Plugin{streamStubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	writer := &collectingStreamWriter{}
	err = service.ExecuteStream(context.Background(), core.CommandEnvelope{
		ID:        "stream-1",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
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
	if len(store.audit) != 1 || store.audit[0].Outcome != "ok" {
		t.Fatalf("expected successful audit entry, got %+v", store.audit)
	}
}

func TestServiceExecuteStreamFallsBackToExecute(t *testing.T) {
	store := &memoryStore{}
	profile := buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}

	service, err := core.NewService(profile, store, []core.Plugin{askStubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	writer := &collectingStreamWriter{}
	err = service.ExecuteStream(context.Background(), core.CommandEnvelope{
		ID:        "stream-2",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if len(writer.chunks) != 0 {
		t.Fatalf("expected no chunks for plain askStubPlugin fallback, got %d", len(writer.chunks))
	}
	if writer.result.Message != "ai answer generated" {
		t.Fatalf("unexpected result message %q", writer.result.Message)
	}
}

func TestServiceExecuteStreamFallsBackToSingleChunkAnswer(t *testing.T) {
	store := &memoryStore{}
	profile := buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}

	service, err := core.NewService(profile, store, []core.Plugin{answerOnlyPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	writer := &collectingStreamWriter{}
	err = service.ExecuteStream(context.Background(), core.CommandEnvelope{
		ID:        "stream-3",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	if len(writer.chunks) != 1 {
		t.Fatalf("expected one fallback chunk, got %d", len(writer.chunks))
	}
	if writer.chunks[0].Delta != "fallback answer" {
		t.Fatalf("unexpected chunk delta %q", writer.chunks[0].Delta)
	}
}

func TestServiceExecuteStreamRejectsMissingCapability(t *testing.T) {
	store := &memoryStore{}
	service, err := core.NewService(buildprofile.Profile{ID: "test-profile"}, store, []core.Plugin{streamStubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	writer := &collectingStreamWriter{}
	err = service.ExecuteStream(context.Background(), core.CommandEnvelope{
		ID:        "stream-4",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() unexpected error = %v", err)
	}
	if writer.result.Status != "error" {
		t.Fatalf("expected error status, got %q", writer.result.Status)
	}
	if len(store.audit) != 1 || store.audit[0].Outcome != "rejected" {
		t.Fatalf("expected rejected audit entry, got %+v", store.audit)
	}
}

func TestServiceExecuteStreamBlocksSensitiveAskPrompt(t *testing.T) {
	store := &memoryStore{}
	profile := buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}

	service, err := core.NewService(profile, store, []core.Plugin{streamStubPlugin{}})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	writer := &collectingStreamWriter{}
	err = service.ExecuteStream(context.Background(), core.CommandEnvelope{
		ID:        "stream-5",
		Transport: "http_chat",
		Name:      "ask",
		Args: map[string]string{
			"prompt": "token=supersecret123",
		},
		Principal: core.Principal{ID: "operator", Transport: "http_chat"},
	}, writer)
	if err != nil {
		t.Fatalf("ExecuteStream() unexpected error = %v", err)
	}
	if writer.result.Message != "request blocked by sensitive data policy" {
		t.Fatalf("unexpected result message %q", writer.result.Message)
	}
	if len(writer.chunks) != 0 {
		t.Fatalf("expected no chunks for blocked prompt, got %d", len(writer.chunks))
	}
}
