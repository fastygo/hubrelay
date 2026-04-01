package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookieName = "hubrelay_session"
	sessionTTL        = 24 * time.Hour
)

// SessionAuth provides GUI-style form login backed by an in-memory session store.
type SessionAuth struct {
	user     string
	pass     string
	now      func() time.Time
	mu       sync.RWMutex
	sessions map[string]time.Time
}

func NewSessionAuth(user, pass string) *SessionAuth {
	return &SessionAuth{
		user:     user,
		pass:     pass,
		now:      time.Now,
		sessions: make(map[string]time.Time),
	}
}

func (a *SessionAuth) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if a.HasValidSession(r) {
				next.ServeHTTP(w, r)
				return
			}

			if username, password, ok := r.BasicAuth(); ok && a.ValidCredentials(username, password) {
				a.CreateSession(w)
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("WWW-Authenticate", `Basic realm="HubRelay Dashboard"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			http.Redirect(w, r, loginRedirectPath(r), http.StatusSeeOther)
		})
	}
}

func (a *SessionAuth) HasValidSession(r *http.Request) bool {
	if a == nil {
		return false
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return false
	}

	a.mu.RLock()
	expiresAt, ok := a.sessions[cookie.Value]
	a.mu.RUnlock()
	if !ok {
		return false
	}

	if expiresAt.Before(a.now()) {
		a.mu.Lock()
		delete(a.sessions, cookie.Value)
		a.mu.Unlock()
		return false
	}

	return true
}

func (a *SessionAuth) ValidCredentials(user, pass string) bool {
	if a == nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(user), []byte(a.user)) == 1 &&
		subtle.ConstantTimeCompare([]byte(pass), []byte(a.pass)) == 1
}

func (a *SessionAuth) CreateSession(w http.ResponseWriter) {
	if a == nil {
		return
	}

	token := make([]byte, 32)
	_, _ = rand.Read(token)

	sessionID := hex.EncodeToString(token)
	expiresAt := a.now().Add(sessionTTL)

	a.mu.Lock()
	a.sessions[sessionID] = expiresAt
	a.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func (a *SessionAuth) ClearSession(w http.ResponseWriter, r *http.Request) {
	if a != nil && r != nil {
		if cookie, err := r.Cookie(sessionCookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
			a.mu.Lock()
			delete(a.sessions, cookie.Value)
			a.mu.Unlock()
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func isPublicPath(path string) bool {
	return path == "/login" || strings.HasPrefix(path, "/static/")
}

func loginRedirectPath(r *http.Request) string {
	target := &url.URL{Path: "/login"}
	lang := strings.TrimSpace(r.URL.Query().Get("lang"))
	if lang == "" {
		return target.String()
	}

	query := target.Query()
	query.Set("lang", lang)
	target.RawQuery = query.Encode()
	return target.String()
}
