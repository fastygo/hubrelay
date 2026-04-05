package buildprofile

import (
	"os"
	"strings"
	"sync"
	"time"

	"sshbot/internal/egress"
	"sshbot/pkg/contract"
)

type Capability = contract.Capability

const (
	CapabilityAdapterHTTPChat    Capability = "adapter.http_chat"
	CapabilityAdapterGRPC        Capability = "adapter.grpc"
	CapabilityAdapterUnixSocket  Capability = "adapter.unix_socket"
	CapabilityAdapterEmail       Capability = "adapter.email"
	CapabilityPluginSystemInfo   Capability = "plugin.system.info"
	CapabilityPluginCapabilities Capability = "plugin.system.capabilities"
	CapabilityPluginAudit        Capability = "plugin.system.audit"
	CapabilityEgressStatus       Capability = "plugin.egress.status"
	CapabilityAIChat             Capability = "ai.chat"
	CapabilityProxySession       Capability = "proxy.session"
)

type HTTPChatConfig struct {
	Enabled     bool
	BindAddress string
}

type GRPCConfig struct {
	Enabled     bool
	BindAddress string
}

type EmailConfig struct {
	Enabled  bool
	Provider string
	Mode     string
}

type UnixSocketConfig struct {
	Enabled    bool
	SocketPath string
}

type OpenAIConfig struct {
	Enabled     bool
	Provider    string
	APIKey      string
	BaseURL     string
	Model       string
	APIMode     string
	ChatHistory bool
	HasAPIKey   bool
}

type ProxySessionConfig struct {
	Enabled bool
	Force   bool
}

type PrivateEgressConfig struct {
	Required   bool
	Interface  string
	TestHost   string
	FailClosed bool
}

type EgressConfig struct {
	UseManager    bool
	CheckInterval time.Duration
	Gateways      []egress.GatewayConfig
}

type Profile struct {
	ID            string
	DisplayName   string
	Capabilities  []Capability
	Config        map[string]string
	HTTPChat      HTTPChatConfig
	GRPC          GRPCConfig
	UnixSocket    UnixSocketConfig
	Email         EmailConfig
	OpenAI        OpenAIConfig
	ProxySession  ProxySessionConfig
	PrivateEgress PrivateEgressConfig
	Egress        EgressConfig
}

var (
	currentProfileID               = "tunnel-email-openai"
	currentDisplayName             = "Tunnel chat + Yandex mail + OpenAI"
	currentHTTPBind                = "127.0.0.1:5500"
	currentGRPCEnabled             = "false"
	currentGRPCBind                = "0.0.0.0:5501"
	currentEmailEnabled            = "true"
	currentEmailProvider           = "yandex"
	currentEmailMode               = "scaffold"
	currentUnixSocketEnabled       = "false"
	currentUnixSocketPath          = "/run/hubrelay/hubrelay.sock"
	currentOpenAIEnabled           = "true"
	currentAIProvider              = "openai"
	currentAIAPIKey                = ""
	currentAIBaseURL               = ""
	currentAIModel                 = "gpt-4.1-mini"
	currentAIAPIMode               = "chat_completions"
	currentChatHistory             = "false"
	currentProxySession            = "true"
	currentProxyForce              = "true"
	currentPrivateEgressRequired   = "false"
	currentPrivateEgressInterface  = ""
	currentPrivateEgressTestHost   = ""
	currentPrivateEgressFailClosed = "false"
	currentEgressGateways          = ""
	currentEgressCheckInterval     = ""
)

func (p Profile) Has(capability Capability) bool {
	for _, item := range p.Capabilities {
		if item == capability {
			return true
		}
	}
	return false
}

func (p Profile) RuntimeProfile() contract.RuntimeProfile {
	capabilities := make([]contract.Capability, 0, len(p.Capabilities))
	for _, capability := range p.Capabilities {
		capabilities = append(capabilities, capability)
	}
	return contract.RuntimeProfile{
		ID:           p.ID,
		DisplayName:  p.DisplayName,
		Capabilities: capabilities,
		Config:       cloneConfigMap(p.Config),
	}
}

