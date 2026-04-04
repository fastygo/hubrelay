# Profile: tunnel-email-openai

## Intent

Loopback operator control plane with:

- HTTP chat
- email ingress
- OpenAI-compatible AI
- safe system inspection plugins

## Capabilities

- `adapter.http_chat`
- `adapter.email`
- `plugin.system.info`
- `plugin.system.capabilities`
- `plugin.system.audit`
- `ai.chat` / `ai.openai`
- optional `proxy.session`

## Runtime constraints

- no runtime plugin/adapter install
- no Telegram/VK
- no unrestricted shell
- no destructive Docker control

## Deployment expectations

- bind loopback
- mount persistent `bbolt`
- run with read-only root FS
- preserve state across redeploy
- default chat/completions and strict proxy for this profile
