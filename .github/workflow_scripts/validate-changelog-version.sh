#!/bin/bash
# validate-changelog-version.sh
#
# Validates that CHANGELOG.md contains a section for the version in VERSION file.
# Exits non-zero with a clear error if not found — used as a gate in production releases.
#
# Usage:
#   validate-changelog-version.sh [--changelog <path>] [--version-file <path>]

set -euo pipefail

CHANGELOG_FILE="CHANGELOG.md"
VERSION_FILE="VERSION"

while [[ $# -gt 0 ]]; do
  case $1 in
    --changelog)    CHANGELOG_FILE="$2"; shift 2 ;;
    --version-file) VERSION_FILE="$2";   shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

if [[ ! -f "$VERSION_FILE" ]]; then
  echo "Version file not found: ${VERSION_FILE}" >&2
  exit 1
fi

if [[ ! -f "$CHANGELOG_FILE" ]]; then
  echo "Changelog file not found: ${CHANGELOG_FILE}" >&2
  exit 1
fi

VERSION=$(tr -d '[:space:]' < "$VERSION_FILE")

if [[ -z "$VERSION" ]]; then
  echo "Could not read version from ${VERSION_FILE}" >&2
  exit 1
fi

if grep -qE "^## \[${VERSION}\]" "$CHANGELOG_FILE"; then
  echo "OK: CHANGELOG.md contains entry for v${VERSION}"
  exit 0
else
  echo "FAIL: CHANGELOG.md has no entry for v${VERSION}" >&2
  echo "" >&2
  echo "  The production release requires a changelog entry." >&2
  echo "  Run the 'Create Release PR' workflow to generate it, then merge the PR before releasing." >&2
  exit 1
fi
