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
cd apps/dashboard
go mod tidy
npm install
npm run sync:ui8kit
npm run build:css
templ generate ./...
APP_DATA_SOURCE=fixture APP_AUTH_DISABLED=true APP_BIND=127.0.0.1:8080 go run ./cmd/server
```

```bash
APP_DATA_SOURCE=live APP_BIND=127.0.0.1:8080 APP_ADMIN_USER=admin APP_ADMIN_PASS=<change_me> go run ./cmd/server
```

Fixture mode is enough for full UI/UX iteration:

- all main pages render from fixtures (`health`, `capabilities`, `ask`, `egress`, `audit`)
- no HubRelay instance is required
- auth can be temporarily disabled with `APP_AUTH_DISABLED=true`

Pages to validate:

```bash
curl -sI http://127.0.0.1:8080/
curl -sI http://127.0.0.1:8080/capabilities
curl -sI http://127.0.0.1:8080/ask
curl -sI http://127.0.0.1:8080/egress
curl -sI http://127.0.0.1:8080/audit
```

## References

- [Fixtures and runtime modes](docs/fixtures-and-runtime.md)
- Deployment: [`.paas/README.md`](../../.paas/README.md)
