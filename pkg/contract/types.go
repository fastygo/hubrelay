package contract

import (
	"context"
	"time"
)

type Capability = string

type Principal struct {
	Version   string            `json:"version,omitempty"`
	ID        string            `json:"id"`
	Display   string            `json:"display"`
	Transport string            `json:"transport"`
	Scope     string            `json:"scope,omitempty"`
	Roles     []string          `json:"roles"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type CommandEnvelope struct {
	Version     string            `json:"version,omitempty"`
	ID          string            `json:"id"`
	Transport   string            `json:"transport"`
	Name        string            `json:"name"`
	Args        map[string]string `json:"args,omitempty"`
	RawText     string            `json:"raw_text,omitempty"`
	Principal   Principal         `json:"principal"`
	RequestedAt time.Time         `json:"requested_at"`
}

type CommandResult struct {
	Kind            string         `json:"kind,omitempty"`
	Code            string         `json:"code,omitempty"`
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

type RuntimeProfile struct {
	ID           string            `json:"id"`
	DisplayName  string            `json:"display_name"`
	Capabilities []Capability      `json:"capabilities,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
}

type PluginDescriptor struct {
	Name                 string
	Summary              string
	RequiredCapabilities []Capability
}

type Store interface {
	EnsureSchema(RuntimeProfile) error
	UpsertPrincipal(Principal) error
	SaveSession(SessionState) error
	RecordAudit(AuditEntry) error
	ListRecentAudit(limit int) ([]AuditEntry, error)
	Close() error
}

type CommandContext struct {
	Profile RuntimeProfile
	Store   Store
	Config  map[string]string
}

type Plugin interface {
	Descriptor() PluginDescriptor
	Execute(context.Context, CommandContext, CommandEnvelope) (CommandResult, error)
}

type Adapter interface {
	Name() string
	Start(context.Context) error
}
