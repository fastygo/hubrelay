package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Audit(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	entries, err := runtime.Source.Audit(ctx, parseLimit(r, 25))
	render(w, r, http.StatusOK, views.AuditPage(runtime.Presenter.AuditPage(entries, err)))
}
