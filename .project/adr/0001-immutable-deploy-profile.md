# ADR 0001: Immutable Deploy Profile

## Status
Accepted

## Context
The bot must run in isolated environments where runtime surface area should stay stable after deployment.
The user explicitly requires that adapters cannot be enabled, disabled, or added at runtime.

## Decision
The deployed image defines the complete capability profile of the running instance.
This includes:
- transport adapters,
- plugin families,
- AI backend presence,
- network listeners,
- background workers.

Runtime state may only tune behavior inside that already deployed capability set.

## Consequences
- Adding a new adapter requires a redeploy.
- The running instance can reliably describe its capabilities.
- Attack surface is bounded by the deployed profile.
- UI and API flows must reject requests that reference unavailable adapters or capabilities.
