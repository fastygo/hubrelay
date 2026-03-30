# Security and privacy

## Secret boundary

| Stored in image / profile at deploy | Never stored in BoltDB (by design) |
| --- | --- |
| AI API keys | Same keys, bearer tokens, passwords |
| Optional registry credentials for deploy tooling | Chat transcripts |
| Compile-time adapter flags | Proxy lists from operators (memory/sessionStorage only) |

**Why**: BoltDB is backup-friendly and may be copied for debugging. Secrets in the DB would leak with every snapshot.

## Chat history

Browser chat persistence is controlled by `INPUT_CHAT_HISTORY`:

- `true`: `localStorage` on the client (still not on server).
- `false`: in-memory for the tab only.

**Why**: server-side transcript storage is an explicit future product choice; MVP avoids new privacy obligations.

## Sensitive data scanner

`ask` prompts pass through regex, dictionary, and entropy-style checks on both:

- **client** (warning modal), and  
- **server** (hard block with audit-safe message).

**Why**: reduce accidental pasting of credentials into third-party inference APIs. This does **not** replace data classification, DLP, or key management.

## Logging

Operational logs may include model names, base URL hosts, proxy session identifiers, and error strings from providers.

**Why**: you need diagnosability; ensure log shipping and retention match your policy. **Do not** log full API keys (current code paths avoid printing the key; keep it that way in future changes).

## Documentation hygiene

- Use placeholders: `<YOUR_AI_API_KEY>`, `<SERVER_HOST>`.
- Do not commit `.env` files with real secrets.
- Redact screenshots of `capabilities` if they include internal hostnames you consider confidential.

## Read-only containers

Production compose often sets `read_only: true` with a writable mount for BoltDB only.

**Why**: limits post-exploitation writes to filesystem. Misconfigured volumes surface as startup errors—fix the mount, not the security model.

## Further reading

- [`.project/security-model.md`](../../.project/security-model.md)
- [`.project/adr/0002-secrets-outside-runtime-state.md`](../../.project/adr/0002-secrets-outside-runtime-state.md)
