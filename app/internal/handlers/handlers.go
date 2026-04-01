package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"hubrelay-dashboard/internal/relay"
)

type App struct {
	Relay *relay.Client
}

func New(client *relay.Client) *App {
	return &App{Relay: client}
}

func render(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	var buf bytes.Buffer
	if err := component.Render(r.Context(), &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 30*time.Second)
}

func parseLimit(r *http.Request, fallback int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return fallback
	}
	return limit
}
