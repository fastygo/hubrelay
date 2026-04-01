package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	data, err := runtime.Source.Health(ctx)
	render(w, r, http.StatusOK, views.HealthPage(runtime.Presenter.HealthPage(data, err)))
}
