# OS-Level Egress Enforcement Runbook

## Purpose

This runbook defines the host-level controls that make HubRelay fail closed when no approved WireGuard egress path is healthy.

The application-level egress manager decides which gateway is active.
The operating system enforces that the `hubrelay` service user cannot silently fall back to the public WAN.

## Scope

This document applies to the HubRelay host only.

It covers:
- service-account isolation,
- policy routing by UID,
- `nftables` kill-switch rules,
- DNS restrictions,
- blackhole fallback behavior,
- safe validation commands.

## Security Intent

The required security posture is:
- HubRelay listens only on loopback or unix socket.
- HubRelay outbound traffic is tagged by service user identity, not by process goodwill.
- Approved egress is limited to:
  - `lo`,
  - `wg-*` interfaces,
  - UDP packets needed to establish WireGuard tunnels to known peer endpoints.
- When no healthy gateway exists, outbound traffic must fail closed.

## Service User

Create a dedicated local account for the HubRelay process:

```bash
useradd --system --home /nonexistent --shell /usr/sbin/nologin hubrelay
id hubrelay
```

Expected result:
- HubRelay runs as its own UID.
- Policy routing can target that UID only.
- Other processes on the host are not forced through the same routing table.

## systemd Unit Expectations

The service should run as the dedicated user:

```ini
[Service]
User=hubrelay
Group=hubrelay
WorkingDirectory=/opt/hubrelay
ExecStart=/usr/local/bin/hubrelay
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/hubrelay /run/hubrelay
RuntimeDirectory=hubrelay
RuntimeDirectoryMode=0750
```

Important notes:
- `RuntimeDirectory` is useful for the unix socket path.
- `ReadWritePaths` should be narrowed to the minimum required runtime directories.
- Do not run HubRelay as `root`.

## Policy Routing

Reserve a dedicated routing table for HubRelay traffic.

Example:

```bash
echo "88 hubrelay" >> /etc/iproute2/rt_tables
ip rule add uidrange 1001-1001 lookup 88
ip route add blackhole default table 88
```

Replace `1001` with the actual UID returned by `id -u hubrelay`.

Design intent:
- all traffic from the `hubrelay` user resolves in table `88`,
- the default route in that table is `blackhole`,
- only explicit routes added for healthy `wg-*` interfaces can carry outbound traffic.

## Per-Gateway Routes

When a gateway is healthy, add explicit routes for the provider or approved destination prefixes into table `88`.

Example for `wg-b1`:

```bash
ip route replace 0.0.0.0/0 dev wg-b1 table 88
```

For stricter setups, prefer destination allowlists instead of a full default route:

```bash
ip route replace 203.0.113.0/24 dev wg-b1 table 88
ip route replace 198.51.100.0/24 dev wg-b1 table 88
```

If no gateway is healthy:

```bash
ip route replace blackhole default table 88
```

## WireGuard Endpoint Exception

Do not send WireGuard encapsulation traffic itself into table `88`.

The host must keep a route in the main table for the public peer endpoints, otherwise the tunnel can loop or stall.

Typical requirement:
- UDP to each WireGuard peer public IP and port stays reachable through the normal host uplink,
- only the post-decryption HubRelay application traffic is forced through `wg-*`.

## nftables Kill Switch

Create a dedicated output filter for the `hubrelay` UID.

Example policy:

```nft
table inet hubrelay_egress {
  set wg_peer_ipv4 {
    type ipv4_addr
    elements = { 198.51.100.10, 198.51.100.11 }
  }

  chain output {
    type filter hook output priority 0;
    policy accept;

    meta skuid != 1001 return

    oifname "lo" accept
    oifname "wg-*" accept

    ip daddr @wg_peer_ipv4 udp dport 51820 accept

    udp dport 53 drop
    tcp dport 53 drop

    reject with icmpx type admin-prohibited
  }
}
```

Key properties:
- loopback remains available,
- traffic already routed to `wg-*` is allowed,
- UDP tunnel maintenance to WireGuard peer endpoints is allowed,
- direct DNS leaks are blocked,
- all other outbound traffic from the HubRelay UID is rejected.

## DNS Policy

Do not allow arbitrary DNS resolution from the `hubrelay` user.

Recommended options:
1. Use a resolver reachable only through the approved path.
2. Pin `/etc/resolv.conf` or `systemd-resolved` to controlled resolvers.
3. Block port `53` outside explicit allowed resolvers in `nftables`.

If DNS must stay local:
- allow only the local stub resolver,
- make sure the stub itself forwards through approved resolvers,
- verify that no fallback public DNS is configured.

## Blackhole Mode

Blackhole mode is the required safe state when:
- no WireGuard gateway is healthy,
- health checks are stale,
- the route controller cannot determine the active gateway.

Operational behavior:
- table `88` keeps `blackhole default`,
- `nftables` still blocks direct WAN escape,
- HubRelay requests fail quickly instead of leaking over the public interface.

## Validation Commands

### Service identity

```bash
id hubrelay
systemctl status hubrelay.service --no-pager
systemctl cat hubrelay.service
```

### Routing

```bash
ip rule show
ip route show table 88
ip route get 1.1.1.1 uid $(id -u hubrelay)
ip route get <provider_ip> uid $(id -u hubrelay)
```

### Firewall

```bash
nft list ruleset
nft list table inet hubrelay_egress
```

### WireGuard

```bash
wg show
ip -br a show type wireguard
```

### Runtime probes

```bash
curl -s http://127.0.0.1:5500/healthz
curl -s http://127.0.0.1:5500/api/command \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"egress-status"}'
```

## Failure Patterns

### HubRelay can reach the internet when WireGuard is down

This means the kill switch is incomplete.

Check:
- `ip rule` for the HubRelay UID,
- default route in table `88`,
- `nftables` output hook for `meta skuid`,
- whether HubRelay is really running as the `hubrelay` user.

### WireGuard is up but provider calls still fail

Check:
- active route in table `88`,
- allowed provider prefixes or default route for the active `wg-*`,
- DNS resolution path,
- NAT and forwarding on the egress gateway.

### DNS works over WAN despite policy

Check:
- stub resolver behavior,
- TCP/UDP port `53` rules,
- any resolver exceptions added for the HubRelay UID.

## Operational Rule

The application decides *which* gateway should be used.
The OS decides *what is impossible*.

If there is any disagreement between them, prefer a host configuration that drops traffic rather than one that permits fallback.
