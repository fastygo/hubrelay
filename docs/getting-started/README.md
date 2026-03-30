# Getting started

## Install Go

**Why Go**: single static binary, strong standard library, straightforward cross-compilation for Linux servers.

1. Install a **recent stable Go** matching `go.mod` (see `go` version directive in the repo root).
2. Verify:

```bash
go version
```

Official install: [https://go.dev/dl/](https://go.dev/dl/)

## Clone and build

```bash
git clone <YOUR_GIT_REMOTE_URL>
cd sshbot
go build -o bot ./cmd/bot
```

Run tests:

```bash
go test ./...
```

## Runtime dependencies

- **BoltDB file**: created automatically under `data/bot.db` unless you set `BOT_DB_FILE`.
- **Network**: required for `ask` and proxy health checks toward your AI provider.
- **Docker** (optional): for parity with production, use `docker compose` as described in [Deploy](../deploy/README.md).

## First run (local)

The bot reads **deploy-time** settings from the compiled profile. For local development, the same `INPUT_*` names used at deploy time are applied once at startup (see `internal/buildprofile`).

```bash
export INPUT_AI_API_KEY="<YOUR_AI_API_KEY>"
export INPUT_AI_BASE_URL="https://api.example.com/v1"
export INPUT_AI_MODEL="<YOUR_MODEL_ID>"
export INPUT_AI_API_MODE="chat_completions"
export INPUT_PROXY_SESSION_FORCE="false"
export INPUT_PROXY_SESSION_ENABLED="false"

go run ./cmd/bot
```

Open `http://127.0.0.1:5500` (or the bind address shown in `capabilities`).

**Why disable proxy flags locally**: unless you are testing SOCKS, disabling proxy session avoids needing a lease for every `ask`.

## Verify wiring

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

Confirm `ai_has_api_key` is `true` before relying on the browser chat.

## Next steps

- [Local testing](../local-testing/README.md) — smaller `provider-smoke` tool.
- [Providers and AI](../providers-and-ai/README.md) — base URLs and provider quirks.
