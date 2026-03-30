# Network, SSH tunnel, and SOCKS proxy

## Loopback bind

The default profile binds HTTP chat to **loopback** (`127.0.0.1` or as set by `INPUT_BOT_HTTP_BIND`). The process listens only on the server itself.

**Why**: exposing an operator panel on `0.0.0.0` without additional auth layers is easy to misconfigure. Loopback forces you to use SSH or an equivalent private path.

## SSH local port forward

On your workstation:

```bash
ssh -N -L 5500:127.0.0.1:5500 -i <SSH_PRIVATE_KEY> user@<SERVER_HOST>
```

Then open `http://127.0.0.1:5500` in a browser on that workstation.

**Why this works**: SSH encrypts and authenticates the tunnel; the browser talks to local loopback only.

## When the browser cannot reach the AI provider

The **browser** does not call OpenAI directly. The **server process** does. Traffic path:

```
Browser → SSH tunnel → bot HTTP → plugin ask → AI provider
```

If the server must egress via SOCKS (provider allow-lists, isolation), the bot can require a **proxy session**.

## Proxy session (SOCKS pool)

**What**: operators paste a list of `host:port` SOCKS proxies in the UI (or API). The server probes them, picks a **sticky lease** per session id, and can fail over on transport errors.

**Where state lives**:

- **Server**: in-memory manager (lost on restart).
- **Browser**: `sessionStorage` for the session id and selection (not BoltDB, not `localStorage` mixed with chat unless you choose otherwise in product UX).

**Why not BoltDB**: proxy pools are ephemeral tactical routing; persisting them would mix operational noise with durable audit data.

## `INPUT_PROXY_SESSION_FORCE`

When `true`, workload outbound (AI calls) must use an **active proxy lease**. When `false`, requests may go **direct** if no `proxy_session_id` is supplied.

**Why force exists**: some deployments want a hard guarantee that the hub never accidentally calls the provider from the server’s bare IP. That is a policy choice, not a security silver bullet—proxy operators must still be trusted.

## Health checks vs workload

Proxy **GET /models** style probes are **control plane** checks. They validate reachability through a candidate SOCKS hop; they do not replace end-to-end verification of `POST /chat/completions` for your exact model.

**Why**: cheaper and faster; full completions remain the ground truth (use smoke or a short `ask` after selecting a proxy).

## Related docs

- [Deploy](../deploy/README.md) for compile-time defaults of proxy flags.
- [`.project/outbound-model.md`](../../.project/outbound-model.md) for the long-term egress abstraction.
