package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/safety"
)

var ErrUnknownCommand = errors.New("unknown command")

type Service struct {
	profile               buildprofile.Profile
	store                 Store
	plugins               map[string]Plugin
	activeGatewayProvider ActiveGatewayProvider
}

type ActiveGatewayProvider interface {
	ActiveGatewayName() string
}

type ServiceOption func(*Service)

func WithActiveGatewayProvider(provider ActiveGatewayProvider) ServiceOption {
	return func(service *Service) {
		service.activeGatewayProvider = provider
	}
}

type dispatchState struct {
	envelope CommandEnvelope
	plugin   Plugin
	result   *CommandResult
	err      error
	outcome  string
}

func NewService(profile buildprofile.Profile, store Store, plugins []Plugin, options ...ServiceOption) (*Service, error) {
	if err := store.EnsureSchema(profile); err != nil {
		return nil, err
	}

	service := &Service{
		profile: profile,
		store:   store,
		plugins: make(map[string]Plugin, len(plugins)),
	}

	for _, plugin := range plugins {
		name := strings.ToLower(strings.TrimSpace(plugin.Descriptor().Name))
		if name == "" {
			return nil, errors.New("plugin name cannot be empty")
		}
		service.plugins[name] = plugin
	}

	for _, option := range options {
		if option != nil {
			option(service)
		}
	}

	return service, nil
}

func (s *Service) Profile() buildprofile.Profile {
	return s.profile
}

func (s *Service) Execute(ctx context.Context, envelope CommandEnvelope) (CommandResult, error) {
	state, err := s.prepareDispatch(envelope)
	if err != nil {
		return CommandResult{}, err
	}
	if state.result != nil {
		s.recordAudit(state.envelope, *state.result, state.outcome)
		return *state.result, state.err
	}

	result, err := state.plugin.Execute(ctx, CommandContext{
		Profile: s.profile,
		Store:   s.store,
	}, state.envelope)
	if err != nil {
		result.Status = "error"
		if result.Message == "" {
			result.Message = err.Error()
		}
		s.recordAudit(state.envelope, result, "failed")
		return result, err
	}

	if result.Status == "" {
		result.Status = "ok"
	}
	if result.Status == "error" {
		s.recordAudit(state.envelope, result, "rejected")
		return result, nil
	}
	s.recordAudit(state.envelope, result, "ok")
	return result, nil
}

func (s *Service) ExecuteStream(ctx context.Context, envelope CommandEnvelope, writer StreamWriter) error {
	state, err := s.prepareDispatch(envelope)
	if err != nil {
		return err
	}

	recording := &recordingStreamWriter{writer: writer}

	if state.result != nil {
		recording.SetResult(*state.result)
		s.recordAudit(state.envelope, *state.result, state.outcome)
		return recording.Flush()
	}

	commandCtx := CommandContext{
		Profile: s.profile,
		Store:   s.store,
	}

	var executeErr error
	if plugin, ok := state.plugin.(StreamingPlugin); ok {
		executeErr = plugin.ExecuteStream(ctx, commandCtx, state.envelope, recording)
	} else {
		var result CommandResult
		result, executeErr = state.plugin.Execute(ctx, commandCtx, state.envelope)
		if executeErr == nil {
			if chunk, ok := resultStreamChunk(result); ok {
				if chunkErr := recording.WriteChunk(chunk); chunkErr != nil {
					executeErr = chunkErr
				}
			}
		}
		recording.SetResult(result)
	}

	result := recording.Result()
	if executeErr != nil {
		result.Status = "error"
		if result.Message == "" {
			result.Message = executeErr.Error()
		}
		recording.SetResult(result)
		s.recordAudit(state.envelope, result, "failed")
		if flushErr := recording.Flush(); flushErr != nil {
			return flushErr
		}
		return executeErr
	}

	if result.Status == "" {
		result.Status = "ok"
		recording.SetResult(result)
	}

	outcome := "ok"
	if result.Status == "error" {
		outcome = "rejected"
	}
	s.recordAudit(state.envelope, result, outcome)
	return recording.Flush()
}

