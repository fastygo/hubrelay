# Deploy

## Strategy

Deployments produce **one immutable image** per profile. Adapters are not installed at runtime.

**Why**: predictable blast radius and reviewable artifacts.

## PaaS tooling (`.paas`)

This repository ships **extension definitions** consumed by an external `paas` CLI:

| Extension | When to use |
| --- | --- |
| `bootstrap-direct` | Fresh host or full reinstall; upload source, build on server |
| `deploy-direct` | Routine update; rebuild on server |
| `deploy` | Build and push via container registry |

Authoritative tables of `INPUT_*` variables and workflows: [`.paas/README.md`](../../.paas/README.md).

Sanitised operator checklist (no real hosts): [`.paas/.check-deploy.md`](../../.paas/.check-deploy.md).

## Typical inputs (conceptual)

Values are **examples only**; set real ones via your secrets manager or shell exports, never commit them.

| Input group | Meaning |
| --- | --- |
| `INPUT_BOT_*` | Compose project name, loopback URL, bind address, profile id, display name |
| `INPUT_BOT_EMAIL_*` | Email adapter compile-time flags |
| `INPUT_BOT_OPENAI_ENABLED` | Whether AI capability is compiled in |
| `INPUT_AI_*` | Provider label, **API key**, optional base URL, model, API mode |
| `INPUT_CHAT_HISTORY` | Browser persistence mode |
| `INPUT_PROXY_SESSION_*` | Enable SOCKS UI/API; force all AI egress through a lease |
| Registry inputs | Only for `deploy` (`INPUT_REGISTRY_*`, repository, tag) |

**Why `INPUT_PROXY_SESSION_FORCE` defaults to true in config examples**: starts from a stricter egress baseline; turn off only when you accept direct server egress.

## Docker and read-only root

The compose file targets **read-only container root** with a **writable volume** for BoltDB.

**Why**: container filesystem is disposable; state survives on the host mount.

## Verification after deploy

On the server (loopback):

```bash
curl -s http://127.0.0.1:5500/healthz
# POST capabilities — see .paas/README.md
```

Through **SSH tunnel** from your laptop: same URLs against `http://127.0.0.1:5500`.

## Local parity

Before touching production, run [Local testing](../local-testing/README.md) and `provider-smoke` with the **same** base URL and model you will bake into the image.

## Environment quirks (Windows)

When invoking `paas.exe`, polluted environments (for example paths with parentheses) can break shell parsing. A minimal **`env -i`** wrapper with only required variables is often safer—see patterns in `.paas/README.md` and `.paas/.check-deploy.md`.
