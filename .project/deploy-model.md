# Deploy Model

Deploy target is an immutable image/binary with externalized runtime state.

## Bootstrap

1. define target profile
2. build/push image or binary
3. first startup creates `bbolt` and initializes schema
4. no runtime config file inside image

## Redeploy

- new image may replace old
- mounted storage persists
- profile changes happen via image/build args only
- state migrations are explicit

## Runtime expectations

- read-only root FS
- writable mount for state (`BOT_DB_FILE`)
- loopback listener preferred
- profile defines transport/adapter behavior
