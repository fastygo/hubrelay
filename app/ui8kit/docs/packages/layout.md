# Package `layout`

Import:

```go
import "github.com/fastygo/ui8kit/layout"
```

## Purpose

The `layout` package provides a dashboard-ready page shell: optional full-page shell, sidebar navigation, top header, and main content region.

It is designed for authenticated or admin-style interfaces where app sections need a consistent chrome with responsive behavior.

## Types

- `NavItem` — `Path`, `Label`, `Icon` (Latty icon name without the `latty-` prefix).
- `SidebarProps` — `Items`, `Active` (current path for active state), `Mobile` (controls sidebar rendering mode).
- `HeaderProps` — `Title` and optional extra header actions.
- `ShellProps` — title, language, brand, nav items, active route, optional `<head>` extras, optional header actions, and CSS path.

## Layout model

`Shell` builds the following structural order:

```text
html
└── body.kit-shell-body
    ├── mobile sheet input + portal (mobile only)
    └── .kit-shell-layout-row
        ├── .kit-shell-desktop-aside
        └── .kit-shell-main-column
            ├── .kit-header
            └── .kit-shell-main
                └── page content
```

### Desktop behavior

- Desktop uses `.kit-shell-desktop-aside` with full navigation and fixed width `md` breakpoint.
- Main content uses `.kit-shell-main-column` and fills remaining width.
- Header remains visible at the top of the main column.

### Mobile behavior

- Mobile navigation is handled by `.kit-shell-mobile-sheet-*` classes.
- The menu trigger is a `<label for="ui8kit-mobile-sheet">` in `Header`.
- Sheet input is hidden (`sr-only`) but controls open/close state.
- `label` controls in overlay and close button toggle the same input state.

## Shell component (CSS-only)

The `Shell` mobile navigation is intentionally implemented **without custom JavaScript for opening/closing**.

It uses these primitives:

- Hidden checkbox `#ui8kit-mobile-sheet`.
- Fixed portal `.kit-shell-mobile-sheet-portal`.
- Overlay label `.kit-shell-mobile-sheet-overlay`.
- Slide-in panel `.kit-shell-mobile-sheet-panel`.
- Close label `for="ui8kit-mobile-sheet"`.

Behavior:

- Menu is closed when checked is `false`.
- Menu opens when checked is `true`.
- Sidebar closes when overlay or close control is clicked.
- Body scroll lock is applied via `body.kit-shell-body:has(...)` at mobile widths.

No dedicated `open`/`close` JavaScript function is used in `layout` for this interaction.

Theme behavior is intentionally left to the consuming application. The layout renders the theme toggle button, but the client-side logic that reads stored theme preference, updates the icon, and toggles `.dark` should be provided as an external script by the app. Theme toggle copy can be passed through `ThemeToggle`.

The header can also render an optional extra action component before the theme toggle. This is useful for app-level controls such as locale switching.

## API reference

### `ShellProps`

| Field | Description |
|-------|-------------|
| `Title` | `<title>` value and header title. |
| `Lang` | `<html lang>` value; defaults to `en`. |
| `BrandName` | Sidebar title in desktop/mobile sheet. Empty value defaults to `App`. |
| `Active` | Path string matched against `NavItem.Path` for active link styling. |
| `NavItems` | Slice of `NavItem` items. |
| `CSSPath` | Stylesheet `href`; default `/static/css/app.css`. |
| `HeadExtra` | Optional `templ.Component` appended inside `<head>` (analytics, fonts, external scripts, custom links). |
| `HeaderExtra` | Optional `templ.Component` rendered in the header before the theme toggle. |
| `ThemeToggle` | Optional theme toggle labels (`Label`, `SwitchToDarkLabel`, `SwitchToLightLabel`). |

### `SidebarProps`

- `Items`: Navigation items.
- `Active`: Current path for active states.
- `Mobile`: `true` to render mobile sidebar links inside the sheet.

### `HeaderProps`

- `Title`: Title text in the header area.
- `Extra`: Optional `templ.Component` rendered before the theme toggle.
- `ThemeToggle`: Optional theme toggle labels forwarded to the theme button data attributes.

## Accessibility and semantics

- Mobile trigger is keyboard-focusable label button with `aria-label` and `aria-controls`.
- Header theme toggle uses `aria-pressed`.
- Sidebar links use semantic anchor tags.
- `aria-modal` is present on panel and `role="dialog"` is used for the sheet container.

## Styling and extension points

- Shared layout classes are defined in `styles/shell.css` using `kit-shell-*` namespaces.
- Prefer reusing `kit-shell-*` classes and utility prop variants before custom local classes.
- For app-specific overrides, add classes from your app CSS layer rather than editing kit internals.

## Best practices

- Keep navigation items as data (`[]layout.NavItem`) and avoid conditional template branches where possible.
- Use `Header`/`Sidebar` independently only when needed.
- Keep `ShellProps.CSSPath` explicit if your app serves CSS from a custom route.
