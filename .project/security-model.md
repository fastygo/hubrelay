# Security Model

Operational security expectations for HubRelay and docs-only docs.

## Default policy

- immutable capabilities drive exposed profile
- no raw shell in default profile
- principals are normalized before command dispatch
- every external action is audited
- outbound follows shared policy layer

## Command classes

- Read-only: `system.info`, `capabilities`, `audit`
- Controlled: typed state changes with confirmation where needed
- High-risk: destructive ops; only in explicit profiles

## AI boundary

AI output must map to typed plugin calls.
No free-form shell execution.

## Outbound boundary

- shared policy before provider/API/webhook calls
- optional SOCKS path is profile-controlled
- proxy health checks remain control-plane concern
