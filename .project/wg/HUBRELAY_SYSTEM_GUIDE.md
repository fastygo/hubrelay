# HubRelay system guide (WireGuard private egress)

This document explains **how the HubRelay binary works**, how **configuration** reaches it at runtime, and how **Linux networking** (WireGuard, policy routing, firewalls) interacts with the application. It is written for newcomers who need an end-to-end picture.

For shorter SRE runbooks, see:

- [`.project/wg/README.md`](README.md) — operational commands and health checks
- [`.project/wg/egress-gateway/README.md`](egress-gateway/README.md) — Marvin (egress gateway) only

---

## 1. What HubRelay is

HubRelay is a **single Go process** (`cmd/bot/main.go`). It:

1. Loads an **immutable profile** (identity, adapters, AI settings, proxy/private-egress policy).
2. Opens a **BoltDB** file for audit and small runtime state (`BOT_DB_FILE`).
3. Registers **plugins** (e.g. `capabilities`, `system.info`, `ask`).
4. Starts **adapters** (loopback **HTTP chat**, optional **email**).
5. Accepts **commands** over HTTP (typically after you open an **SSH tunnel** from your laptop).

There is no separate “worker” process. Everything runs in one binary.

```
Laptop ──SSH -L 5500:127.0.0.1:5500──▶ Jarvis 127.0.0.1:5500
                                              │
                                              ▼
                                        HubRelay (HTTP adapter)
                                              │
                                              ▼
                                        Command bus → plugins (e.g. ask)
                                              │
                                              ▼
                                        Outbound HTTPS to AI provider
```

---

## 2. Configuration: three layers

Understanding configuration is critical: some values are “burned in” at build time, others come from the environment **once** at process start.

### 2.1 Compile-time (`-ldflags`)

When you build with `go build -ldflags "-X sshbot/internal/buildprofile.currentProfileID=..."`, those strings replace **package-level variables** in `internal/buildprofile/profile.go`.

Examples of variables that can be set this way:

- `currentProfileID`, `currentDisplayName`, `currentHTTPBind`
- Email/OpenAI toggles and defaults
- `currentAIAPIKey`, `currentAIBaseURL`, `currentAIModel`, etc.
- `currentProxySession`, `currentProxyForce`
- `currentPrivateEgressRequired`, `currentPrivateEgressInterface`, `currentPrivateEgressTestHost`, `currentPrivateEgressFailClosed`

**Dockerfile** and some **paas** extensions pass the same set as `--build-arg` → `-X` flags.

### 2.2 Environment overrides (`INPUT_*`)

On startup, `buildprofile.Current()` calls `applyEnvOverrides()` **once** (`sync.Once`). If an `INPUT_*` variable is **non-empty**, it **replaces** the corresponding string holder before the profile is assembled.

Relevant keys (partial list from `applyEnvOverrides`):

| Variable | Effect |
|----------|--------|
| `INPUT_AI_API_KEY` | Provider API key (sensitive) |
| `INPUT_AI_BASE_URL` | Provider base URL (e.g. `https://api.cerebras.ai/v1`) |
| `INPUT_AI_MODEL` | Model name |
| `INPUT_AI_PROVIDER` | Provider label (`openai`, etc.) |
| `INPUT_AI_API_MODE` | `chat_completions` or `responses` |
| `INPUT_PROXY_SESSION_ENABLED` | Enables proxy-session UI/API |
| `INPUT_PROXY_SESSION_FORCE` | Requires outbound AI calls to use a leased SOCKS proxy |
| `INPUT_CHAT_HISTORY` | Feature toggle for browser chat history |
| `INPUT_PRIVATE_EGRESS_REQUIRED` | If true, run private egress validation before each AI request |
| `INPUT_PRIVATE_EGRESS_INTERFACE` | Expected interface name (e.g. `wg0`) |
| `INPUT_PRIVATE_EGRESS_TEST_HOST` | Host or IP used to verify routing (e.g. gateway `10.88.0.1`) |
| `INPUT_PRIVATE_EGRESS_FAIL_CLOSED` | If true, failed validation **blocks** `ask`; if false, logs and continues |

**Important:** `INPUT_*` overrides are read **only at first `Current()` call**. Changing env without restarting the process does nothing.

### 2.3 Process environment (`BOT_DB_FILE`)

`BOT_DB_FILE` is read directly in `main` from `os.Getenv`, **not** through `buildprofile`. Default if unset: `data/bot.db`.

Host-run deployments often set:

- `BOT_DB_FILE=/var/lib/hubrelay/bot.db`

---

## 3. The profile object

