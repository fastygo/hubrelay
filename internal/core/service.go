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
	profile buildprofile.Profile
	store   Store
	plugins map[string]Plugin
}

func NewService(profile buildprofile.Profile, store Store, plugins []Plugin) (*Service, error) {
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

	return service, nil
}

func (s *Service) Profile() buildprofile.Profile {
	return s.profile
}

func (s *Service) Execute(ctx context.Context, envelope CommandEnvelope) (CommandResult, error) {
	envelope.Name = strings.ToLower(strings.TrimSpace(envelope.Name))
	envelope.Transport = strings.ToLower(strings.TrimSpace(envelope.Transport))
	if envelope.RequestedAt.IsZero() {
		envelope.RequestedAt = time.Now().UTC()
	}

	if err := s.store.UpsertPrincipal(envelope.Principal); err != nil {
		return CommandResult{}, err
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
		return CommandResult{}, err
	}

	plugin, ok := s.plugins[envelope.Name]
	if !ok {
		result := CommandResult{
			Status:  "error",
			Message: fmt.Sprintf("unknown command: %s", envelope.Name),
		}
		s.recordAudit(envelope, result, "rejected")
		return result, ErrUnknownCommand
	}

	descriptor := plugin.Descriptor()
	for _, capability := range descriptor.RequiredCapabilities {
		if !s.profile.Has(capability) {
			result := CommandResult{
				Status:  "error",
				Message: fmt.Sprintf("capability %s is not available in profile %s", capability, s.profile.ID),
			}
			s.recordAudit(envelope, result, "rejected")
			return result, nil
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
		s.recordAudit(envelope, result, "rejected")
		return result, nil
	}

	result, err := plugin.Execute(ctx, CommandContext{
		Profile: s.profile,
		Store:   s.store,
	}, envelope)
	if err != nil {
		result.Status = "error"
		if result.Message == "" {
			result.Message = err.Error()
		}
		s.recordAudit(envelope, result, "failed")
		return result, err
	}

	if result.Status == "" {
		result.Status = "ok"
	}
	if result.Status == "error" {
		s.recordAudit(envelope, result, "rejected")
		return result, nil
	}
	s.recordAudit(envelope, result, "ok")
	return result, nil
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
