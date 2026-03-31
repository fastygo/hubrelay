# Documentation index

This folder is the **canonical English guide** for operators and contributors. It explains **why** design choices exist, not only **what** to run.

| Section | Purpose |
| --- | --- |
| [Overview](overview/README.md) | Problem space, goals, non-goals |
| [Getting started](getting-started/README.md) | Go toolchain, clone, build, first run |
| [Architecture](architecture/README.md) | How the bot is structured; link to `.project` |
| [Local testing](local-testing/README.md) | Provider smoke CLI, local bot, verification curls |
| [Providers and AI](providers-and-ai/README.md) | OpenAI-compatible endpoints, models, base URL pitfalls |
| [Network, tunnel, and proxy](network-tunnel-and-proxy/README.md) | SSH loopback access, SOCKS sessions, outbound policy |
| [SDK](sdk.md) | Go client package for HTTP and unix socket access |
| [Egress](egress.md) | Multi-egress configuration, health model, and failover behavior |
| [Security and privacy](security-and-privacy/README.md) | Secrets boundary, sensitive-data handling, safe docs practice |
| [Deploy](deploy/README.md) | PaaS extensions, inputs, read-only containers |
| [Operations](operations/README.md) | Health checks, logs, common failures |

**Project design notebooks** (shorter ADRs and contracts) live under [`.project`](../.project/README.md).

**Deploy automation** details stay next to the tooling in [`.paas/README.md`](../.paas/README.md). This repo also keeps a sanitised deploy checklist at [`.paas/.check-deploy.md`](../.paas/.check-deploy.md) and local smoke steps at [`.paas/.check-smoke.md`](../.paas/.check-smoke.md).

---

## No sensitive data in documentation

- Do not commit real API keys, hostnames you consider private, or personal SSH paths.
- Use placeholders such as `<SERVER_HOST>`, `<SSH_PRIVATE_KEY>`, `<YOUR_AI_API_KEY>`.
- Treat `capabilities` output as the source of truth for whether a key is present (`ai_has_api_key`), never paste secrets into tickets.
