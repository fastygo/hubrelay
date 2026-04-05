package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	hubrelay "github.com/fastygo/hubrelay-sdk"
	"hubrelay-dashboard/internal/config"
)

type Client struct {
	sdk       *hubrelay.Client
	principal hubrelay.Principal
}

type AuditEntry struct {
	Command   string    `json:"command"`
	Principal string    `json:"principal"`
	Transport string    `json:"transport"`
	Outcome   string    `json:"outcome"`
	Message   string    `json:"message"`
	At        time.Time `json:"at"`
}

func New(cfg config.Config) (*Client, error) {
	principal := hubrelay.Principal{
		ID:      "dashboard",
		Display: "dashboard",
		Roles:   []string{"operator"},
	}

	client, err := hubrelay.NewClient(
		cfg.Transport,
		hubrelay.WithPrincipal(principal),
		hubrelay.WithBaseURL(cfg.HubRelayBaseURL),
		hubrelay.WithGRPCTarget(cfg.HubRelayGRPCTarget),
		hubrelay.WithSocketPath(cfg.HubRelaySocketPath),
	)
	if err != nil {
		return nil, fmt.Errorf("unsupported transport %q: %w", cfg.Transport, err)
	}

	return &Client{
		sdk:       client,
		principal: principal,
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.sdk == nil {
		return nil
	}
	return c.sdk.Close()
}

func (c *Client) Discover(ctx context.Context) (hubrelay.DiscoveryResponse, error) {
	return c.sdk.Discover(ctx)
}

func (c *Client) Health(ctx context.Context) (hubrelay.HealthResponse, error) {
	return c.sdk.Health(ctx)
}

func (c *Client) Capabilities(ctx context.Context) (hubrelay.CapabilitiesResponse, error) {
	return c.sdk.Capabilities(ctx, c.principal)
}

func (c *Client) Ask(ctx context.Context, prompt, model string) (hubrelay.CommandResult, error) {
	args := map[string]string{
		"prompt": prompt,
	}
	if model != "" {
		args["model"] = model
	}

	return c.sdk.Execute(ctx, hubrelay.CommandRequest{
		Principal: c.principal,
		Command:   "ask",
		Args:      args,
	})
}

func (c *Client) AskStream(ctx context.Context, prompt, model string) (hubrelay.ResultStream, error) {
	args := map[string]string{
		"prompt": prompt,
	}
	if model != "" {
		args["model"] = model
	}

	return c.sdk.ExecuteStream(ctx, hubrelay.CommandRequest{
		Principal: c.principal,
		Command:   "ask",
		Args:      args,
	})
}

func (c *Client) EgressStatus(ctx context.Context) (hubrelay.EgressStatusResponse, error) {
	return c.sdk.EgressStatus(ctx, c.principal)
}

func (c *Client) Audit(ctx context.Context, limit int) ([]AuditEntry, error) {
	result, err := c.sdk.Execute(ctx, hubrelay.CommandRequest{
		Principal: c.principal,
		Command:   "audit",
		Args: map[string]string{
			"limit": fmt.Sprintf("%d", limit),
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Status == "error" {
		return nil, errors.New(result.Message)
	}

	var payload struct {
		Entries []AuditEntry `json:"entries"`
	}
	body, err := json.Marshal(result.Data)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload.Entries, nil
}
