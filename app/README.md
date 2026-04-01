# HubRelay Dashboard

`app/` is a standalone BFF web application for HubRelay.

It renders HTML with `templ` + `github.com/fastygo/ui8kit`, talks to HubRelay through `sshbot/sdk/hubrelay`, and exposes browser-facing SSE for streaming ask responses.

The development module currently uses a local `replace` for `github.com/fastygo/ui8kit` because the latest published module does not ship generated `*_templ.go` files. The Docker build clones the upstream repository inside the container so it does not depend on your local checkout path.

## Features

- Health and discovery view
- Capabilities view
- Ask page with SSE streaming
- Egress status view
- Audit view
- HTTP and unix socket transport modes

## Configuration

- `HUBRELAY_TRANSPORT=http|unix`
- `HUBRELAY_BASE_URL=http://127.0.0.1:5500`
- `HUBRELAY_SOCKET_PATH=/run/hubrelay/hubrelay.sock`
- `APP_BIND=0.0.0.0:8080`

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

## Docker

Build from the repository root so the local `sshbot` module is available to the app module replace:

```bash
docker build -f app/Dockerfile -t hubrelay-dashboard .
```
