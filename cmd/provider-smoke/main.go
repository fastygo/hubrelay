package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"sshbot/internal/ai"
	"sshbot/internal/buildprofile"
	"sshbot/internal/outbound"
)

type smokeConfig struct {
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	APIMode    string
	Prompt     string
	System     string
	UserID     string
	TimeoutSec int
}

type smokeOutput struct {
	Provider   string `json:"provider"`
	BaseURL    string `json:"base_url,omitempty"`
	Model      string `json:"model"`
	APIMode    string `json:"api_mode"`
	ResponseID string `json:"response_id,omitempty"`
	Answer     string `json:"answer,omitempty"`
}

func main() {
	config, err := loadConfig(os.Getenv)
	if err != nil {
		fail(err)
	}

	provider, err := ai.NewOpenAICompatibleProvider(
		config.Provider,
		config.APIKey,
		config.BaseURL,
		config.Model,
		config.APIMode,
		outbound.Policy{RequireProxy: false},
		buildprofile.PrivateEgressConfig{},
		nil,
		nil,
	)
	if err != nil {
		fail(fmt.Errorf("configure provider: %w", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSec)*time.Second)
	defer cancel()

	response, err := provider.Ask(ctx, ai.AskRequest{
		Prompt:    config.Prompt,
		System:    config.System,
		Model:     config.Model,
		SessionID: "provider-smoke",
		UserID:    config.UserID,
	})
	if err != nil {
		fail(fmt.Errorf("provider request failed: %w", err))
	}

	writeJSON(smokeOutput{
		Provider:   response.Provider,
		BaseURL:    config.BaseURL,
		Model:      response.Model,
		APIMode:    config.APIMode,
		ResponseID: response.ResponseID,
		Answer:     response.Answer,
	})
}

func loadConfig(getenv func(string) string) (smokeConfig, error) {
	timeoutSec := 60
	if raw := strings.TrimSpace(getenv("SMOKE_TIMEOUT_SEC")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return smokeConfig{}, errors.New("SMOKE_TIMEOUT_SEC must be a positive integer")
		}
		timeoutSec = value
	}

	config := smokeConfig{
		Provider:   defaultValue(getenv("SMOKE_AI_PROVIDER"), "openai"),
		BaseURL:    strings.TrimSpace(getenv("SMOKE_AI_BASE_URL")),
		APIKey:     strings.TrimSpace(getenv("SMOKE_AI_API_KEY")),
		Model:      strings.TrimSpace(getenv("SMOKE_AI_MODEL")),
		APIMode:    defaultValue(getenv("SMOKE_AI_API_MODE"), "chat_completions"),
		Prompt:     defaultValue(getenv("SMOKE_PROMPT"), "Say hello from provider-smoke."),
		System:     strings.TrimSpace(getenv("SMOKE_SYSTEM")),
		UserID:     defaultValue(getenv("SMOKE_USER_ID"), "provider-smoke"),
		TimeoutSec: timeoutSec,
	}

	if config.APIKey == "" {
		return smokeConfig{}, errors.New("SMOKE_AI_API_KEY is required")
	}
	if config.Model == "" {
		return smokeConfig{}, errors.New("SMOKE_AI_MODEL is required")
	}
	return config, nil
}

func defaultValue(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func writeJSON(payload any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "provider-smoke error: %v\n", err)
	os.Exit(1)
}
