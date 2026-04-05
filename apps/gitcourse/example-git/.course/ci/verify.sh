#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

status_for_lesson() {
  if [[ "$1" == "true" ]]; then
    printf "done"
  else
    printf "pending"
  fi
}

check_exists() {
  local target="$1"
  [[ -e "$target" ]]
}

check_contains() {
  local target="$1"
  local needle="$2"
  [[ -f "$target" ]] && grep -q "$needle" "$target"
}

check_build() {
  npm run build >/dev/null 2>&1
}

lesson001_node_modules=false
lesson001_build=false
lesson002_component=false
lesson002_nav=false
lesson003_dependency=false
lesson003_router=false
lesson004_catalog=false
lesson005_fetch=false
lesson006_tailwind=false
lesson007_build=false

check_exists node_modules && lesson001_node_modules=true
check_build && lesson001_build=true && lesson007_build=true
check_exists src/components/Header.tsx && lesson002_component=true
check_contains src/components/Header.tsx "nav" && lesson002_nav=true
check_contains package.json "react-router-dom" && lesson003_dependency=true
check_contains src/App.tsx "BrowserRouter" && lesson003_router=true
check_contains src/App.tsx "Catalog" && lesson004_catalog=true
check_contains src/App.tsx "fetch(" && lesson005_fetch=true
check_contains src/App.tsx "grid" && lesson006_tailwind=true

cat <<JSON
{
  "course_id": "vite-react-starter",
  "lessons": [
    {
      "id": "001",
      "status": "$(status_for_lesson "$([[ "$lesson001_node_modules" == true && "$lesson001_build" == true ]] && echo true || echo false)")",
      "checks": {
        "node_modules": $lesson001_node_modules,
        "build_ok": $lesson001_build
      },
      "messages": {}
    },
    {
      "id": "002",
      "status": "$(status_for_lesson "$([[ "$lesson002_component" == true && "$lesson002_nav" == true ]] && echo true || echo false)")",
      "checks": {
        "header_component": $lesson002_component,
        "header_nav": $lesson002_nav
      },
      "messages": {}
    },
    {
      "id": "003",
      "status": "$(status_for_lesson "$([[ "$lesson003_dependency" == true && "$lesson003_router" == true ]] && echo true || echo false)")",
      "checks": {
        "router_dependency": $lesson003_dependency,
        "router_usage": $lesson003_router
      },
      "messages": {}
    },
    {
      "id": "004",
      "status": "$(status_for_lesson "$lesson004_catalog")",
      "checks": {
        "catalog_data": $lesson004_catalog
      },
      "messages": {}
    },
    {
      "id": "005",
      "status": "$(status_for_lesson "$lesson005_fetch")",
      "checks": {
        "fetch_usage": $lesson005_fetch
      },
      "messages": {}
    },
    {
      "id": "006",
      "status": "$(status_for_lesson "$lesson006_tailwind")",
      "checks": {
        "tailwind_classes": $lesson006_tailwind
      },
      "messages": {}
    },
    {
      "id": "007",
      "status": "$(status_for_lesson "$lesson007_build")",
      "checks": {
        "final_build": $lesson007_build
      },
      "messages": {}
    }
  ]
}
JSON
