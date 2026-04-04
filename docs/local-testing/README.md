# Local testing

Use a three-layer flow: smoke API, full bot, and dashboard overlay checks.

## 1) Provider smoke (`cmd/provider-smoke`)

```bash
export SMOKE_AI_API_KEY="<YOUR_AI_API_KEY>"
export SMOKE_AI_BASE_URL="https://api.example.com/v1"
export SMOKE_AI_MODEL="<YOUR_MODEL_ID>"
export SMOKE_AI_API_MODE="chat_completions"
export SMOKE_PROMPT="Reply with one short sentence."
export SMOKE_SYSTEM="false" # optional

go run ./cmd/provider-smoke
```

Optional: `SMOKE_SYSTEM`, `SMOKE_USER_ID`, `SMOKE_TIMEOUT_SEC`.

Copy-paste checklist: [`.paas/.check-smoke.md`](../../.paas/.check-smoke.md).

## 2) Full bot locally (`cmd/bot`)

Use `INPUT_*` variables as in [Getting started](../getting-started/README.md), then run:

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"ask","args":{"prompt":"hello"}}'
```

## 2.1) Dashboard against local bot

```bash
APP_DATA_SOURCE=live go run ./apps/dashboard/cmd/server
APP_DATA_SOURCE=fixture APP_AUTH_DISABLED=true go run ./apps/dashboard/cmd/server
```

## Troubleshooting

1. Start bot and dashboard in separate terminals so one failing service does not mask the other.
2. If the dashboard page is empty but returns HTTP 200:

```bash
rm -rf apps/dashboard/static/build
cd apps/dashboard
npm run build:css
templ generate ./...
go run ./cmd/server
```

On Windows:

```bash
Remove-Item -Recurse -Force apps/dashboard/static/build
```

3. If API calls to `POST /api/command` fail with connection refused:

```bash
ss -ltnp | grep 5500
lsof -i :5500
```

On Windows where `ss`/`lsof` are missing:

```bash
netstat -ano | findstr 5500
```

4. If provider smoke succeeds but bot ask fails, validate profile alignment:

```bash
echo "INPUT_AI_MODEL=$INPUT_AI_MODEL"
echo "INPUT_AI_BASE_URL=$INPUT_AI_BASE_URL"
echo "INPUT_AI_API_MODE=$INPUT_AI_API_MODE"
```

## 4) Proxy session API (optional)

```bash
curl -s -X POST http://127.0.0.1:5500/api/proxy/session \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","proxies":"<HOST>:<PORT>"}'
```
