# Operations

## Health

```bash
curl -s http://127.0.0.1:5500/healthz
```

Expect JSON including adapter name and profile id.

## Capabilities inspection

```bash
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

Use this to confirm:

- `ai_has_api_key` — key was compiled into the image (boolean only, never the secret),
- `ai_model`, `ai_base_url`, `ai_api_mode`,
- `proxy_force`, `proxy_session`,
- list of capabilities.

**Why**: faster than reading logs when debugging “ask returns generic error”.

## Process logs

Startup logs include a one-line **profile summary** (model, base URL host path, proxy flags) and `[ask]` / `[ai]` lines when AI logging is enabled.

**Why**: correlates operator actions with provider errors without exposing secrets.

## BoltDB location

Default: `data/bot.db` relative to the working directory, or `BOT_DB_FILE`.

Back up the file when you care about audit history; restoring it should be done with the bot stopped.

## Common failures

| Symptom | Likely cause |
| --- | --- |
| `ai_has_api_key: false` | Key not passed at image build / wrong `INPUT_*` |
| `proxy session is required` | `INPUT_PROXY_SESSION_FORCE=true` but no lease in UI/API |
| `404` from provider | Wrong base URL join, wrong model id, or key lacks entitlement |
| Connection reset on tunnel | SSH session died or bind port clash locally |

## Escalation path

1. `provider-smoke` with same URL/model/key class as production.
2. `capabilities` on running bot.
3. Short `ask` via `curl` (isolates browser from server).
4. Provider dashboard / key scope with vendor.