// Current returns the immutable runtime profile compiled into the image.
// When compile-time values are empty, environment variables (INPUT_AI_*)
// are checked so that local `go run ./cmd/bot` works without -ldflags.
func Current() Profile {
	applyEnvOverrides()
	hasAIKey := strings.TrimSpace(currentAIAPIKey) != ""
	capabilities := []Capability{
		CapabilityPluginSystemInfo,
		CapabilityPluginCapabilities,
		CapabilityPluginAudit,
	}
	if isTruthy(currentEmailEnabled) {
		capabilities = append(capabilities, CapabilityAdapterEmail)
	}
	if isTruthy(currentUnixSocketEnabled) {
		capabilities = append(capabilities, CapabilityAdapterUnixSocket)
	}
	if isTruthy(currentOpenAIEnabled) && hasAIKey {
		capabilities = append(capabilities, CapabilityAIChat)
		if provider := strings.TrimSpace(strings.ToLower(currentAIProvider)); provider != "" {
			capabilities = append(capabilities, Capability("ai."+provider))
		}
	}
	if currentHTTPBind != "" {
		capabilities = append(capabilities, CapabilityAdapterHTTPChat)
	}
	if isTruthy(currentGRPCEnabled) && strings.TrimSpace(currentGRPCBind) != "" {
		capabilities = append(capabilities, CapabilityAdapterGRPC)
	}
	if isTruthy(currentProxySession) {
		capabilities = append(capabilities, CapabilityProxySession)
	}
	egressGateways, err := egress.ParseGatewayConfigs(currentEgressGateways)
	if err == nil && len(egressGateways) > 0 {
		capabilities = append(capabilities, CapabilityEgressStatus)
	}

	return Profile{
		ID:           currentProfileID,
		DisplayName:  currentDisplayName,
		Capabilities: capabilities,
		Config: map[string]string{
			"http_bind":                  currentHTTPBind,
			"grpc_enabled":               currentGRPCEnabled,
			"grpc_bind":                  strings.TrimSpace(currentGRPCBind),
			"email_enabled":              currentEmailEnabled,
			"email_provider":             currentEmailProvider,
			"email_mode":                 currentEmailMode,
			"unix_socket_enabled":        currentUnixSocketEnabled,
			"unix_socket_path":           strings.TrimSpace(currentUnixSocketPath),
			"ai_enabled":                 currentOpenAIEnabled,
			"ai_provider":                strings.TrimSpace(strings.ToLower(currentAIProvider)),
			"ai_base_url":                strings.TrimSpace(currentAIBaseURL),
			"ai_model":                   strings.TrimSpace(currentAIModel),
			"ai_api_mode":                normalizeAIAPIMode(currentAIAPIMode),
			"ai_has_api_key":             boolString(hasAIKey),
			"chat_history":               currentChatHistory,
			"proxy_session":              currentProxySession,
			"proxy_force":                currentProxyForce,
			"private_egress_required":    currentPrivateEgressRequired,
			"private_egress_interface":   strings.TrimSpace(currentPrivateEgressInterface),
			"private_egress_test_host":   strings.TrimSpace(currentPrivateEgressTestHost),
			"private_egress_fail_closed": currentPrivateEgressFailClosed,
			"egress_gateways":            currentEgressGateways,
			"egress_check_interval":      currentEgressCheckInterval,
		},
		HTTPChat: HTTPChatConfig{
			Enabled:     currentHTTPBind != "",
			BindAddress: currentHTTPBind,
		},
		GRPC: GRPCConfig{
			Enabled:     isTruthy(currentGRPCEnabled),
			BindAddress: strings.TrimSpace(currentGRPCBind),
		},
		UnixSocket: UnixSocketConfig{
			Enabled:    isTruthy(currentUnixSocketEnabled),
			SocketPath: strings.TrimSpace(currentUnixSocketPath),
		},
		Email: EmailConfig{
			Enabled:  isTruthy(currentEmailEnabled),
			Provider: currentEmailProvider,
			Mode:     currentEmailMode,
		},
		OpenAI: OpenAIConfig{
			Enabled:     isTruthy(currentOpenAIEnabled),
			Provider:    strings.TrimSpace(strings.ToLower(currentAIProvider)),
			APIKey:      currentAIAPIKey,
			BaseURL:     strings.TrimSpace(currentAIBaseURL),
			Model:       strings.TrimSpace(currentAIModel),
			APIMode:     normalizeAIAPIMode(currentAIAPIMode),
			ChatHistory: isTruthy(currentChatHistory),
			HasAPIKey:   hasAIKey,
		},
		ProxySession: ProxySessionConfig{
			Enabled: isTruthy(currentProxySession),
			Force:   isTruthy(currentProxyForce),
		},
		PrivateEgress: PrivateEgressConfig{
			Required:   isTruthy(currentPrivateEgressRequired),
			Interface:  strings.TrimSpace(currentPrivateEgressInterface),
			TestHost:   strings.TrimSpace(currentPrivateEgressTestHost),
			FailClosed: isTruthy(currentPrivateEgressFailClosed),
		},
		Egress: EgressConfig{
			UseManager:    err == nil && len(egressGateways) > 0,
			CheckInterval: egress.NormalizeCheckInterval(currentEgressCheckInterval),
			Gateways:      egressGateways,
		},
	}
}

