# Streaming

SSE output is the streaming variant of `ask` and future commands.

## Endpoints

- `POST /api/command` (JSON)
- `POST /api/command/stream` (SSE)

Request body is identical for both endpoints.

## SSE format

Headers:

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

Event pattern:

- `chunk`: partial text delta
- `done`: final command result
- `error`: failure terminal event

## Contract surface

- Existing plugin sync path remains compatible.
- Streaming plugin implements `ExecuteStream(...)` and emits to a `StreamWriter`.
- If plugin does not stream, it falls back to sync execution.

## Example

```bash
curl -N http://127.0.0.1:5500/api/command/stream \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"ask","args":{"prompt":"hello"}}'
```

## Client and SDK

Use `sdk/hubrelay` for both JSON and SSE paths. Provider streaming still passes the same policy and safety checks before execution.
