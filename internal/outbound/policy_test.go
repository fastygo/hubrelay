package outbound

import (
	"context"
	"testing"
	"time"

	proxymgr "sshbot/internal/proxy"
)

func TestPolicyRequiresProxySession(t *testing.T) {
	policy := Policy{RequireProxy: true}

	if _, err := policy.ResolveProxyAddress("", nil); !errorsIs(err, ErrProxyManagerRequired, ErrProxyRequired) {
		t.Fatalf("expected proxy policy error, got %v", err)
	}
}

func TestPolicyUsesActiveLease(t *testing.T) {
	manager := proxymgr.NewManager(func(_ context.Context, address string) (time.Duration, error) {
		if address == "10.0.0.2:1080" {
			return 10 * time.Millisecond, nil
		}
		return 25 * time.Millisecond, nil
	})

	session, err := manager.CreateSession("operator", []string{"10.0.0.1:1080\n10.0.0.2:1080"})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if _, err := manager.CheckSession(context.Background(), session.ID); err != nil {
		t.Fatalf("CheckSession() error = %v", err)
	}
	if _, err := manager.SelectProxy(session.ID, "fastest"); err != nil {
		t.Fatalf("SelectProxy() error = %v", err)
	}

	address, err := (Policy{RequireProxy: true}).ResolveProxyAddress(session.ID, ProxyLeaseResolver{Manager: manager})
	if err != nil {
		t.Fatalf("ResolveProxyAddress() error = %v", err)
	}
	if address != "10.0.0.2:1080" {
		t.Fatalf("unexpected proxy address %q", address)
	}
}

func TestPolicyAllowsDirectWhenNotForced(t *testing.T) {
	address, err := (Policy{RequireProxy: false}).ResolveProxyAddress("", nil)
	if err != nil {
		t.Fatalf("expected direct path to be allowed, got %v", err)
	}
	if address != "" {
		t.Fatalf("expected empty address for direct path, got %q", address)
	}
}

func errorsIs(err error, candidates ...error) bool {
	for _, candidate := range candidates {
		if err == candidate {
			return true
		}
	}
	return false
}
