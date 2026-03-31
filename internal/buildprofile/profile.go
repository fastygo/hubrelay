package buildprofile

import (
	"os"
	"strings"
	"sync"
	"time"

	"sshbot/internal/egress"
)

type Capability string

const (
	CapabilityAdapterHTTPChat    Capability = "adapter.http_chat"
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
	HTTPChat      HTTPChatConfig
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
		HTTPChat: HTTPChatConfig{
			Enabled:     currentHTTPBind != "",
			BindAddress: currentHTTPBind,
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