var envOnce sync.Once

func envOverrideAlways(current *string, key string) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		*current = value
	}
}

func applyEnvOverrides() {
	envOnce.Do(func() {
		envOverrideAlways(&currentAIAPIKey, "INPUT_AI_API_KEY")
		envOverrideAlways(&currentAIBaseURL, "INPUT_AI_BASE_URL")
		envOverrideAlways(&currentAIModel, "INPUT_AI_MODEL")
		envOverrideAlways(&currentAIProvider, "INPUT_AI_PROVIDER")
		envOverrideAlways(&currentAIAPIMode, "INPUT_AI_API_MODE")
		envOverrideAlways(&currentGRPCEnabled, "INPUT_GRPC_ENABLED")
		envOverrideAlways(&currentGRPCBind, "INPUT_GRPC_BIND")
		envOverrideAlways(&currentProxyForce, "INPUT_PROXY_SESSION_FORCE")
		envOverrideAlways(&currentProxySession, "INPUT_PROXY_SESSION_ENABLED")
		envOverrideAlways(&currentUnixSocketEnabled, "INPUT_UNIX_SOCKET_ENABLED")
		envOverrideAlways(&currentUnixSocketPath, "INPUT_UNIX_SOCKET_PATH")
		envOverrideAlways(&currentChatHistory, "INPUT_CHAT_HISTORY")
		envOverrideAlways(&currentPrivateEgressRequired, "INPUT_PRIVATE_EGRESS_REQUIRED")
		envOverrideAlways(&currentPrivateEgressInterface, "INPUT_PRIVATE_EGRESS_INTERFACE")
		envOverrideAlways(&currentPrivateEgressTestHost, "INPUT_PRIVATE_EGRESS_TEST_HOST")
		envOverrideAlways(&currentPrivateEgressFailClosed, "INPUT_PRIVATE_EGRESS_FAIL_CLOSED")
		envOverrideAlways(&currentEgressGateways, "INPUT_EGRESS_GATEWAYS")
		envOverrideAlways(&currentEgressCheckInterval, "INPUT_EGRESS_CHECK_INTERVAL")
	})
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func normalizeAIAPIMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "responses":
		return "responses"
	case "chat", "chat_completions", "chat/completions", "":
		return "chat_completions"
	default:
		return "chat_completions"
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func cloneConfigMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
