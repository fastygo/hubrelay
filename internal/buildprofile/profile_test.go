package buildprofile

import (
	"os"
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
	previousPrivateEgressRequired := currentPrivateEgressRequired
	previousPrivateEgressInterface := currentPrivateEgressInterface
	previousPrivateEgressTestHost := currentPrivateEgressTestHost
	previousPrivateEgressFailClosed := currentPrivateEgressFailClosed
	previousOnce := envOnce
	t.Cleanup(func() {
		currentAIAPIKey = previousAIKey
		currentProxyForce = previousProxyForce
		currentAIAPIMode = previousAIAPIMode
		currentOpenAIEnabled = previousOpenAIEnabled
		currentPrivateEgressRequired = previousPrivateEgressRequired
		currentPrivateEgressInterface = previousPrivateEgressInterface
		currentPrivateEgressTestHost = previousPrivateEgressTestHost
		currentPrivateEgressFailClosed = previousPrivateEgressFailClosed
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
	if profile.PrivateEgress.Required {
		t.Fatalf("expected private egress to be disabled by default")
	}
}

func TestCurrentReadsPrivateEgressEnvOverrides(t *testing.T) {
	previousRequired := currentPrivateEgressRequired
	previousInterface := currentPrivateEgressInterface
	previousTestHost := currentPrivateEgressTestHost
	previousFailClosed := currentPrivateEgressFailClosed
	previousOnce := envOnce
	t.Cleanup(func() {
		currentPrivateEgressRequired = previousRequired
		currentPrivateEgressInterface = previousInterface
		currentPrivateEgressTestHost = previousTestHost
		currentPrivateEgressFailClosed = previousFailClosed
		envOnce = previousOnce
		_ = os.Unsetenv("INPUT_PRIVATE_EGRESS_REQUIRED")
		_ = os.Unsetenv("INPUT_PRIVATE_EGRESS_INTERFACE")
		_ = os.Unsetenv("INPUT_PRIVATE_EGRESS_TEST_HOST")
		_ = os.Unsetenv("INPUT_PRIVATE_EGRESS_FAIL_CLOSED")
	})
	envOnce = sync.Once{}
	currentPrivateEgressRequired = "false"
	currentPrivateEgressInterface = ""
	currentPrivateEgressTestHost = ""
	currentPrivateEgressFailClosed = "false"
	_ = os.Setenv("INPUT_PRIVATE_EGRESS_REQUIRED", "true")
	_ = os.Setenv("INPUT_PRIVATE_EGRESS_INTERFACE", "wg0")
	_ = os.Setenv("INPUT_PRIVATE_EGRESS_TEST_HOST", "10.88.0.1")
	_ = os.Setenv("INPUT_PRIVATE_EGRESS_FAIL_CLOSED", "true")

	profile := Current()
	if !profile.PrivateEgress.Required {
		t.Fatalf("expected private egress to be enabled")
	}
	if profile.PrivateEgress.Interface != "wg0" {
		t.Fatalf("expected private egress interface wg0, got %q", profile.PrivateEgress.Interface)
	}
	if profile.PrivateEgress.TestHost != "10.88.0.1" {
		t.Fatalf("expected private egress test host to flow through, got %q", profile.PrivateEgress.TestHost)
	}
	if !profile.PrivateEgress.FailClosed {
		t.Fatalf("expected fail-closed private egress")
	}
}
