package course

import (
	"net/http"
	"strings"

	"gitcourse/internal/handlers"
	appmodule "gitcourse/internal/module"
)

type Module struct {
	app *handlers.App
}

func New(app *handlers.App) *Module {
	return &Module{app: app}
}

func (m *Module) ID() string { return "course" }
func (m *Module) Name() string { return "Course" }

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		m.app.CatalogPage(w, r)
	})

	mux.HandleFunc("/courses/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		m.app.AddCourse(w, r)
	})

	mux.HandleFunc("/courses/", func(w http.ResponseWriter, r *http.Request) {
		pathValue := strings.TrimPrefix(r.URL.Path, "/courses/")
		switch {
		case strings.HasSuffix(pathValue, "/enroll"):
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			id := strings.TrimSuffix(pathValue, "/enroll")
			m.app.EnrollCourse(w, r, strings.Trim(id, "/"))
		case strings.HasSuffix(pathValue, "/remove"):
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			id := strings.TrimSuffix(pathValue, "/remove")
			m.app.RemoveCourse(w, r, strings.Trim(id, "/"))
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/course/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		trimmed := strings.TrimPrefix(r.URL.Path, "/course/")
		parts := strings.Split(strings.Trim(trimmed, "/"), "/")
		if len(parts) == 1 {
			m.app.CoursePage(w, r, parts[0])
			return
		}
		if len(parts) == 3 && parts[1] == "lesson" {
			m.app.LessonPage(w, r, parts[0], parts[2])
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/api/webhook/progress", m.app.ProgressWebhook)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Courses",
		Path:  "/",
		Icon:  "book-open",
		Order: 10,
	}}
}