`buildprofile.Current()` returns a `Profile` struct used across the app:

- **Capabilities** — what the server advertises (`capabilities` command): HTTP chat, email, AI, proxy session, plugins, etc.
- **`HTTPChat`** — bind address (e.g. `127.0.0.1:5500`). If empty, HTTP chat is disabled.
- **`OpenAI`** — provider name, key, base URL, model, API mode; `HasAPIKey` is derived from non-empty key.
- **`ProxySession`** — whether SOCKS session UI exists and whether AI **must** use a proxy (`Policy.RequireProxy`).
- **`PrivateEgress`** — whether to validate private egress before AI, which interface, test host, and fail-closed behavior.

The profile is **immutable for the lifetime of the process** in the sense that there is no hot reload: restart to apply new build flags or env.

---

## 4. Program flow inside the binary

### 4.1 `main` (`cmd/bot/main.go`)

1. `profile := buildprofile.Current()`
2. Validates `proxy_force` requires `proxy_session` enabled (fatal if misconfigured).
3. Opens BoltDB at `BOT_DB_FILE`.
4. If proxy sessions are enabled **and** AI is enabled with a key, creates `proxymgr.Manager`.
5. Always constructs `outbound.NewInterfaceEgressChecker()` for optional private egress checks.
6. If AI is enabled with a key, builds `ai.NewOpenAICompatibleProvider(...)` with:
   - outbound `Policy{RequireProxy: profile.ProxySession.Force}`
   - `profile.PrivateEgress` + `privateEgressChecker` + `proxyManager`
7. Registers `ask` plugin with that provider.
8. Builds `core.NewService` and starts adapters:
   - **HTTP chat** on `profile.HTTPChat.BindAddress`
   - **Email** adapter if enabled
9. Waits for SIGINT/SIGTERM or adapter failure.

### 4.2 HTTP adapter and commands

The HTTP adapter exposes JSON endpoints (including `/api/command`). Commands are dispatched through the core service with role/capability checks. The **`ask`** command invokes the OpenAI-compatible provider.

### 4.3 Private egress check (application layer)

Implementation: `internal/outbound/private_egress.go` + call site in `internal/ai/openai.go`.

When `PrivateEgress.Required` is true, before `Ask`:

1. If `FailClosed` and no checker — error.
2. **`Check(interfaceName, testHost)`**:
   - Resolve interface by name; ensure it exists and **`FlagUp`**.
   - If `testHost` is empty — stop after interface check (success).
   - Resolve `testHost` to IP(s) (DNS or literal IP).
   - For each target IP, the checker opens a **UDP “connection”** to `target:9` (discard port) with `DialContext` to learn which **local source IP** the kernel would use.
   - If that local IP is one of the IPs assigned to the named interface (e.g. `10.88.0.3` on `wg0`), validation succeeds.
   - Otherwise → `ErrPrivateEgressRouteMismatch`.

If validation fails and `FailClosed` is true → **`ask` returns an error immediately** (no full HTTP round trip to the provider). If `FailClosed` is false → log and continue.

**Note:** This is **routing introspection**, not a full TCP test to the AI provider. It confirms “kernel would egress toward this host using an address on `wg0`”.

### 4.4 Proxy sessions vs WireGuard

- **Proxy sessions** (`INPUT_PROXY_SESSION_*`): application-level SOCKS leases; `Policy.RequireProxy` forces the HTTP client to dial via proxy.
- **WireGuard**: **host** networking; the process does not “enable WG”. It only **validates** that `wg0` exists and routes look right **if** you enable private egress checks.

For WG-based egress, typical production intent:

- `INPUT_PROXY_SESSION_ENABLED=false`
- `INPUT_PROXY_SESSION_FORCE=false`
- `INPUT_PRIVATE_EGRESS_REQUIRED=true` with interface `wg0` and a **gateway** IP as `TEST_HOST`

---

## 5. Network model: Jarvis, Marvin, and “fail closed”

### 5.1 Roles

| Host | Role |
|------|------|
| **Jarvis** | HubRelay host (Netherlands in the reference design). Runs `hubrelay` as a dedicated Linux user. Listens HTTP on loopback only. |
| **Marvin** | WireGuard peer + **NAT egress gateway** (e.g. US VDS). Forwards tunneled traffic to the Internet with masquerade. |

### 5.2 Intended traffic path for AI (host-run)

For the `hubrelay` Linux user (example uid **995**):

