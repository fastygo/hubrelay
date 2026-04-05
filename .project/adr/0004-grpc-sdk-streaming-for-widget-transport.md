# ADR 0004: gRPC transport for dashboard SDK widget paths

## Status

Accepted

## Context

`apps/dashboard` currently consumes bot data through `sdk/hubrelay` and HTTP (`/api/command`, `/api/command/stream`).
For pointwise dashboard widgets and metric surfaces, the team needs a strict, typed service transport without rewriting all browser pages to websocket.

## Decision

Introduce gRPC as a secondary transport for non-browser service integration while keeping HTTP as the default/baseline contract.

The dashboard must stay transport-agnostic through:

- `apps/dashboard/internal/relay` (SDK client backend)
- `apps/dashboard/internal/source` (command/query abstraction)
- existing browser UX entry points for widget-level pages

Concretely:

1. Keep HTTP JSON compatibility for `/api/command` and `/api/command/stream`.
2. Add gRPC client/server support in `sdk/hubrelay` and bot service.
3. Add transport selection by config so BFF can switch between HTTP, Unix socket, and gRPC.
4. Ensure ask streaming remains served to browser as SSE (or equivalent page-level stream adapter) while gRPC is used under the hood where applicable.

## Consequences

- Existing integrations remain stable due to HTTP baseline compatibility.
- Dashboard widgets and metrics can use typed transport with server-stream semantics per endpoint.
- gRPC rollout remains scoped: no full-page websocket migration required in v1.
- Future consumers (besides dashboard) can reuse the same gRPC contract directly via SDK.
