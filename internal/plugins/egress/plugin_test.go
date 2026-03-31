package egress

import (
	"context"
	"testing"

	"sshbot/internal/core"
	"sshbot/internal/egress"
)

type staticHealthChecker struct {
	statuses map[string]egress.GatewayStatus
}

func (c staticHealthChecker) CheckGateway(_ context.Context, gateway egress.Gateway) egress.GatewayStatus {
	return c.statuses[gateway.Name]
}

func TestStatusPluginReturnsGatewayStatuses(t *testing.T) {
	manager := egress.NewManager(egress.Config{
		Gateways: []egress.GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
			{Name: "wg-b2", Interface: "wg-b2", Priority: 20, Enabled: true},
		},
	}, staticHealthChecker{
		statuses: map[string]egress.GatewayStatus{
			"wg-b1": {Healthy: true, HealthLevel: "healthy"},
			"wg-b2": {Healthy: false, HealthLevel: "transport", LastError: "probe failed"},
		},
	})
	manager.Refresh(context.Background())

	result, err := NewStatusPlugin(manager).Execute(context.Background(), core.CommandContext{}, core.CommandEnvelope{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != "ok" {
		t.Fatalf("expected ok status, got %q", result.Status)
	}

	gateways, ok := result.Data["gateways"].([]map[string]any)
	if !ok || len(gateways) != 2 {
		t.Fatalf("expected two gateway records, got %#v", result.Data["gateways"])
	}
	if result.Message == "" {
		t.Fatalf("expected result message")
	}
}
