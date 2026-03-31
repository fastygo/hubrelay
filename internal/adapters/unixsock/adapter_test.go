package unixsock

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
)

type testStore struct{}

func (testStore) EnsureSchema(buildprofile.Profile) error        { return nil }
func (testStore) UpsertPrincipal(core.Principal) error           { return nil }
func (testStore) SaveSession(core.SessionState) error            { return nil }
func (testStore) RecordAudit(core.AuditEntry) error              { return nil }
func (testStore) ListRecentAudit(int) ([]core.AuditEntry, error) { return nil, nil }
func (testStore) Close() error                                   { return nil }

func TestAdapterServesHTTPOverUnixSocket(t *testing.T) {
	socketDir := t.TempDir()
	socketPath := filepath.Join(socketDir, "hubrelay.sock")

	service, err := core.NewService(buildprofile.Profile{
		ID: "test-profile",
	}, testStore{}, nil)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	adapter := New(socketPath, service, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- adapter.Start(ctx)
	}()

	waitForSocket(t, socketPath)

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
	}

	response, err := client.Get("http://unix/")
	if err != nil {
		t.Fatalf("GET / over unix socket failed: %v", err)
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("json.Decode() error = %v", err)
	}
	if payload["service"] != "hubrelay" || payload["profile"] != "test-profile" {
		t.Fatalf("unexpected discovery payload %+v", payload)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("adapter returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("adapter did not shut down")
	}

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Fatalf("expected socket file to be removed, stat err=%v", err)
	}
}

func waitForSocket(t *testing.T, socketPath string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(socketPath); err == nil {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("socket %s was not created", socketPath)
}
