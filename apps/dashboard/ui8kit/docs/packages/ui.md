# Package `ui`

Import:

```go
import "github.com/fastygo/ui8kit/ui"
```

## Purpose

The `ui` package provides presentational components: layout primitives, typography, actions, form fields, and icons. Each component is a `templ` template function that returns `templ.Component`.

## Components

| Component | Description |
|-----------|-------------|
| `Box` | Generic container; uses inner `<div>` with merged utility classes |
| `Stack` | Vertical stack: `flex flex-col gap-4` plus utilities |
| `Group` | Horizontal group: `flex` row with optional `Grow` |
| `Container` | Centered max-width container (`max-w-7xl mx-auto px-4`) |
| `Block` | Alias for `Box` (`BlockProps` = `BoxProps`) |
| `Button` | Button or link (`Href` set → `<a>`) with variants and sizes |
| `Badge` | Small status label |
| `Text` | Paragraph with typography props |
| `Title` | Heading `h1`–`h6` via `Order` |
| `Field` | Input, textarea, select, checkbox/radio via `Component` and `Type` |
| `Icon` | Latty icon span (`latty latty-{name}`) |

## Props pattern

All visual props are defined in `props.go` in this package. Embed `utils.UtilityProps` where applicable:

```go
ui.Button(ui.ButtonProps{
    UtilityProps: utils.UtilityProps{P: "2", Rounded: "md"},
    Variant:      "primary",
    Size:         "sm",
}, "Save")
```

## Tailwind and `UtilityProps` literals

When `UtilityProps` values are built dynamically at runtime, Tailwind cannot always detect final classes (for example `p-2 rounded-md`) during source scanning.

UI8Kit provides `scripts/gen-ui8kit-css.go` to build a safelist CSS file:

```bash
# 1) generate *_templ.go first
templ generate

# 2) generate styles/ui8kit.css (includes classes from UtilityProps literals in *_templ.go)
go run ./scripts/gen-ui8kit-css.go
```

Or run both steps with one command:

```bash
./scripts/gen-css.sh
```

You can also use standard Go generation hooks:

```bash
go generate ./...
```

The generator scans:

- `utils/props.go` for static utility mappings
- `*_templ.go` for literal `utils.UtilityProps{...}` values

As a result, `UtilityProps: utils.UtilityProps{P: "2", Rounded: "md"}` is materialized into concrete safelist selectors like `.p-2` and `.rounded-md`.

## Button variants

Common values for `Variant`: `primary`, `destructive`, `outline`, `secondary`, `ghost`, `link`. Unknown values fall through to raw class fragments (see `utils.ButtonStyleVariant`).

## Field modes

- Default: single-line `<input>`; set `Type` to `email`, `password`, etc.
- `Component: "textarea"` — multi-line; use `Rows`.
- `Component: "select"` — provide `Options` and `Value`.
- Checkbox / radio: set `Type` to `checkbox` or `radio`; styling uses control-specific classes.

## Children

Components that wrap content (`Box`, `Stack`, `Group`, `Container`, `Block`) use templ’s `{ children... }` slot in the template source. Call them from other `.templ` files with block syntax as shown in the main [README](../../README.md).
