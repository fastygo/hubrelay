# Shell component

This page describes the `layout.Shell` component behavior in one place.

## What Shell assembles

- Full HTML document wrapper with `<head>` and `<body>`.
- Global stylesheet link from `ShellProps.CSSPath` (default `/static/css/app.css`).
- Responsive dashboard structure: desktop sidebar + header + main column.
- Mobile Sheet overlay with fixed panel.

## Why this Shell uses CSS-only navigation

The mobile menu is intentionally implemented without navigation JavaScript.

It uses:

- hidden checkbox input `#ui8kit-mobile-sheet`
- label-driven toggle (`for="ui8kit-mobile-sheet"`)
- fixed portal wrapper `.kit-shell-mobile-sheet-portal`
- overlay label `.kit-shell-mobile-sheet-overlay`
- sliding panel `.kit-shell-mobile-sheet-panel`

This pattern allows pure CSS state transitions and predictable close behavior.

## State handling

- Checked input = sidebar open.
- Unchecked input = sidebar closed.
- Both overlay and close label target the same checkbox.
- On mobile, `body` scroll lock can be enforced via `:has()` selector.

## Accessibility considerations

- Trigger control has `aria-label`, `aria-haspopup`, and `aria-controls`.
- Header keeps accessible landmarks.
- Panel has `role="dialog"`, `aria-modal="true"`.
- Keep focus order natural and test keyboard use after content updates.

## Interaction and extension strategy

- Keep theme switching in a separate script (`themeScript`) and avoid coupling it to menu state.
- For further animation or focus-management enhancement, prefer CSS transitions and minimal JS hooks if needed.
- Prefer overriding `kit-shell-*` in app `app.css` over rewriting the markup shape.
