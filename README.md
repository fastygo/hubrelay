# HubRelay

HubRelay is a headless command hub with controlled egress relay.

It does not serve a browser UI and it does not own your application data.
It accepts a command, gates it, relays it through the approved path, and returns what came back.

## What it does

HubRelay runs a private command bus on the server and exposes JSON and streaming endpoints over controlled transports.

```text
client app / CLI / dashboard
        |
        v
sdk/hubrelay
        |
        v
HubRelay binary
        |
        +--> core.Service
        +--> plugins
        +--> audit (BoltDB)
        +--> egress manager
                |
                +--> wg-b1 / wg-b2 / ...
                +--> blackhole when nothing is healthy
```

Primary client model:
- Go SDK over HTTP
- Go SDK over unix socket

Debug or lab access can still use SSH local forwarding to the private HTTP listener.

## Design choices

| Decision | Reasoning |
| --- | --- |
| Headless binary | Keep transport/UI concerns outside the server process. |
| SDK-first access | Future dashboards, CLIs, and agents share one typed client contract. |
| Immutable build profile | Runtime surface does not grow after deploy. |
| BoltDB on a mounted volume | State survives restarts without turning the image mutable. |
| Multi-egress manager | Failover is explicit, auditable, and health-driven. |
| OS-level enforcement | Host routing and firewall policy prevent WAN fallback. |

## Get running

Run tests:

```bash
go test ./...
```

Local run:

```bash
export INPUT_AI_API_KEY="<YOUR_AI_API_KEY>"
export INPUT_AI_BASE_URL="https://api.example.com/v1"
export INPUT_AI_MODEL="<YOUR_MODEL_ID>"
export INPUT_AI_API_MODE="chat_completions"
export INPUT_PROXY_SESSION_FORCE="false"
export INPUT_PROXY_SESSION_ENABLED="false"
go run ./cmd/bot
```

Basic verification:

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s http://127.0.0.1:5500/api/command \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

## SDK

Use the Go SDK from `sdk/hubrelay`:

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
    Args: map[string]string{
        "prompt": "hello",
    },
})
```

See:
- [`docs/sdk.md`](docs/sdk.md)
- [`docs/egress.md`](docs/egress.md)
- [`docs/streaming.md`](docs/streaming.md)

## Docs

| Path | What you find there |
| --- | --- |
| [`docs/`](docs/README.md) | Canonical English guide for operators and contributors |
| [`.project/`](.project/README.md) | Architecture notes, ADRs, threat model |
| [`.project/wg/os-enforcement.md`](.project/wg/os-enforcement.md) | Host-level kill switch and policy routing runbook |
| [`.paas/README.md`](.paas/README.md) | Deploy extensions and `INPUT_*` reference |

## Storage

BoltDB lives at `data/bot.db` by default. Set `BOT_DB_FILE` to change the path.

Secrets are never written to BoltDB. They exist only in runtime configuration and provider requests.

## Security

Do not commit real API keys, internal hostnames, or customer payloads. Use placeholder values in docs and examples.

HubRelay should validate the private path.
The operating system should enforce it.

---

HubRelay — command hub with controlled egress relay. It does not generate answers, serve pages, or own your data. It accepts a command, gates it, relays it through a private path, and returns what came back. Everything else is a plugin.
