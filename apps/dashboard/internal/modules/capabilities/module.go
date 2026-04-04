package capabilities

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
	return "capabilities"
}

func (m *Module) Name() string {
	return "Capabilities"
}

func (m *Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/capabilities", m.app.Capabilities)
}

func (m *Module) NavItems() []appmodule.NavItem {
	return []appmodule.NavItem{{
		Label: "Capabilities",
		Path:  "/capabilities",
		Icon:  "sparkles",
		Order: 10,
	}}
}
