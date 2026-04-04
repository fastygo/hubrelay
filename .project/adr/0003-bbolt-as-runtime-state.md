# ADR 0003: bbolt As Runtime State

## Status

Accepted

## Context

Runtime state must survive binary replacement while keeping history.

## Decision

`bbolt` is the single mutable store.

Included:
- principals
- sessions
- audit history
- workflow and plugin state

Excluded:
- secrets
- adapter inventory
- build profile metadata

## Consequence

- state survives redeploy
- read-only root FS remains practical
- schema changes require explicit migrations
