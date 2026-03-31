package outbound

import (
	"context"
	"testing"

	"sshbot/internal/egress"
)

type staticHealthChecker struct {
	statuses map[string]egress.GatewayStatus
}

func (c staticHealthChecker) CheckGateway(_ context.Context, gateway egress.Gateway) egress.GatewayStatus {
	return c.statuses[gateway.Name]
}

func TestEgressManagerCheckerDelegatesToActiveGateway(t *testing.T) {
	manager := egress.NewManager(egress.Config{
		Gateways: []egress.GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
		},
	}, staticHealthChecker{
		statuses: map[string]egress.GatewayStatus{
			"wg-b1": {Healthy: true, HealthLevel: "healthy"},
		},
	})
	manager.Refresh(context.Background())

	checker := NewEgressManagerChecker(manager)
	if err := checker.Check(context.Background(), "", ""); err != nil {
		t.Fatalf("expected manager-backed check to pass, got %v", err)
	}
}

func TestEgressManagerCheckerFailsClosedWithoutHealthyGateway(t *testing.T) {
	manager := egress.NewManager(egress.Config{
		Gateways: []egress.GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
		},
	}, staticHealthChecker{
		statuses: map[string]egress.GatewayStatus{
			"wg-b1": {Healthy: false, HealthLevel: "wg", LastError: "transport failed"},
		},
	})
	manager.Refresh(context.Background())

	checker := NewEgressManagerChecker(manager)
	if err := checker.Check(context.Background(), "", ""); err != ErrNoHealthyEgressGateway {
		t.Fatalf("expected fail-closed error, got %v", err)
	}
}
