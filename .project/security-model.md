# Security Model

Operator-facing privacy and secret-boundary narrative: [`docs/security-and-privacy`](../docs/security-and-privacy/README.md).

## Default Policy
- immutable capabilities define the maximum runtime surface,
- no raw shell execution in the default profile,
- every external action is audited,
- adapters must map requests to normalized principals before dispatch,
- workload outbound traffic must pass through a shared outbound policy layer.

## Command Classes
### Read-only inspection
- system info,
- capabilities,
- audit visibility.

### Controlled operations
- future state-changing actions with typed inputs,
- optional confirmation before execution.

### High-risk actions
- destructive host changes,
- service restarts,
- Docker mutations.

These are excluded from the first profile unless a later deploy profile explicitly includes them.

## AI Boundary
- the AI layer may select approved commands only,
- AI output must be translated into typed plugin calls,
- free-form shell text is not an execution interface.

## Outbound Boundary
- outbound policy is a workload concern, not an AI-only concern,
- future integrations with APIs, GraphQL endpoints, webhooks, and platforms must not open ad hoc clients without policy resolution,
- strict proxy or future VPN routing should be enforced by the shared outbound layer,
- proxy pool checks are exempt because they are part of routing control-plane logic rather than workload egress.
