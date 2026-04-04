# HubRelay system guide (WireGuard private egress)

Concise view of process flow, profile config, and egress enforcement.

## Runtime model

`cmd/bot/main.go` is one process:

1. load `buildprofile.Current()`
2. open `BOT_DB_FILE`
3. init adapters/plugins
4. apply outbound policy
5. serve loopback command API (`127.0.0.1` default)

Capabilities are immutable for process life.

## Config precedence

- `-ldflags` profile values
- one-time `INPUT_*` bootstrap at first startup
- `BOT_DB_FILE` direct env read (default `data/bot.db`)

Required runtime knobs:

- `INPUT_AI_*`
- `INPUT_PRIVATE_EGRESS_*`
- `INPUT_PROXY_SESSION_*`

Restart required after changes.

## Command flow

`HTTP API -> core -> plugin -> outbound policy -> provider`

Streaming path:

- `POST /api/command` (JSON)
- `POST /api/command/stream` (SSE)

## Private egress check

Before `ask` when required:

- interface exists and is up (`wg0` expected)
- probe route resolves through interface-address space for `test_host`
- fail-closed mismatch returns error before provider call

## WireGuard vs proxy session

- `INPUT_PROXY_SESSION_*` = optional SOCKS proxy control in app
- WireGuard = host routing requirement

Recommended baseline:

```bash
INPUT_PROXY_SESSION_ENABLED=false
INPUT_PROXY_SESSION_FORCE=false
INPUT_PRIVATE_EGRESS_REQUIRED=true
INPUT_PRIVATE_EGRESS_INTERFACE=wg0
INPUT_PRIVATE_EGRESS_TEST_HOST=10.88.0.1
INPUT_PRIVATE_EGRESS_FAIL_CLOSED=true
```

## Operator path

```bash
ssh -N -L 5500:127.0.0.1:5500 <operator>@<jarvis_public_ip>
curl -s http://127.0.0.1:5500/healthz
```

## Host deploy

- binary: `/opt/hubrelay/bot`
- env file: `/opt/hubrelay/env`
- db: `BOT_DB_FILE=/var/lib/hubrelay/bot.db`
- unit: systemd, user `hubrelay`, `After=wg-quick@wg0`

Restart on change:

```bash
systemctl restart hubrelay.service
```

## Checks

```bash
curl -s -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'

wg show
ip rule show
ip route show table 88
systemctl status hubrelay.service --no-pager
ss -ltnup | grep 5500
```

## Related

- `.project/wg/README.md`
- `.project/wg/egress-gateway/README.md`
- `docs/network-tunnel-and-proxy/README.md`
- `.project/outbound-model.md`

