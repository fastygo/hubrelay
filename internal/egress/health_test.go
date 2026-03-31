package egress

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func TestCheckerPassesAllThreeLevels(t *testing.T) {
	checker := &Checker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg-b1", Flags: net.FlagUp}, nil
		},
		interfaceIPs: func(*net.Interface) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.3")}, nil
		},
		lookupIPs: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.1")}, nil
		},
		localIPForIP: func(context.Context, net.IP) (net.IP, error) {
			return net.ParseIP("10.88.0.3"), nil
		},
		now:              func() time.Time { return time.Unix(1700000000, 0).UTC() },
		transportTimeout: time.Second,
		businessTimeout:  time.Second,
	}
	checker.transportProbe = func(context.Context, Gateway) error { return nil }
	checker.businessProbe = func(context.Context, Gateway) error { return nil }

	status := checker.CheckGateway(context.Background(), Gateway{
		Name:      "wg-b1",
		Interface: "wg-b1",
		TestHost:  "10.88.0.1",
		ProbeURL:  "https://api.example.com/v1/models",
		Enabled:   true,
	})

	if !status.Healthy {
		t.Fatalf("expected gateway to be healthy, got %+v", status)
	}
	if status.HealthLevel != healthLevelHealthy {
		t.Fatalf("expected healthy level, got %q", status.HealthLevel)
	}
}

func TestCheckerStopsAtWGFailure(t *testing.T) {
	checker := &Checker{
		interfaceByName: func(string) (*net.Interface, error) {
			return nil, errors.New("missing interface")
		},
		now: func() time.Time { return time.Unix(1700000000, 0).UTC() },
	}
	checker.transportProbe = func(context.Context, Gateway) error {
		t.Fatal("transport probe should not run")
		return nil
	}
	checker.businessProbe = func(context.Context, Gateway) error {
		t.Fatal("business probe should not run")
		return nil
	}

	status := checker.CheckGateway(context.Background(), Gateway{
		Name:      "wg-b1",
		Interface: "wg-b1",
		ProbeURL:  "https://api.example.com/v1/models",
		Enabled:   true,
	})

	if status.Healthy {
		t.Fatalf("expected unhealthy gateway, got %+v", status)
	}
	if status.HealthLevel != healthLevelUnknown {
		t.Fatalf("expected unknown level after wg failure, got %q", status.HealthLevel)
	}
	if !strings.Contains(status.LastError, "missing interface") {
		t.Fatalf("expected wg error to be preserved, got %q", status.LastError)
	}
}

func TestCheckerStopsAtTransportFailure(t *testing.T) {
	checker := &Checker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg-b1", Flags: net.FlagUp}, nil
		},
		interfaceIPs: func(*net.Interface) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.3")}, nil
		},
		lookupIPs: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.1")}, nil
		},
		localIPForIP: func(context.Context, net.IP) (net.IP, error) {
			return net.ParseIP("10.88.0.3"), nil
		},
		now: func() time.Time { return time.Unix(1700000000, 0).UTC() },
	}
	checker.transportProbe = func(context.Context, Gateway) error {
		return errors.New("tls timeout")
	}
	checker.businessProbe = func(context.Context, Gateway) error {
		t.Fatal("business probe should not run")
		return nil
	}

	status := checker.CheckGateway(context.Background(), Gateway{
		Name:      "wg-b1",
		Interface: "wg-b1",
		TestHost:  "10.88.0.1",
		ProbeURL:  "https://api.example.com/v1/models",
		Enabled:   true,
	})

	if status.Healthy {
		t.Fatalf("expected unhealthy gateway, got %+v", status)
	}
	if status.HealthLevel != healthLevelWG {
		t.Fatalf("expected wg level after transport failure, got %q", status.HealthLevel)
	}
	if !strings.Contains(status.LastError, "tls timeout") {
		t.Fatalf("expected transport error to be preserved, got %q", status.LastError)
	}
}

func TestCheckerStopsAtBusinessFailure(t *testing.T) {
	checker := &Checker{
		interfaceByName: func(string) (*net.Interface, error) {
			return &net.Interface{Name: "wg-b1", Flags: net.FlagUp}, nil
		},
		interfaceIPs: func(*net.Interface) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.3")}, nil
		},
		lookupIPs: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("10.88.0.1")}, nil
		},
		localIPForIP: func(context.Context, net.IP) (net.IP, error) {
			return net.ParseIP("10.88.0.3"), nil
		},
		now: func() time.Time { return time.Unix(1700000000, 0).UTC() },
	}
	checker.transportProbe = func(context.Context, Gateway) error { return nil }
	checker.businessProbe = func(context.Context, Gateway) error { return errors.New("status 503") }

	status := checker.CheckGateway(context.Background(), Gateway{
		Name:      "wg-b1",
		Interface: "wg-b1",
		TestHost:  "10.88.0.1",
		ProbeURL:  "https://api.example.com/v1/models",
		Enabled:   true,
	})

	if status.Healthy {
		t.Fatalf("expected unhealthy gateway, got %+v", status)
	}
	if status.HealthLevel != healthLevelTransport {
		t.Fatalf("expected transport level after business failure, got %q", status.HealthLevel)
	}
	if !strings.Contains(status.LastError, "status 503") {
		t.Fatalf("expected business error to be preserved, got %q", status.LastError)
	}
}
