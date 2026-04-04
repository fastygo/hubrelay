# ADR 0001: Immutable Deploy Profile

## Status

Accepted

## Context

Runtime surface must be stable after deploy.

## Decision

Deployed image defines fixed capability profile:

- adapters
- plugin families
- AI backend presence
- listeners
- workers

## Consequence

- new adapter/plugin requires redeploy
- instance capability set is deterministic
- UI/API reject unavailable capabilities
