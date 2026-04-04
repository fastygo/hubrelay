# Getting started

## Install Go

```bash
go version
```

Install from: https://go.dev/dl/

## Repository layout

- `sshbot` root: daemon (`cmd/bot`)
- `sdk/hubrelay`: Go client
- `hubcore`: shared dashboard library
- `apps/dashboard`: dashboard service
- `apps/dashboard/ui8kit`: UI toolkit

## Clone and build

```bash
git clone <YOUR_GIT_REMOTE_URL>
cd sshbot
go build -o bot ./cmd/bot
go build -o dashboard ./apps/dashboard/cmd/server
```

## Runtime dependencies

- `data/bot.db` (or `BOT_DB_FILE`)
- network access for `ask` and provider checks
- optional `docker compose` parity: [Deploy](../deploy/README.md)

## Before first run

```bash
cd apps/dashboard
go mod tidy
go mod download
npm install
npm run sync:ui8kit
npm run build:css
templ generate ./...
go test ./...
```

## First run

```bash
export INPUT_AI_API_KEY="<YOUR_AI_API_KEY>"
export INPUT_AI_BASE_URL="https://api.example.com/v1"
export INPUT_AI_MODEL="<YOUR_MODEL_ID>"
export INPUT_AI_API_MODE="chat_completions"
export INPUT_PROXY_SESSION_FORCE="false"
export INPUT_PROXY_SESSION_ENABLED="false"

go run ./cmd/bot
```

```bash
cd apps/dashboard
go run ./cmd/server
```

## Verify

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

## Next

- [Local testing](../local-testing/README.md)
- [Providers and AI](../providers-and-ai/README.md)
- [SDK](../sdk.md)
