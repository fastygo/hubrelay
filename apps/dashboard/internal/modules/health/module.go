package health

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
	return "health"
}

func (m *Module) Name() string {
	return "Health"
}

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/", m.app.Health)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Health",
		Path:  "/",
		Icon:  "server",
		Order: 0,
	}}
}
