# Overview

HubRelay is a control-plane binary with the following invariants:

- immutable runtime profile at deploy time,
- normalized command envelope + capability gates,
- mutable runtime state in BoltDB,
- operator access through loopback + SSH tunnel.

## Monorepo structure

- `sshbot` — core daemon and plugin bus (`cmd/bot`).
- `sdk/hubrelay` — typed Go client module.
- `hubcore` — shared dashboard library (import-time only).
- `apps/dashboard` — operator dashboard service (`cmd/server`).
- `apps/dashboard/ui8kit` — UI toolkit.

`hubcore` is linked into each binary module and is not started as a service.

Current service binaries:

- `cmd/bot`
- `apps/dashboard/cmd/server`

Each binary is deployed as an independent system unit.

## Next step

- [Getting started](../getting-started/README.md)
- [Network, tunnel, and proxy](../network-tunnel-and-proxy/README.md)
