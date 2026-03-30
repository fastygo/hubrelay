package main

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	config, err := loadConfig(func(key string) string {
		switch key {
		case "SMOKE_AI_API_KEY":
			return "test-key"
		case "SMOKE_AI_MODEL":
			return "gpt-oss-120b"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if config.Provider != "openai" {
		t.Fatalf("expected default provider openai, got %q", config.Provider)
	}
	if config.APIMode != "chat_completions" {
		t.Fatalf("expected default api mode chat_completions, got %q", config.APIMode)
	}
	if config.TimeoutSec != 60 {
		t.Fatalf("expected default timeout 60, got %d", config.TimeoutSec)
	}
}

func TestLoadConfigRequiresKeyAndModel(t *testing.T) {
	if _, err := loadConfig(func(string) string { return "" }); err == nil {
		t.Fatalf("expected missing env validation error")
	}
}

func TestLoadConfigRejectsBadTimeout(t *testing.T) {
	_, err := loadConfig(func(key string) string {
		switch key {
		case "SMOKE_AI_API_KEY":
			return "test-key"
		case "SMOKE_AI_MODEL":
			return "gpt-oss-120b"
		case "SMOKE_TIMEOUT_SEC":
			return "bad"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatalf("expected timeout validation error")
	}
}
