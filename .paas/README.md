# HubRelay PAAS Deployment

This `.paas` package supports two current deployment scenarios:

- `deploy-hostrun` deploys the bot runtime to `/opt/hubrelay` as `hubrelay.service`
- `deploy-app` deploys the dashboard runtime to `/opt/hubrelay-dashboard` as `hubrelay-dashboard.service`

Both flows depend on `.paas/parse_config.sh`. Keep that file in the repository.

## Key files

- `config.yml` defines shared `INPUT_*` defaults
- `extensions/deploy-hostrun.yml` deploys the bot binary and runtime env
- `extensions/deploy-app.yml` deploys the dashboard binary and static assets
- `deploy-hostrun-clean.sh` runs the bot deploy through `env -i`
- `deploy-app-clean.sh` runs the dashboard deploy through `env -i`

## Scenario 1: deploy the bot

Use this when you need the HubRelay bot API and optional gRPC adapter.

```bash
export HUBRELAY_HOST='176.124.209.3'
export HUBRELAY_USER='root'
export HUBRELAY_SSH_KEY='C:/Users/alexe/.ssh/appserv'

export INPUT_AI_API_KEY='<OPENAI_API_KEY>'
export INPUT_AI_BASE_URL='https://api.cerebras.ai/v1'
export INPUT_AI_MODEL='gpt-oss-120b'

# optional gRPC transport
export INPUT_GRPC_ENABLED='true'
export INPUT_GRPC_BIND='0.0.0.0:5501'

bash ./.paas/deploy-hostrun-clean.sh
```

What the script does:

- resolves `INPUT_*` through `parse_config.sh`
- builds and uploads `./cmd/bot` to `/opt/hubrelay/bot`
- refreshes `hubrelay.service`
- verifies HTTP health on `INPUT_BOT_URL`
- verifies the gRPC listener when `INPUT_GRPC_ENABLED=true`
- runs a gRPC `system-info` smoke call through an SSH tunnel when gRPC is enabled

Verify after deploy:

```bash
ssh -i "$HUBRELAY_SSH_KEY" "${HUBRELAY_USER}@${HUBRELAY_HOST}"
systemctl status hubrelay.service --no-pager
curl http://127.0.0.1:5500/healthz
curl -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

SSH tunnel access from your workstation:

```bash
ssh -N \
  -L 5500:127.0.0.1:5500 \
  -L 5501:127.0.0.1:5501 \
  -i "$HUBRELAY_SSH_KEY" \
  "${HUBRELAY_USER}@${HUBRELAY_HOST}"
```

After the tunnel is open:

- HTTP API is available at `http://127.0.0.1:5500`
- gRPC target is available at `127.0.0.1:5501`
- example gRPC smoke command: `go run ./cmd/grpc-system-info --target 127.0.0.1:5501`

## Scenario 2: deploy the dashboard

Use this when you need the admin UI. The bot must already be deployed.

```bash
export HUBRELAY_HOST='176.124.209.3'
export HUBRELAY_USER='root'
export HUBRELAY_SSH_KEY='C:/Users/alexe/.ssh/appserv'

export APP_HOST="$HUBRELAY_HOST"
export APP_USER="$HUBRELAY_USER"
export APP_SSH_KEY="$HUBRELAY_SSH_KEY"

export INPUT_APP_ADMIN_PASS='<CHANGE_ME>'

# choose one upstream mode
export INPUT_HUBRELAY_TRANSPORT='http'
export INPUT_HUBRELAY_BASE_URL='http://127.0.0.1:5500'

# or
# export INPUT_HUBRELAY_TRANSPORT='grpc'
# export INPUT_HUBRELAY_GRPC_TARGET='127.0.0.1:5501'

bash ./.paas/deploy-app-clean.sh
```

What the script does:

- builds `apps/dashboard/cmd/server` locally
- uploads the dashboard binary and static assets
- renders `dashboard.env`
- restarts `hubrelay-dashboard.service`
- verifies `/login`, root redirect, static assets, and authenticated pages
- verifies the configured HubRelay upstream path for `http`, `grpc`, or `unix`

Verify after deploy:

```bash
ssh -i "$APP_SSH_KEY" "${APP_USER}@${APP_HOST}"
systemctl status hubrelay-dashboard.service --no-pager
curl -I http://127.0.0.1:8080/login
curl -sS -u "${INPUT_APP_ADMIN_USER:-admin}:${INPUT_APP_ADMIN_PASS}" \
  http://127.0.0.1:8080/capabilities
```

SSH tunnel access to the dashboard:

```bash
ssh -N -L 18080:127.0.0.1:8080 -i "$APP_SSH_KEY" "${APP_USER}@${APP_HOST}"
```

Then open `http://127.0.0.1:18080/login`.

## Important notes

- `deploy-hostrun` and `deploy-app` are independent deploy flows with different artifacts
- keep `INPUT_APP_ADMIN_PASS` outside git and pass it only through env
- when WireGuard is not needed, keep `INPUT_BOT_APP_WG_ENABLED=false`
- gRPC is optional for the bot, but when enabled it should be tunnelled the same way as HTTP for private access
