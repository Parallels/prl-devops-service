#!/bin/bash
# get-latest-changelog.sh
#
# Reads CHANGELOG.md, finds the most recent version section, and outputs its content.
#
# Usage:
#   get-latest-changelog.sh [options]
#
# Options:
#   --changelog     Path to CHANGELOG.md (default: CHANGELOG.md)
#   --file          Output file path (default: release_notes.md)
#   --output-to-file  Write to file instead of stdout
#   -m | --mode     GENERATE (default) or HIGHEST_VERSION

set -euo pipefail

CHANGELOG_FILE="CHANGELOG.md"
OUTPUT_FILE="release_notes.md"
OUTPUT_TO_FILE="false"
MODE="GENERATE"

while [[ $# -gt 0 ]]; do
  case $1 in
    -m|--mode)         MODE="$2";           shift 2 ;;
    --changelog)       CHANGELOG_FILE="$2"; shift 2 ;;
    --file)            OUTPUT_FILE="$2";    shift 2 ;;
    --output-to-file)  OUTPUT_TO_FILE="true"; shift ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

get_highest_version() {
  grep -E '^## \[[0-9]+\.[0-9]+\.[0-9]+\]' "$CHANGELOG_FILE" \
    | sed 's/^## \[//;s/\].*//' \
    | sort -Vr \
    | head -1
}

get_content_for_version() {
  local version="$1"
  awk "/^## \[${version}\]/{found=1; next} found && /^## \[/{exit} found{print}" "$CHANGELOG_FILE" \
    | sed '/^[[:space:]]*$/{ /./!d }' # collapse multiple blank lines
}

run_generate() {
  local version
  version=$(get_highest_version)

  if [[ -z "$version" ]]; then
    echo "No version entries found in ${CHANGELOG_FILE}" >&2
    exit 1
  fi

  local content
  content=$(get_content_for_version "$version")

  local output
  output="# Release Notes for v${version}

${content}"

  if [[ "$OUTPUT_TO_FILE" == "true" ]]; then
    echo "$output" > "$OUTPUT_FILE"
  else
    echo "$output"
  fi
}

case "$MODE" in
  GENERATE)        run_generate ;;
  HIGHEST_VERSION) get_highest_version ;;
  *) echo "Unknown mode: $MODE" >&2; exit 1 ;;
esac
