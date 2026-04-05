package hubrelay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type commandPayload struct {
	PrincipalID string            `json:"principal_id"`
	Roles       []string          `json:"roles,omitempty"`
	Command     string            `json:"command"`
	Args        map[string]string `json:"args,omitempty"`
}

var (
	ErrUnsupportedTransport   = errors.New("unsupported transport")
	ErrTransportNotImplemented = errors.New("transport not implemented yet")
	ErrUnexpectedStreamType   = errors.New("unexpected stream type")
)

type commandTransport struct {
	raw              rawTransport
	defaultPrincipal Principal
}

func newCommandTransport(raw rawTransport, principal Principal) CommandTransport {
	return &commandTransport{
		raw:              raw,
		defaultPrincipal: principal,
	}
}

func (t *commandTransport) Discover(ctx context.Context) (DiscoveryResponse, error) {
	response, err := t.raw.Do(ctx, http.MethodGet, "/", nil)
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

func (t *commandTransport) Health(ctx context.Context) (HealthResponse, error) {
	response, err := t.raw.Do(ctx, http.MethodGet, "/healthz", nil)
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

func (t *commandTransport) Capabilities(ctx context.Context, principal Principal) (CapabilitiesResponse, error) {
	result, err := t.Execute(ctx, CommandRequest{
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

func (t *commandTransport) Execute(ctx context.Context, req CommandRequest) (CommandResult, error) {
	payload := commandPayload{
		Command: req.Command,
		Args:    req.Args,
	}
	principal := t.principalOrDefault(req.Principal)
	payload.PrincipalID = principal.ID
	payload.Roles = append([]string(nil), principal.Roles...)

	body, err := json.Marshal(payload)
	if err != nil {
		return CommandResult{}, err
	}

	response, err := t.raw.Do(ctx, http.MethodPost, "/api/command", bytes.NewReader(body))
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

func (t *commandTransport) ExecuteStream(ctx context.Context, req CommandRequest) (ResultStream, error) {
	payload := commandPayload{
		Command: req.Command,
		Args:    req.Args,
	}
	principal := t.principalOrDefault(req.Principal)
	payload.PrincipalID = principal.ID
	payload.Roles = append([]string(nil), principal.Roles...)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := t.raw.Do(ctx, http.MethodPost, "/api/command/stream", bytes.NewReader(body))
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

func (t *commandTransport) EgressStatus(ctx context.Context, principal Principal) (EgressStatusResponse, error) {
	result, err := t.Execute(ctx, CommandRequest{
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

func (t *commandTransport) Close() error {
	if t == nil || t.raw == nil {
		return nil
	}
	return t.raw.Close()
}

func (t *commandTransport) principalOrDefault(principal Principal) Principal {
	if strings.TrimSpace(principal.ID) == "" {
		principal = t.defaultPrincipal
	}
	if strings.TrimSpace(principal.Display) == "" {
		principal.Display = principal.ID
	}
	return principal
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
