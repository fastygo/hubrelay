package egress

import (
	"context"
	"strings"
	"testing"
)

type stubHealthChecker struct {
	statuses map[string]GatewayStatus
}

func (c stubHealthChecker) CheckGateway(_ context.Context, gateway Gateway) GatewayStatus {
	status := c.statuses[gateway.Name]
	status.Name = gateway.Name
	status.Interface = gateway.Interface
	status.Priority = gateway.Priority
	status.Enabled = gateway.Enabled
	return status
}

func TestManagerSelectsHighestPriorityHealthyGateway(t *testing.T) {
	manager := NewManager(Config{
		Gateways: []GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
			{Name: "wg-b2", Interface: "wg-b2", Priority: 20, Enabled: true},
		},
	}, stubHealthChecker{
		statuses: map[string]GatewayStatus{
			"wg-b1": {Healthy: true, HealthLevel: healthLevelHealthy},
			"wg-b2": {Healthy: true, HealthLevel: healthLevelHealthy},
		},
	})

	manager.Refresh(context.Background())

	active := manager.Active()
	if active == nil || active.Name != "wg-b1" {
		t.Fatalf("expected wg-b1 to be active, got %+v", active)
	}
}

func TestManagerFallsBackToNextHealthyGateway(t *testing.T) {
	checker := stubHealthChecker{
		statuses: map[string]GatewayStatus{
			"wg-b1": {Healthy: true, HealthLevel: healthLevelHealthy},
			"wg-b2": {Healthy: true, HealthLevel: healthLevelHealthy},
		},
	}
	manager := NewManager(Config{
		Gateways: []GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
			{Name: "wg-b2", Interface: "wg-b2", Priority: 20, Enabled: true},
		},
	}, checker)

	manager.Refresh(context.Background())
	checker.statuses["wg-b1"] = GatewayStatus{
		Healthy:     false,
		HealthLevel: healthLevelWG,
		LastError:   "transport failed",
	}
	manager.checker = checker
	manager.Refresh(context.Background())

	active := manager.Active()
	if active == nil || active.Name != "wg-b2" {
		t.Fatalf("expected wg-b2 to be active after failover, got %+v", active)
	}
}

func TestManagerReturnsNilWhenAllGatewaysUnhealthy(t *testing.T) {
	manager := NewManager(Config{
		Gateways: []GatewayConfig{
			{Name: "wg-b1", Interface: "wg-b1", Priority: 10, Enabled: true},
			{Name: "wg-b2", Interface: "wg-b2", Priority: 20, Enabled: true},
		},
	}, stubHealthChecker{
		statuses: map[string]GatewayStatus{
			"wg-b1": {Healthy: false, HealthLevel: healthLevelWG, LastError: "transport failed"},
			"wg-b2": {Healthy: false, HealthLevel: healthLevelUnknown, LastError: "wg failed"},
		},
	})

	manager.Refresh(context.Background())

	if active := manager.Active(); active != nil {
		t.Fatalf("expected no active gateway, got %+v", active)
	}
}

func TestNormalizeGatewayConfigsRejectsDuplicateNames(t *testing.T) {
	_, err := NormalizeGatewayConfigs([]GatewayConfig{
		{Name: "wg-b1", Interface: "wg-b1", Enabled: true},
		{Name: "wg-b1", Interface: "wg-b2", Enabled: true},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}
