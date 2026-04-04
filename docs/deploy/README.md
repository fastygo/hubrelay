# Deploy

The repository deploys **one standalone binary per module**.

- HubRelay runtime: `go build ./cmd/bot`
- Dashboard: `go build ./apps/dashboard/cmd/server`
- `hubcore` is linked at build time and is not a runtime process.

Deployment orchestration is kept in `.paas` as private automation:

- `.paas/config.yml` contains default `INPUT_*` values.
- `.paas/extensions/*.yml` define required/optional inputs per scenario.
- `.paas/parse_config.sh` renders all effective environment variables for a scenario.
- `.paas/deploy-hostrun-clean.sh` and `.paas/deploy-app-clean.sh` perform end-to-end deploys.

## PaaS entry points

- PaaS extensions: `bootstrap-direct`, `deploy-direct`, `deploy`
- Authoritative inputs, env, and command contracts:
  - [`.paas/README.md`](../../.paas/README.md)
  - `.paas/deploy-hostrun-clean.sh` (bot deployment)
  - `.paas/deploy-app-clean.sh` (dashboard deployment)
  - `.paas/parse_config.sh` (render `INPUT_*` env for both scenarios)
- Operator runbooks:
  - [`.paas/.check-deploy.md`](../../.paas/.check-deploy.md)
  - [`.paas/.check-smoke.md`](../../.paas/.check-smoke.md)

Example: dump effective deployment env (without comments):

```bash
bash ./.paas/parse_config.sh --extension deploy-hostrun --no-comments
```

## Before packaging

```bash
go test ./...
go test ./hubcore/...
go test ./sdk/hubrelay/...
go test ./apps/dashboard/...
```

```bash
go test ./cmd/bot
go test ./apps/dashboard/cmd/server
```

## Post-deploy checks

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

## Related docs

- [Local testing](../local-testing/README.md)
- [Operations](../operations/README.md)
- `.project/README.md`
