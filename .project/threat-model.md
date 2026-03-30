# Threat Model

## Assets
- server availability,
- runtime state in `bbolt`,
- immutable secrets embedded in the deployed profile,
- operator trust,
- audit trail integrity.

## Trust Boundaries
- external transport adapters cross into the bot core,
- the bot core crosses into safe plugins,
- plugins cross into host inspection or controlled operations,
- the runtime database crosses redeploy boundaries.

## Main Risks
- spoofed or replayed external requests,
- unauthorized command execution through chat-like channels,
- privilege escalation by asking the AI layer to escape typed tools,
- secret disclosure through logs or diagnostic commands,
- state corruption during abrupt container restart,
- adapter abuse causing denial of service.

## Mitigations
- per-adapter authentication and principal mapping,
- immutable capability checks before command dispatch,
- no raw shell access in the default profile,
- structured audit records for every externally triggered action,
- schema versioning and transactional `bbolt` writes,
- rate limiting and bounded command payload sizes.
