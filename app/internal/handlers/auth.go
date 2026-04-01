package handlers

import (
	"net/http"
	"strings"

	"hubrelay-dashboard/internal/presenter"
	"hubrelay-dashboard/views"
)

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	if a.AuthDisabled || a.Auth == nil {
		http.Redirect(w, r, runtime.Presenter.HomePath(), http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if a.Auth.HasValidSession(r) {
			http.Redirect(w, r, runtime.Presenter.HomePath(), http.StatusSeeOther)
			return
		}
		if username, password, ok := r.BasicAuth(); ok && a.Auth.ValidCredentials(username, password) {
			a.Auth.CreateSession(w)
			http.Redirect(w, r, runtime.Presenter.HomePath(), http.StatusSeeOther)
			return
		}
		render(w, r, http.StatusOK, views.LoginPage(runtime.Presenter.LoginPage(presenter.LoginPageState{})))
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			render(w, r, http.StatusOK, views.LoginPage(runtime.Presenter.LoginPage(presenter.LoginPageState{
				Error: runtime.Presenter.LoginInvalidFormError(),
			})))
			return
		}
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		if username == "" || strings.TrimSpace(password) == "" {
			render(w, r, http.StatusOK, views.LoginPage(runtime.Presenter.LoginPage(presenter.LoginPageState{
				Username: username,
				Error:    runtime.Presenter.LoginInvalidFormError(),
			})))
			return
		}
		if !a.Auth.ValidCredentials(username, password) {
			render(w, r, http.StatusOK, views.LoginPage(runtime.Presenter.LoginPage(presenter.LoginPageState{
				Username: username,
				Error:    runtime.Presenter.LoginInvalidCredentialsError(),
			})))
			return
		}
		a.Auth.CreateSession(w)
		http.Redirect(w, r, runtime.Presenter.HomePath(), http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.Auth != nil {
		a.Auth.ClearSession(w, r)
	}
	http.Redirect(w, r, runtime.Presenter.LoginPath(), http.StatusSeeOther)
}
