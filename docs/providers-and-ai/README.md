# Providers and AI

## Compatible providers

HubRelay uses the OpenAI-compatible Go SDK with `WithBaseURL`.

- OpenAI
- OpenRouter
- Cerebras and other compatible providers

## API mode

`INPUT_AI_API_MODE` defaults to `chat_completions`; `responses` is optional.

## Base URL

Normalize provider base URL format before requests:

- keep vendor prefix as documented (usually includes `/v1`)
- slash normalization is handled internally

## API keys and entitlements

Personal and organization keys may produce different model availability. Use `cmd/provider-smoke` for first-party error visibility.

## Runtime check matrix

- `provider-smoke` — key/path/model validation with smallest surface
- `cmd/bot` — full path: adapters, policy, audit, UI

See also [Local testing](../local-testing/README.md).
