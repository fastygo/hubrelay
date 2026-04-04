# Egress

HubRelay supports explicit multi-gateway egress selection.

## Configuration

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

Inputs: `INPUT_EGRESS_GATEWAYS`, `INPUT_EGRESS_CHECK_INTERVAL`, `INPUT_PRIVATE_EGRESS_REQUIRED`, `INPUT_PRIVATE_EGRESS_FAIL_CLOSED`, `INPUT_UNIX_SOCKET_ENABLED`, `INPUT_UNIX_SOCKET_PATH`.

If `INPUT_EGRESS_GATEWAYS` is empty, runtime keeps legacy single-interface behavior.

## Health levels

- `unknown`
- `wg`
- `transport`
- `healthy`
- `disabled`

A gateway is healthy only when WG, transport, and business probe checks pass.

## Failover

- lower `priority` wins
- only healthy candidates are used
- unhealthy active gateway triggers promotion to next healthy gateway
- when none are healthy, fail to blackhole mode

## `egress-status`

```bash
curl -s http://127.0.0.1:5500/api/command \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"egress-status"}'
```

The response includes active gateway, health level, check timestamps, and per-gateway state.

## Host-level enforcement

Application checks are not enough on their own. OS enforcement is described in:

- [`.project/wg/os-enforcement.md`](../.project/wg/os-enforcement.md)
