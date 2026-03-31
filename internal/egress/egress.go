package egress

import (
	"context"
	"log"
	"sync"
	"time"
)

type Gateway struct {
	Name      string
	Interface string
	TestHost  string
	ProbeURL  string
	Priority  int
	Enabled   bool
}

type LevelStatus struct {
	OK        bool      `json:"ok"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type HealthLevels struct {
	WG        LevelStatus `json:"wg"`
	Transport LevelStatus `json:"transport"`
	Business  LevelStatus `json:"business"`
}

type GatewayStatus struct {
	Name             string       `json:"name"`
	Interface        string       `json:"interface"`
	Priority         int          `json:"priority"`
	Enabled          bool         `json:"enabled"`
	Healthy          bool         `json:"healthy"`
	HealthLevel      string       `json:"health_level"`
	Levels           HealthLevels `json:"levels"`
	LastCheckAt      time.Time    `json:"last_check_at,omitempty"`
	LastTransitionAt time.Time    `json:"last_transition_at,omitempty"`
	LastError        string       `json:"last_error,omitempty"`
	Active           bool         `json:"active"`
}

type HealthChecker interface {
	CheckGateway(context.Context, Gateway) GatewayStatus
}

type Manager struct {
	mu            sync.RWMutex
	gateways      []Gateway
	statuses      map[string]GatewayStatus
	checker       HealthChecker
	checkInterval time.Duration
	logf          func(string, ...any)
}

func NewManager(cfg Config, checker HealthChecker) *Manager {
	if checker == nil {
		checker = NewChecker()
	}

	manager := &Manager{
		gateways:      make([]Gateway, 0, len(cfg.Gateways)),
		statuses:      make(map[string]GatewayStatus, len(cfg.Gateways)),
		checker:       checker,
		checkInterval: cfg.CheckInterval,
		logf:          log.Printf,
	}
	if manager.checkInterval <= 0 {
		manager.checkInterval = DefaultCheckInterval
	}

	for _, gateway := range cfg.Gateways {
		item := Gateway{
			Name:      gateway.Name,
			Interface: gateway.Interface,
			TestHost:  gateway.TestHost,
			ProbeURL:  gateway.ProbeURL,
			Priority:  gateway.Priority,
			Enabled:   gateway.Enabled,
		}
		manager.gateways = append(manager.gateways, item)
		manager.statuses[item.Name] = GatewayStatus{
			Name:        item.Name,
			Interface:   item.Interface,
			Priority:    item.Priority,
			Enabled:     item.Enabled,
			HealthLevel: healthLevelUnknown,
		}
	}

	return manager
}

func (m *Manager) Start(ctx context.Context) {
	m.Refresh(ctx)

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Refresh(ctx)
		}
	}
}

func (m *Manager) Refresh(ctx context.Context) {
	if m == nil {
		return
	}

	now := time.Now().UTC()
	next := make(map[string]GatewayStatus, len(m.gateways))
	for _, gateway := range m.gateways {
		status := m.checker.CheckGateway(ctx, gateway)
		status.Name = gateway.Name
		status.Interface = gateway.Interface
		status.Priority = gateway.Priority
		status.Enabled = gateway.Enabled
		status.LastCheckAt = now
		next[gateway.Name] = status
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, gateway := range m.gateways {
		status := next[gateway.Name]
		previous, hadPrevious := m.statuses[gateway.Name]
		if !hadPrevious || stateChanged(previous, status) {
			status.LastTransitionAt = now
			if m.logf != nil {
				m.logf("[egress] gateway=%s interface=%s healthy=%v level=%s error=%q",
					status.Name, status.Interface, status.Healthy, status.HealthLevel, status.LastError)
			}
		} else {
			status.LastTransitionAt = previous.LastTransitionAt
		}
		m.statuses[gateway.Name] = status
	}
}

func (m *Manager) Active() *Gateway {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, gateway := range m.gateways {
		status, ok := m.statuses[gateway.Name]
		if ok && status.Enabled && status.Healthy {
			selected := gateway
			return &selected
		}
	}
	return nil
}

func (m *Manager) ActiveGatewayName() string {
	active := m.Active()
	if active == nil {
		return ""
	}
	return active.Name
}

func (m *Manager) All() []GatewayStatus {
	if m == nil {
		return nil
	}

	activeName := m.ActiveGatewayName()

	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]GatewayStatus, 0, len(m.gateways))
	for _, gateway := range m.gateways {
		status := m.statuses[gateway.Name]
		status.Active = status.Name == activeName && status.Healthy
		items = append(items, status)
	}
	return items
}

func stateChanged(previous, current GatewayStatus) bool {
	return previous.Enabled != current.Enabled ||
		previous.Healthy != current.Healthy ||
		previous.HealthLevel != current.HealthLevel ||
		previous.LastError != current.LastError
}
