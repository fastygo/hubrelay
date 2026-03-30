package ai

import "testing"

func TestNormalizeAPIModeDefaultsToChatCompletions(t *testing.T) {
	if got := normalizeAPIMode(""); got != "chat_completions" {
		t.Fatalf("expected chat_completions default, got %q", got)
	}
	if got := normalizeAPIMode("responses"); got != "responses" {
		t.Fatalf("expected responses mode, got %q", got)
	}
}
