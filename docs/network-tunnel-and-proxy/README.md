# Network, SSH tunnel, and SOCKS proxy

## Loopback bind

Default daemon HTTP bind is loopback (`127.0.0.1` unless overridden by `INPUT_BOT_HTTP_BIND`).

## SSH local forward

```bash
ssh -N -L 5500:127.0.0.1:5500 -i <SSH_PRIVATE_KEY> user@<SERVER_HOST>
```

Call API locally at port `5500`.

## Operator path

```
Client → SSH tunnel → HubRelay API → plugin ask → AI provider
```

## Proxy session flow

- operators send SOCKS list through API
- bot probes and binds a sticky session lease
- proxy required only when `INPUT_PROXY_SESSION_FORCE=true`

## Related

- [Deploy inputs for proxy](../deploy/README.md)
- [Outbound model](../../.project/outbound-model.md)
