package ai

import "context"

type AskRequest struct {
	Prompt         string
	System         string
	Model          string
	SessionID      string
	UserID         string
	ProxySessionID string
}

type AskResponse struct {
	Provider   string
	Model      string
	Answer     string
	ResponseID string
}

type Provider interface {
	Name() string
	Ask(context.Context, AskRequest) (AskResponse, error)
}
