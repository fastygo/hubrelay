#!/usr/bin/env bash
set -euo pipefail

LLMS=llms.txt
META=llms-metadata.json

if [[ ! -f "$LLMS" ]]; then
  echo "missing: $LLMS"
  exit 1
fi

shopt -s globstar nullglob

# Required section markers
for token in "Mandatory read order" "Open sections map" "Onboarding command matrix" "Update policy"; do
  if ! grep -q "$token" "$LLMS"; then
    echo "missing section token in $LLMS: $token"
    exit 1
  fi
done

# Must keep private dirs excluded
for token in ".manual/*" ".paas/*"; do
  if ! grep -q "$token" "$LLMS"; then
    echo "missing private-dir exclusion token in $LLMS: $token"
    exit 1
  fi
done

# Must expose baseline entrypoints
for entry in "/README.md" "/docs/README.md" "/docs/overview/README.md" "/.project/README.md" "/.roadmap/checklist.md"; do
  if ! grep -q "$entry" "$LLMS"; then
    echo "missing entrypoint in $LLMS: $entry"
    exit 1
  fi
done

# Extract path from a bullet line with backticks (`path`)
extract_path() {
  local line="$1"
  if [[ "$line" != -* ]] || [[ "$line" != *\`* ]]; then
    return 1
  fi

  local path
  path="$(printf '%s\n' "$line" | awk -F'`' '{print $2}')"

  if [[ -z "$path" ]]; then
    return 1
  fi

  echo "$path"
  return 0
}

check_path_exists() {
  local target="$1"

  # Ignore non-path inline values
  if [[ "$target" == http*://* ]] || [[ "$target" == "HubRelay"* ]] || [[ "$target" == *.go ]] || [[ "$target" == *"" ]]; then
    return 0
  fi

  if [[ "$target" == /* ]]; then
    target=".${target}"
  fi

  # Skip module/term tokens without path context
  if [[ "$target" != */* ]] && [[ "$target" != *.md ]] && [[ "$target" != *.json ]] && [[ "$target" != *"/*" ]]; then
    return 0
  fi

  if [[ "$target" == *\** ]]; then
    local matches=( $target )
    if (( ${#matches[@]} == 0 )); then
      return 1
    fi
    return 0
  fi

  [[ -e "$target" ]]
}

while IFS= read -r line; do
  if [[ "$line" == -* && "$line" == *\`* ]]; then
    path="$(extract_path "$line" || true)"
    [[ -z "$path" ]] && continue

    if ! check_path_exists "$path"; then
      echo "missing referenced path in $LLMS: $path"
      exit 1
    fi
  fi
done < "$LLMS"

if [[ -f "$META" ]]; then
  if ! grep -q '"raw_path"' "$META"; then
    echo "missing raw_path in $META"
    exit 1
  fi
else
  echo "warning: $META not found (optional)"
fi

echo "llms navigator checks passed"
