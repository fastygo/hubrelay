package httpchat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
)

type testStore struct{}

func (testStore) EnsureSchema(core.RuntimeProfile) error         { return nil }
func (testStore) UpsertPrincipal(core.Principal) error           { return nil }
func (testStore) SaveSession(core.SessionState) error            { return nil }
func (testStore) RecordAudit(core.AuditEntry) error              { return nil }
func (testStore) ListRecentAudit(int) ([]core.AuditEntry, error) { return nil, nil }
func (testStore) Close() error                                   { return nil }

type streamPlugin struct{}

func (streamPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name: "ask",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}
}

func (streamPlugin) Execute(context.Context, core.CommandContext, core.CommandEnvelope) (core.CommandResult, error) {
	return core.CommandResult{
		Status:  "ok",
		Message: "ai answer generated",
		Data: map[string]any{
			"answer": "hello world",
		},
	}, nil
}

func (streamPlugin) ExecuteStream(_ context.Context, _ core.CommandContext, _ core.CommandEnvelope, writer core.StreamWriter) error {
	if err := writer.WriteChunk(core.StreamChunk{Delta: "hello"}); err != nil {
		return err
	}
	if err := writer.WriteChunk(core.StreamChunk{Delta: " world"}); err != nil {
		return err
	}
	writer.SetResult(core.CommandResult{
		Status:  "ok",
		Message: "ai answer generated",
		Data: map[string]any{
			"provider": "openai",
		},
	})
	return nil
}

func newTestAdapter(t *testing.T, profile buildprofile.Profile, plugins []core.Plugin, proxy *proxymgr.Manager) *Adapter {
	t.Helper()
	service, err := core.NewService(profile, testStore{}, plugins)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return New("127.0.0.1:5500", service, proxy)
}

func TestRootRouteReturnsDiscoveryJSON(t *testing.T) {
	adapter := newTestAdapter(t, buildprofile.Profile{
		ID: "test-profile",
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
	}, nil, nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	adapter.buildMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected JSON discovery response, got %q", contentType)
	}
	if strings.Contains(recorder.Body.String(), "<html") {
		t.Fatalf("expected no HTML in discovery response, got %s", recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload["service"] != "hubrelay" || payload["profile"] != "test-profile" || payload["status"] != "ok" {
		t.Fatalf("unexpected discovery payload %+v", payload)
	}
}

func TestProxySessionAPIFlow(t *testing.T) {
	manager := proxymgr.NewManager(func(_ context.Context, address string) (time.Duration, error) {
		if address == "10.0.0.2:1080" {
			return 15 * time.Millisecond, nil
		}
		return 40 * time.Millisecond, nil
	})

	adapter := newTestAdapter(t, buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAdapterHTTPChat,
			buildprofile.CapabilityAIChat,
			buildprofile.CapabilityProxySession,
		},
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
		OpenAI: buildprofile.OpenAIConfig{
			Enabled:   true,
			Provider:  "openai",
			Model:     "gpt-4.1-mini",
			HasAPIKey: true,
		},
		ProxySession: buildprofile.ProxySessionConfig{
			Enabled: true,
		},
	}, nil, manager)

	mux := adapter.buildMux()

	createBody := bytes.NewBufferString(`{"principal_id":"operator-local","proxies":"10.0.0.1:1080\n10.0.0.2:1080"}`)
	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/proxy/session", createBody)
	mux.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("unexpected create status %d: %s", createRecorder.Code, createRecorder.Body.String())
	}

	var created struct {
		Session struct {
			ID string `json:"id"`
		} `json:"session"`
	}
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("Unmarshal(create) error = %v", err)
	}
	if created.Session.ID == "" {
		t.Fatalf("expected proxy session id")
	}

	checkRecorder := httptest.NewRecorder()
	checkRequest := httptest.NewRequest(http.MethodPost, "/api/proxy/session/"+created.Session.ID+"/check", nil)
	mux.ServeHTTP(checkRecorder, checkRequest)
	if checkRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected check status %d: %s", checkRecorder.Code, checkRecorder.Body.String())
	}

	selectBody := bytes.NewBufferString(`{"proxy":"fastest"}`)
	selectRecorder := httptest.NewRecorder()
	selectRequest := httptest.NewRequest(http.MethodPost, "/api/proxy/session/"+created.Session.ID+"/select", selectBody)
	mux.ServeHTTP(selectRecorder, selectRequest)
	if selectRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected select status %d: %s", selectRecorder.Code, selectRecorder.Body.String())
	}
	if !strings.Contains(selectRecorder.Body.String(), `"selected_proxy":"10.0.0.2:1080"`) {
		t.Fatalf("expected fastest proxy to become active, got %s", selectRecorder.Body.String())
	}
}

func TestBuildMuxAvoidsRouteConflicts(t *testing.T) {
	adapter := newTestAdapter(t, buildprofile.Profile{
		ID: "test-profile",
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
	}, nil, nil)

	mux := adapter.buildMux()

	indexRecorder := httptest.NewRecorder()
	indexRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(indexRecorder, indexRequest)
	if indexRecorder.Code != http.StatusOK {
		t.Fatalf("expected root path to be served, got %d", indexRecorder.Code)
	}
	if strings.Contains(indexRecorder.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("expected root handler to return JSON instead of HTML")
	}

	proxyRecorder := httptest.NewRecorder()
	proxyRequest := httptest.NewRequest(http.MethodGet, "/api/proxy/session/test", nil)
	mux.ServeHTTP(proxyRecorder, proxyRequest)
	if proxyRecorder.Code == http.StatusOK && strings.Contains(proxyRecorder.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("expected API route not to be shadowed by root handler")
	}
}

func TestHandleCommandStreamEmitsSSEChunksAndDone(t *testing.T) {
	adapter := newTestAdapter(t, buildprofile.Profile{
		ID: "test-profile",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAdapterHTTPChat,
			buildprofile.CapabilityAIChat,
		},
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
	}, []core.Plugin{streamPlugin{}}, nil)

	body := bytes.NewBufferString(`{"principal_id":"operator-local","roles":["operator"],"command":"ask","args":{"prompt":"hello"}}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/command/stream", body)

	adapter.buildMux().ServeHTTP(recorder, request)

	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected SSE content type, got %q", contentType)
	}
	responseBody := recorder.Body.String()
	if !strings.Contains(responseBody, "event: chunk") {
		t.Fatalf("expected chunk event, got %s", responseBody)
	}
	if !strings.Contains(responseBody, `"delta":"hello"`) {
		t.Fatalf("expected first delta in response, got %s", responseBody)
	}
	if !strings.Contains(responseBody, "event: done") {
		t.Fatalf("expected done event, got %s", responseBody)
	}
	if !strings.Contains(responseBody, `"provider":"openai"`) {
		t.Fatalf("expected final provider metadata, got %s", responseBody)
	}
}

func TestHandleCommandStreamEmitsErrorEvent(t *testing.T) {
	adapter := newTestAdapter(t, buildprofile.Profile{ID: "test-profile"}, []core.Plugin{streamPlugin{}}, nil)

	body := bytes.NewBufferString(`{"principal_id":"operator-local","roles":["operator"],"command":"ask","args":{"prompt":"hello"}}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/command/stream", body)

	adapter.buildMux().ServeHTTP(recorder, request)

	responseBody := recorder.Body.String()
	if !strings.Contains(responseBody, "event: error") {
		t.Fatalf("expected error event, got %s", responseBody)
	}
	if !strings.Contains(responseBody, `"status":"error"`) {
		t.Fatalf("expected error payload, got %s", responseBody)
	}
}