func (s *Service) prepareDispatch(envelope CommandEnvelope) (dispatchState, error) {
	envelope.Name = strings.ToLower(strings.TrimSpace(envelope.Name))
	envelope.Transport = strings.ToLower(strings.TrimSpace(envelope.Transport))
	if envelope.RequestedAt.IsZero() {
		envelope.RequestedAt = time.Now().UTC()
	}

	if err := s.store.UpsertPrincipal(envelope.Principal); err != nil {
		return dispatchState{}, err
	}

	session := SessionState{
		ID:         envelope.Transport + ":" + envelope.Principal.ID,
		Principal:  envelope.Principal.ID,
		Transport:  envelope.Transport,
		LastSeenAt: envelope.RequestedAt,
		Values: map[string]string{
			"last_command": envelope.Name,
		},
	}
	if err := s.store.SaveSession(session); err != nil {
		return dispatchState{}, err
	}

	plugin, ok := s.plugins[envelope.Name]
	if !ok {
		result := CommandResult{
			Status:  "error",
			Message: fmt.Sprintf("unknown command: %s", envelope.Name),
		}
		return dispatchState{
			envelope: envelope,
			result:   &result,
			err:      ErrUnknownCommand,
			outcome:  "rejected",
		}, nil
	}

	descriptor := plugin.Descriptor()
	for _, capability := range descriptor.RequiredCapabilities {
		if !s.profile.Has(capability) {
			result := CommandResult{
				Status:  "error",
				Message: fmt.Sprintf("capability %s is not available in profile %s", capability, s.profile.ID),
			}
			return dispatchState{
				envelope: envelope,
				result:   &result,
				outcome:  "rejected",
			}, nil
		}
	}

	if blocked := scanSensitiveAsk(envelope); blocked.Blocked {
		result := CommandResult{
			Status:  "error",
			Message: "request blocked by sensitive data policy",
			Data: map[string]any{
				"policy_code": "sensitive_data",
				"rule_codes":  blocked.RuleCodes(),
				"findings":    len(blocked.Findings),
			},
		}
		return dispatchState{
			envelope: envelope,
			result:   &result,
			outcome:  "rejected",
		}, nil
	}

	return dispatchState{
		envelope: envelope,
		plugin:   plugin,
	}, nil
}

func (s *Service) recordAudit(envelope CommandEnvelope, result CommandResult, outcome string) {
	entry := AuditEntry{
		ID:          fmt.Sprintf("%s:%d", envelope.ID, time.Now().UTC().UnixNano()),
		CommandID:   envelope.ID,
		PrincipalID: envelope.Principal.ID,
		Transport:   envelope.Transport,
		CommandName: envelope.Name,
		Outcome:     outcome,
		Message:     result.Message,
		Metadata: map[string]string{
			"profile_id":   s.profile.ID,
			"profile_name": s.profile.DisplayName,
			"roles":        strings.Join(slices.Clone(envelope.Principal.Roles), ","),
		},
		RecordedAt: time.Now().UTC(),
	}
	if s.activeGatewayProvider != nil {
		if gatewayName := strings.TrimSpace(s.activeGatewayProvider.ActiveGatewayName()); gatewayName != "" {
			entry.Metadata["egress_gateway"] = gatewayName
		}
	}
	_ = s.store.RecordAudit(entry)
}

func scanSensitiveAsk(envelope CommandEnvelope) safety.Report {
	if envelope.Name != "ask" {
		return safety.Report{}
	}
	prompt := strings.TrimSpace(envelope.Args["prompt"])
	if prompt == "" {
		prompt = strings.TrimSpace(envelope.RawText)
		if strings.HasPrefix(strings.ToLower(prompt), "ask ") {
			prompt = strings.TrimSpace(prompt[4:])
		}
	}
	return safety.ScanText(prompt)
}

type recordingStreamWriter struct {
	writer StreamWriter
	result CommandResult
}

func (w *recordingStreamWriter) WriteChunk(chunk StreamChunk) error {
	return w.writer.WriteChunk(chunk)
}

func (w *recordingStreamWriter) SetResult(result CommandResult) {
	w.result = cloneCommandResult(result)
	w.writer.SetResult(result)
}

func (w *recordingStreamWriter) Flush() error {
	return w.writer.Flush()
}

func (w *recordingStreamWriter) Result() CommandResult {
	return cloneCommandResult(w.result)
}

func resultStreamChunk(result CommandResult) (StreamChunk, bool) {
	if result.Data == nil {
		return StreamChunk{}, false
	}
	answer, ok := result.Data["answer"].(string)
	if !ok || answer == "" {
		return StreamChunk{}, false
	}
	return StreamChunk{Delta: answer}, true
}
