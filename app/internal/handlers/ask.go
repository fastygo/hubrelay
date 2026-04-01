package handlers

import (
	"net/http"
	"strings"

	"hubrelay-dashboard/views"
)

func (a *App) AskPage(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, views.AskPage(views.AskPageData{
		Prompt: strings.TrimSpace(r.URL.Query().Get("prompt")),
		Model:  strings.TrimSpace(r.URL.Query().Get("model")),
	}))
}

func (a *App) AskSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, r, http.StatusBadRequest, views.AskPage(views.AskPageData{
			Error: "invalid form submission",
		}))
		return
	}

	mode := strings.TrimSpace(r.FormValue("mode"))
	if mode != "sync" {
		http.Redirect(w, r, "/ask", http.StatusSeeOther)
		return
	}

	prompt := strings.TrimSpace(r.FormValue("prompt"))
	model := strings.TrimSpace(r.FormValue("model"))
	if prompt == "" {
		render(w, r, http.StatusBadRequest, views.AskPage(views.AskPageData{
			Prompt: prompt,
			Model:  model,
			Error:  "prompt is required",
		}))
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	result, err := a.Relay.Ask(ctx, prompt, model)
	if err != nil {
		render(w, r, http.StatusOK, views.AskPage(views.AskPageData{
			Prompt: prompt,
			Model:  model,
			Error:  err.Error(),
		}))
		return
	}

	render(w, r, http.StatusOK, views.AskPage(views.AskPageData{
		Prompt: prompt,
		Model:  model,
		Result: &result,
	}))
}
