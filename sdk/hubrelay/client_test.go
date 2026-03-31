package hubrelay

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testCommandPayload struct {
	PrincipalID string            `json:"principal_id"`
	Roles       []string          `json:"roles"`
	Command     string            `json:"command"`
	Args        map[string]string `json:"args"`
}

func TestHTTPClientExecuteHealthAndEgressStatus(t *testing.T) {
	server := httptest.NewServer(newTestHandler(t))
	defer server.Close()

	client := NewHTTPClient(server.URL, WithPrincipal(Principal{
		ID:    "operator-local",
		Roles: []string{"operator"},
	}))
	defer client.Close()

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "ok" || health.Profile != "test-profile" {
		t.Fatalf("unexpected health response %+v", health)
	}

	result, err := client.Execute(context.Background(), CommandRequest{
		Command: "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != "ok" || result.Message != "command executed" {
		t.Fatalf("unexpected execute result %+v", result)
	}

	egress, err := client.EgressStatus(context.Background(), Principal{})
	if err != nil {
		t.Fatalf("EgressStatus() error = %v", err)
	}
	if len(egress.Gateways) != 1 || egress.Gateways[0].Name != "wg-b1" || !egress.Gateways[0].Active {
		t.Fatalf("unexpected egress response %+v", egress)
	}
}

func TestHTTPClientExecuteStreamParsesSSE(t *testing.T) {
	server := httptest.NewServer(newTestHandler(t))
	defer server.Close()

	client := NewHTTPClient(server.URL, WithPrincipal(Principal{
		ID:    "operator-local",
		Roles: []string{"operator"},
	}))
	defer client.Close()

	stream, err := client.ExecuteStream(context.Background(), CommandRequest{
		Command: "ask",
		Args: map[string]string{
			"prompt": "hello",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteStream() error = %v", err)
	}
	defer stream.Close()

	var deltas []string
	for stream.Next() {
		deltas = append(deltas, stream.Chunk().Delta)
	}

	result, err := stream.Result()
	if err != nil {
		t.Fatalf("Result() error = %v", err)
	}
	if strings.Join(deltas, "") != "hello world" {
		t.Fatalf("unexpected deltas %v", deltas)
	}
	if result.Status != "ok" || result.Data["provider"] != "openai" {
		t.Fatalf("unexpected stream result %+v", result)
	}
}

func TestCapabilitiesDecodesStructuredResponse(t *testing.T) {
	server := httptest.NewServer(newTestHandler(t))
	defer server.Close()

	client := NewHTTPClient(server.URL, WithPrincipal(Principal{
		ID:    "operator-local",
		Roles: []string{"operator"},
	}))
	defer client.Close()

	response, err := client.Capabilities(context.Background(), Principal{})
	if err != nil {
		t.Fatalf("Capabilities() error = %v", err)
	}
	if response.ProfileID != "test-profile" || len(response.Capabilities) != 2 {
		t.Fatalf("unexpected capabilities response %+v", response)
	}
}

func TestUnixClientUsesSocketTransport(t *testing.T) {
	socketDir := t.TempDir()
	socketPath := filepath.Join(socketDir, "hubrelay.sock")
	server := &http.Server{Handler: newTestHandler(t)}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
	}()

	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := NewUnixClient(socketPath, WithPrincipal(Principal{
		ID:    "operator-local",
		Roles: []string{"operator"},
	}))
	defer client.Close()

	discovery, err := client.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if discovery.Service != "hubrelay" || discovery.Profile != "test-profile" {
		t.Fatalf("unexpected discovery response %+v", discovery)
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(DiscoveryResponse{
			Service: "hubrelay",
			Profile: "test-profile",
			Status:  "ok",
		})
	})
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(HealthResponse{
			Status:  "ok",
			Adapter: "test",
			Profile: "test-profile",
		})
	})
	mux.HandleFunc("POST /api/command", func(writer http.ResponseWriter, request *http.Request) {
		var payload testCommandPayload
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload.PrincipalID != "operator-local" {
			t.Fatalf("unexpected principal id %q", payload.PrincipalID)
		}

		var result CommandResult
		switch payload.Command {
		case "capabilities":
			result = CommandResult{
				Status:  "ok",
				Message: "capabilities returned",
				Data: map[string]any{
					"profile_id":   "test-profile",
					"display_name": "Test Profile",
					"capabilities": []string{"adapter.http_chat", "plugin.egress.status"},
				},
			}
		case "egress-status":
			result = CommandResult{
				Status:  "ok",
				Message: "egress returned",
				Data: map[string]any{
					"gateways": []map[string]any{
						{
							"name":         "wg-b1",
							"interface":    "wg-b1",
							"healthy":      true,
							"health_level": "healthy",
							"active":       true,
						},
					},
				},
			}
		default:
			result = CommandResult{
				Status:  "ok",
				Message: "command executed",
				Data: map[string]any{
					"command": payload.Command,
				},
			}
		}
		_ = json.NewEncoder(writer).Encode(result)
	})
	mux.HandleFunc("POST /api/command/stream", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(writer, "event: chunk\ndata: {\"delta\":\"hello\"}\n\n")
		_, _ = fmt.Fprint(writer, "event: chunk\ndata: {\"delta\":\" world\"}\n\n")
		_, _ = fmt.Fprint(writer, "event: done\ndata: {\"status\":\"ok\",\"message\":\"done\",\"data\":{\"provider\":\"openai\"}}\n\n")
	})
	return mux
}
