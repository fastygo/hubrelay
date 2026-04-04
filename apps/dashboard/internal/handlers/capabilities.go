package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Capabilities(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	data, err := runtime.Source.Capabilities(ctx)
	render(w, r, http.StatusOK, views.CapabilitiesPage(runtime.Presenter.CapabilitiesPage(data, err)))
}
