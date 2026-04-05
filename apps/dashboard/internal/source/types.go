package source

import (
	"context"
	"time"
)

type Source interface {
	Health(ctx context.Context) (HealthData, error)
	Capabilities(ctx context.Context) (CapabilitiesData, error)
	Ask(ctx context.Context, prompt, model string) (CommandResult, error)
	AskStream(ctx context.Context, prompt, model string) (AskStream, error)
	Egress(ctx context.Context) (EgressData, error)
	Audit(ctx context.Context, limit int) ([]AuditEntry, error)
}

type AskStream interface {
	Next() bool
	Chunk() StreamChunk
	Result() (CommandResult, error)
	Close() error
}

type CommandResult struct {
	Status          string         `json:"status"`
	Message         string         `json:"message"`
	Data            map[string]any `json:"data,omitempty"`
	RequiresConfirm bool           `json:"requires_confirm,omitempty"`
}

type HealthData struct {
	Discovery DiscoveryData `json:"discovery"`
	Health    AdapterHealth `json:"health"`
}

type DiscoveryData struct {
	Service string `json:"service"`
	Profile string `json:"profile"`
	Status  string `json:"status"`
}

type AdapterHealth struct {
	Status  string `json:"status"`
	Adapter string `json:"adapter"`
	Profile string `json:"profile"`
}

type CapabilitiesData struct {
	ProfileID    string            `json:"profileID"`
	DisplayName  string            `json:"displayName"`
	Capabilities []string          `json:"capabilities"`
	Config       map[string]string `json:"config,omitempty"`
	HTTPBind     string            `json:"httpBind"`
	EmailEnabled bool              `json:"emailEnabled"`
	AIEnabled    bool              `json:"aiEnabled"`
	AIProvider   string            `json:"aiProvider"`
	AIBaseURL    string            `json:"aiBaseURL"`
	AIModel      string            `json:"aiModel"`
	AIAPIMode    string            `json:"aiApiMode"`
	ChatHistory  bool              `json:"chatHistory"`
	AIHasAPIKey  bool              `json:"aiHasApiKey"`
	ProxySession bool              `json:"proxySession"`
	ProxyForce   bool              `json:"proxyForce"`
}

type StreamChunk struct {
	Delta    string         `json:"delta"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type EgressData struct {
	Gateways []GatewayData `json:"gateways"`
}

type GatewayData struct {
	Name             string        `json:"name"`
	Interface        string        `json:"interface"`
	Priority         int           `json:"priority"`
	Enabled          bool          `json:"enabled"`
	Healthy          bool          `json:"healthy"`
	HealthLevel      string        `json:"healthLevel"`
	Levels           GatewayLevels `json:"levels"`
	LastCheckAt      time.Time     `json:"lastCheckAt"`
	LastTransitionAt time.Time     `json:"lastTransitionAt"`
	LastError        string        `json:"lastError"`
	Active           bool          `json:"active"`
}

type GatewayLevels struct {
	WG        LevelStatus `json:"wg"`
	Transport LevelStatus `json:"transport"`
	Business  LevelStatus `json:"business"`
}

type LevelStatus struct {
	OK        bool      `json:"ok"`
	CheckedAt time.Time `json:"checkedAt"`
	Error     string    `json:"error"`
}

type AuditEntry struct {
	Command   string    `json:"command"`
	Principal string    `json:"principal"`
	Transport string    `json:"transport"`
	Outcome   string    `json:"outcome"`
	Message   string    `json:"message"`
	At        time.Time `json:"at"`
}
