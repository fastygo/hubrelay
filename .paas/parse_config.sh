#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_FILE="${ROOT_DIR}/.paas/config.yml"

usage() {
  cat <<'EOF'
Usage:
  bash ./.paas/parse_config.sh --extension <name> [--config <path>] [--no-comments]

Description:
  Prints a full export list for the requested extension in the form:
    export INPUT_FOO=bar

Resolution order:
  1. Current INPUT_* environment variables
  2. defaults: from ./.paas/config.yml
  3. default: values from the extension inputs
  4. Empty string when nothing is set
EOF
}

normalize_config_value() {
  local value="$1"
  if [[ "${value}" == '""' ]]; then
    printf ''
  elif [[ "${value}" == \"*\" && "${value}" == *\" ]]; then
    value="${value#\"}"
    value="${value%\"}"
    printf '%s' "${value}"
  else
    printf '%s' "${value}"
  fi
}

input_name_to_key() {
  printf 'INPUT_%s' "$(printf '%s' "$1" | tr '[:lower:]-' '[:upper:]_')"
}

quote_for_export() {
  local value="$1"
  local escaped=""
  printf -v escaped '%q' "${value}"
  printf '%s' "${escaped}"
}

EXTENSION=""
SHOW_COMMENTS="true"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --extension)
      [[ $# -ge 2 ]] || { echo "Missing value for --extension" >&2; exit 1; }
      EXTENSION="$2"
      shift 2
      ;;
    --config)
      [[ $# -ge 2 ]] || { echo "Missing value for --config" >&2; exit 1; }
      CONFIG_FILE="$2"
      shift 2
      ;;
    --no-comments)
      SHOW_COMMENTS="false"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${EXTENSION}" ]]; then
  usage
  exit 1
fi

if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "Config file not found: ${CONFIG_FILE}" >&2
  exit 1
fi

EXTENSIONS_DIR_REL="$(awk -F': *' '/^extensions_dir:/ {print $2; exit}' "${CONFIG_FILE}")"
EXTENSIONS_DIR_REL="${EXTENSIONS_DIR_REL:-.paas/extensions}"
if [[ "${EXTENSIONS_DIR_REL}" = /* || "${EXTENSIONS_DIR_REL}" =~ ^[A-Za-z]:/ ]]; then
  EXTENSIONS_DIR="${EXTENSIONS_DIR_REL}"
else
  EXTENSIONS_DIR="${ROOT_DIR}/${EXTENSIONS_DIR_REL}"
fi

EXTENSION_FILE="${EXTENSIONS_DIR}/${EXTENSION%.yml}.yml"
if [[ ! -f "${EXTENSION_FILE}" ]]; then
  echo "Extension file not found: ${EXTENSION_FILE}" >&2
  exit 1
fi

declare -A config_defaults=()
declare -A extension_defaults=()
declare -A extension_has_default=()
declare -A extension_required=()
declare -A resolved_values=()
declare -A value_sources=()
declare -a extension_key_order=()

while IFS=$'\t' read -r key raw_value; do
  [[ -z "${key}" ]] && continue
  config_defaults["${key}"]="$(normalize_config_value "${raw_value}")"
done < <(
  awk '
    /^defaults:[[:space:]]*$/ { in_defaults=1; next }
    in_defaults && /^[^[:space:]]/ { in_defaults=0 }
    in_defaults && /^  [A-Za-z0-9_]+:/ {
      line=$0
      sub(/^  /, "", line)
      key=line
      sub(/:.*/, "", key)
      value=line
      sub(/^[^:]+:[[:space:]]*/, "", value)
      print key "\t" value
    }
  ' "${CONFIG_FILE}"
)

current_name=""
current_default=""
current_has_default="false"
current_required="false"
in_inputs=0

finalize_extension_input() {
  if [[ -z "${current_name}" ]]; then
    return
  fi

  local key
  key="$(input_name_to_key "${current_name}")"
  extension_required["${key}"]="${current_required}"
  if [[ "${current_has_default}" == "true" ]]; then
    extension_defaults["${key}"]="${current_default}"
    extension_has_default["${key}"]="true"
  fi
  extension_key_order+=("${key}")
  current_name=""
  current_default=""
  current_has_default="false"
  current_required="false"
}

while IFS= read -r raw_line || [[ -n "${raw_line}" ]]; do
  line="${raw_line%$'\r'}"
  if [[ "${line}" =~ ^inputs:[[:space:]]*$ ]]; then
    in_inputs=1
    continue
  fi
  if [[ "${in_inputs}" -eq 1 && "${line}" =~ ^steps:[[:space:]]*$ ]]; then
    finalize_extension_input
    break
  fi
  if [[ "${in_inputs}" -ne 1 ]]; then
    continue
  fi

  if [[ "${line}" =~ ^[[:space:]]*-[[:space:]]name:[[:space:]]*(.+)$ ]]; then
    finalize_extension_input
    current_name="${BASH_REMATCH[1]}"
    current_default=""
    current_has_default="false"
    current_required="false"
    continue
  fi
  if [[ -z "${current_name}" ]]; then
    continue
  fi
  if [[ "${line}" =~ ^[[:space:]]+required:[[:space:]]*true[[:space:]]*$ ]]; then
    current_required="true"
    continue
  fi
  if [[ "${line}" =~ ^[[:space:]]+default:[[:space:]]*(.*)$ ]]; then
    current_default="$(normalize_config_value "${BASH_REMATCH[1]}")"
    current_has_default="true"
  fi
done < "${EXTENSION_FILE}"

finalize_extension_input

if [[ "${SHOW_COMMENTS}" == "true" ]]; then
  echo "# Extension: ${EXTENSION%.yml}"
  echo "# Extension file: ${EXTENSION_FILE}"
  echo "# Config file: ${CONFIG_FILE}"
  echo "# Resolution order: env -> config defaults -> extension defaults -> empty"
  echo
fi

for key in "${extension_key_order[@]}"; do
  value=""
  source="empty"

  if [[ "${!key+x}" == x ]]; then
    value="${!key}"
    source="env"
  elif [[ -v config_defaults["${key}"] ]]; then
    value="${config_defaults["${key}"]}"
    source="config"
  elif [[ "${extension_has_default["${key}"]:-false}" == "true" ]]; then
    value="${extension_defaults["${key}"]}"
    source="extension default"
  fi

  resolved_values["${key}"]="${value}"
  value_sources["${key}"]="${source}"

  if [[ "${SHOW_COMMENTS}" == "true" ]]; then
    comment="${source}"
    if [[ "${extension_required["${key}"]:-false}" == "true" ]]; then
      comment="${comment}, required"
    fi
    printf 'export %s=%s  # %s\n' "${key}" "$(quote_for_export "${value}")" "${comment}"
  else
    printf 'export %s=%s\n' "${key}" "$(quote_for_export "${value}")"
  fi
done
