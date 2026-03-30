# Autonomous Bot Roadmap

## Purpose
This project builds an autonomous server bot that behaves like an internal control plane.
It must stay useful across network restrictions, survive container recreation, and keep mutable runtime state outside the container image.

## Locked Invariants
- Deploy inputs define an immutable runtime profile.
- Adapters are fixed by the deployed image and cannot be added at runtime.
- Secrets and access tokens are baked into the deployed image/profile and never written to `bbolt`.
- Mutable runtime state lives only in `bbolt`.
- The container must support `docker --read-only` with writable state limited to mounted runtime storage.
- Operator access can happen through multiple transport adapters, but every adapter maps into the same core command model.
- Workload outbound traffic must go through a shared policy layer rather than plugin-local transport rules.

## Phase Map
### Phase 1: Architecture and documentation
- Record ADRs for immutable profiles, secret boundaries, and runtime persistence.
- Define threat model and trust boundaries.
- Define the first deploy profile: `tunnel_chat + yandex_mail + openai`.

### Phase 2: Core runtime
- Implement a transport-agnostic command bus.
- Implement capability-gated plugins.
- Implement `bbolt` storage for runtime state, sessions, ACL, and audit.

### Phase 3: First deliverable profile
- Implement loopback HTTP tunnel chat.
- Implement a mail adapter scaffold.
- Implement safe system inspection plugins.
- Expose a `capabilities` command to describe the immutable build profile.

### Phase 4: Controlled operations
- Add confirmation-aware operational plugins.
- Extend audit coverage and replay visibility.
- Add stricter policy gates for destructive actions.

### Phase 5: Hub workflows and outbound integrations
- Introduce reusable outbound contracts for platform clients and workflow actions.
- Keep proxy-based egress as an implementation detail behind shared policy.
- Allow future replacement of proxy routing with VPN-backed egress without changing plugin contracts.

## Recent engineering themes (shipped in tree)
- **Outbound policy package** (`internal/outbound`) shared above `ask` and future plugins.
- **Provider smoke CLI** (`cmd/provider-smoke`) for minimal AI-path debugging.
- **Local dev ergonomics**: `INPUT_*` env merge once at startup for `go run ./cmd/bot` without repeating `-ldflags`.
- **HTTP operator surface**: body limits, server timeouts, safer dynamic HTML, cached index template parse.
- **AI observability**: structured logs for `ask` / provider calls (no secrets in log lines).
- **OpenAI-compatible base URL** normalisation (trailing slash) to avoid vendor `404` misroutes.

Canonical narrative docs: [`docs/`](../docs/README.md).

## Working Rules
- Runtime changes may adjust state and policy only inside the deployed capability set.
- Redeploy may replace the image, ports, and adapter set, but must preserve `bbolt` state unless an explicit migration says otherwise.
- Unsafe raw shell execution is out of scope for the default profile.
