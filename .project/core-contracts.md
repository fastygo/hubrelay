# Core Contracts

Primary execution contracts exposed by the runtime.

## Principal

Normalized actor identity:

- `id`, `display`, `transport`, `roles`, `metadata`

## CommandEnvelope

Transport-neutral request:

- `id`, `transport`, `name`, `args`, `raw_text`, `principal`, `requested_at`

## CommandResult

Normalized response:

- `status`, `message`, `data`, `requires_confirm`

## Plugin

One command + one capability gate per plugin.

Required:
- `Name`
- `RequiredCapabilities`
- `Execute(ctx)`
- transport-agnostic behavior

## OutboundPolicy

Shared decision layer for all outbound work:

- route by immutable policy
- return `direct | proxy | blocked`
- never hard-bind to AI plugin logic

## LeaseResolver

Resolve proxy lease/session for workload ID.

- return active lease or clear error
- no plugin/adapter coupling

## Adapter

Transport boundary that:

- maps incoming message to `CommandEnvelope`
- validates principal
- returns `CommandResult`

## AuditEntry

Immutable record for every command outcome:
- success / reject / failure
