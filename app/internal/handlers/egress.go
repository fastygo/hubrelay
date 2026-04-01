package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Egress(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	response, err := a.Relay.EgressStatus(ctx)
	if err != nil {
		render(w, r, http.StatusOK, views.EgressPage(views.EgressPageData{
			Error: err.Error(),
		}))
		return
	}

	render(w, r, http.StatusOK, views.EgressPage(views.EgressPageData{
		Gateways: response.Gateways,
	}))
}
