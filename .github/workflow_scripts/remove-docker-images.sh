#!/bin/bash

set -euo pipefail

MODE="UNKNOWN"
PATTERN=""
NO_CONFIRM="false"
IMAGE_REPOSITORY="${IMAGE_REPOSITORY:-ghcr.io/parallels/prl-devops-service}"

while [[ $# -gt 0 ]]; do
  case $1 in
  rm)
    MODE="REMOVE"
    shift
    ;;
  ls)
    MODE="LIST"
    shift
    ;;
  --pattern)
    PATTERN=$2
    shift
    shift
    ;;
  --filter)
    PATTERN=$2
    shift
    shift
    ;;
  --no-confirm)
    NO_CONFIRM="true"
    shift
    ;;
  --repository)
    IMAGE_REPOSITORY=$2
    shift
    shift
    ;;
  *)
    echo "Invalid argument: $1" >&2
    exit 1
    ;;
  esac
done

if [ "$MODE" == "UNKNOWN" ]; then
  echo "You need to specify the mode (rm, ls) with the first argument"
  exit 1
fi

if [ "$PATTERN" == "" ]; then
  PATTERN=".*"
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh is required to manage GHCR package versions" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required to parse GHCR package versions" >&2
  exit 1
fi

REPOSITORY_PATH=${IMAGE_REPOSITORY#ghcr.io/}
GHCR_OWNER=${REPOSITORY_PATH%%/*}
GHCR_PACKAGE=${REPOSITORY_PATH#*/}

if [ "$GHCR_OWNER" == "$GHCR_PACKAGE" ] || [ -z "$GHCR_OWNER" ] || [ -z "$GHCR_PACKAGE" ]; then
  echo "Invalid GHCR repository: $IMAGE_REPOSITORY" >&2
  exit 1
fi

VERSIONS_API="/orgs/${GHCR_OWNER}/packages/container/${GHCR_PACKAGE}/versions"

function matching_versions() {
  gh api --paginate "${VERSIONS_API}?per_page=100" |
    jq -r --arg pattern "$PATTERN" '
      .[] |
      {
        id: .id,
        tags: (.metadata.container.tags // [])
      } |
      select(.tags | any(test($pattern))) |
      "\(.id)\t\(.tags | map(select(test($pattern))) | join(","))\t\(.tags | join(","))"
    '
}

function list() {
  matching_versions | while IFS=$'\t' read -r _ matched_tags _; do
    echo "$matched_tags" | tr ',' '\n'
  done
}

function remove() {
  if [ "$NO_CONFIRM" == "false" ]; then
    echo "WARNING: You are about to permanently delete GHCR package versions from $IMAGE_REPOSITORY"
    echo "         Matching tags pattern: $PATTERN"
    echo "         This action is irreversible"
    read -r -p "Are you sure you want to continue? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
      echo "Operation aborted."
      exit 1
    fi
  fi

  lines=$(matching_versions)
  if [ -z "$lines" ]; then
    echo "No images found matching pattern: $PATTERN"
    exit 0
  fi

  echo "$lines" | awk -F '\t' '!seen[$1]++ {print $1 "\t" $3}' | while IFS=$'\t' read -r version_id tags; do
    if [ -n "$version_id" ]; then
      echo "Deleting package version $version_id from $IMAGE_REPOSITORY (tags: $tags)"
      gh api --method DELETE "${VERSIONS_API}/${version_id}"
    fi
  done
}

if [ "$MODE" == "LIST" ]; then
  list
elif [ "$MODE" == "REMOVE" ]; then
  remove "$PATTERN"
fi
