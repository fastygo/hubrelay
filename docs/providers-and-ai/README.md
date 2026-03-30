# Providers and AI

## OpenAI-compatible mode

The bot uses the official **OpenAI Go SDK** with `WithBaseURL` so the same code path works for:

- OpenAI
- OpenRouter
- Cerebras and other OpenAI-compatible hosts

**Why one SDK**: fewer bespoke HTTP clients; provider differences stay in URL, model ID, and key scope.

## Default API mode: `chat_completions`

`INPUT_AI_API_MODE` defaults to `chat_completions` because it is the most portable surface across vendors. Optional `responses` mode exists for providers that implement that API shape.

## Base URL and trailing slashes

When you pass `INPUT_AI_BASE_URL` or compile `currentAIBaseURL`, the provider layer **normalises** the URL so OpenAI-compatible clients join paths correctly (many SDKs require a trailing slash before appending `chat/completions`).

**Why this matters**: a missing slash can produce wrong URLs and confusing `404` responses even when `curl` examples in vendor docs look correct.

Always set the base URL **exactly as your vendor documents** for their SDK (usually including `/v1`); the code adds the slash if needed.

## API keys: personal vs organisation

Many providers issue **different entitlements** per account type. A key from a personal workspace may return `404` or `model not found` for models that require an **organisation** or **enterprise** project.

**Why**: this is provider-side authorisation, not a bug in the bot. Use `cmd/provider-smoke` to see the raw HTTP error before debugging the HTTP UI.

## Models

Set `INPUT_AI_MODEL` at deploy time to your vendor’s model string. The GUI allows an optional per-request override in the chat form; empty override falls back to the profile default.

## Sensitive prompts

Both the browser and the server run **sensitive-data heuristics** for `ask` to reduce accidental exfiltration of secrets.

**Why**: operators paste real keys into chat by mistake; warnings plus server-side blocking limit damage. Do not rely on this as the only control — still use least-privilege keys and rotation.

## Smoke vs full bot

| Tool | Use when |
| --- | --- |
| `provider-smoke` | Verifying key, URL, model with the smallest surface |
| `cmd/bot` | Verifying adapters, proxy policy, audit, and UI |

See [Local testing](../local-testing/README.md).
