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
