# Packages overview

UI8Kit is split into four importable packages.

| Package | Import path | Role |
|---------|-------------|------|
| `ui` | `github.com/fastygo/ui8kit/ui` | Low-level building blocks and form controls |
| `layout` | `github.com/fastygo/ui8kit/layout` | Full-page shell with sidebar, header, and main |
| `utils` | `github.com/fastygo/ui8kit/utils` | Class utilities and design-token helpers |
| `styles` | `github.com/fastygo/ui8kit/styles` | Embedded CSS (`embed.FS`) for theme and icons |

The repository root package `github.com/fastygo/ui8kit` exposes only metadata; components live in subpackages.

## Detailed references

- [ui](ui.md)
- [layout](layout.md)
- [Shell component](layout-shell.md)
- [utils](utils.md)
- [styles](styles.md)

## Dependency graph

- `ui` and `layout` depend on `utils` and `github.com/a-h/templ`.
- `styles` has no templ dependency; it is plain Go + embedded CSS files.
- Your app imports `ui` / `layout` and typically serves a compiled CSS bundle from Tailwind input that imports `styles/*.css`.
