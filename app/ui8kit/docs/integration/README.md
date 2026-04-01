# Integration

These guides describe how to connect UI8Kit to a real Go application: compiling CSS with Tailwind v4 and serving assets so `layout.Shell` and browsers stay in sync.

## Guides

- [Tailwind CSS v4 setup](tailwind-setup.md) — `package.json`, `input.css`, watch scripts, syncing UI8Kit styles
- [How to sync UI8Kit CSS in apps](sync-ui8kit-css.md) — run `app/scripts/sync-ui8kit-css.sh` and verify
- [HTTP server and static assets](http-server.md) — routes, `embed.FS`, matching default CSS paths

## Summary

1. Add UI8Kit with `go get github.com/fastygo/ui8kit`.
2. Add Tailwind CLI to your app (dev dependency) and create `static/css/input.css`.
3. Sync UI8Kit CSS into your app (`sync-ui8kit-css.sh` in consuming repo).
4. Run Tailwind CLI to produce `static/css/app.css`.
5. Serve `/static/css/app.css` (or set `layout.ShellProps.CSSPath` to your URL).
