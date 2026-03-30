# ADR 0003: bbolt As Runtime State

## Status
Accepted

## Context
The bot should survive container deletion, image replacement, and port changes while retaining operator state and audit history.

## Decision
`bbolt` is the only mutable persistence layer for runtime state.
Its contents include:
- principals and policy flags,
- session state,
- audit history,
- workflow state,
- plugin-specific mutable state.

It excludes:
- secrets,
- adapter inventory,
- build profile metadata that is already fixed by the image.

## Consequences
- Redeploys preserve operator-visible history and state.
- The container can run with a read-only root filesystem.
- Schema changes must be versioned and migrated explicitly.
