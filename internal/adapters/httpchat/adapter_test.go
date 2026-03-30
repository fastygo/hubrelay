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

func (testStore) EnsureSchema(buildprofile.Profile) error        { return nil }
func (testStore) UpsertPrincipal(core.Principal) error           { return nil }
func (testStore) SaveSession(core.SessionState) error            { return nil }
func (testStore) RecordAudit(core.AuditEntry) error              { return nil }
func (testStore) ListRecentAudit(int) ([]core.AuditEntry, error) { return nil, nil }
func (testStore) Close() error                                   { return nil }

func TestHandleIndexRendersChatControls(t *testing.T) {
	profile := buildprofile.Profile{
		ID:          "test-profile",
		DisplayName: "Test",
		Capabilities: []buildprofile.Capability{
			buildprofile.CapabilityAdapterHTTPChat,
			buildprofile.CapabilityAIChat,
		},
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
		OpenAI: buildprofile.OpenAIConfig{
			Enabled:     true,
			Provider:    "openai",
			Model:       "gpt-4.1-mini",
			ChatHistory: true,
			HasAPIKey:   true,
		},
		ProxySession: buildprofile.ProxySessionConfig{
			Enabled: true,
			Force:   true,
		},
	}

	service, err := core.NewService(profile, testStore{}, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter := New("127.0.0.1:5500", service, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	adapter.handleIndex(recorder, request)

	body := recorder.Body.String()
	if !strings.Contains(body, "MVP Chat") {
		t.Fatalf("expected MVP Chat block in HTML")
	}
	if !strings.Contains(body, "Export chat JSON") {
		t.Fatalf("expected export button in HTML")
	}
	if !strings.Contains(body, "chatHistory: true") {
		t.Fatalf("expected chat history config in HTML")
	}
	if !strings.Contains(body, "proxyEnabled: false") {
		t.Fatalf("expected proxy feature to stay disabled without runtime manager")
	}
}

func TestHandleIndexRendersProxyControlsWhenEnabled(t *testing.T) {
	manager := proxymgr.NewManager(func(_ context.Context, _ string) (time.Duration, error) {
		return 10 * time.Millisecond, nil
	})

	profile := buildprofile.Profile{
		ID:          "test-profile",
		DisplayName: "Test",
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
			Force:   true,
		},
	}

	service, err := core.NewService(profile, testStore{}, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter := New("127.0.0.1:5500", service, manager)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	adapter.handleIndex(recorder, request)

	body := recorder.Body.String()
	if !strings.Contains(body, "Configure Proxy Session") {
		t.Fatalf("expected proxy modal entry point in HTML")
	}
	if !strings.Contains(body, "sessionStorage") {
		t.Fatalf("expected browser sessionStorage mention in HTML")
	}
	if !strings.Contains(body, "proxyEnabled: true") {
		t.Fatalf("expected proxy feature to be enabled in page config")
	}
	if !strings.Contains(body, "proxyForced: true") {
		t.Fatalf("expected forced proxy config in page")
	}
	if !strings.Contains(body, "Review Sensitive Data") {
		t.Fatalf("expected sensitive data warning dialog in HTML")
	}
	if !strings.Contains(body, "scanSensitiveInput") {
		t.Fatalf("expected sensitive data scanner hook in HTML")
	}
}

func TestProxySessionAPIFlow(t *testing.T) {
	manager := proxymgr.NewManager(func(_ context.Context, address string) (time.Duration, error) {
		if address == "10.0.0.2:1080" {
			return 15 * time.Millisecond, nil
		}
		return 40 * time.Millisecond, nil
	})

	profile := buildprofile.Profile{
		ID:          "test-profile",
		DisplayName: "Test",
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
	}

	service, err := core.NewService(profile, testStore{}, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter := New("127.0.0.1:5500", service, manager)

	createBody := bytes.NewBufferString(`{"principal_id":"operator-local","proxies":"10.0.0.1:1080\n10.0.0.2:1080"}`)
	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest("POST", "/api/proxy/session", createBody)
	adapter.handleProxyCollection(createRecorder, createRequest)
	if createRecorder.Code != 201 {
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
	checkRequest := httptest.NewRequest("POST", "/api/proxy/session/"+created.Session.ID+"/check", nil)
	adapter.handleProxySession(checkRecorder, checkRequest)
	if checkRecorder.Code != 200 {
		t.Fatalf("unexpected check status %d: %s", checkRecorder.Code, checkRecorder.Body.String())
	}

	selectBody := bytes.NewBufferString(`{"proxy":"fastest"}`)
	selectRecorder := httptest.NewRecorder()
	selectRequest := httptest.NewRequest("POST", "/api/proxy/session/"+created.Session.ID+"/select", selectBody)
	adapter.handleProxySession(selectRecorder, selectRequest)
	if selectRecorder.Code != 200 {
		t.Fatalf("unexpected select status %d: %s", selectRecorder.Code, selectRecorder.Body.String())
	}
	if !strings.Contains(selectRecorder.Body.String(), `"selected_proxy":"10.0.0.2:1080"`) {
		t.Fatalf("expected fastest proxy to become active, got %s", selectRecorder.Body.String())
	}

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest("GET", "/api/proxy/session/"+created.Session.ID, nil)
	adapter.handleProxySession(getRecorder, getRequest)
	if getRecorder.Code != 200 {
		t.Fatalf("unexpected get status %d: %s", getRecorder.Code, getRecorder.Body.String())
	}
	if !strings.Contains(getRecorder.Body.String(), `"address":"10.0.0.2:1080"`) {
		t.Fatalf("expected active lease in get response, got %s", getRecorder.Body.String())
	}
}

func TestBuildMuxAvoidsRouteConflicts(t *testing.T) {
	profile := buildprofile.Profile{
		ID: "test-profile",
		HTTPChat: buildprofile.HTTPChatConfig{
			Enabled:     true,
			BindAddress: "127.0.0.1:5500",
		},
	}

	service, err := core.NewService(profile, testStore{}, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter := New("127.0.0.1:5500", service, nil)
	mux := adapter.buildMux()

	indexRecorder := httptest.NewRecorder()
	indexRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	mux.ServeHTTP(indexRecorder, indexRequest)
	if indexRecorder.Code != http.StatusOK {
		t.Fatalf("expected root path to be served, got %d", indexRecorder.Code)
	}

	proxyRecorder := httptest.NewRecorder()
	proxyRequest := httptest.NewRequest(http.MethodGet, "/api/proxy/session/test", nil)
	mux.ServeHTTP(proxyRecorder, proxyRequest)
	if proxyRecorder.Code == http.StatusOK && strings.Contains(proxyRecorder.Body.String(), "<!DOCTYPE html>") {
		t.Fatalf("expected API route not to be shadowed by root GET handler")
	}
}
