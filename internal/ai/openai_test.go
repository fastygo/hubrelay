package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sshbot/internal/buildprofile"
	"sshbot/internal/outbound"
)

func TestNormalizeAPIModeDefaultsToChatCompletions(t *testing.T) {
	if got := normalizeAPIMode(""); got != "chat_completions" {
		t.Fatalf("expected chat_completions default, got %q", got)
	}
	if got := normalizeAPIMode("responses"); got != "responses" {
		t.Fatalf("expected responses mode, got %q", got)
	}
}

type fakePrivateEgressChecker struct {
	err error
}

func (f fakePrivateEgressChecker) Check(context.Context, string, string) error {
	return f.err
}

func TestAskFailsClosedWhenPrivateEgressCheckFails(t *testing.T) {
	provider, err := NewOpenAICompatibleProvider(
		"openai",
		"test-key",
		"https://api.example.com/v1",
		"gpt-test",
		"chat_completions",
		outbound.Policy{},
		buildprofile.PrivateEgressConfig{
			Required:   true,
			Interface:  "wg0",
			TestHost:   "10.88.0.1",
			FailClosed: true,
		},
		fakePrivateEgressChecker{err: errors.New("private route unavailable")},
		nil,
	)
	if err != nil {
		t.Fatalf("NewOpenAICompatibleProvider() error = %v", err)
	}

	_, err = provider.Ask(context.Background(), AskRequest{Prompt: "hello"})
	if err == nil {
		t.Fatal("expected private egress failure")
	}
	if !strings.Contains(err.Error(), "private route unavailable") {
		t.Fatalf("unexpected error %v", err)
	}
}
