# Operations

## Runtime checks

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"ask","args":{"prompt":"hello"}}'
```

## Logs to watch

`[profile]` startup summary, `[ask]`, `[ai]`.

## BoltDB

- Default path: `data/bot.db`
- Override: `BOT_DB_FILE`

## Monorepo checks after structure changes

```bash
go test ./...
go test ./apps/dashboard/...
go test ./sdk/hubrelay/...
go test ./hubcore/...
go test ./cmd/bot
go test ./apps/dashboard/cmd/server
```

## Common failures

- `ai_has_api_key: false` — missing key input in build/config (`INPUT_*`).
- `proxy session is required` — `INPUT_PROXY_SESSION_FORCE=true` without active lease.
- `404` from provider — wrong base URL/model or key entitlement mismatch.
- Tunnel reset during checks — SSH ended or local port conflict.

## Escalation

1. `go run ./cmd/provider-smoke`
2. `curl` health + capabilities
3. short `ask` call via `curl`
4. provider dashboard / key scope check
