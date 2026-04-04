# OS egress enforcement for HubRelay

Host-level enforcement prevents fallback outside approved WG path.
The app selects active gateway; the OS enforces what is allowed.

## Scope

- service identity (`hubrelay` user)
- policy routing by UID
- nftables output guardrail
- WireGuard endpoint exception
- DNS restrictions
- validation and failure triage

## Service user and systemd hardening

```bash
useradd --system --home /nonexistent --shell /usr/sbin/nologin hubrelay
```

Run service as dedicated user:

```ini
[Service]
User=hubrelay
Group=hubrelay
ProtectSystem=strict
NoNewPrivileges=true
RuntimeDirectory=hubrelay
ReadWritePaths=/var/lib/hubrelay /run/hubrelay
```

## Policy routing

Default table for service traffic must be safe by default.

```bash
echo "88 hubrelay" >> /etc/iproute2/rt_tables
ip rule add uidrange <hubrelay_uid>-<hubrelay_uid> lookup 88
ip route add blackhole default table 88
```

When a gateway is healthy, add explicit WG routes; otherwise keep blackhole.

```bash
ip route replace 0.0.0.0/0 dev <wg-iface> table 88
# fail-safe
ip route replace blackhole default table 88
```

## WireGuard endpoint exception

WG encapsulation to peer endpoint must stay in main table.

```bash
ip rule add to <peer-ip> lookup main priority <higher>
ip route get 51820 uid $(id -u hubrelay)
```

## nftables output policy

Allow only loopback, approved `wg-*`, and WireGuard maintenance packets.

```nft
table inet hubrelay_egress {
  chain output {
    type filter hook output priority 0;
    policy accept;
    meta skuid != <hubrelay_uid> return
    oifname "lo" accept
    oifname "wg-*" accept
    ip daddr <wg_peer_ips> udp dport 51820 accept
    udp dport 53 drop
    tcp dport 53 drop
    reject with icmpx type admin-prohibited
  }
}
```

## DNS policy

Block direct DNS for `hubrelay` service unless explicitly through approved resolver path.

## Validation

```bash
id hubrelay
ip rule show
ip route show table 88
ip route get 1.1.1.1 uid $(id -u hubrelay)
nft list table inet hubrelay_egress
curl -s http://127.0.0.1:5500/healthz
curl -s http://127.0.0.1:5500/api/command -X POST -H "Content-Type: application/json" -d '{"principal_id":"operator-local","roles":["operator"],"command":"egress-status"}'
```

## Failure checks

- If traffic leaks to WAN: validate `ip rule`, table route, and nftables UID block.
- If provider calls fail with WireGuard healthy: check destination routes and gateway prefixes.
- If DNS leaks: verify resolver flow and port-53 blocks.

Host policy should fail-closed before app-level fallback.

