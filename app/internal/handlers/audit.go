package handlers

import (
	"net/http"

	"hubrelay-dashboard/views"
)

func (a *App) Audit(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	entries, err := a.Relay.Audit(ctx, parseLimit(r, 25))
	if err != nil {
		render(w, r, http.StatusOK, views.AuditPage(views.AuditPageData{
			Error: err.Error(),
		}))
		return
	}

	items := make([]views.AuditEntryView, 0, len(entries))
	for _, entry := range entries {
		items = append(items, views.AuditEntryView{
			Command:   entry.Command,
			Principal: entry.Principal,
			Transport: entry.Transport,
			Outcome:   entry.Outcome,
			Message:   entry.Message,
			At:        entry.At,
		})
	}

	render(w, r, http.StatusOK, views.AuditPage(views.AuditPageData{
		Entries: items,
	}))
}
