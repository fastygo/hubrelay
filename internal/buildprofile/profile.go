package buildprofile

import (
	"os"
	"strings"
	"sync"
)

type Capability string

const (
	CapabilityAdapterHTTPChat    Capability = "adapter.http_chat"
	CapabilityAdapterEmail       Capability = "adapter.email"
	CapabilityPluginSystemInfo   Capability = "plugin.system.info"
	CapabilityPluginCapabilities Capability = "plugin.system.capabilities"
	CapabilityPluginAudit        Capability = "plugin.system.audit"
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

type Profile struct {
	ID           string
	DisplayName  string
	Capabilities []Capability
	HTTPChat     HTTPChatConfig
	Email        EmailConfig
	OpenAI       OpenAIConfig
	ProxySession ProxySessionConfig
}

var (
	currentProfileID     = "tunnel-email-openai"
	currentDisplayName   = "Tunnel chat + Yandex mail + OpenAI"
	currentHTTPBind      = "127.0.0.1:5500"
	currentEmailEnabled  = "true"
	currentEmailProvider = "yandex"
	currentEmailMode     = "scaffold"
	currentOpenAIEnabled = "true"
	currentAIProvider    = "openai"
	currentAIAPIKey      = ""
	currentAIBaseURL     = ""
	currentAIModel       = "gpt-4.1-mini"
	currentAIAPIMode     = "chat_completions"
	currentChatHistory   = "false"
	currentProxySession  = "true"
	currentProxyForce    = "true"
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

	return Profile{
		ID:           currentProfileID,
		DisplayName:  currentDisplayName,
		Capabilities: capabilities,
		HTTPChat: HTTPChatConfig{
			Enabled:     currentHTTPBind != "",
			BindAddress: currentHTTPBind,
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
		envOverrideAlways(&currentChatHistory, "INPUT_CHAT_HISTORY")
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
