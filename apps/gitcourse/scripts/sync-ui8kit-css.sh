#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_DIR="$ROOT_DIR/static/css/ui8kit"

UI8KIT_DIR="$(cd "$ROOT_DIR/../dashboard/ui8kit" && pwd)"

mkdir -p "$TARGET_DIR"
cp "$UI8KIT_DIR/styles/base.css" "$TARGET_DIR/base.css"
cp "$UI8KIT_DIR/styles/shell.css" "$TARGET_DIR/shell.css"
cp "$UI8KIT_DIR/styles/components.css" "$TARGET_DIR/components.css"
cp "$UI8KIT_DIR/styles/latty.css" "$TARGET_DIR/latty.css"

echo "ui8kit CSS synced to $TARGET_DIR"
