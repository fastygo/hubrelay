package egress

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
	return "egress"
}

func (m *Module) Name() string {
	return "Egress"
}

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/egress", m.app.Egress)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Egress",
		Path:  "/egress",
		Icon:  "shield",
		Order: 30,
	}}
}
