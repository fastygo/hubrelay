# Architecture

Monorepo is organized as explicit module boundaries.

## Module model

- `sshbot` root module: core runtime (`go run ./cmd/bot`).
- `sdk/hubrelay`: typed client module.
- `hubcore`: shared dashboard library (import-time only).
- `apps/dashboard`: dashboard service (`go run ./apps/dashboard/cmd/server`).
- `apps/dashboard/ui8kit`: reusable UI primitives.

## Library + binary contract

- Shared modules are libraries.
- Every executable imports them at build time.
- Each service is one standalone binary.

```text
hubrelay daemon        -> cmd/bot
dashboard service      -> apps/dashboard/cmd/server
```

## Request flow

`adapter -> CommandEnvelope -> core service -> plugin -> outbound policy -> external target`

## Core contracts

- Principal: transport-independent actor identity.
- CommandEnvelope: normalized command request.
- CommandResult: typed response + status flags.
- Plugin: capability-gated command handler.
- Adapter: transport boundary to command envelope.
- OutboundPolicy: shared egress control.
- AuditEntry: immutable execution record.

## Operational map

- Detailed notes: [`.project/architecture.md`](../../.project/architecture.md)
- Outbound policy contracts: [`.project/outbound-model.md`](../../.project/outbound-model.md)
- Security model: [`.project/security-model.md`](../../.project/security-model.md)
