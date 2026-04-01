package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hubrelay-dashboard/internal/content"
	"hubrelay-dashboard/internal/middleware"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	locales, err := content.AvailableLocales()
	if err != nil {
		t.Fatalf("AvailableLocales() error = %v", err)
	}

	catalogs := make(map[string]content.Catalog, len(locales))
	for _, locale := range locales {
		catalogs[locale], err = content.Load(locale)
		if err != nil {
			t.Fatalf("Load(%s) error = %v", locale, err)
		}
	}

	return New(catalogs, locales, nil, nil, middleware.NewSessionAuth("admin", "secret"), false)
}

func TestLoginPageRendersLocalizedCopy(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/login?lang=ru", nil)
	recorder := httptest.NewRecorder()
	app.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected OK status, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Вход") {
		t.Fatalf("expected russian login title, got:\n%s", body)
	}
	if !strings.Contains(body, `action="/login?lang=ru"`) {
		t.Fatalf("expected localized form action, got:\n%s", body)
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/login?lang=es", strings.NewReader("username=admin&password=wrong"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	app.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected inline form error with OK status, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Nombre de usuario o contrase") {
		t.Fatalf("expected spanish invalid credentials error, got:\n%s", body)
	}
	if !strings.Contains(body, `value="admin"`) {
		t.Fatalf("expected username to stay in the form, got:\n%s", body)
	}
}

func TestLoginCreatesSessionAndRedirectsHome(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/login?lang=ru", strings.NewReader("username=admin&password=secret"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	app.Login(recorder, request)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status, got %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/?lang=ru" {
		t.Fatalf("expected localized home redirect, got %q", location)
	}
	if !strings.Contains(recorder.Header().Get("Set-Cookie"), "hubrelay_session=") {
		t.Fatalf("expected session cookie, got %q", recorder.Header().Get("Set-Cookie"))
	}
}

func TestLogoutClearsSessionAndRedirectsToLogin(t *testing.T) {
	app := newTestApp(t)

	loginRequest := httptest.NewRequest(http.MethodPost, "/login?lang=ru", strings.NewReader("username=admin&password=secret"))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRecorder := httptest.NewRecorder()
	app.Login(loginRecorder, loginRequest)

	logoutRequest := httptest.NewRequest(http.MethodPost, "/logout?lang=ru", nil)
	for _, cookie := range loginRecorder.Result().Cookies() {
		logoutRequest.AddCookie(cookie)
	}
	logoutRecorder := httptest.NewRecorder()
	app.Logout(logoutRecorder, logoutRequest)

	if logoutRecorder.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status, got %d", logoutRecorder.Code)
	}
	if location := logoutRecorder.Header().Get("Location"); location != "/login?lang=ru" {
		t.Fatalf("expected localized login redirect, got %q", location)
	}
	if !strings.Contains(logoutRecorder.Header().Get("Set-Cookie"), "Max-Age=0") && !strings.Contains(logoutRecorder.Header().Get("Set-Cookie"), "Max-Age=-1") {
		t.Fatalf("expected cookie expiration, got %q", logoutRecorder.Header().Get("Set-Cookie"))
	}
}
