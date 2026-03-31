package hubrelay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type commandPayload struct {
	PrincipalID string            `json:"principal_id"`
	Roles       []string          `json:"roles,omitempty"`
	Command     string            `json:"command"`
	Args        map[string]string `json:"args,omitempty"`
}

func (c *Client) Discover(ctx context.Context) (DiscoveryResponse, error) {
	response, err := c.transport.Do(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return DiscoveryResponse{}, err
	}
	defer response.Body.Close()

	var discovery DiscoveryResponse
	if err := decodeJSON(response.Body, &discovery); err != nil {
		return DiscoveryResponse{}, err
	}
	return discovery, nil
}

func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	response, err := c.transport.Do(ctx, http.MethodGet, "/healthz", nil)
	if err != nil {
		return HealthResponse{}, err
	}
	defer response.Body.Close()

	var health HealthResponse
	if err := decodeJSON(response.Body, &health); err != nil {
		return HealthResponse{}, err
	}
	return health, nil
}

func (c *Client) Capabilities(ctx context.Context, principal Principal) (CapabilitiesResponse, error) {
	result, err := c.Execute(ctx, CommandRequest{
		Principal: principal,
		Command:   "capabilities",
	})
	if err != nil {
		return CapabilitiesResponse{}, err
	}
	if result.Status == "error" {
		return CapabilitiesResponse{}, errors.New(result.Message)
	}

	var response CapabilitiesResponse
	if err := decodeResultData(result.Data, &response); err != nil {
		return CapabilitiesResponse{}, err
	}
	return response, nil
}

func (c *Client) Execute(ctx context.Context, req CommandRequest) (CommandResult, error) {
	payload := commandPayload{
		Command: req.Command,
		Args:    req.Args,
	}
	principal := c.principalOrDefault(req.Principal)
	payload.PrincipalID = principal.ID
	payload.Roles = append([]string(nil), principal.Roles...)

	body, err := json.Marshal(payload)
	if err != nil {
		return CommandResult{}, err
	}

	response, err := c.transport.Do(ctx, http.MethodPost, "/api/command", bytes.NewReader(body))
	if err != nil {
		return CommandResult{}, err
	}
	defer response.Body.Close()

	var result CommandResult
	if err := decodeJSON(response.Body, &result); err != nil {
		return CommandResult{}, err
	}
	return result, nil
}

func (c *Client) ExecuteStream(ctx context.Context, req CommandRequest) (*Stream, error) {
	payload := commandPayload{
		Command: req.Command,
		Args:    req.Args,
	}
	principal := c.principalOrDefault(req.Principal)
	payload.PrincipalID = principal.ID
	payload.Roles = append([]string(nil), principal.Roles...)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := c.transport.Do(ctx, http.MethodPost, "/api/command/stream", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= http.StatusBadRequest {
		defer response.Body.Close()
		var result CommandResult
		if err := decodeJSON(response.Body, &result); err != nil {
			return nil, err
		}
		return nil, errors.New(result.Message)
	}
	return newStream(response), nil
}

func (c *Client) EgressStatus(ctx context.Context, principal Principal) (EgressStatusResponse, error) {
	result, err := c.Execute(ctx, CommandRequest{
		Principal: principal,
		Command:   "egress-status",
	})
	if err != nil {
		return EgressStatusResponse{}, err
	}
	if result.Status == "error" {
		return EgressStatusResponse{}, errors.New(result.Message)
	}

	var response EgressStatusResponse
	if err := decodeResultData(result.Data, &response); err != nil {
		return EgressStatusResponse{}, err
	}
	return response, nil
}

func decodeResultData(data map[string]any, target any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func decodeJSON(reader io.Reader, target any) error {
	return json.NewDecoder(reader).Decode(target)
}
