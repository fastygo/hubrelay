<h1 align="center">HubRelay</h1>

<p align="center">
  A single command bus on the server.<br/>
  Your laptop talks to it over SSH. Integrations plug in as adapters.
</p>

<p align="center">
  <sub>Go · BoltDB · loopback HTTP · immutable profiles · SOCKS proxy sessions · OpenAI-compatible AI</sub>
</p>

---

## What it does

HubRelay sits on a host and exposes a typed command interface. You connect through an SSH tunnel, issue commands from the browser or `curl`, and the server runs them through a capability-gated plugin chain. Outbound calls (AI, APIs) go through a shared egress policy so you control what leaves the box and how.

```
laptop ──ssh -L──▶ 127.0.0.1:5500 ──▶ command bus ──▶ plugins
                                            │
                                        audit (BoltDB)
```

Nothing listens on a public port. No secrets in the database. No adapter surprises after deploy.

## Design choices

| Decision | Reasoning |
| --- | --- |
| Profile compiled into the image | Adapters and capabilities are reviewed once. The running container cannot grow new surface. |
| BoltDB on a mounted volume | State survives `docker compose down && up`. Image root stays `read_only`. |
| Loopback bind + `ssh -L` | Your workstation is the trust boundary, not a firewall rule you hope is right. |
| Shared outbound policy | AI is the first consumer. The same lease/proxy/direct logic covers anything that dials out. |
| Env override for local dev | `INPUT_*` vars merge at startup so `go run ./cmd/bot` works without `-ldflags` on your laptop. |

## Get running

```bash
go test ./...
```

### Local (dev)

```bash
export INPUT_AI_API_KEY="<YOUR_AI_API_KEY>"
export INPUT_AI_BASE_URL="https://api.example.com/v1"
export INPUT_AI_MODEL="<YOUR_MODEL_ID>"
export INPUT_AI_API_MODE="chat_completions"
export INPUT_PROXY_SESSION_FORCE="false"
export INPUT_PROXY_SESSION_ENABLED="false"
go run ./cmd/bot
```

Then open `http://127.0.0.1:5500`.

### Provider smoke (just the AI path, no server)

```bash
export SMOKE_AI_API_KEY="<YOUR_AI_API_KEY>"
export SMOKE_AI_BASE_URL="https://api.example.com/v1"
export SMOKE_AI_MODEL="<YOUR_MODEL_ID>"
go run ./cmd/provider-smoke
```

Prints the raw provider response or error to stderr. Useful before you blame the UI.

### Docker

```bash
docker compose up --build
curl -s http://127.0.0.1:5500/healthz
```

## Docs

| Path | What you find there |
| --- | --- |
| [`docs/`](docs/README.md) | Full guide: Go install, providers, tunnel, proxy, deploy |
| [`.project/`](.project/README.md) | Architecture notes, ADRs, threat model |
| [`.paas/README.md`](.paas/README.md) | Deploy extensions and every `INPUT_*` explained |
| [`.paas/.check-deploy.md`](.paas/.check-deploy.md) | Deploy checklist with sanitised placeholders |
| [`.paas/.check-smoke.md`](.paas/.check-smoke.md) | Local smoke curls |

## Storage

BoltDB lives at `data/bot.db` by default. Set `BOT_DB_FILE` to change the path.

Secrets are never written to BoltDB. They exist only inside the compiled profile.

## Security

Do not commit real API keys or internal hostnames. Use `<PLACEHOLDER>` style in docs and issues.

---

<sub>The Go module is still called `sshbot`. Think of it as technical debt with a `TODO` — it works, it ships, it will get renamed right after that one other thing.</sub>
