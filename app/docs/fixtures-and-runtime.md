# Fixtures and runtime modes

This document explains how the dashboard uses localized fixture files, localized mock data, and the live HubRelay SDK runtime.

## Locale catalogs

UI copy lives in locale directories under `app/fixtures/`.

Examples:

- `fixtures/en/*.json`
- `fixtures/es/*.json`
- `fixtures/ru/*.json`

These files contain:

- page titles
- descriptions
- button labels
- section titles
- table headers
- shell navigation labels
- locale metadata
- shell navigation labels

The server discovers locales by scanning the fixture directories at startup and chooses one per request using the `lang` query parameter.

English is required and remains the default locale.

## Locale discovery and MVP toggle behavior

The dashboard orders locales like this:

1. English first
2. all remaining discovered locales after it

All discovered locales are available to the server runtime, so adding `fixtures/es` or `fixtures/it` is enough to make those locale catalogs loadable.

This repository includes `fixtures/es` as a proof-of-concept to validate the discovery flow and confirm that locale handling no longer depends on `en` + `ru` branches spread across runtime code.

For the current MVP, the header language toggle switches only between the first two discovered locales. This keeps the interaction simple until the footer/dropdown version is designed.

## Default locale and browser persistence

English is always the default locale.

The browser stores the user-selected locale in `localStorage` under the `hubrelay-language` key. The application shell script synchronizes that value with the current URL:

- the default locale keeps the canonical URL without `lang`
- any non-default locale uses `?lang=<locale>`

This allows the server to keep rendering translated HTML while the client still remembers the last language choice.

When there is no stored locale yet, the shell script also checks `navigator.languages` / `navigator.language` and tries to match the user's preferred browser locale against the discovered fixture locales.

Priority order is:

1. explicit `lang` query parameter
2. stored locale in `localStorage`
3. auto-detected browser locale
4. default locale (`en`)

## Client scripts

Shell-level client behavior is handled by `static/js/app-shell.js`.

That file currently owns:

- theme initialization and toggle behavior
- locale persistence in `localStorage`
- locale switch URL synchronization

Streaming behavior for the ask page remains in `static/js/stream.js`.

Theme labels are localized through fixture data and passed to the shell as button data attributes, so `app-shell.js` does not hardcode user-facing locale-dependent copy for theme switching.

## Mock payloads

Demo data lives separately from UI copy:

- `fixtures/en/mocks/*.json`
- `fixtures/<locale>/mocks/*.json`

Use these files for:

- preview mode
- UI iteration without a live HubRelay instance
- deterministic screenshots and demos
- local testing of translated page states

Keep these files focused on representative payloads, not exhaustive backend coverage.

### About `fixtures/en/mocks`

`fixtures/en/mocks` is the reference English mock set. It is the easiest place to add or review new preview payloads first because:

- English is the default locale
- new pages usually land in English before translation
- it defines the baseline structure for additional locales

When you add a new page or a new section:

1. add English UI copy in `fixtures/en/*.json`
2. add English preview data in `fixtures/en/mocks/*.json`
3. mirror the same structure in `fixtures/<locale>/*.json` and `fixtures/<locale>/mocks/*.json` for every additional language you support

## Runtime modes

The dashboard supports two runtime modes.

### `APP_DATA_SOURCE=fixture`

This mode reads demo payloads from fixture sources and never talks to HubRelay.

Use it when:

- designing or reviewing UI
- working offline
- testing translations
- validating shell behavior without backend dependencies

### `APP_DATA_SOURCE=live`

This is the default mode.

In live mode the dashboard talks to HubRelay through `sshbot/sdk/hubrelay`:

1. HTTP handlers call the app `source` layer.
2. The live source calls `internal/relay`.
3. `internal/relay` wraps `sdk/hubrelay`.
4. The SDK talks to the real HubRelay server over HTTP or unix socket.

This means fixture copy and localized page labels still come from the dashboard repository, but operational data comes from the real HubRelay runtime.

## Live runtime notes

When `APP_DATA_SOURCE=live` is enabled:

- page chrome and labels still come from fixtures
- validation errors generated in the app can be localized
- transport and command errors returned by HubRelay may still reflect live backend messages
- `ask` SSE streaming still uses the SDK-backed server path

Relevant configuration:

- `APP_DATA_SOURCE=live|fixture`
- `HUBRELAY_TRANSPORT=http|unix`
- `HUBRELAY_BASE_URL=http://127.0.0.1:5500`
- `HUBRELAY_SOCKET_PATH=/run/hubrelay/hubrelay.sock`

## Recommended workflow

For UI work:

```bash
APP_DATA_SOURCE=fixture go run ./cmd/server
```

For real integration work:

```bash
APP_DATA_SOURCE=live go run ./cmd/server
```

If you change templates or styles, also run:

```bash
templ generate ./...
npm run sync:ui8kit
npm run build:css
```
