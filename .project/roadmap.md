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

## Current monorepo layout after refactor

- `sshbot` root module: core daemon and runtime contracts.
- `sdk/hubrelay`: protocol client module.
- `hubcore`: reusable dashboard library (import-time only, no runtime process).
- `apps/dashboard`: dashboard binary module (`hubrelay-dashboard`).
- `apps/dashboard/ui8kit`: design system module.

Execution model: library modules (`hubcore`, `sdk/hubrelay`, `ui8kit`) are linked into binaries.
Each service is deployed as its own executable and its own system unit; there is no shared hubcore runtime.

The rule is consistent across repo evolution:

- any new service module should import `hubcore` as needed,
- build as one standalone executable,
- deploy as one unit file.
This layout keeps production invariants stable while lowering coupling between control plane and operator UI.

## Working Rules
- Runtime changes may adjust state and policy only inside the deployed capability set.
- Redeploy may replace the image, ports, and adapter set, but must preserve `bbolt` state unless an explicit migration says otherwise.
- Unsafe raw shell execution is out of scope for the default profile.
