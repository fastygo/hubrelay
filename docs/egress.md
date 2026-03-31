# Egress

HubRelay now supports explicit multi-egress configuration instead of a single unnamed WireGuard path.

## Configuration

Multi-egress gateways are provided through `INPUT_EGRESS_GATEWAYS` as a JSON array:

```json
[
  {
    "name": "wg-b1",
    "interface": "wg-b1",
    "test_host": "10.88.0.1",
    "probe_url": "https://api.example.com/v1/models",
    "priority": 10,
    "enabled": true
  },
  {
    "name": "wg-b2",
    "interface": "wg-b2",
    "test_host": "10.89.0.1",
    "probe_url": "https://api.example.com/v1/models",
    "priority": 20,
    "enabled": true
  }
]
```

Related inputs:
- `INPUT_EGRESS_CHECK_INTERVAL`
- `INPUT_PRIVATE_EGRESS_REQUIRED`
- `INPUT_PRIVATE_EGRESS_FAIL_CLOSED`
- `INPUT_UNIX_SOCKET_ENABLED`
- `INPUT_UNIX_SOCKET_PATH`

If `INPUT_EGRESS_GATEWAYS` is empty, HubRelay keeps the legacy single-interface private-egress behavior.

## Health Model

Each configured gateway is healthy only when all three levels pass:

1. WG: interface exists, is up, and the test host resolves through that interface.
2. Transport: TCP connect and TLS handshake succeed through the selected interface.
3. Business: a safe HTTP probe to the configured endpoint succeeds.

Health levels exposed by `egress-status`:
- `unknown`
- `wg`
- `transport`
- `healthy`
- `disabled`

## Failover

Selection rules:
- lower `priority` wins,
- only healthy gateways are candidates,
- when the active gateway degrades, the manager promotes the next healthy gateway,
- if none are healthy, the result is blackhole mode.

Blackhole mode means:
- no active gateway is returned by the egress manager,
- application checks fail closed,
- the host runbook should keep table `88` on `blackhole default`.

## Observability

Use the `egress-status` command to inspect runtime state:

```bash
curl -s http://127.0.0.1:5500/api/command \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"egress-status"}'
```

The result includes:
- all configured gateways,
- active gateway flag,
- health level,
- last check time,
- last transition time,
- per-level status for WG, transport, and business probes.

Audit entries also include `egress_gateway` when a command is dispatched while a gateway is active.

## OS Enforcement

Application checks are only part of the control model.

For host-level enforcement, see:
- [`.project/wg/os-enforcement.md`](../.project/wg/os-enforcement.md)

That runbook covers:
- dedicated `hubrelay` user,
- UID-based routing,
- `nftables` kill switch,
- DNS restrictions,
- blackhole fallback routing.
