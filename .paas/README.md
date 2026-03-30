# Hub Bot Deployment Guide

Operator narrative (installation, providers, “why”): see **[`docs/`](../docs/README.md)**.

This `.paas` folder manages one thing only:

- the internal-only hub bot runtime.

The deployment strategy is intentionally simple:

- no managed application flows,
- no admin dashboard flows,
- no runtime adapter installation,
- one immutable bot image per deploy profile.

## Runtime model

- the bot listens only on the server loopback interface,
- the default internal URL is `http://127.0.0.1:5500`,
- first operator tests use `curl`,
- the first browser UX uses an SSH tunnel to the same loopback endpoint,
- mutable state stays in the mounted `bbolt` database,
- deploy inputs define the immutable image profile.

## Supported flows

| Extension | Purpose |
| --- | --- |
| `bootstrap-direct` | First install on a clean server or a full reinstall |
| `deploy-direct` | Update the bot by building the image on the server |
| `deploy` | Update the bot through a registry-backed image |

## Required inputs

### Shared runtime inputs

| Input | Meaning |
| --- | --- |
| `INPUT_BOT_NAME` | Docker Compose project and runtime directory name |
| `INPUT_BOT_URL` | Internal bot URL, normally `http://127.0.0.1:5500` |
| `INPUT_BOT_PROFILE_ID` | Immutable profile baked into the built image |
| `INPUT_BOT_DISPLAY_NAME` | Human-readable name for the immutable profile |
| `INPUT_BOT_HTTP_BIND` | Bind address compiled into the image; `127.0.0.1:5500` for strict loopback, or `0.0.0.0:5500` when the runtime must accept in-network connections (still use firewall policy) |
| `INPUT_BOT_EMAIL_ENABLED` | Whether the image exposes the email adapter |
| `INPUT_BOT_EMAIL_PROVIDER` | Email provider name baked into the image |
| `INPUT_BOT_EMAIL_MODE` | Email adapter mode baked into the image |
| `INPUT_BOT_OPENAI_ENABLED` | Whether the image declares the OpenAI capability |
| `INPUT_AI_PROVIDER` | AI backend provider such as `openai`, `openrouter`, or `cerebras` |
| `INPUT_AI_API_KEY` | Deploy-time API key baked into the immutable runtime profile |
| `INPUT_AI_BASE_URL` | Optional OpenAI-compatible base URL override |
| `INPUT_AI_MODEL` | Default AI model for the `ask` command and browser chat |
| `INPUT_AI_API_MODE` | OpenAI-compatible API path: `chat_completions` by default, optional `responses` |
| `INPUT_CHAT_HISTORY` | Browser-only chat history flag: `true` uses localStorage, `false` uses tab memory only |
| `INPUT_PROXY_SESSION_ENABLED` | Enables the in-memory proxy session UI/API for AI requests |
| `INPUT_PROXY_SESSION_FORCE` | Requires outbound provider traffic to use an active proxy session lease; recommended default is `true` |
| `INPUT_TAG` | Optional explicit image tag instead of `sha-<commit>` |

### Registry inputs

| Input | Meaning |
| --- | --- |
| `INPUT_REGISTRY_HOST` | Registry host for `deploy` |
| `INPUT_IMAGE_REPOSITORY` | Registry repository for `deploy` |
| `INPUT_REGISTRY_USERNAME` | Registry username for `deploy` |
| `INPUT_REGISTRY_PASSWORD` | Registry password for `deploy` |

## Recommended `.paas/config.yml`

```yaml
server: production

defaults:
  INPUT_BOT_NAME: hub-bot
  INPUT_BOT_URL: http://127.0.0.1:5500
  INPUT_BOT_PROFILE_ID: tunnel-email-openai
  INPUT_BOT_DISPLAY_NAME: Tunnel chat + Yandex mail + OpenAI
  INPUT_BOT_HTTP_BIND: 127.0.0.1:5500
  INPUT_BOT_EMAIL_ENABLED: "true"
  INPUT_BOT_EMAIL_PROVIDER: yandex
  INPUT_BOT_EMAIL_MODE: scaffold
  INPUT_BOT_OPENAI_ENABLED: "true"
  INPUT_AI_PROVIDER: openai
  INPUT_AI_BASE_URL: ""
  INPUT_AI_MODEL: gpt-4.1-mini
  INPUT_AI_API_MODE: chat_completions
  INPUT_CHAT_HISTORY: "false"
  INPUT_PROXY_SESSION_ENABLED: "true"
  INPUT_PROXY_SESSION_FORCE: "true"
  INPUT_REGISTRY_HOST: <REGISTRY_HOST>
  INPUT_IMAGE_REPOSITORY: <REGISTRY_NAMESPACE>/<REPOSITORY>
  # INPUT_TAG: ""

extensions_dir: .paas/extensions
```

Keep secrets outside git:

```bash
export INPUT_REGISTRY_USERNAME="<REGISTRY_USERNAME>"
export INPUT_REGISTRY_PASSWORD="<REGISTRY_PASSWORD>"
export INPUT_AI_API_KEY="<AI_PROVIDER_KEY>"
```

The AI provider key remains deploy-time only and is not stored in `bbolt`.
The default AI request path is `chat/completions`, which is often the safest OpenAI-compatible option for providers like OpenRouter or Cerebras.
The recommended default for `INPUT_PROXY_SESSION_FORCE` is `true` so new outbound-capable integrations start from the stricter policy baseline.

