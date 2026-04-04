package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (a *App) AskStream(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	prompt := strings.TrimSpace(r.URL.Query().Get("prompt"))
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	if prompt == "" {
		writeSSEHeaders(w)
		writeSSEEvent(w, "error", map[string]any{
			"status":  "error",
			"message": runtime.Presenter.AskPromptRequiredError(),
		})
		flusher.Flush()
		return
	}

	ctx := r.Context()
	stream, err := runtime.Source.AskStream(ctx, prompt, model)
	if err != nil {
		writeSSEHeaders(w)
		writeSSEEvent(w, "error", map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		flusher.Flush()
		return
	}
	defer stream.Close()

	writeSSEHeaders(w)
	flusher.Flush()

	for stream.Next() {
		writeSSEEvent(w, "chunk", stream.Chunk())
		flusher.Flush()
	}

	result, resultErr := stream.Result()
	if resultErr != nil {
		writeSSEEvent(w, "error", map[string]any{
			"status":  "error",
			"message": resultErr.Error(),
		})
		flusher.Flush()
		return
	}

	writeSSEEvent(w, "done", result)
	flusher.Flush()
}

func writeSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func writeSSEEvent(w http.ResponseWriter, event string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		body = []byte(`{"status":"error","message":"failed to encode event"}`)
	}
	_, _ = w.Write([]byte("event: " + event + "\n"))
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(body)
	_, _ = w.Write([]byte("\n\n"))
}
