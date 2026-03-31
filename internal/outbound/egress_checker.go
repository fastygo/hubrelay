package outbound

import (
	"context"
	"errors"

	"sshbot/internal/egress"
)

var ErrNoHealthyEgressGateway = errors.New("no healthy egress gateway is available")

type EgressManagerChecker struct {
	manager *egress.Manager
}

func NewEgressManagerChecker(manager *egress.Manager) *EgressManagerChecker {
	return &EgressManagerChecker{manager: manager}
}

func (c *EgressManagerChecker) Check(_ context.Context, _, _ string) error {
	if c == nil || c.manager == nil {
		return ErrNoHealthyEgressGateway
	}
	if c.manager.Active() == nil {
		return ErrNoHealthyEgressGateway
	}
	return nil
}
