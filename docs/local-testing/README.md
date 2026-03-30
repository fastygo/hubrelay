# Local testing

Why test in layers: isolate **provider credentials** from **HTTP stack**, then validate the **full bot**.

## 1. Provider smoke (`cmd/provider-smoke`)

Minimal binary that calls the same `internal/ai` stack as the bot **without** BoltDB, adapters, or proxy UI.

**Why**: fastest feedback when a key, model, or base URL is wrong. Errors print to stderr with full provider messages.

```bash
export SMOKE_AI_API_KEY="<YOUR_AI_API_KEY>"
export SMOKE_AI_BASE_URL="https://api.example.com/v1"
export SMOKE_AI_MODEL="<YOUR_MODEL_ID>"
export SMOKE_AI_API_MODE="chat_completions"
export SMOKE_PROMPT="Reply with one short sentence."

go run ./cmd/provider-smoke
```

Optional: `SMOKE_SYSTEM`, `SMOKE_USER_ID`, `SMOKE_TIMEOUT_SEC`.

See also [`.paas/.check-smoke.md`](../../.paas/.check-smoke.md) for a copy-paste checklist.

## 2. Full bot locally (`cmd/bot`)

**Why**: exercises plugins, `/api/command`, sensitive-data policy, and (if enabled) proxy endpoints.

Use `INPUT_*` variables as in [Getting started](../getting-started/README.md). Then run the same `capabilities` and `ask` curls as in deploy verification.

## 3. Optional checks

| Check | Command |
| --- | --- |
| Health | `curl -s http://127.0.0.1:5500/healthz` |
| Capabilities | `POST /api/command` with `"command":"capabilities"` |
| Ask | `POST /api/command` with `"command":"ask"` and `args.prompt` |

## 4. Proxy session API (when enabled)

If `INPUT_PROXY_SESSION_ENABLED=true`, you can create a session for lab testing. Use non-production addresses in your own environment only; do not commit real proxy lists.

```bash
curl -s -X POST http://127.0.0.1:5500/api/proxy/session \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","proxies":"<HOST>:<PORT>"}'
```

**Why**: the API contract is part of the product; real infrastructure values belong in private runbooks, not in the repo.
