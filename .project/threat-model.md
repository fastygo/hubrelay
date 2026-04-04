# Threat Model

## Assets

- availability
- runtime state
- embedded secrets policy
- operator trust
- audit integrity

## Attack boundaries

- external adapters -> core
- core -> plugins
- plugins -> host/system integrations
- runtime state across redeploy

## Main risks

- spoofed requests and replay
- unauthorized command execution
- AI tool escape attempts
- secret leakage in logs
- db corruption on abrupt stop
- adapter-level DoS

## Main mitigations

- principal-based auth per adapter
- immutable capability checks
- strict plugin command permissions
- audit trail for externally triggered actions
- schema versioning + transactional writes
- payload size and request rate limits
