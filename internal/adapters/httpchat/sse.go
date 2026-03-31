package httpchat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"sshbot/internal/core"
)

type sseStreamWriter struct {
	ctx     context.Context
	writer  http.ResponseWriter
	flusher http.Flusher

	mu     sync.Mutex
	result core.CommandResult
}

func newSSEStreamWriter(ctx context.Context, writer http.ResponseWriter) (*sseStreamWriter, error) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming is not supported by this response writer")
	}
	return &sseStreamWriter{
		ctx:     ctx,
		writer:  writer,
		flusher: flusher,
	}, nil
}

func (w *sseStreamWriter) WriteChunk(chunk core.StreamChunk) error {
	return w.writeEvent("chunk", chunk)
}

func (w *sseStreamWriter) SetResult(result core.CommandResult) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.result = cloneResult(result)
}

func (w *sseStreamWriter) Flush() error {
	result := w.Result()
	if result.Status == "error" {
		return w.writeEvent("error", result)
	}
	return w.writeEvent("done", result)
}

func (w *sseStreamWriter) Result() core.CommandResult {
	w.mu.Lock()
	defer w.mu.Unlock()
	return cloneResult(w.result)
}

func (w *sseStreamWriter) writeEvent(name string, payload any) error {
	select {
	case <-w.ctx.Done():
		return w.ctx.Err()
	default:
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.writer, "event: %s\ndata: %s\n\n", name, body); err != nil {
		return err
	}
	w.flusher.Flush()
	return nil
}

func cloneResult(result core.CommandResult) core.CommandResult {
	cloned := core.CommandResult{
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
