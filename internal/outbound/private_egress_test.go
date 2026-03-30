package outbound

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestInterfaceEgressCheckerRejectsDownInterface(t *testing.T) {
	checker := &InterfaceEgressChecker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg0", Flags: 0}, nil
		},
		lookupIPs:    lookupHostIPs,
		localIPForIP: localIPForTarget,
	}

	if err := checker.Check(context.Background(), "wg0", "10.88.0.1"); !errors.Is(err, ErrPrivateEgressInterfaceDown) {
		t.Fatalf("expected interface down error, got %v", err)
	}
}

func TestInterfaceEgressCheckerAcceptsMatchingLocalRoute(t *testing.T) {
	checker := &InterfaceEgressChecker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg0", Flags: net.FlagUp}, nil
		},
		lookupIPs: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.1")}, nil
		},
		localIPForIP: func(context.Context, net.IP) (net.IP, error) {
			return net.ParseIP("10.88.0.3"), nil
		},
	}

	originalInterfaceIPs := interfaceIPs
	interfaceIPs = func(*net.Interface) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.88.0.3")}, nil
	}
	t.Cleanup(func() {
		interfaceIPs = originalInterfaceIPs
	})

	if err := checker.Check(context.Background(), "wg0", "10.88.0.1"); err != nil {
		t.Fatalf("expected private egress check to pass, got %v", err)
	}
}

func TestInterfaceEgressCheckerRejectsRouteMismatch(t *testing.T) {
	checker := &InterfaceEgressChecker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg0", Flags: net.FlagUp}, nil
		},
		lookupIPs: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.1")}, nil
		},
		localIPForIP: func(context.Context, net.IP) (net.IP, error) {
			return net.ParseIP("192.0.2.10"), nil
		},
	}

	originalInterfaceIPs := interfaceIPs
	interfaceIPs = func(*net.Interface) ([]net.IP, error) {
		return []net.IP{net.ParseIP("10.88.0.3")}, nil
	}
	t.Cleanup(func() {
		interfaceIPs = originalInterfaceIPs
	})

	if err := checker.Check(context.Background(), "wg0", "10.88.0.1"); !errors.Is(err, ErrPrivateEgressRouteMismatch) {
		t.Fatalf("expected route mismatch error, got %v", err)
	}
}
