# Sync UI8Kit styles in an application

This guide documents how application projects should sync UI8Kit CSS assets before building `app.css`.

## Script

- Path in consuming app: `app/scripts/sync-ui8kit-css.sh`

```bash
bash ./scripts/sync-ui8kit-css.sh
```

## What the script does

1. Resolves current UI8Kit module location with `go list -m -f '{{.Dir}}' github.com/fastygo/ui8kit`.
2. Copies theme and component CSS into:
   - `static/css/ui8kit/base.css`
   - `static/css/ui8kit/components.css`
   - `static/css/ui8kit/shell.css`
   - `static/css/ui8kit/latty.css`
3. Prints confirmation: `ui8kit CSS synced to ...`.

## Recommended command flow

From the app root:

```bash
templ generate ./...
npm run sync:ui8kit
npm run build:css
```

In this repository, the npm script is defined in `package.json` as:

```json
{
  "sync:ui8kit": "bash ./scripts/sync-ui8kit-css.sh"
}
```

## Where should the imports point

`static/css/input.css` should import all copied files:

```css
@import "./ui8kit/base.css";
@import "./ui8kit/components.css";
@import "./ui8kit/shell.css";
@import "./ui8kit/latty.css";
```

## Verification checklist

- Check that `app.css` includes shell and component classes after rebuild.
- Open a page with navigation and verify `.kit-shell-mobile-sheet-*` exists in DOM and transitions work.
- Confirm `layout` header still loads `/static/css/app.css` or your custom `CSSPath`.

## CI recommendation

Add these commands to your CI job if you keep generated static files in version control:

```bash
templ generate ./...
go run ./scripts/gen-ui8kit-css.go
npm run sync:ui8kit
npm run build:css
```

For monorepo workflows, run sync immediately before CSS build and restart server after file replacement.
