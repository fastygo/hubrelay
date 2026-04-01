package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"hubrelay-dashboard/internal/content"
	"hubrelay-dashboard/internal/middleware"
	"hubrelay-dashboard/internal/presenter"
	"hubrelay-dashboard/internal/source"
)

type App struct {
	DefaultLocale  string
	Auth           *middleware.SessionAuth
	AuthDisabled   bool
	Presenters     map[string]*presenter.Presenter
	LiveSource     source.Source
	FixtureSources map[string]source.Source
}

type requestRuntime struct {
	Locale    string
	Presenter *presenter.Presenter
	Source    source.Source
}

func New(catalogs map[string]content.Catalog, localeOrder []string, liveSource source.Source, fixtureSources map[string]source.Source, auth *middleware.SessionAuth, authDisabled bool) *App {
	presenters := make(map[string]*presenter.Presenter, len(catalogs))
	toggleLocales := append([]string(nil), localeOrder...)
	if len(toggleLocales) > 2 {
		toggleLocales = toggleLocales[:2]
	}
	for locale, catalog := range catalogs {
		presenters[locale] = presenter.New(catalog, presenter.Config{
			AuthEnabled:      !authDisabled,
			DefaultLocale:    content.DefaultLocale,
			AvailableLocales: localeOrder,
			ToggleLocales:    toggleLocales,
			Catalogs:         catalogs,
		})
	}

	return &App{
		DefaultLocale:  content.DefaultLocale,
		Auth:           auth,
		AuthDisabled:   authDisabled,
		Presenters:     presenters,
		LiveSource:     liveSource,
		FixtureSources: fixtureSources,
	}
}

func (a *App) runtimeFor(r *http.Request) requestRuntime {
	locale := a.localeFromRequest(r)

	runtime := requestRuntime{
		Locale:    locale,
		Presenter: a.Presenters[locale],
	}
	if runtime.Presenter == nil {
		runtime.Locale = a.DefaultLocale
		runtime.Presenter = a.Presenters[a.DefaultLocale]
	}

	if a.LiveSource != nil {
		runtime.Source = a.LiveSource
		return runtime
	}

	runtime.Source = a.FixtureSources[runtime.Locale]
	if runtime.Source == nil {
		runtime.Source = a.FixtureSources[a.DefaultLocale]
	}
	return runtime
}

func (a *App) localeFromRequest(r *http.Request) string {
	locale := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang")))
	if _, ok := a.Presenters[locale]; ok {
		return locale
	}
	return a.DefaultLocale
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
