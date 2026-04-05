package handlers

import (
	"net/http"
	"strings"

	"gitcourse/internal/presenter"
	"gitcourse/views"
)

func (a *App) AskPage(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
		Prompt:  strings.TrimSpace(r.URL.Query().Get("prompt")),
		Model:   strings.TrimSpace(r.URL.Query().Get("model")),
		Context: strings.TrimSpace(r.URL.Query().Get("context")),
	})))
}

func (a *App) AskSubmit(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	if err := r.ParseForm(); err != nil {
		render(w, r, http.StatusBadRequest, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Error: "invalid form",
		})))
		return
	}

	mode := strings.TrimSpace(r.FormValue("mode"))
	prompt := strings.TrimSpace(r.FormValue("prompt"))
	model := strings.TrimSpace(r.FormValue("model"))
	contextValue := strings.TrimSpace(r.FormValue("context"))
	if prompt == "" {
		render(w, r, http.StatusBadRequest, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Prompt:  prompt,
			Model:   model,
			Context: contextValue,
			Error:   runtime.Presenter.AskPromptRequiredError(),
		})))
		return
	}

	if mode != "sync" {
		render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Prompt:  prompt,
			Model:   model,
			Context: contextValue,
		})))
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	result, err := runtime.Source.Ask(ctx, prompt, model, contextValue)
	if err != nil {
		render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
			Prompt:  prompt,
			Model:   model,
			Context: contextValue,
			Error:   err.Error(),
		})))
		return
	}

	render(w, r, http.StatusOK, views.AskPage(runtime.Presenter.AskPage(presenter.AskPageState{
		Prompt:  prompt,
		Model:   model,
		Context: contextValue,
		Result:  &result,
	})))
}
