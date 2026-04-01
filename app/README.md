# HubRelay Dashboard

`app/` is a standalone BFF web application for HubRelay.

It renders HTML with `templ` + `github.com/fastygo/ui8kit`, talks to HubRelay through `sshbot/sdk/hubrelay`, and exposes browser-facing SSE for streaming ask responses.

The development module currently uses a local `replace` for `github.com/fastygo/ui8kit` because the latest published module does not ship generated `*_templ.go` files. The Docker build clones the upstream repository inside the container so it does not depend on your local checkout path.

## Features

- Session-based operator login with in-memory cookies
- Health and discovery view
- Capabilities view
- Ask page with SSE streaming
- Egress status view
- Audit view
- HTTP and unix socket transport modes

## Configuration

- `APP_DATA_SOURCE=live|fixture`
- `HUBRELAY_TRANSPORT=http|unix`
- `HUBRELAY_BASE_URL=http://127.0.0.1:5500`
- `HUBRELAY_SOCKET_PATH=/run/hubrelay/hubrelay.sock`
- `APP_BIND=0.0.0.0:8080`
- `APP_ADMIN_USER=admin`
- `APP_ADMIN_PASS=admin@123`
- `APP_AUTH_DISABLED=false`

`APP_DATA_SOURCE=live` is the default and talks to a running HubRelay instance.

When auth is enabled, browser requests are redirected to `/login` and API-style requests can still use HTTP Basic Auth. A successful Basic Auth request also creates a browser session cookie, matching the GUI dashboard flow.

For compatibility with the existing GUI environment, `PAAS_ADMIN_USER`, `PAAS_ADMIN_PASS`, and `DASHBOARD_AUTH_DISABLED` are also accepted as fallbacks.

`APP_DATA_SOURCE=fixture` loads page copy from locale fixtures and demo payloads from locale-specific `mocks` directories, which is useful for preview mode and UI iteration without a live backend.

Locales are discovered from `app/fixtures/*` directories at startup.

English is always the default locale. The language toggle stores the chosen locale in browser `localStorage` and synchronizes it with the `lang` query parameter for server-rendered pages.

The current MVP toggle switches between the first two discovered locales, with English forced to the first position when present.

`fixtures/es` is included as a proof-of-concept locale to verify that locale discovery no longer depends on hardcoded `ru` logic.

On the first visit, when no locale is stored yet, the shell script also tries to match the user's browser locale against the discovered fixture locales.

## Docs

- [Fixtures and runtime modes](docs/fixtures-and-runtime.md)

## Development

From `app/`:

```bash
go mod tidy
npm install
npm run sync:ui8kit
npm run build:css
templ generate ./...
go run ./cmd/server
```

For fixture-backed preview mode:

```bash
APP_DATA_SOURCE=fixture go run ./cmd/server
```

To iterate on the UI without logging in:

```bash
APP_AUTH_DISABLED=true go run ./cmd/server
```

## Docker

Build from the repository root so the local `sshbot` module is available to the app module replace:

```bash
docker build -f app/Dockerfile -t hubrelay-dashboard .
```
