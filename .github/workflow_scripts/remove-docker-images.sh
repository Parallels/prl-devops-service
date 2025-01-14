#!/bin/bash

MODE="UNKNOWN"
PATTERN=""
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

function list() {
  OUT=$(hub-tool tag ls cjlapao/prl-devops-service --format json)
  LINES=$(echo "$OUT" | jq -r '.[].Name')
  echo "$LINES" | while IFS= read -r line; do
    if [[ $line =~ $PATTERN ]]; then
      echo "$line"
    fi
  done
}

function remove() {
  echo "WARNING: You are about to permanently delete images that match the pattern: $PATTERN"
  echo "         This action is irreversible"
  read -r -p "Are you sure you want to continue? (yes/no): " confirm
  if [ "$confirm" != "yes" ]; then
    echo "Operation aborted."
    exit 1
  fi

  lines=$(list)
  if [ -z "$lines" ]; then
    echo "No images found matching pattern: $PATTERN"
    exit 0
  fi

  echo "$lines" | while IFS= read -r line; do
    if [ -n "$line" ]; then
      echo "Deleting image $line"
      hub-tool tag rm "$line" -f
    fi
  done
}

if [ "$MODE" == "LIST" ]; then
  list
elif [ "$MODE" == "REMOVE" ]; then
  remove "$PATTERN"
fi
