package ask

import (
	"net/http"

	"hubrelay-dashboard/internal/handlers"
	appmodule "hubrelay-dashboard/internal/module"
)

type Module struct {
	app *handlers.App
}

func New(app *handlers.App) *Module {
	return &Module{app: app}
}

func (m *Module) ID() string {
	return "ask"
}

func (m *Module) Name() string {
	return "Ask"
}

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			m.app.AskPage(w, r)
		case http.MethodPost:
			m.app.AskSubmit(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/ask/stream", m.app.AskStream)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Ask",
		Path:  "/ask",
		Icon:  "message-circle",
		Order: 20,
	}}
}
