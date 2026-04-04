package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionAuthMiddlewareRedirectsHTMLRequestsToLogin(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	protected := auth.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/ask", nil)
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}
}

func TestSessionAuthMiddlewarePreservesLocaleOnRedirect(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	protected := auth.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/capabilities?lang=ru", nil)
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	if location := recorder.Header().Get("Location"); location != "/login?lang=ru" {
		t.Fatalf("expected redirect to preserve locale, got %q", location)
	}
}

func TestSessionAuthMiddlewareAllowsStaticAssets(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	protected := auth.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil)
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected static asset request to bypass auth, got %d", recorder.Code)
	}
}

func TestSessionAuthMiddlewareAllowsBasicAuthAndCreatesSession(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	protected := auth.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/ask", nil)
	request.SetBasicAuth("admin", "secret")
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected OK status, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Set-Cookie"), sessionCookieName+"=") {
		t.Fatalf("expected session cookie to be set, got %q", recorder.Header().Get("Set-Cookie"))
	}
}

func TestSessionAuthMiddlewareRejectsAPIRequestsWithoutCredentials(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	protected := auth.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()
	protected.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", recorder.Code)
	}
	if header := recorder.Header().Get("WWW-Authenticate"); header == "" {
		t.Fatal("expected WWW-Authenticate header")
	}
}

func TestSessionAuthClearSessionExpiresCookie(t *testing.T) {
	auth := NewSessionAuth("admin", "secret")
	loginRecorder := httptest.NewRecorder()
	auth.CreateSession(loginRecorder)

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	for _, cookie := range loginRecorder.Result().Cookies() {
		request.AddCookie(cookie)
	}

	logoutRecorder := httptest.NewRecorder()
	auth.ClearSession(logoutRecorder, request)

	if cookie := logoutRecorder.Header().Get("Set-Cookie"); !strings.Contains(cookie, sessionCookieName+"=") {
		t.Fatalf("expected session cookie to be cleared, got %q", cookie)
	}
	if auth.HasValidSession(request) {
		t.Fatal("expected session to be removed from memory")
	}
}
