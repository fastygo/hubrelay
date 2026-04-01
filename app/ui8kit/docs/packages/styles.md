# Package `styles`

Import:

```go
import "github.com/fastygo/ui8kit/styles"
```

## Purpose

Ships the design system CSS as embedded files so you can serve or read them without maintaining a separate icon repository.

`styles` contains compile-time embedded assets that are safe for embedding and simple HTTP serving.

## Embedded files

The following files are included in the binary via `embed.FS`:

| File | Role |
|------|------|
| `base.css` | Theme variables (`light`/`dark`), `@theme inline` bridge, base layer |
| `shell.css` | Layout shell and chrome (`kit-shell-*`, `kit-header-*`) |
| `components.css` | Shared UI patterns (`kit-*` cards/tables/chips), form helpers |
| `latty.css` | Latty icon masks and `.latty-*` classes |

## Styling approach

UI8Kit keeps styling in three stable layers plus icon definitions.

- `base.css` owns tokens and global defaults.
- `components.css` owns reusable primitives.
- `shell.css` owns full-page page shell and responsive behavior.
- `latty.css` owns icon masks and icon sizing.

Higher-level applications add their own `app-*` utility classes in their own CSS source (`static/css/input.css` + app layer).

## Access and serving

```go
data, err := styles.FS.ReadFile("base.css")
```

Or serve all style files:

```go
http.Handle("/static/css/ui8kit/", http.StripPrefix("/static/css/ui8kit/",
	http.FileServer(http.FS(styles.FS))))
```

## Relationship to Tailwind

UI8Kit CSS files are imported into the app pipeline and then combined with app-level utility output into a single bundle.

- Keep `@import "tailwindcss";` at the top of your Tailwind input.
- Import UI8Kit files in your app entry so layer order remains deterministic.
- Keep `.kit-*` class definitions in kit CSS layers and app-specific utilities in your app layer.

## Dark mode

`base.css` defines `.dark` variables and uses `@custom-variant dark (&:is(.dark *));`.

The layout script toggles `class="dark"` on `<html>`.

## Icons

`latty.css` defines CSS variables for SVG mask icons and `.latty-*` utility selectors.

Ensure `latty.css` is part of final generated CSS if you use `layout` or `ui.Icon`.
