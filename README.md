<h1 align="center">HubRelay</h1>

<p align="center"><strong>Operator hub for engineers.</strong></p>

<p align="center">
  One command bus on the server — your laptop reaches it over SSH; integrations plug in as adapters.<br/>
  <sub>Go · immutable deploy profile · BoltDB runtime state · loopback HTTP · optional AI & SOCKS egress</sub>
</p>

---

**HubRelay** is built for **SREs, platform engineers, and operators** who need a **stable control surface** on a host: inspect the box, run typed commands, optionally talk to an **OpenAI-compatible** API, and route outbound work through **proxy sessions** when policy requires it — without bolting secrets into BoltDB or widening the network blast radius by default.

| Principle | Why it matters |
| --- | --- |
| **Profile baked at build** | Adapters and capabilities are fixed per image — review once, deploy predictably. |
| **State on disk, not in the image** | BoltDB on a volume survives container replacement; `read_only` root stays viable. |
| **Loopback first** | HTTP UI binds to loopback; you use **`ssh -L`** so your workstation is the trust boundary. |
| **Shared outbound policy** | AI is the first consumer; the same egress rules can cover future integrations. |

---

## Quick start

```bash
go test ./...
```

**Local development** (profile merges from `INPUT_*` env — no `-ldflags` required):

```bash
export INPUT_AI_API_KEY="<YOUR_AI_API_KEY>"
export INPUT_AI_BASE_URL="https://api.example.com/v1"
export INPUT_AI_MODEL="<YOUR_MODEL_ID>"
export INPUT_AI_API_MODE="chat_completions"
export INPUT_PROXY_SESSION_FORCE="false"
export INPUT_PROXY_SESSION_ENABLED="false"
go run ./cmd/bot
```

Open `http://127.0.0.1:5500`.

**Minimal AI sanity check** (no HTTP server):

```bash
export SMOKE_AI_API_KEY="<YOUR_AI_API_KEY>"
export SMOKE_AI_BASE_URL="https://api.example.com/v1"
export SMOKE_AI_MODEL="<YOUR_MODEL_ID>"
go run ./cmd/provider-smoke
```

**Docker** (production-like):

```bash
docker compose up --build
curl -s http://127.0.0.1:5500/healthz
```

---

## Documentation

| Resource | Description |
| --- | --- |
| [docs/](docs/README.md) | End-to-end guide (toolchain → providers → tunnel → deploy) |
| [.project/](.project/README.md) | Architecture, ADRs, threat & storage models |
| [.paas/README.md](.paas/README.md) | Deploy extensions and `INPUT_*` reference |
| [.paas/.check-deploy.md](.paas/.check-deploy.md) | Deploy checklist (sanitised placeholders) |
| [.paas/.check-smoke.md](.paas/.check-smoke.md) | Local smoke & curl checks |

---

## Runtime storage

Default BoltDB path: `data/bot.db`. Override with `BOT_DB_FILE`.

---

## Security

Do not commit real API keys or private hostnames. Use placeholders in issues and docs.

---

*Repository module path remains `sshbot` until you rename the Go module.*
