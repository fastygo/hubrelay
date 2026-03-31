package ai

import "context"

type AskStreamChunk struct {
	Delta        string
	FinishReason string
}

type StreamCallback func(AskStreamChunk) error

type StreamingProvider interface {
	Provider
	AskStream(context.Context, AskRequest, StreamCallback) (AskResponse, error)
}
