# Architecture (summary)

This section orients new readers. Authoritative diagrams and contracts also live under [`.project`](../../.project/README.md).

## Layers

| Layer | Responsibility |
| --- | --- |
| **Adapters** | Map transports (HTTP, email, …) to `CommandEnvelope` |
| **Core / command bus** | ACL, capabilities, audit, sensitive-data gates |
| **Plugins** | Implement commands (`capabilities`, `system-info`, `ask`, …) |
| **Outbound policy** | Decide direct vs SOCKS lease for workload egress |
| **AI provider** | OpenAI-compatible client (`chat_completions` or `responses` mode) |
| **Storage** | BoltDB for principals, sessions, audit — not secrets |
| **Build profile** | Immutable capability set compiled into the image |

## Why adapters are fixed at deploy time

**Why**: reduces attack surface and configuration drift. Changing channels requires a **rebuild and redeploy**, which is intentional for a hardened hub.

## Why capabilities gate plugins

**Why**: every command declares `RequiredCapabilities`. The core rejects requests that the image never promised. That keeps partial or mis-built images from silently exposing features.

## Data flow (mental model)

```
Client → Adapter → CommandEnvelope → Core → Plugin(s)
                              ↓
                         Audit (BoltDB)
```

When a plugin calls an external API, it goes through **outbound policy** and the configured HTTP client (direct or SOCKS).

## HTTP chat adapter (operational hardening)

The loopback server applies:

- Bounded request bodies (DoS mitigation),
- Read/write timeouts on the HTTP server,
- Safer rendering in the embedded page (dynamic strings escaped before `innerHTML`),

**Why**: even an internal-only listener can be reached by anyone who can open the forwarded port on the workstation; treat defense-in-depth as normal.

## Build profile and local dev

Compile-time `ldflags` set the production profile. For `go run ./cmd/bot`, `INPUT_*` environment variables are merged **once** at startup (`sync.Once`) so local development does not require repeating `-ldflags`.

**Why**: developer ergonomics without changing production’s “single immutable image” story.

## Deeper reading

- [Outbound model](../../.project/outbound-model.md)
- [Security model](../../.project/security-model.md)
- [Core contracts](../../.project/core-contracts.md)
