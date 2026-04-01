package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Capabilities(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	response, err := a.Relay.Capabilities(ctx)
	if err != nil {
		render(w, r, http.StatusOK, views.CapabilitiesPage(views.CapabilitiesPageData{
			Error: err.Error(),
		}))
		return
	}

	render(w, r, http.StatusOK, views.CapabilitiesPage(views.CapabilitiesPageData{
		Capabilities: response,
	}))
}
