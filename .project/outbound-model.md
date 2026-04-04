# Outbound Model

Runtime policy is shared for all workload egress, not per-plugin.

## Current minimal contracts

- `Policy`
- `LeaseResolver`
- `ProxyLeaseResolver`

## Policy inputs

- `principal_id`
- `workload_id`
- `capability`
- `target_kind`
- `target_name`
- `proxy_session_id`
- `transport`

## Decision output

- `mode: direct | proxy | blocked`
- `proxy_address`
- `reason`

## Resolution order

1. validate immutable profile capability
2. check if proxy is required
3. resolve proxy session if needed
4. return a single shared routing decision

## Current consumers

- AI provider calls (`ask`)

## Hard rule

No outbound-capable module may bypass `internal/outbound`.
