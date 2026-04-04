package hubcore

import "net/http"

type Module interface {
	ID() string
	Name() string
	Routes(mux *http.ServeMux)
	NavItems() []NavItem
}

type NavItem struct {
	Label string
	Path  string
	Icon  string
	Order int
}
