package egress

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const DefaultCheckInterval = 30 * time.Second

type GatewayConfig struct {
	Name      string `json:"name"`
	Interface string `json:"interface"`
	TestHost  string `json:"test_host"`
	ProbeURL  string `json:"probe_url"`
	Priority  int    `json:"priority"`
	Enabled   bool   `json:"enabled"`
}

type Config struct {
	Gateways      []GatewayConfig
	CheckInterval time.Duration
}

func ParseGatewayConfigs(raw string) ([]GatewayConfig, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var gateways []GatewayConfig
	if err := json.Unmarshal([]byte(trimmed), &gateways); err != nil {
		return nil, fmt.Errorf("parse INPUT_EGRESS_GATEWAYS: %w", err)
	}
	return NormalizeGatewayConfigs(gateways)
}

func NormalizeGatewayConfigs(gateways []GatewayConfig) ([]GatewayConfig, error) {
	if len(gateways) == 0 {
		return nil, nil
	}

	normalized := make([]GatewayConfig, 0, len(gateways))
	seen := make(map[string]struct{}, len(gateways))
	for index, gateway := range gateways {
		gateway.Name = strings.TrimSpace(gateway.Name)
		gateway.Interface = strings.TrimSpace(gateway.Interface)
		gateway.TestHost = strings.TrimSpace(gateway.TestHost)
		gateway.ProbeURL = strings.TrimSpace(gateway.ProbeURL)
		if gateway.Name == "" {
			return nil, fmt.Errorf("gateway[%d] name is required", index)
		}
		if gateway.Interface == "" {
			return nil, fmt.Errorf("gateway[%d] interface is required", index)
		}
		if _, exists := seen[gateway.Name]; exists {
			return nil, fmt.Errorf("gateway %q is duplicated", gateway.Name)
		}
		seen[gateway.Name] = struct{}{}
		normalized = append(normalized, gateway)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].Priority != normalized[j].Priority {
			return normalized[i].Priority < normalized[j].Priority
		}
		return normalized[i].Name < normalized[j].Name
	})
	return normalized, nil
}

func NormalizeCheckInterval(value string) time.Duration {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return DefaultCheckInterval
	}
	parsed, err := time.ParseDuration(trimmed)
	if err != nil || parsed <= 0 {
		return DefaultCheckInterval
	}
	return parsed
}
