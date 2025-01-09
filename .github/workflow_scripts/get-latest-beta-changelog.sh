#!/bin/bash

CHANGELOG_FILE="CHANGELOG.md"
OUTPUT_FILE="release_notes.md"
OUTPUT_TO_FILE="FALSE"
REPO=""

while [[ $# -gt 0 ]]; do
  case $1 in
  -m)
    MODE=$2
    shift
    shift
    ;;
  --version)
    VERSION=$2
    shift
    shift
    ;;
  -v)
    VERSION=$2
    shift
    shift
    ;;
  --CHANGELOG_FILE)
    CHANGELOG_FILE=$2
    shift
    shift
    ;;
  --repo)
    REPO=$2
    shift
    shift
    ;;
  --file)
    OUTPUT_FILE shift
    shift
    ;;
  --output-to-file)
    OUTPUT_TO_FILE="TRUE"
    shift
    ;;
  *)
    echo "Invalid argument: $1" >&2
    exit 1
    ;;
  esac
done

function generate_release_notes() {
  # Get the content for the highest version
  content=$(./.github/workflow_scripts/generate-changelog.sh --repo "$REPO" --mode RELEASE)

  # Write the content to the output file
  if [ "$OUTPUT_TO_FILE" == "TRUE" ]; then
    echo -e "# Release Notes for v$VERSION\n\n$content" >$OUTPUT_FILE
  else
    echo -e "# Release Notes for v$VERSION\n\n$content"
  fi
}

if [ -z "$REPO" ]; then
  echo "Error: --repo is not set" >&2
  exit 1
fi

generate_release_notes
