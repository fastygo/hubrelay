# UI8Kit design principles

This document defines the architecture and styling rules that guide all UI8Kit changes.

## Core philosophy

- Build consistent dashboard UI with an **8px grid mindset**.
- Keep spacing and sizing aligned to `4` and `8` steps where possible.
- Keep visuals simple and calm: avoid unnecessary gradients, heavy effects, and decorative motion.
- Use semantic design tokens from `styles/base.css` as the single source of truth for colors, borders, and surfaces.
- Keep behavior and rendering separate: `.templ` should stay structural and readable.

## Style architecture

UI8Kit styles are split by responsibility.

- `styles/base.css` owns global design tokens, theme variables, and base resets.
- `styles/components.css` owns shared visual primitives (`kit-card`, tables, chips, labels).
- `styles/shell.css` owns page chrome layout (`kit-shell-*`, `kit-header-*`) and dashboard structure.
- `styles/latty.css` owns icon masks and shared icon utility classes.
- `app.css` is expected to combine Tailwind output and imported UI8Kit layers.

Prefer semantic class names over ad-hoc utility combinations in `.templ` code.

## Template rules (`.templ`)

- Avoid direct utility usage when a component prop or utility prop can express the style.
- Avoid custom inline styles in templates.
- Prefer `utils.UtilityProps` and component props as the first styling API.
- If a template currently uses raw utility strings, treat it as migration debt and move it toward prop-driven composition.
- Use semantic `kit-` classes for reusable kit visuals instead of repeating utility clusters.

## `props` and `variants` workflow

`utils/props.go` and `utils/variants.go` form the styling API layer.

- `props.go` provides semantic prop fields and class composition entry points.
- `variants.go` provides named semantic variants (`primary`, `outline`, `kpi`, `lg`, and others) backed by utility strings.
- Keep props minimal and reusable so app teams can compose consistent DOM with predictable names.
- Use `FieldVariant`, `ButtonStyleVariant`, `TypographyClasses`, and `CardVariant` before introducing new local utility patterns.
- Extend these files when a new reusable visual pattern is needed across the kit, not per template.

## When props are not enough

If a new reusable pattern cannot be expressed with existing props:

1. Add or extend semantic props in `utils/props.go`.
2. Add variant names in `utils/variants.go` for controlled, discoverable options.
3. Add semantic classes in CSS with `@apply` when the pattern has new structure:
   - UI8Kit shared primitives: `kit-*` in `styles/components.css` or `styles/shell.css`.
   - App-only extensions: `app-*` in application CSS (for example `static/css/app.css` source).
4. Keep component templates declarative and low-noise after the abstraction is added.

## Layout and interaction principles

- Keep structure and interaction roles clear in layout components.
- Prefer CSS-only interaction patterns for simple UI state transitions where possible.
- The `layout` package should remain accessible without custom JavaScript for core navigation.
- Keep extra scripts localized and small when interactions cannot be solved with CSS.

## Accessibility and theming

- Dark mode support is mandatory and defaults through `.dark` variables in `base.css`.
- ARIA attributes should be present on interactive controls and regions.
- Preserve accessible focus and keyboard visibility.
- Keep icon classes consistent; add icons through `latty.css`.

## Enforcement workflow

- Generate outputs and styling artifacts before validation:
  - `templ generate`
  - `go run ./scripts/gen-ui8kit-css.go` (or `./scripts/gen-css.sh`, `go generate ./...`)
- During review, reject direct utility sprawl in `.templ` that bypasses props/variants without justification.
