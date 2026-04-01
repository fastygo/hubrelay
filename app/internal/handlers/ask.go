package handlers

import (
	"net/http"
	"strings"

	"hubrelay-dashboard/internal/presenter"
	"hubrelay-dashboard/views"
)

func (a *App) AskPage(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
		Prompt: strings.TrimSpace(r.URL.Query().Get("prompt")),
		Model:  strings.TrimSpace(r.URL.Query().Get("model")),
	})))
}

func (a *App) AskSubmit(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	if err := r.ParseForm(); err != nil {
		render(w, r, http.StatusBadRequest, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Error: runtime.Presenter.AskInvalidFormError(),
		})))
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
		render(w, r, http.StatusBadRequest, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Prompt: prompt,
			Model:  model,
			Error:  runtime.Presenter.AskPromptRequiredError(),
		})))
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	result, err := runtime.Source.Ask(ctx, prompt, model)
	if err != nil {
		render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Prompt: prompt,
			Model:  model,
			Error:  err.Error(),
		})))
		return
	}

	render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
		Prompt: prompt,
		Model:  model,
		Result: &result,
	})))
}
