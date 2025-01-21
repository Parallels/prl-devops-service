#!/bin/bash
WEBHOOK_URL=""
VERSION=""
BETA="FALSE"
while [[ $# -gt 0 ]]; do
  case $1 in
  --webhook-url)
    WEBHOOK_URL=$2
    shift
    shift
    ;;
  --version)
    VERSION=$2
    shift
    shift
    ;;
  --beta)
    BETA="TRUE"
    shift
    ;;
  *)
    echo "Invalid argument: $1" >&2
    exit 1
    ;;
  esac
done

# Validate required parameters
if [ -z "$WEBHOOK_URL" ]; then
  echo "Error: webhook-url is required"
  exit 1
fi

if [ -z "$VERSION" ]; then
  echo "Error: version is required"
  exit 1
fi

# Get the latest changelog content
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
CHANGELOG_CONTENT=$("$SCRIPT_DIR/get-latest-changelog.sh")

# Escape special characters for JSON
CHANGELOG_CONTENT=$(echo "$CHANGELOG_CONTENT" | jq -Rs .)
CHANGELOG_CONTENT=${CHANGELOG_CONTENT#\"} # Remove leading quote
CHANGELOG_CONTENT=${CHANGELOG_CONTENT%\"} # Remove trailing quote

# Check if changelog content is longer than 4096 characters
if [ ${#CHANGELOG_CONTENT} -gt 4096 ]; then
  CHANGELOG_CONTENT="${CHANGELOG_CONTENT:0:3900}..."
  CHANGELOG_CONTENT+=$"\nFor the complete changelog, visit: https://github.com/Parallels/terraform-provider-parallels-desktop/releases/tag/v${VERSION}"
fi

TITLE="📢 New Release v${VERSION}"
if [ "$BETA" = "TRUE" ]; then
  TITLE="🧪 New Beta Release v${VERSION}"
fi

# Create the JSON payload
JSON_PAYLOAD=$(
  cat <<EOF
{
  "embeds": [{
  "title": "${TITLE}",
  "description": "${CHANGELOG_CONTENT}",
  "color": 3447003
  }]
}
EOF
)
echo "$JSON_PAYLOAD"

# Send the webhook request
curl -H "Content-Type: application/json" \
  -d "$JSON_PAYLOAD" \
  "$WEBHOOK_URL"

if [ $? -eq 0 ]; then
  echo "Successfully posted changelog to Discord"
else
  echo "Failed to post changelog to Discord webhook"
  exit 1
fi
