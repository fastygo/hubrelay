# UI8Kit documentation

Welcome to the UI8Kit documentation. UI8Kit is a **Go + templ + Tailwind CSS** component kit in the style of shadcn/ui.

## Contents

### Getting started

- [Overview](getting-started/README.md) — what UI8Kit is and when to use it
- [Installation](getting-started/installation.md) — Go module, templ CLI, optional Node.js for Tailwind
- [Project structure](getting-started/project-structure.md) — how to lay out an app that uses UI8Kit
- [Design principles](design-principles.md) — 8px-grid mindset, prop-first styling, accessibility rules

### Packages

- [Packages overview](packages/README.md)
- [ui](packages/ui.md) — Box, Stack, Group, Container, Button, Badge, Text, Title, Field, Icon
- [layout](packages/layout.md) — Shell, Header, Sidebar, CSS-only mobile behavior
- [utils](packages/utils.md) — UtilityProps, `Cn`, variant helpers
- [styles](packages/styles.md) — embedded CSS, theme tokens, Latty icons

### Tooling and integration

- [Scripts](scripts.md) — generator helpers in `ui8kit/scripts`
- [Integration overview](integration/README.md)
- [Tailwind CSS v4 setup](integration/tailwind-setup.md) — compiler, input.css, scanning `.go` / `.templ`
- [How to sync UI8Kit CSS in an app](integration/sync-ui8kit-css.md) — sync and copy script workflow
- [HTTP server and static assets](integration/http-server.md) — serving CSS, matching `layout` defaults

### Testing

- [Testing](testing.md) — test suite overview, architecture, CI integration, writing new tests

### Versioning

- [Versioning and releases](versioning.md) — semver, release script, git tags, Go proxy

## Quick links

| Topic | Document |
|--------|----------|
| `go get` and imports | [Installation](getting-started/installation.md) |
| npm + `@tailwindcss/cli` | [Tailwind setup](integration/tailwind-setup.md) |
| `UtilityProps` | [utils](packages/utils.md) |
| Styling and refactor guardrails | [Design principles](design-principles.md) |
| Full-page layout | [layout](packages/layout.md) |
| Scripts for style generation | [Scripts](scripts.md) |

## Module path

```text
github.com/fastygo/ui8kit
```

## License

MIT — see the repository [LICENSE](../LICENSE) file.
