# HubRelay Dashboard

`apps/dashboard/` is a standalone BFF web application that uses `sshbot/sdk/hubrelay` and `github.com/fastygo/ui8kit`.

## Features

- session login (memory cookies)
- capabilities view
- ask with SSE streaming
- egress status view
- audit and discovery views
- HTTP + unix socket transport modes

## Runtime config

- `APP_DATA_SOURCE=live|fixture`
- `HUBRELAY_TRANSPORT=http|unix`
- `HUBRELAY_BASE_URL=http://127.0.0.1:5500`
- `HUBRELAY_SOCKET_PATH=/run/hubrelay/hubrelay.sock`
- `APP_BIND=0.0.0.0:8080`
- `APP_ADMIN_USER`, `APP_ADMIN_PASS`
- `APP_AUTH_DISABLED`

`APP_DATA_SOURCE=live` (default) talks to HubRelay.
`APP_DATA_SOURCE=fixture` uses local mock payloads.

## Local run

```bash
go mod tidy
npm install
npm run sync:ui8kit
npm run build:css
templ generate ./...
go run ./cmd/server
```

```bash
APP_DATA_SOURCE=fixture go run ./cmd/server
APP_AUTH_DISABLED=true go run ./cmd/server
```

## References

- [Fixtures and runtime modes](docs/fixtures-and-runtime.md)
- Deployment: [`.paas/README.md`](../../.paas/README.md)
