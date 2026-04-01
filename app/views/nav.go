package views

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/fastygo/ui8kit/layout"
)

func navItems() []layout.NavItem {
	return []layout.NavItem{
		{Path: "/", Label: "Health", Icon: "server"},
		{Path: "/capabilities", Label: "Capabilities", Icon: "sparkles"},
		{Path: "/ask", Label: "Ask", Icon: "message-circle"},
		{Path: "/egress", Label: "Egress", Icon: "shield"},
		{Path: "/audit", Label: "Audit", Icon: "history"},
	}
}

func askPageHead() templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<script src="/static/js/stream.js" defer></script>`)
		return err
	})
}
