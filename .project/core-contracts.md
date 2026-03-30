# Core Contracts

See also the English operator guide [`docs/`](../docs/README.md) for how these contracts show up in HTTP and deploy flows.

## Principal
Normalized actor identity independent from transport.

Fields:
- `id`
- `display`
- `transport`
- `roles`
- `metadata`

## CommandEnvelope
Transport-neutral execution request.

Fields:
- `id`
- `transport`
- `name`
- `args`
- `raw_text`
- `principal`
- `requested_at`

## CommandResult
Normalized plugin response.

Fields:
- `status`
- `message`
- `data`
- `requires_confirm`

## Plugin
A capability-gated handler for one command name.

Required behaviors:
- declare command name,
- declare required capabilities,
- execute with typed context,
- never assume direct transport details.

## OutboundPolicy
A shared workload egress decision layer.

Required behaviors:
- accept transport-neutral outbound intent,
- enforce immutable profile restrictions,
- resolve whether traffic is `direct`, `proxy`, or `blocked`,
- avoid binding future integrations to AI-specific rules.

## LeaseResolver
A narrow contract that resolves workload egress routing state for a given session identifier.

Required behaviors:
- resolve the active lease for a session,
- return an explicit error when no lease is available,
- stay independent from plugin or adapter business logic.

## Adapter
A transport boundary that:
- authenticates or maps a principal,
- converts external messages into `CommandEnvelope`,
- returns `CommandResult` to the source channel.

## AuditEntry
Immutable execution record written for successful, rejected, and failed commands.
