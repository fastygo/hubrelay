# ADR 0002: Secrets Outside Runtime State

## Status
Accepted

## Context
The user requires tokens and secrets to be fixed at deploy time and unavailable from runtime persistence.
The runtime database must remain portable across redeploys without becoming the source of truth for secrets.

## Decision
Secrets are part of the immutable deployed image/profile and are not stored in `bbolt`.
`bbolt` stores only mutable runtime state such as sessions, audit entries, principal records, workflow state, and plugin data.

## Consequences
- Backing up `bbolt` does not leak transport or AI credentials.
- Secret rotation requires redeploy or image replacement.
- Runtime inspection commands must report secret presence only as capability metadata, never as secret values.
