# Marvin egress gateway

`Marvin` is the dedicated WireGuard + NAT host for HubRelay private egress.
It accepts tunnel traffic from Jarvis peers and forwards only as configured in host policies.

## Runtime target

- UDP WireGuard on `51820`
- `wg0` peer interface
- forwarding enabled and NAT on tunnel subnet
- service boot: `wg-quick@wg0`

## Quick setup state

```bash
systemctl status wg-quick@wg0
wg show
ip -br a
ip -br a show wg0
sysctl net.ipv4.ip_forward
```

## Core checks

```bash
wg show
systemctl status nftables --no-pager
nft list ruleset
ss -lunp | grep 51820
```

## Validate host egress

```bash
ip route get 10.88.0.1
curl -s http://ifconfig.me
```

Expected: active tunnel interface, incrementing traffic counters, and NAT path for tunnel subnet.

## Failure map

- no handshake: verify peer keys, endpoint, UDP 51820 reachability
- handshake exists, no tunnel traffic: check AllowedIPs + forwarding + NAT rules
- peer can reach gateway but not Internet: check masquerade and public uplink rules

## Safe operations

```bash
systemctl restart wg-quick@wg0
systemctl restart nftables
journalctl -u wg-quick@wg0 -n 100 --no-pager
journalctl -u nftables -n 100 --no-pager
```

## Related

- [HUBRELAY_SYSTEM_GUIDE.md](../HUBRELAY_SYSTEM_GUIDE.md)
- [.project/wg/README.md](../README.md)

