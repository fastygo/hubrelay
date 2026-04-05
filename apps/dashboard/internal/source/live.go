package source

import (
	"context"
	"strings"

	hubrelay "github.com/fastygo/hubrelay-sdk"
	"hubrelay-dashboard/internal/relay"
)

type Live struct {
	relay *relay.Client
}

func NewLive(client *relay.Client) *Live {
	return &Live{relay: client}
}

func (s *Live) Health(ctx context.Context) (HealthData, error) {
	var out HealthData

	discovery, err := s.relay.Discover(ctx)
	if err != nil {
		return out, err
	}
	out.Discovery = DiscoveryData{
		Service: discovery.Service,
		Profile: discovery.Profile,
		Status:  discovery.Status,
	}

	health, err := s.relay.Health(ctx)
	if err != nil {
		return out, err
	}
	out.Health = AdapterHealth{
		Status:  health.Status,
		Adapter: health.Adapter,
		Profile: health.Profile,
	}

	return out, nil
}

func (s *Live) Capabilities(ctx context.Context) (CapabilitiesData, error) {
	response, err := s.relay.Capabilities(ctx)
	if err != nil {
		return CapabilitiesData{}, err
	}

	return CapabilitiesData{
		ProfileID:    response.ProfileID,
		DisplayName:  response.DisplayName,
		Capabilities: append([]string(nil), response.Capabilities...),
		Config:       cloneStringMap(response.Config),
		HTTPBind:     firstString(response.HTTPBind, response.Config["http_bind"]),
		EmailEnabled: boolFromConfig(response.EmailEnabled, response.Config, "email_enabled"),
		AIEnabled:    boolFromConfig(response.AIEnabled, response.Config, "ai_enabled"),
		AIProvider:   firstString(response.AIProvider, response.Config["ai_provider"]),
		AIBaseURL:    firstString(response.AIBaseURL, response.Config["ai_base_url"]),
		AIModel:      firstString(response.AIModel, response.Config["ai_model"]),
		AIAPIMode:    firstString(response.AIAPIMode, response.Config["ai_api_mode"]),
		ChatHistory:  boolFromConfig(response.ChatHistory, response.Config, "chat_history"),
		AIHasAPIKey:  boolFromConfig(response.AIHasAPIKey, response.Config, "ai_has_api_key"),
		ProxySession: boolFromConfig(response.ProxySession, response.Config, "proxy_session"),
		ProxyForce:   boolFromConfig(response.ProxyForce, response.Config, "proxy_force"),
	}, nil
}

func (s *Live) Ask(ctx context.Context, prompt, model string) (CommandResult, error) {
	result, err := s.relay.Ask(ctx, prompt, model)
	if err != nil {
		return CommandResult{}, err
	}
	return mapCommandResult(result), nil
}

func (s *Live) AskStream(ctx context.Context, prompt, model string) (AskStream, error) {
	stream, err := s.relay.AskStream(ctx, prompt, model)
	if err != nil {
		return nil, err
	}
	return &liveAskStream{stream: stream}, nil
}

func (s *Live) Egress(ctx context.Context) (EgressData, error) {
	response, err := s.relay.EgressStatus(ctx)
	if err != nil {
		return EgressData{}, err
	}

	gateways := make([]GatewayData, 0, len(response.Gateways))
	for _, gateway := range response.Gateways {
		gateways = append(gateways, mapGateway(gateway))
	}

	return EgressData{Gateways: gateways}, nil
}

func (s *Live) Audit(ctx context.Context, limit int) ([]AuditEntry, error) {
	entries, err := s.relay.Audit(ctx, limit)
	if err != nil {
		return nil, err
	}

	items := make([]AuditEntry, 0, len(entries))
	for _, entry := range entries {
		items = append(items, AuditEntry{
			Command:   entry.Command,
			Principal: entry.Principal,
			Transport: entry.Transport,
			Outcome:   entry.Outcome,
			Message:   entry.Message,
			At:        entry.At,
		})
	}

	return items, nil
}

type liveAskStream struct {
	stream hubrelay.ResultStream
}

func (s *liveAskStream) Next() bool {
	return s.stream.Next()
}

func (s *liveAskStream) Chunk() StreamChunk {
	chunk := s.stream.Chunk()
	return StreamChunk{
		Delta:    chunk.Delta,
		Metadata: cloneMap(chunk.Metadata),
	}
}

func (s *liveAskStream) Result() (CommandResult, error) {
	result, err := s.stream.Result()
	return mapCommandResult(result), err
}

func (s *liveAskStream) Close() error {
	return s.stream.Close()
}

func mapCommandResult(result hubrelay.CommandResult) CommandResult {
	return CommandResult{
		Status:          result.Status,
		Message:         result.Message,
		Data:            cloneMap(result.Data),
		RequiresConfirm: result.RequiresConfirm,
	}
}

func mapGateway(gateway hubrelay.GatewayStatus) GatewayData {
	return GatewayData{
		Name:        gateway.Name,
		Interface:   gateway.Interface,
		Priority:    gateway.Priority,
		Enabled:     gateway.Enabled,
		Healthy:     gateway.Healthy,
		HealthLevel: gateway.HealthLevel,
		Levels: GatewayLevels{
			WG: LevelStatus{
				OK:        gateway.Levels.WG.OK,
				CheckedAt: gateway.Levels.WG.CheckedAt,
				Error:     gateway.Levels.WG.Error,
			},
			Transport: LevelStatus{
				OK:        gateway.Levels.Transport.OK,
				CheckedAt: gateway.Levels.Transport.CheckedAt,
				Error:     gateway.Levels.Transport.Error,
			},
			Business: LevelStatus{
				OK:        gateway.Levels.Business.OK,
				CheckedAt: gateway.Levels.Business.CheckedAt,
				Error:     gateway.Levels.Business.Error,
			},
		},
		LastCheckAt:      gateway.LastCheckAt,
		LastTransitionAt: gateway.LastTransitionAt,
		LastError:        gateway.LastError,
		Active:           gateway.Active,
	}
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}

	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func boolFromConfig(current bool, config map[string]string, key string) bool {
	if current {
		return true
	}
	value, ok := config[key]
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
