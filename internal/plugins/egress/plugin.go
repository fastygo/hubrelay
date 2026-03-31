package egress

import (
	"context"
	"fmt"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	egressmgr "sshbot/internal/egress"
)

type StatusPlugin struct {
	manager *egressmgr.Manager
}

func NewStatusPlugin(manager *egressmgr.Manager) StatusPlugin {
	return StatusPlugin{manager: manager}
}

func (p StatusPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name:    "egress-status",
		Summary: "Return current egress gateway health and failover state",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityEgressStatus,
		},
	}
}

func (p StatusPlugin) Execute(_ context.Context, _ core.CommandContext, _ core.CommandEnvelope) (core.CommandResult, error) {
	if p.manager == nil {
		return core.CommandResult{
			Status:  "error",
			Message: "egress manager is not configured",
		}, nil
	}

	statuses := p.manager.All()
	items := make([]map[string]any, 0, len(statuses))
	for _, status := range statuses {
		items = append(items, map[string]any{
			"name":               status.Name,
			"interface":          status.Interface,
			"priority":           status.Priority,
			"enabled":            status.Enabled,
			"healthy":            status.Healthy,
			"health_level":       status.HealthLevel,
			"last_check_at":      status.LastCheckAt,
			"last_transition_at": status.LastTransitionAt,
			"last_error":         status.LastError,
			"active":             status.Active,
			"levels": map[string]any{
				"wg":        status.Levels.WG,
				"transport": status.Levels.Transport,
				"business":  status.Levels.Business,
			},
		})
	}

	return core.CommandResult{
		Status:  "ok",
		Message: fmt.Sprintf("returned %d egress gateways", len(items)),
		Data: map[string]any{
			"gateways": items,
		},
	}, nil
}