1. **Policy routing** sends this user’s **outbound** traffic to **routing table 88** (example).
2. Table 88 has a default route via **`wg0`**.
3. **nftables** may set **`fwmark`** on packets from that uid so marked traffic also uses table 88.
4. Encrypted WireGuard frames go to **Marvin’s public IP:51820/UDP**.
5. Marvin decaps, NATs, sends to the AI provider; return traffic comes back through the tunnel.

If **`wg0` is down** or mis-routed:

- At the **kernel** level: connections stall or fail (no “silent” leak to the Dutch default path if policy is strict).
- At the **app** level: private egress check fails → **`ask` fails** when `FAIL_CLOSED=true`.

### 5.3 Critical routing detail: WireGuard endpoint loop

If you route **all** traffic from `hubrelay` through `default dev wg0` **without** an exception, **WireGuard’s own encapsulated UDP** to the peer endpoint can be forced back into `wg0`, causing a **routing loop** and broken connectivity.

The fix used in production is an **`ip rule` with higher priority** (lower number) than the uid rule:

- Example: `from all to <MARVIN_PUBLIC_IP> lookup main`

So **UDP to the WG endpoint** uses the **main** table (physical interface), while **everything else** for uid 995 still uses table 88 → `wg0`.

Scripts `hubrelay-net-up` / `hubrelay-net-down` on Jarvis should install and remove this rule together with other policy routing.

See [`.project/wg/README.md`](README.md) for concrete `ip rule` / `ip route show table 88` checks.

---

## 6. How you operate it day to day

### 6.1 Operator access (no public HubRelay port)

From your laptop:

```bash
ssh -N -L 5500:127.0.0.1:5500 -i ~/.ssh/key user@jarvis
```

Open `http://127.0.0.1:5500`. Traffic to that URL hits **Jarvis loopback**, not the public Internet.

### 6.2 Host-run systemd service

Typical layout on Jarvis:

| Path | Purpose |
|------|---------|
| `/opt/hubrelay/bot` | Binary |
| `/opt/hubrelay/env` | `INPUT_*` for runtime secrets and toggles (`EnvironmentFile=`) |
| `/var/lib/hubrelay/bot.db` | BoltDB (`BOT_DB_FILE`) |
| `/etc/systemd/system/hubrelay.service` | `User=hubrelay`, `After=wg-quick@wg0`, etc. |

Restart after changing env or binary: `systemctl restart hubrelay.service`.

### 6.3 Docker vs host-run (WG)

| Mode | Sees `wg0`? | Private egress check |
|------|-------------|----------------------|
| **Host-run** | Yes (host network namespace) | Meaningful |
| **Default Docker bridge** | No | Fails or is meaningless; AI may use container default route |
| **Docker `network_mode: host`** | Yes | Same caveats as host-run regarding uid/routing |

The repository’s **deploy-hostrun** path cross-compiles a Linux binary locally and uploads it; **deploy-direct** builds a Docker image — choose the path that matches how you enforce routing.

---

## 7. Cross-compilation and local dev

- **Production binary** for Jarvis is often built as `GOOS=linux GOARCH=amd64` on a developer machine; that artifact **does not run on Windows**.
- **Local dev** on Windows or macOS: `go run ./cmd/bot` with `INPUT_*` exported — `buildprofile` picks up env without `-ldflags`.

When using **Git Bash** on Windows with `env -i` wrappers (some deploy docs), ensure **`LOCALAPPDATA`** (and optionally `GOPATH`) are passed through so Go can locate **`GOCACHE`**.

---

## 8. Quick mental model

| Layer | Responsibility |
|-------|------------------|
| **SSH + loopback bind** | Only trusted operators reach the command surface |
| **Linux routing + nftables** | Force `hubrelay` egress through `wg0`; exception for WG endpoint UDP |
| **WireGuard + Marvin NAT** | Encrypted transit + exit IP at gateway |
| **`INPUT_PRIVATE_EGRESS_*`** | App refuses `ask` if interface/route checks fail (when required + fail-closed) |
| **`INPUT_PROXY_SESSION_*`** | Optional SOCKS path inside the app (orthogonal to WG) |
| **BoltDB** | Audit/state; **not** for secrets |

---

## 9. Where to read code

| Topic | Location |
|-------|----------|
| Entry point | `cmd/bot/main.go` |
| Profile and env | `internal/buildprofile/profile.go` |
| Private egress logic | `internal/outbound/private_egress.go` |
| AI + preflight | `internal/ai/openai.go` |
| Proxy policy | `internal/outbound/policy.go` |
| Project overview | `README.md` |
| Deploy checklist | `.paas/.check-deploy.md` |

This guide intentionally stays **sanitized**: use placeholders in docs, never commit real API keys or internal hostnames.
