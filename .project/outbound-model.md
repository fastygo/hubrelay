# Outbound Model

## Intent
The bot is evolving from a chat-first operator assistant into a hub that may call external APIs, SaaS platforms, internal services, and later VPN-backed routes.
Because of that, outbound policy must live above any single plugin such as `ask`.

## Boundary
- proxy health checks are a separate control-plane concern and are not governed by normal workload outbound rules,
- workload outbound traffic must be routed through a shared policy layer,
- future adapters, plugins, and workflow actions must not open direct HTTP clients without consulting that layer.

## Current Minimal Runtime Shape
Today the shared runtime shape is represented by:
- `Policy`
- `LeaseResolver`
- `ProxyLeaseResolver`

This is intentionally small so the first AI integration stays simple while future integrations can adopt the same contract.

## Recommended Interface Direction

### `OutboundRequest`
Represents a workload egress decision request.

Suggested fields:
- `principal_id`
- `workload_id`
- `capability`
- `target_kind`
- `target_name`
- `proxy_session_id`
- `transport`

Notes:
- `target_kind` can later distinguish `ai_provider`, `http_api`, `webhook`, `smtp`, `imap`, `graphql`, `websocket`.
- `capability` ties the outbound decision back to immutable deploy profile scope.

### `ResolvedEgress`
Represents the routing decision returned by the shared policy layer.

Suggested fields:
- `mode`
- `proxy_address`
- `reason`

Expected values:
- `mode=direct`
- `mode=proxy`
- `mode=blocked`

## Recommended Go Contracts
The current codebase already has a small base contract in `internal/outbound`.
The next interface shape should grow toward:

```go
type LeaseResolver interface {
    ResolveProxyAddress(sessionID string) (string, error)
}

type Policy interface {
    Resolve(OutboundRequest) (ResolvedEgress, error)
}

type ClientFactory interface {
    HTTPClient(OutboundRequest) (*http.Client, ResolvedEgress, error)
}
```

## Resolution Rules
1. Validate immutable profile policy.
2. Check whether outbound for this workload requires a proxy.
3. If proxy is required, resolve the workload lease from `proxy_session_id`.
4. Return a routing decision instead of letting each plugin improvise transport rules.

## Why This Layer Matters
- keeps `ask` from becoming the hidden owner of network policy,
- gives future workflow nodes the same enforcement path,
- allows proxy to be replaced later with VPN or another egress mechanism,
- makes audit and policy reasoning possible without binding everything to one provider SDK.

## Near-Term Adoption
The first consumers should be:
- AI provider calls,
- future HTTP API plugins,
- future GraphQL or webhook actions.

The rule should be:
- no new outbound-capable module bypasses `internal/outbound`.
