package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"gitcourse/internal/course"
	"gitcourse/internal/presenter"
	"gitcourse/views"
)

func (a *App) CatalogPage(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	courses, err := runtime.Source.Courses(ctx)
	if err != nil {
		render(w, r, http.StatusOK, views.CatalogPage(runtime.Presenter.CatalogPage(presenter.CatalogPageState{
			Error: err.Error(),
		})))
		return
	}

	render(w, r, http.StatusOK, views.CatalogPage(runtime.Presenter.CatalogPage(presenter.CatalogPageState{
		RepoURL: strings.TrimSpace(r.URL.Query().Get("repo_url")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		Success: strings.TrimSpace(r.URL.Query().Get("success")),
		Courses: courses,
	})))
}

func (a *App) CoursePage(w http.ResponseWriter, r *http.Request, id string) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	view, err := runtime.Source.Course(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	render(w, r, http.StatusOK, views.CoursePage(runtime.Presenter.CoursePage(presenter.CoursePageState{
		View:  view,
		Error: strings.TrimSpace(r.URL.Query().Get("error")),
	})))
}

func (a *App) LessonPage(w http.ResponseWriter, r *http.Request, courseID, lessonID string) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()

	view, err := runtime.Source.Course(ctx, courseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	render(w, r, http.StatusOK, views.LessonPage(runtime.Presenter.LessonPage(presenter.LessonPageState{
		Course:   view,
		LessonID: lessonID,
		Error:    strings.TrimSpace(r.URL.Query().Get("error")),
	})))
}

func (a *App) AddCourse(w http.ResponseWriter, r *http.Request) {
	runtime := a.runtimeFor(r)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/?error=invalid+form", http.StatusSeeOther)
		return
	}

	repoURL := strings.TrimSpace(r.FormValue("repo_url"))
	if repoURL == "" {
		http.Redirect(w, r, "/?error=repo+URL+is+required", http.StatusSeeOther)
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()
	if err := runtime.Source.AddCourse(ctx, repoURL); err != nil {
		http.Redirect(w, r, "/?error="+urlQueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/?success=course+saved", http.StatusSeeOther)
}

func (a *App) EnrollCourse(w http.ResponseWriter, r *http.Request, id string) {
	runtime := a.runtimeFor(r)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/course/"+id+"?error=invalid+form", http.StatusSeeOther)
		return
	}
	studentRepo := strings.TrimSpace(r.FormValue("student_repo"))
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := runtime.Source.Enroll(ctx, id, studentRepo); err != nil {
		http.Redirect(w, r, "/course/"+id+"?error="+urlQueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/course/"+id, http.StatusSeeOther)
}

func (a *App) RemoveCourse(w http.ResponseWriter, r *http.Request, id string) {
	runtime := a.runtimeFor(r)
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := runtime.Source.RemoveCourse(ctx, id); err != nil {
		http.Redirect(w, r, "/?error="+urlQueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/?success=course+removed", http.StatusSeeOther)
}

func (a *App) ProgressWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.WebhookToken != "" {
		token := strings.TrimSpace(r.Header.Get("X-Webhook-Token"))
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("token"))
		}
		if token != a.WebhookToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	var payload course.Progress
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func urlQueryEscape(value string) string {
	replacer := strings.NewReplacer(" ", "+", "&", "", "?", "", "#", "")
	return replacer.Replace(value)
}
