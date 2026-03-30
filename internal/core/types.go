package core

import (
	"context"
	"time"

	"sshbot/internal/buildprofile"
)

type Principal struct {
	ID        string            `json:"id"`
	Display   string            `json:"display"`
	Transport string            `json:"transport"`
	Roles     []string          `json:"roles"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type CommandEnvelope struct {
	ID          string            `json:"id"`
	Transport   string            `json:"transport"`
	Name        string            `json:"name"`
	Args        map[string]string `json:"args,omitempty"`
	RawText     string            `json:"raw_text,omitempty"`
	Principal   Principal         `json:"principal"`
	RequestedAt time.Time         `json:"requested_at"`
}

type CommandResult struct {
	Status          string         `json:"status"`
	Message         string         `json:"message"`
	Data            map[string]any `json:"data,omitempty"`
	RequiresConfirm bool           `json:"requires_confirm,omitempty"`
}

type AuditEntry struct {
	ID          string            `json:"id"`
	CommandID   string            `json:"command_id"`
	PrincipalID string            `json:"principal_id"`
	Transport   string            `json:"transport"`
	CommandName string            `json:"command_name"`
	Outcome     string            `json:"outcome"`
	Message     string            `json:"message"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RecordedAt  time.Time         `json:"recorded_at"`
}

type SessionState struct {
	ID         string            `json:"id"`
	Principal  string            `json:"principal"`
	Transport  string            `json:"transport"`
	LastSeenAt time.Time         `json:"last_seen_at"`
	Values     map[string]string `json:"values,omitempty"`
}

type PluginDescriptor struct {
	Name                 string
	Summary              string
	RequiredCapabilities []buildprofile.Capability
}

type CommandContext struct {
	Profile buildprofile.Profile
	Store   Store
}

type Plugin interface {
	Descriptor() PluginDescriptor
	Execute(context.Context, CommandContext, CommandEnvelope) (CommandResult, error)
}

type Adapter interface {
	Name() string
	Start(context.Context) error
}

type Store interface {
	EnsureSchema(buildprofile.Profile) error
	UpsertPrincipal(Principal) error
	SaveSession(SessionState) error
	RecordAudit(AuditEntry) error
	ListRecentAudit(limit int) ([]AuditEntry, error)
	Close() error
}
