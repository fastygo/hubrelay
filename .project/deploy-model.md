# Deploy Model

## Bootstrap
- local deploy inputs define the target profile,
- deploy pipeline produces or selects the immutable image,
- first startup creates the `bbolt` file and initializes schema,
- no runtime config file is required on the host.

## Redeploy
- a new image may replace the old one,
- the mounted `bbolt` file is retained,
- adapters and secrets change only if the new image profile changes,
- runtime state is migrated explicitly when schema version changes.

## Read-Only Runtime
- container root filesystem is read-only,
- writable state is limited to mounted runtime storage,
- loopback HTTP listeners are preferred for tunnel-based access,
- adapter-specific network behavior must be declared by the profile.

## Failure Expectations
- deleting the container must not delete runtime state,
- changing ports must not mutate runtime data,
- losing an adapter endpoint must not corrupt `bbolt`,
- database recovery is handled through host-level backup and restore.
