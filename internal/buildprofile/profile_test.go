package buildprofile

import (
	"sync"
	"testing"
)

func TestNormalizeAIAPIModeDefaultsToChatCompletions(t *testing.T) {
	if got := normalizeAIAPIMode(""); got != "chat_completions" {
		t.Fatalf("expected chat_completions default, got %q", got)
	}
	if got := normalizeAIAPIMode("chat/completions"); got != "chat_completions" {
		t.Fatalf("expected slash alias to normalize, got %q", got)
	}
	if got := normalizeAIAPIMode("responses"); got != "responses" {
		t.Fatalf("expected responses mode, got %q", got)
	}
	if got := normalizeAIAPIMode("unexpected"); got != "chat_completions" {
		t.Fatalf("expected fallback to chat_completions, got %q", got)
	}
}

func TestCurrentDefaultsForceProxyAndChatCompletions(t *testing.T) {
	previousAIKey := currentAIAPIKey
	previousProxyForce := currentProxyForce
	previousAIAPIMode := currentAIAPIMode
	previousOpenAIEnabled := currentOpenAIEnabled
	previousOnce := envOnce
	t.Cleanup(func() {
		currentAIAPIKey = previousAIKey
		currentProxyForce = previousProxyForce
		currentAIAPIMode = previousAIAPIMode
		currentOpenAIEnabled = previousOpenAIEnabled
		envOnce = previousOnce
	})
	envOnce = sync.Once{}

	currentAIAPIKey = "test-key"
	currentProxyForce = "true"
	currentAIAPIMode = ""
	currentOpenAIEnabled = "true"

	profile := Current()
	if !profile.ProxySession.Force {
		t.Fatalf("expected proxy force to be enabled by default")
	}
	if profile.OpenAI.APIMode != "chat_completions" {
		t.Fatalf("expected chat_completions default, got %q", profile.OpenAI.APIMode)
	}
}
