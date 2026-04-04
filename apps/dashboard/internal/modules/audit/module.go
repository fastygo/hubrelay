package audit

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
	return "audit"
}

func (m *Module) Name() string {
	return "Audit"
}

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/audit", m.app.Audit)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Audit",
		Path:  "/audit",
		Icon:  "history",
		Order: 40,
	}}
}
