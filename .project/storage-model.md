# Storage Model

`bbolt` stores mutable runtime state only.

## Buckets

- `meta`
- `principals`
- `sessions`
- `audit`
- `plugin_state`

## Excluded from `bbolt`

- secrets
- adapter inventory
- build profile metadata
- image identity

## Migration rules

- monotonic schema versions
- preserve existing buckets on restart
- migrate before command handling
- destructive changes require explicit migration
