# Documentation index

Canonical English documentation for operators and contributors.

| Section | Purpose |
| --- | --- |
| [Overview](overview/README.md) | Core invariants and module layout |
| [Getting started](getting-started/README.md) | Toolchain, first run, smoke and local checks |
| [Architecture](architecture/README.md) | Boundaries, modules, and runtime flow |
| [Local testing](local-testing/README.md) | Smoke, bot checks, and dashboard checks |
| [Providers and AI](providers-and-ai/README.md) | Provider config and model/path behavior |
| [Network, tunnel, and proxy](network-tunnel-and-proxy/README.md) | SSH access and SOCKS session model |
| [SDK](sdk.md) | Go client usage |
| [Egress](egress.md) | Multi-egress config, health, and failover |
| [Streaming](streaming.md) | SSE command output contract |
| [Security and privacy](security-and-privacy/README.md) | Secrets and logging boundaries |
| [Deploy](deploy/README.md) | Build units and deployment checks |
| [Operations](operations/README.md) | Smoke checks, logs, escalation |

## Monorepo map

- Core runtime: `cmd/bot` in `sshbot`
- Client SDK: `sdk/hubrelay`
- Shared dashboard library: `hubcore` (import-time only)
- Dashboard service: `apps/dashboard`
- UI kit: `apps/dashboard/ui8kit`

## Deployment links

- [`.paas/README.md`](../.paas/README.md)
- [`.paas/.check-deploy.md`](../.paas/.check-deploy.md)
- [`.paas/.check-smoke.md`](../.paas/.check-smoke.md)
