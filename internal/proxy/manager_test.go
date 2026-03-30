package proxy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManagerCheckSelectAndFailover(t *testing.T) {
	manager := NewManager(func(_ context.Context, address string) (time.Duration, error) {
		switch address {
		case "10.0.0.1:1080":
			return 90 * time.Millisecond, nil
		case "10.0.0.2:1080":
			return 40 * time.Millisecond, nil
		default:
			return 0, errors.New("down")
		}
	})

	session, err := manager.CreateSession("operator-local", []string{
		"10.0.0.1:1080\n10.0.0.2:1080\n10.0.0.3:1080",
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	session, err = manager.CheckSession(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("CheckSession() error = %v", err)
	}
	if got := session.Candidates[1].Status; got != StatusHealthy {
		t.Fatalf("expected healthy proxy, got %q", got)
	}

	session, err = manager.SelectProxy(session.ID, "fastest")
	if err != nil {
		t.Fatalf("SelectProxy(fastest) error = %v", err)
	}
	if session.SelectedProxy != "10.0.0.2:1080" {
		t.Fatalf("unexpected selected proxy %q", session.SelectedProxy)
	}

	lease, err := manager.AcquireLease(session.ID)
	if err != nil {
		t.Fatalf("AcquireLease() error = %v", err)
	}
	if lease.Address != "10.0.0.2:1080" {
		t.Fatalf("unexpected lease %q", lease.Address)
	}

	session, err = manager.ReportFailure(session.ID, "10.0.0.2:1080", "dial tcp timeout")
	if err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}
	if session.SelectedProxy != "10.0.0.1:1080" {
		t.Fatalf("expected failover to second healthy proxy, got %q", session.SelectedProxy)
	}
}

func TestManagerRejectsInvalidProxyAddress(t *testing.T) {
	manager := NewManager(nil)

	if _, err := manager.CreateSession("operator-local", []string{"bad-address"}); err == nil {
		t.Fatalf("expected invalid address error")
	}
}
