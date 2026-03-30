# Profile: tunnel-email-openai

## Intent
Provide an operator-facing control plane that works through:
- loopback HTTP chat intended for SSH tunneling,
- Yandex mail command ingress and response,
- an embedded OpenAI-backed AI layer,
- safe system inspection plugins.

## Built-In Capabilities
- `adapter.http_chat`
- `adapter.email`
- `plugin.system.info`
- `plugin.system.capabilities`
- `plugin.system.audit`
- `ai.chat` (gated on a non-empty deploy-time AI key) and `ai.openai` (provider label)
- `proxy.session` (optional SOCKS session UI/API; compile-time toggles)

## Exclusions
- no runtime adapter installation,
- no Telegram or VK adapter in this profile,
- no unrestricted shell plugin,
- no destructive Docker control by default.

## Runtime Expectations
- bind HTTP chat to loopback only,
- mount `bbolt` data from the host,
- run with read-only root filesystem,
- preserve runtime state across redeploy,
- default AI traffic to OpenAI-compatible `chat/completions`,
- default workload outbound policy to strict proxy enforcement for this profile.
