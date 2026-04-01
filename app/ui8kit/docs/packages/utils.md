# Package `utils`

Import:

```go
import "github.com/fastygo/ui8kit/utils"
```

## Purpose

Shared building blocks for class composition.

`utils` contains two related responsibilities:

- structural prop composition helpers (`UtilityProps` and `Resolve()`)
- semantic variant functions (`ButtonStyleVariant`, `FieldVariant`, `CardVariant`, etc.)

These keep UI8Kit templates readable and consistent.

## `Cn`

`Cn` concatenates non-empty class strings with a single space.

```go
utils.Cn("px-4", "", "py-2", "bg-card") // "px-4 py-2 bg-card"
```

## `UtilityProps`

`UtilityProps` mirrors common Tailwind property families (`p`, `m`, `w`, `h`, `rounded`, `bg`, and so on).

Calling `Resolve()` converts populated fields into a class string.

Key examples:

| Fields | Output |
|--------|--------|
| `P: "4"` | `p-4` |
| `Flex: "col"` | `flex flex-col` |
| `Gap: "lg"` | `gap-6` (semantic gap alias) |
| `Rounded: "lg"` | `rounded-lg` |
| `Rounded: "default"` | `rounded` |
| `Border: "true"` | `border` |
| `Hidden: true` | `hidden` |

Embed `UtilityProps` in component props when style customization is part of component API.

## Variant helpers

Use these helpers before writing raw utility strings:

- `ButtonStyleVariant`, `ButtonSizeVariant` for button styles.
- `BadgeStyleVariant`, `BadgeSizeVariant` for badges.
- `FieldVariant`, `FieldSizeVariant` for inputs and textareas.
- `FieldControlVariant`, `FieldControlSizeVariant` for checkbox/radio controls.
- `TypographyClasses` for text and title settings.
- `CardVariant` for shared card surfaces (`kit-card`, `kit-card--*`).

Variant functions are implemented in `utils/variants.go`.

## `props.go` and `variants.go` integration

- `utils/props.go` is the prop API surface. Prefer adding new style-relevant fields here before inventing custom local class logic.
- `utils/variants.go` stores stable reusable style maps and aliases.
- If a new reusable visual pattern is required across package boundaries, add it to `variants.go` first.
- The generator in `scripts/gen-ui8kit-css.go` scans `props.go` and generated `*_templ.go` files for utility class usage.

## Extensibility

Passing unknown `Variant` values often falls back to raw classes.

Prefer documented variant names for stable upgrades and predictable behavior.
