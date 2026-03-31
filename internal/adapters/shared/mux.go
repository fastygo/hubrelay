package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
)

const maxRequestBodyBytes = 1 << 20 // 1 MiB

type Handler struct {
	transportName string
	service       *core.Service
	proxy         *proxymgr.Manager
}

type commandRequest struct {
	PrincipalID string            `json:"principal_id"`
	Roles       []string          `json:"roles"`
	Command     string            `json:"command"`
	Args        map[string]string `json:"args,omitempty"`
}

type proxyCreateRequest struct {
	PrincipalID string `json:"principal_id"`
	Proxies     string `json:"proxies"`
}

type proxySelectRequest struct {
	Proxy string `json:"proxy"`
}

func NewMux(service *core.Service, proxy *proxymgr.Manager, transportName string) *http.ServeMux {
	handler := &Handler{
		transportName: strings.TrimSpace(transportName),
		service:       service,
		proxy:         proxy,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handler.handleRoot)
	mux.HandleFunc("GET /healthz", handler.handleHealth)
	mux.HandleFunc("POST /api/command", handler.handleCommand)
	mux.HandleFunc("POST /api/command/stream", handler.handleCommandStream)
	mux.HandleFunc("POST /api/proxy/session", handler.handleProxyCollection)
	mux.HandleFunc("GET /api/proxy/session/", handler.handleProxySession)
	mux.HandleFunc("POST /api/proxy/session/", handler.handleProxySession)
	return mux
}

func (h *Handler) handleRoot(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"service": "hubrelay",
		"profile": h.service.Profile().ID,
		"status":  "ok",
	})
}

func (h *Handler) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"adapter": h.transportName,
		"profile": h.service.Profile().ID,
	})
}

func (h *Handler) handleCommand(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	payload, ok := decodeCommandRequest(writer, request)
	if !ok {
		return
	}

	result, err := h.service.Execute(request.Context(), h.commandEnvelope(*payload))
	if err != nil && result.Message == "" {
		result.Message = err.Error()
	}

	statusCode := http.StatusOK
	if result.Status == "error" {
		statusCode = http.StatusBadRequest
	}
	writeJSON(writer, statusCode, result)
}

func (h *Handler) handleCommandStream(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	payload, ok := decodeCommandRequest(writer, request)
	if !ok {
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	streamWriter, err := newSSEStreamWriter(request.Context(), writer)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := h.service.ExecuteStream(request.Context(), h.commandEnvelope(*payload), streamWriter); err != nil {
		log.Printf("command stream failed: %v", err)
	}
}

func decodeCommandRequest(writer http.ResponseWriter, request *http.Request) (*commandRequest, bool) {
	var payload commandRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": "invalid JSON payload",
		})
		return nil, false
	}
	return &payload, true
}

func (h *Handler) commandEnvelope(payload commandRequest) core.CommandEnvelope {
	name, args := parseCommand(payload.Command)
	if len(payload.Args) > 0 {
		if args == nil {
			args = make(map[string]string, len(payload.Args))
		}
		for key, value := range payload.Args {
			args[key] = value
		}
	}
	return core.CommandEnvelope{
		ID:        fmt.Sprintf("%s-%d", h.transportName, time.Now().UTC().UnixNano()),
		Transport: h.transportName,
		Name:      name,
		Args:      args,
		RawText:   payload.Command,
		Principal: core.Principal{
			ID:        payload.PrincipalID,
			Display:   payload.PrincipalID,
			Transport: h.transportName,
			Roles:     payload.Roles,
		},
		RequestedAt: time.Now().UTC(),
	}
}

func (h *Handler) handleProxyCollection(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.NotFound(writer, request)
		return
	}
	if !h.proxyEnabled() {
		http.NotFound(writer, request)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	var payload proxyCreateRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": "invalid proxy session payload",
		})
		return
	}

	session, err := h.proxy.CreateSession(payload.PrincipalID, []string{payload.Proxies})
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusCreated, map[string]any{
		"status":  "ok",
		"message": "proxy session created",
		"session": session,
		"lease":   proxyLeaseFromSession(session),
	})
}

func (h *Handler) handleProxySession(writer http.ResponseWriter, request *http.Request) {
	if !h.proxyEnabled() {
		http.NotFound(writer, request)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)

	sessionID, action, ok := parseProxyPath(request.URL.Path)
	if !ok {
		http.NotFound(writer, request)
		return
	}

	var (
		session proxymgr.Session
		err     error
	)

	switch {
	case request.Method == http.MethodGet && action == "":
		session, err = h.proxy.GetSession(sessionID)
	case request.Method == http.MethodPost && action == "check":
		session, err = h.proxy.CheckSession(request.Context(), sessionID)
	case request.Method == http.MethodPost && action == "select":
		var payload proxySelectRequest
		if decodeErr := json.NewDecoder(request.Body).Decode(&payload); decodeErr != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]any{
				"status":  "error",
				"message": "invalid proxy selection payload",
			})
			return
		}
		session, err = h.proxy.SelectProxy(sessionID, payload.Proxy)
	default:
		http.NotFound(writer, request)
		return
	}

	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, proxymgr.ErrSessionNotFound) {
			statusCode = http.StatusNotFound
		}
		writeJSON(writer, statusCode, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "proxy session loaded",
		"session": session,
		"lease":   proxyLeaseFromSession(session),
	})
}

func parseCommand(raw string) (string, map[string]string) {
	trimmed := strings.TrimSpace(raw)
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", nil
	}

	args := make(map[string]string)
	if strings.EqualFold(fields[0], "ask") && len(fields) > 1 && !strings.Contains(fields[1], "=") {
		args["prompt"] = strings.TrimSpace(trimmed[len(fields[0]):])
		return strings.ToLower(fields[0]), args
	}
	for _, field := range fields[1:] {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 {
			args[parts[0]] = parts[1]
		}
	}
	return strings.ToLower(fields[0]), args
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}

func (h *Handler) proxyEnabled() bool {
	return h.proxy != nil && h.service.Profile().ProxySession.Enabled
}

func parseProxyPath(path string) (string, string, bool) {
	trimmed := strings.TrimPrefix(path, "/api/proxy/session/")
	if trimmed == path || trimmed == "" {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", true
	}
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}

func proxyLeaseFromSession(session proxymgr.Session) *proxymgr.Lease {
	if session.SelectedProxy == "" {
		return nil
	}
	for _, candidate := range session.Candidates {
		if candidate.Address != session.SelectedProxy || candidate.Status != proxymgr.StatusHealthy {
			continue
		}
		return &proxymgr.Lease{
			SessionID: session.ID,
			Address:   candidate.Address,
			LatencyMS: candidate.LatencyMS,
			UpdatedAt: session.UpdatedAt,
		}
	}
	return nil
}