## One-time server prerequisites

Prepare the server before running any extension:

```bash
ssh root@<SERVER_HOST>
docker --version
docker compose version
mkdir -p /opt/hub-bot
ls -ld /opt/hub-bot
```

The server must provide:

- Docker Engine and `docker compose`,
- writable runtime directory under `/opt`,
- outbound network access if the selected profile depends on external APIs,
- SSH access for source upload in direct flows.

## Input export workflow

Use `parse_config.sh` to print the exact `export INPUT_*` lines needed for one extension:

```bash
bash ./.paas/parse_config.sh --extension bootstrap-direct
```

Apply the generated exports in the current shell and then run `paas.exe` directly:

```bash
eval "$(bash ./.paas/parse_config.sh --extension bootstrap-direct --no-comments)"
./paas.exe run bootstrap-direct
```

This is the recommended flow on Windows / Git Bash and on Linux/macOS because it makes the resolved values explicit before execution.

The export helper applies this priority:

1. exported `INPUT_*` variables already present in the current shell
2. values from `.paas/config.yml`
3. extension defaults from `inputs:`
4. empty string

## What to review before a run

Check the generated exports for:

- the expected bot name,
- the expected loopback bot URL on port `5500`,
- the expected immutable profile ID,
- the expected AI provider, model, and base URL,
- whether browser chat history should survive reload through `localStorage`,
- unexpected adapter defaults,
- unexpected registry values for `deploy`.

Validate the extensions before the first real run:

```bash
./paas.exe validate bootstrap-direct
./paas.exe validate deploy-direct
./paas.exe validate deploy
```

## `bootstrap-direct`

Use this flow on a fresh server or when reinstalling the hub bot runtime from scratch.

```bash
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/<KEY_FILE>

./paas.exe validate bootstrap-direct
bash ./.paas/parse_config.sh --extension bootstrap-direct
eval "$(bash ./.paas/parse_config.sh --extension bootstrap-direct --no-comments)"
./paas.exe run bootstrap-direct
```

What it does:

1. uploads the tracked repository contents to `/tmp/build-<INPUT_BOT_NAME>` on the server,
2. builds the bot image on the server with immutable profile build args,
3. renders the root `docker-compose.yml` into `/opt/<INPUT_BOT_NAME>/docker-compose.yml`,
4. starts or replaces the runtime with `docker compose up -d --remove-orphans`,
5. waits until `GET /healthz` succeeds,
6. runs a smoke command against `POST /api/command`.

## `deploy-direct`

Use this flow for routine updates without a registry hop.

```bash
./paas.exe validate deploy-direct
bash ./.paas/parse_config.sh --extension deploy-direct
eval "$(bash ./.paas/parse_config.sh --extension deploy-direct --no-comments)"
./paas.exe run deploy-direct
```

What it does:

1. uploads source to the server,
2. builds a new image tag,
3. re-renders `/opt/<INPUT_BOT_NAME>/docker-compose.yml`,
4. runs `docker compose up -d --remove-orphans`,
5. verifies `/healthz`,
6. runs a smoke command through `/api/command`.

## `deploy`

Use this flow when you want the bot image pushed to a registry as part of the update.

```bash
export INPUT_REGISTRY_USERNAME="<REGISTRY_USERNAME>"
export INPUT_REGISTRY_PASSWORD="<REGISTRY_PASSWORD>"

./paas.exe validate deploy
bash ./.paas/parse_config.sh --extension deploy
eval "$(bash ./.paas/parse_config.sh --extension deploy --no-comments)"
./paas.exe run deploy
```

What it does:

1. uploads source to the server,
2. logs in to the registry on the server,
3. builds and tags the image,
4. pushes both `sha-<commit>` and `main`,
5. re-renders `/opt/<INPUT_BOT_NAME>/docker-compose.yml` with the registry image,
6. updates the runtime,
7. verifies `/healthz`,
8. runs a smoke command through `/api/command`.

## First real operator tests

Run these on the server:

```bash
curl http://127.0.0.1:5500/healthz

curl -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'

curl -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"system-info"}'

docker compose -p "${INPUT_BOT_NAME}" -f "/opt/${INPUT_BOT_NAME}/docker-compose.yml" ps
docker ps --format '{{.Names}}'
```

## SSH tunnel UX

Run this on the local workstation:

```bash
ssh -N -L 5500:127.0.0.1:5500 -i ~/.ssh/<KEY_FILE> root@<SERVER_HOST>
```

Then open:

```text
http://127.0.0.1:5500
```

This is the first real browser UX for the bot and should be treated as the primary operator path until richer clients are implemented.

## Troubleshooting

### Bot never becomes ready

Check:

```bash
docker compose -p "${INPUT_BOT_NAME}" -f "/opt/${INPUT_BOT_NAME}/docker-compose.yml" logs --tail 120
ss -ltnp | grep 5500
curl http://127.0.0.1:5500/healthz
```

### SSH upload works but remote bash fails

Print the resolved exports for the target extension:

```bash
bash ./.paas/parse_config.sh --extension <extension>
```

Then apply them explicitly before invoking `paas.exe`.

### Read-only container validation

The compose file enables a read-only root filesystem. Runtime writes must go only to the mounted data directory. If startup fails, inspect the logs for an unexpected write target.
