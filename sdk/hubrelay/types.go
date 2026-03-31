package hubrelay

import "time"

type Principal struct {
	ID      string
	Display string
	Roles   []string
}

type CommandRequest struct {
	Principal Principal
	Command   string
	Args      map[string]string
}

type CommandResult struct {
	Status          string         `json:"status"`
	Message         string         `json:"message"`
	Data            map[string]any `json:"data,omitempty"`
	RequiresConfirm bool           `json:"requires_confirm,omitempty"`
}

type DiscoveryResponse struct {
	Service string `json:"service"`
	Profile string `json:"profile"`
	Status  string `json:"status"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Adapter string `json:"adapter,omitempty"`
	Profile string `json:"profile,omitempty"`
}

type CapabilitiesResponse struct {
	ProfileID    string   `json:"profile_id"`
	DisplayName  string   `json:"display_name"`
	Capabilities []string `json:"capabilities"`
	HTTPBind     string   `json:"http_bind,omitempty"`
	EmailEnabled bool     `json:"email_enabled,omitempty"`
	AIEnabled    bool     `json:"ai_enabled,omitempty"`
	AIProvider   string   `json:"ai_provider,omitempty"`
	AIBaseURL    string   `json:"ai_base_url,omitempty"`
	AIModel      string   `json:"ai_model,omitempty"`
	AIAPIMode    string   `json:"ai_api_mode,omitempty"`
	ChatHistory  bool     `json:"chat_history,omitempty"`
	AIHasAPIKey  bool     `json:"ai_has_api_key,omitempty"`
	ProxySession bool     `json:"proxy_session,omitempty"`
	ProxyForce   bool     `json:"proxy_force,omitempty"`
}

type StreamChunk struct {
	Delta    string         `json:"delta"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type LevelStatus struct {
	OK        bool      `json:"ok"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type GatewayLevels struct {
	WG        LevelStatus `json:"wg"`
	Transport LevelStatus `json:"transport"`
	Business  LevelStatus `json:"business"`
}

type GatewayStatus struct {
	Name             string        `json:"name"`
	Interface        string        `json:"interface"`
	Priority         int           `json:"priority"`
	Enabled          bool          `json:"enabled"`
	Healthy          bool          `json:"healthy"`
	HealthLevel      string        `json:"health_level"`
	Levels           GatewayLevels `json:"levels"`
	LastCheckAt      time.Time     `json:"last_check_at,omitempty"`
	LastTransitionAt time.Time     `json:"last_transition_at,omitempty"`
	LastError        string        `json:"last_error,omitempty"`
	Active           bool          `json:"active"`
}

type EgressStatusResponse struct {
	Gateways []GatewayStatus `json:"gateways"`
}
