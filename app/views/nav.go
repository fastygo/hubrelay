package views

import (
	"context"
	"io"

	"github.com/a-h/templ"
)

func askPageHead() templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<script src="/static/js/stream.js" defer></script>`)
		return err
	})
}

func layoutHeadExtra(extra templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, `<script src="/static/js/app-shell.js" defer></script>`); err != nil {
			return err
		}
		if extra == nil {
			return nil
		}
		return extra.Render(ctx, w)
	})
}
