# ADR 0002: Secrets Outside Runtime State

## Status

Accepted

## Context

Secrets must not live in mutable runtime storage.

## Decision

Secrets remain in deployed profile/flags.
`bbolt` stores only runtime state.

## Consequence

- `bbolt` backups are not secret repositories
- secret rotation is redeploy/task-oriented
- runtime APIs expose presence, not secret values
