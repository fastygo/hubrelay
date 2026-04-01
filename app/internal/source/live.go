package source

import (
	"context"

	"hubrelay-dashboard/internal/relay"
	"sshbot/sdk/hubrelay"
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
		HTTPBind:     response.HTTPBind,
		EmailEnabled: response.EmailEnabled,
		AIEnabled:    response.AIEnabled,
		AIProvider:   response.AIProvider,
		AIBaseURL:    response.AIBaseURL,
		AIModel:      response.AIModel,
		AIAPIMode:    response.AIAPIMode,
		ChatHistory:  response.ChatHistory,
		AIHasAPIKey:  response.AIHasAPIKey,
		ProxySession: response.ProxySession,
		ProxyForce:   response.ProxyForce,
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
	stream *hubrelay.Stream
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
