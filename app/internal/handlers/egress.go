package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Egress(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	data, err := runtime.Source.Egress(ctx)
	render(w, r, http.StatusOK, views.EgressPage(runtime.Presenter.EgressPage(data, err)))
}
