package ask

import (
	"context"
	"errors"
	"log"
	"strings"

	"sshbot/internal/ai"
	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
)

type Plugin struct {
	provider ai.Provider
}

func New(provider ai.Provider) Plugin {
	return Plugin{provider: provider}
}

func (p Plugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name:    "ask",
		Summary: "Ask the configured AI provider",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityAIChat,
		},
	}
}

func (p Plugin) Execute(ctx context.Context, cmdCtx core.CommandContext, envelope core.CommandEnvelope) (core.CommandResult, error) {
	if p.provider == nil {
		return core.CommandResult{
			Status:  "error",
			Message: "ai provider is not configured",
		}, errors.New("ai provider is not configured")
	}

	prompt := strings.TrimSpace(envelope.Args["prompt"])
	if prompt == "" {
		prompt = strings.TrimSpace(envelope.RawText)
		if strings.HasPrefix(strings.ToLower(prompt), "ask ") {
			prompt = strings.TrimSpace(prompt[4:])
		}
	}
	if prompt == "" || strings.EqualFold(prompt, "ask") {
		return core.CommandResult{
			Status:  "error",
			Message: "prompt is required",
		}, errors.New("prompt is required")
	}

	model := strings.TrimSpace(envelope.Args["model"])
	proxySessionID := strings.TrimSpace(envelope.Args["proxy_session_id"])
	log.Printf("[ask] provider=%s model=%q base_prompt_len=%d proxy_session=%q",
		p.provider.Name(), model, len(prompt), proxySessionID)

	response, err := p.provider.Ask(ctx, ai.AskRequest{
		Prompt:         prompt,
		System:         strings.TrimSpace(envelope.Args["system"]),
		Model:          model,
		SessionID:      envelope.Transport + ":" + envelope.Principal.ID,
		UserID:         envelope.Principal.ID,
		ProxySessionID: proxySessionID,
	})
	if err != nil {
		log.Printf("[ask] provider error: %v", err)
		return core.CommandResult{
			Status:  "error",
			Message: "ai provider request failed: " + err.Error(),
		}, err
	}

	return core.CommandResult{
		Status:  "ok",
		Message: "ai answer generated",
		Data: map[string]any{
			"answer":      response.Answer,
			"provider":    response.Provider,
			"model":       response.Model,
			"response_id": response.ResponseID,
		},
	}, nil
}

func Builtins(provider ai.Provider) []core.Plugin {
	return []core.Plugin{New(provider)}
}
