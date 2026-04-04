# Security and privacy

## Secret boundary

| Value type | Runtime DB |
| --- | --- |
| AI API keys, bearer tokens, passwords | Never stored |
| Optional registry credentials | Never stored |
| Adapter/runtime flags | Never stored |
| Proxy lists from operators | In-memory / client session only |

## Chat history

`INPUT_CHAT_HISTORY=true` stores chat in browser `localStorage` only.

## Sensitive data controls

`ask` inputs run validator checks on both client and server.

## Logging

Allowed: model names, base URL hosts, proxy session id, provider errors.
Forbidden: full provider keys.

## Container mode

Production is typically `read_only: true` with writable mount for `BoltDB`.

## Links

- [`.project/security-model.md`](../../.project/security-model.md)
- [`.project/adr/0002-secrets-outside-runtime-state.md`](../../.project/adr/0002-secrets-outside-runtime-state.md)

## Docs hygiene

Use placeholders (`<YOUR_AI_API_KEY>`, `<SERVER_HOST>`), do not commit `.env`, redact any internal hostnames in screenshots.
