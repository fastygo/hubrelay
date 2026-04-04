# Fixtures and runtime modes

The dashboard uses locale fixtures for copy and optional mock payloads.

## Locale catalogs

- `fixtures/en/*.json`
- `fixtures/es/*.json`
- `fixtures/ru/*.json`

The server scans `apps/dashboard/fixtures/*` at startup.

## Runtime locale logic

1. explicit `lang` query parameter
2. `hubrelay-language` in `localStorage`
3. browser language
4. `en` fallback

MVP UI currently toggles between first two discovered locales.

## Mock payloads

- `fixtures/<locale>/mocks/*.json`

Use mocks for:

- preview mode
- UI-only iteration
- deterministic demos

## Runtime modes

- `APP_DATA_SOURCE=fixture` — local mocks only, no HubRelay call
- `APP_DATA_SOURCE=live` — call real HubRelay via `sdk/hubrelay`

## Source chain in live mode

1. HTTP handler → source layer
2. source layer → `internal/relay`
3. `internal/relay` → `sdk/hubrelay`
4. SDK → HubRelay over HTTP/unix socket

## Client behavior

- page labels from fixtures
- transport and command errors from live backend
- SSE for ask follows SDK backend behavior

## Recommended commands

```bash
APP_DATA_SOURCE=fixture go run ./apps/dashboard/cmd/server
APP_DATA_SOURCE=live go run ./apps/dashboard/cmd/server

templ generate ./...
npm run sync:ui8kit
npm run build:css
```
