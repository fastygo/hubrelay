# Storage Model

## Scope
`bbolt` stores mutable runtime state only.

## Included Buckets
- `meta`: schema version only.
- `principals`: normalized identity and policy state.
- `sessions`: per-principal transport session snapshots.
- `audit`: immutable command history.
- `plugin_state`: plugin-owned mutable state.

## Explicit Exclusions
- secrets,
- deploy-time token material,
- adapter inventory,
- build profile metadata,
- generated image identity.

## Migration Rules
- schema version is monotonic,
- redeploy must preserve existing buckets,
- migrations run before command handling starts,
- incompatible schema changes require explicit code migration, never silent reset.
