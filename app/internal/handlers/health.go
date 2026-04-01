package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	discovery, err := a.Relay.Discover(ctx)
	if err != nil {
		render(w, r, http.StatusOK, views.HealthPage(views.HealthPageData{
			Error: err.Error(),
		}))
		return
	}

	health, err := a.Relay.Health(ctx)
	if err != nil {
		render(w, r, http.StatusOK, views.HealthPage(views.HealthPageData{
			Discovery: discovery,
			Error:     err.Error(),
		}))
		return
	}

	render(w, r, http.StatusOK, views.HealthPage(views.HealthPageData{
		Discovery: discovery,
		Health:    health,
	}))
}
