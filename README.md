# HubRelay

HubRelay is a headless command bus with controlled outbound egress.

## What it is

- command runtime, no browser UI,
- immutable profile from deploy-time input,
- mutable state in BoltDB,
- operator access via loopback + SSH tunnel.

## Modules and boundaries

- `sshbot` root: core daemon (`cmd/bot`), plugins, policy.
- `sdk/hubrelay`: typed Go client for `/api/command` and streaming.
- `hubcore`: shared dashboard library linked at build time.
- `apps/dashboard`: dashboard service (`apps/dashboard/cmd/server`).
- `apps/dashboard/ui8kit`: UI component toolkit.

`hubcore` is a library and is not deployed as a service.

## Build and test matrix

```bash
go test ./...
go test ./hubcore/...
go test ./sdk/hubrelay/...
go test ./apps/dashboard/...
go test ./apps/dashboard/ui8kit/...
```

```bash
go build ./cmd/bot
go build ./apps/dashboard/cmd/server
```

## Local run

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
go run ./apps/dashboard/cmd/server
```

## Runtime checks

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

## SDK entry

```go
client := hubrelay.NewHTTPClient("http://127.0.0.1:5500",
    hubrelay.WithPrincipal(hubrelay.Principal{
        ID:    "operator-local",
        Roles: []string{"operator"},
    }),
)
defer client.Close()

result, err := client.Execute(ctx, hubrelay.CommandRequest{
    Command: "ask",
    Args: map[string]string{"prompt": "hello"},
})
```

## Docs

- Docs index: [`docs/`](docs/README.md)
- Architecture notes: [`.project/`](.project/README.md)
- Deploy/runbook: [`.paas/README.md`](.paas/README.md)
- Security summary: [`.project/security-model.md`](.project/security-model.md)

## Safety note

Do not commit secrets. Use placeholders such as `<YOUR_AI_API_KEY>`, `<SERVER_HOST>`, `<SSH_PRIVATE_KEY>`.
