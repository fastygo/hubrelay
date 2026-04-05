package relay

import (
	"context"
	"fmt"

	hubrelay "github.com/fastygo/hubrelay-sdk"
	"gitcourse/internal/config"
)

type Client struct {
	sdk       *hubrelay.Client
	principal hubrelay.Principal
}

func New(cfg config.Config) (*Client, error) {
	principal := hubrelay.Principal{
		ID:      "gitcourse",
		Display: "gitcourse",
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

	return &Client{sdk: client, principal: principal}, nil
}

func (c *Client) Close() error {
	if c == nil || c.sdk == nil {
		return nil
	}
	return c.sdk.Close()
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
