# WireGuard private egress for HubRelay

HubRelay runs as a loopback HTTP API (`127.0.0.1`) and is accessed by operators through SSH forwarding.
Private egress uses host-level WireGuard routing (`wg0`) on the HubRelay host; Marvin is the gateway node.

## Traffic model

```text
operator → ssh -L 5500:127.0.0.1:5500 → HubRelay (jarvis) → wg0 → Marvin → AI provider
```

## Core intent

- No public HubRelay command port.
- `ask` follows policy: direct vs proxy path, private egress validation when required.
- `Marvin` is a network gateway only; it never runs the HubRelay process.

## Baseline deployment inputs

```bash
INPUT_PRIVATE_EGRESS_REQUIRED=true
INPUT_PRIVATE_EGRESS_INTERFACE=wg0
INPUT_PRIVATE_EGRESS_TEST_HOST=10.88.0.1
INPUT_PRIVATE_EGRESS_FAIL_CLOSED=true
INPUT_PROXY_SESSION_ENABLED=false
INPUT_PROXY_SESSION_FORCE=false
```

## First checks

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'

wg show
ip -br a show wg0
ip rule show
ip route show table 88
systemctl status hubrelay.service --no-pager
```

## Common failures and quick fix path

- `ask` fails immediately: confirm private egress settings and `INPUT_PRIVATE_EGRESS_*`.
- Tunnel appears up, but no outbound traffic: inspect fwmark/table mapping and `ip route get` for `uid`.
- SSH/API is accessible but no commands: verify service is bound to loopback and restarted after env changes.

```bash
systemctl restart hubrelay.service
systemctl restart wg-quick@wg0
```

## Read next

- [HUBRELAY_SYSTEM_GUIDE.md](HUBRELAY_SYSTEM_GUIDE.md)
- [egress-gateway/README.md](egress-gateway/README.md)
- [Deploy model](../deploy-model.md)

