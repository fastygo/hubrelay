package core

import (
	"context"
	"strings"
	"sync"
)

type StreamChunk struct {
	Delta    string         `json:"delta"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type StreamWriter interface {
	WriteChunk(StreamChunk) error
	SetResult(CommandResult)
	Flush() error
}

type StreamingPlugin interface {
	Plugin
	ExecuteStream(context.Context, CommandContext, CommandEnvelope, StreamWriter) error
}

type BufferedStreamWriter struct {
	mu     sync.Mutex
	chunks []StreamChunk
	result CommandResult
}

func (w *BufferedStreamWriter) WriteChunk(chunk StreamChunk) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.chunks = append(w.chunks, chunk)
	return nil
}

func (w *BufferedStreamWriter) SetResult(result CommandResult) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.result = cloneCommandResult(result)
}

func (w *BufferedStreamWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.result.Status == "" {
		w.result.Status = "ok"
	}

	answer := w.combinedAnswerLocked()
	if answer != "" {
		if w.result.Data == nil {
			w.result.Data = map[string]any{}
		}
		if _, exists := w.result.Data["answer"]; !exists {
			w.result.Data["answer"] = answer
		}
	}

	return nil
}

func (w *BufferedStreamWriter) Result() CommandResult {
	w.mu.Lock()
	defer w.mu.Unlock()
	return cloneCommandResult(w.result)
}

func (w *BufferedStreamWriter) combinedAnswerLocked() string {
	var builder strings.Builder
	for _, chunk := range w.chunks {
		builder.WriteString(chunk.Delta)
	}
	return builder.String()
}

func cloneCommandResult(result CommandResult) CommandResult {
	cloned := CommandResult{
		Status:          result.Status,
		Message:         result.Message,
		RequiresConfirm: result.RequiresConfirm,
	}
	if len(result.Data) > 0 {
		cloned.Data = make(map[string]any, len(result.Data))
		for key, value := range result.Data {
			cloned.Data[key] = value
		}
	}
	return cloned
}
