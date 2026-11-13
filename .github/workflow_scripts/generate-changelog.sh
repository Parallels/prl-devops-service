#!/bin/bash
# generate-changelog.sh
#
# Collects changelog entries from merged PRs since the last release, optionally
# generates an AI-written description via the Claude API, and writes the result
# to CHANGELOG.md (GENERATE mode) or release_notes.md (RELEASE mode).
#
# PR bodies must follow the pull_request_template.md format:
#   ## Changelog
#   - feat: ...
#   - fix: ...
#
# Usage:
#   generate-changelog.sh [options]
#
# Options:
#   -m | --mode          GENERATE (default), RELEASE, or ENTRIES
#   -v | --version       Version string, e.g. 1.2.3
#   -r | --repo          GitHub repository, e.g. Parallels/prl-devops-service
#        --changelog     Path to CHANGELOG.md (default: CHANGELOG.md)
#        --file          Path to release notes output file (default: release_notes.md)
#        --output-to-file  Write output to file instead of stdout
#        --from-changelog  RELEASE only: use the exact version's committed changelog
#        --base          Base branch for merged-PR collection (default: main).
#                        ENTRIES --base dev captures PRs merged into dev.
#        --boundary-label  Label of the PRs that mark the collection boundary,
#                        always searched on main (default: release-request;
#                        dev integration uses: dev-integration)
#        --since         Boundary timestamp (mergedAt), overrides the
#                        boundary-label lookup
#        --no-ai         Skip AI description generation even if ANTHROPIC_API_KEY is set
#        --verbose       Enable verbose logging
#
# Modes:
#   GENERATE  Collect entries (AI-summarised) and insert them into CHANGELOG.md.
#   RELEASE   Write release notes (from CHANGELOG via --from-changelog, or from
#             merged PRs) for a PR body / GitHub Release.
#   ENTRIES   Print the collected raw bullets only — no AI, no file writes.
#             Used by update-dev-integration-pr.yml to capture dev merges.

set -euo pipefail

# ── Defaults ──────────────────────────────────────────────────────────────────
VERBOSE="false"
CHANGELOG_FILE="CHANGELOG.md"
RELEASE_NOTES_FILE="release_notes.md"
OUTPUT_TO_FILE="false"
MODE="GENERATE"
NEW_RELEASE=""
REPO_NAME=""
NO_AI="false"
FROM_CHANGELOG="false"
BASE_BRANCH="main"
BOUNDARY_LABEL="release-request"
SINCE=""

# ── Argument parsing ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case $1 in
    -m|--mode)            MODE="$2";             shift 2 ;;
    -v|--version)         NEW_RELEASE="$2";      shift 2 ;;
    -r|--repo)            REPO_NAME="$2";        shift 2 ;;
    --changelog)          CHANGELOG_FILE="$2";   shift 2 ;;
    --file)               RELEASE_NOTES_FILE="$2"; shift 2 ;;
    --output-to-file)     OUTPUT_TO_FILE="true"; shift ;;
    --from-changelog)     FROM_CHANGELOG="true"; shift ;;
    --base)               BASE_BRANCH="$2";    shift 2 ;;
    --boundary-label)     BOUNDARY_LABEL="$2"; shift 2 ;;
    --since)              SINCE="$2";          shift 2 ;;
    --no-ai)              NO_AI="true";          shift ;;
    --verbose)            VERBOSE="true";        shift ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done

# ── Helpers ───────────────────────────────────────────────────────────────────
log() { [[ "$VERBOSE" == "true" ]] && echo "[changelog] $*" >&2 || true; }

require() {
  command -v "$1" >/dev/null 2>&1 || { echo "Required tool not found: $1" >&2; exit 1; }
}

require gh
require jq

# ── Parse a Dependabot PR title into a changelog entry ───────────────────────
# Titles arrive with a conventional commit prefix, e.g.:
#   "deps(deps): bump the minor-and-patch group across 1 directory with 18 updates"
#   "deps/actions(deps): bump the actions group with 2 updates"
#   "deps(deps-dev): bump the major group with 3 updates"
#   "deps(deps): bump axios from 1.13.2 to 1.13.6"
dependabot_entry_from_title() {
  local title="$1"

  # Strip conventional commit prefix (everything up to and including ": ")
  local bump
  bump=$(echo "$title" | sed -E 's/^[^:]+:[[:space:]]*//')

  # Grouped bump: "bump the <group> group ... with <N> updates"
  if echo "$bump" | grep -qiE "bump the .+ group .* with [0-9]+ updates?"; then
    local group count
    group=$(echo "$bump" | sed -E 's/.*[Bb]ump the ([^ ]+) group.*/\1/')
    count=$(echo "$bump" | sed -E 's/.*with ([0-9]+) updates?.*/\1/')
    echo "- deps: bump ${count} packages (${group})"
    return
  fi

  # Individual bump: "bump <package> from X to Y"
  if echo "$bump" | grep -qiE "^bump .+ from .+ to "; then
    local rest
    rest=$(echo "$bump" | sed -E 's/^[Bb]ump //')
    echo "- deps: bump ${rest}"
    return
  fi

  # Fallback: use the stripped bump text
  echo "- deps: ${bump}"
}

# ── Collect changelog entries from merged PRs ─────────────────────────────────
# Reads the ## Changelog section from regular PR bodies and extracts prefixed
# bullets. For Dependabot PRs, derives an entry from the PR title.
collect_entries() {
  log "Fetching merged PRs for ${REPO_NAME}"

  # Find when the most recent boundary PR (release PR, or dev-integration PR)
  # was merged so we only include newer PRs. Boundary PRs always target main.
  if [[ -n "$SINCE" ]]; then
    LAST_RELEASE_MERGED_AT="$SINCE"
  else
    LAST_RELEASE_MERGED_AT=$(
      gh pr list \
        --repo "$REPO_NAME" \
        --base main \
        --state merged \
        --json mergedAt \
        --search "label:$BOUNDARY_LABEL" \
        --limit 1 \
      | jq -r '.[0].mergedAt // empty'
    ) || return 1
  fi

  if [[ -n "$LAST_RELEASE_MERGED_AT" ]]; then
    log "Last ${BOUNDARY_LABEL} PR merged at: $LAST_RELEASE_MERGED_AT"
    SEARCH_QUERY="merged:>${LAST_RELEASE_MERGED_AT} -label:$BOUNDARY_LABEL"
  else
    log "No previous ${BOUNDARY_LABEL} PR found — collecting all merged PRs"
    SEARCH_QUERY="-label:$BOUNDARY_LABEL"
  fi

  CHANGELIST=$(
    gh pr list \
      --repo "$REPO_NAME" \
      --base "$BASE_BRANCH" \
      --state merged \
      --json number,title,body,author \
      --search "$SEARCH_QUERY" \
      --limit 200
  ) || return 1

  PR_COUNT=$(echo "$CHANGELIST" | jq 'length')
  log "Found ${PR_COUNT} PR(s) to process"

  if [[ "$VERBOSE" == "true" ]]; then
    echo "$CHANGELIST" | jq -r '.[] | "\(.author.login)\t\(.title)"' | \
      while IFS=$'\t' read -r author title; do
        log "  PR author=${author} title=${title}"
      done
  fi

  # Identify Dependabot PRs: author contains "dependabot" OR title contains "bump"
  # Author comes back as "app/dependabot" from gh CLI (not "dependabot[bot]")
  # Use contains() instead of test() regex to avoid jq regex engine differences
  is_dependabot_filter='((.author.login | ascii_downcase) | contains("dependabot")) or ((.title | ascii_downcase) | contains("bump"))'

  # Extract ## Changelog section from regular (non-Dependabot) PRs
  ENTRIES=$(
    echo "$CHANGELIST" | \
    jq -r ".[] | select(($is_dependabot_filter) | not) | .body // \"\"" | \
    awk '
      /^## Changelog/ { in_section=1; next }
      /^##/           { in_section=0 }
      in_section && /^[[:space:]]*-[[:space:]]/ { print }
    ' | \
    sed 's/^[[:space:]]*//' | \
    grep -v '^-[[:space:]]*$' || true
  )

  # Extract entries from Dependabot PRs via title parsing
  DEPENDABOT_ENTRIES=""
  while IFS= read -r title; do
    [[ -z "$title" ]] && continue
    entry=$(dependabot_entry_from_title "$title")
    DEPENDABOT_ENTRIES="${DEPENDABOT_ENTRIES}${entry}"$'\n'
    log "Dependabot entry: ${entry}"
  done < <(echo "$CHANGELIST" | jq -r ".[] | select($is_dependabot_filter) | .title")

  # Combine: regular entries first, then dependency bumps
  ALL_ENTRIES=""
  [[ -n "$ENTRIES" ]] && ALL_ENTRIES="${ENTRIES}"$'\n'
  [[ -n "$DEPENDABOT_ENTRIES" ]] && ALL_ENTRIES="${ALL_ENTRIES}${DEPENDABOT_ENTRIES}"

  echo "$ALL_ENTRIES" | grep -v '^[[:space:]]*$' || true
}

# ── Generate AI description via Claude API ────────────────────────────────────
generate_ai_description() {
  local entries="$1"

  if [[ "$NO_AI" == "true" ]]; then
    log "AI generation skipped (--no-ai)"
    echo ""
    return
  fi

  if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
    log "ANTHROPIC_API_KEY not set — skipping AI description"
    echo ""
    return
  fi

  if [[ -z "$entries" ]]; then
    log "No entries to summarise — skipping AI description"
    echo ""
    return
  fi

  log "Generating AI description via Claude API"

  local prompt
  prompt="You are writing release notes for PRL DevOps Service, a Go-based backend service that automates and manages Parallels Desktop VM fleets.

Based on the changelog entries below, write a concise 2-3 sentence release description that:
- Summarises the most significant user-visible improvements
- Groups related changes naturally (don't list every bullet)
- Is professional, clear, and written for end users and operators
- Does NOT start with 'This release' or 'In this release'

Changelog entries:
${entries}

Reply with only the description text. No headings, no bullet points, no preamble."

  local response
  response=$(
    curl -sf https://api.anthropic.com/v1/messages \
      -H "x-api-key: ${ANTHROPIC_API_KEY}" \
      -H "anthropic-version: 2023-06-01" \
      -H "content-type: application/json" \
      -d "$(jq -n \
        --arg model "claude-haiku-4-5-20251001" \
        --arg prompt "$prompt" \
        '{
          model: $model,
          max_tokens: 256,
          messages: [{ role: "user", content: $prompt }]
        }'
      )"
  ) || { log "Claude API call failed — skipping AI description"; echo ""; return; }

  echo "$response" | jq -r '.content[0].text // empty'
}

# ── Format the full changelog block ──────────────────────────────────────────
build_changelog_block() {
  local version="$1"
  local entries="$2"
  local ai_description="$3"
  local today
  today=$(date '+%Y-%m-%d')

  {
    echo "## [${version}] - ${today}"
    echo ""
    if [[ -n "$ai_description" ]]; then
      echo "$ai_description"
      echo ""
    fi
    if [[ -n "$entries" ]]; then
      echo "$entries"
    else
      echo "- chore: maintenance release"
    fi
    echo ""
  }
}

# ── Insert content into CHANGELOG.md ─────────────────────────────────────────
insert_into_changelog() {
  local block="$1"
  local tmp block_file
  tmp=$(mktemp)
  block_file=$(mktemp)

  # Write block to a temp file — avoids awk -v multiline mangling
  printf '%s\n' "$block" > "$block_file"

  # Find the first existing ## version line and insert before it,
  # or append after the header block if no versions exist yet.
  local insert_after
  insert_after=$(grep -n "^## \[" "$CHANGELOG_FILE" | head -1 | cut -d: -f1 || true)

  if [[ -z "$insert_after" ]]; then
    # No versions yet — append to end of file (strip wc -l leading whitespace)
    insert_after=$(wc -l < "$CHANGELOG_FILE" | tr -d ' ')
  else
    insert_after=$((insert_after - 1))
  fi

  awk -v line="$insert_after" -v bf="$block_file" '
    NR == line { print; while ((getline bline < bf) > 0) print bline; next }
    { print }
  ' "$CHANGELOG_FILE" > "$tmp"

  rm -f "$block_file"
  mv "$tmp" "$CHANGELOG_FILE"
}

# ── Append content to an existing version section ────────────────────────────
append_to_version_section() {
  local version_line="$1"
  local entries="$2"
  local tmp
  tmp=$(mktemp)
  local entries_file
  entries_file=$(mktemp)
  echo "$entries" > "$entries_file"

  # Find the end of this version's section
  local end_line
  end_line=$(awk -v start="$version_line" \
    'NR > start && /^## \[/ { print NR-1; exit }' "$CHANGELOG_FILE")
  [[ -z "$end_line" ]] && end_line=$(wc -l < "$CHANGELOG_FILE" | tr -d ' ')

  awk -v end="$end_line" -v ef="$entries_file" '
    NR == end { print; while ((getline line < ef) > 0) print line; next }
    { print }
  ' "$CHANGELOG_FILE" > "$tmp"

  mv "$tmp" "$CHANGELOG_FILE"
  rm "$entries_file"
}

# ── GENERATE mode: update CHANGELOG.md ───────────────────────────────────────
run_generate() {
  if [[ -z "$NEW_RELEASE" ]]; then
    echo "--version is required in GENERATE mode" >&2; exit 1
  fi
  if [[ -z "$REPO_NAME" ]]; then
    echo "--repo is required" >&2; exit 1
  fi

  local entries ai_description
  entries=$(collect_entries)
  ai_description=$(generate_ai_description "$entries")

  local version_line
  version_line=$(grep -n "^## \[${NEW_RELEASE}\]" "$CHANGELOG_FILE" | cut -d: -f1 || true)

  if [[ -z "$version_line" ]]; then
    log "Creating new section for ${NEW_RELEASE}"
    local block
    block=$(build_changelog_block "$NEW_RELEASE" "$entries" "$ai_description")
    insert_into_changelog "$block"
  else
    log "Appending to existing section for ${NEW_RELEASE}"
    append_to_version_section "$version_line" "$entries"
  fi

  log "CHANGELOG.md updated successfully"
}

# ── RELEASE mode: write release_notes.md (for PR body / GitHub Release) ──────
run_release() {
  if [[ -z "$REPO_NAME" ]]; then
    echo "--repo is required" >&2; exit 1
  fi

  local entries ai_description
  if [[ "$FROM_CHANGELOG" == "true" ]]; then
    # Once the release PR merges, collect_entries starts AFTER that PR. The
    # committed section retains the changes and summary collected before it.
    if [[ -z "$NEW_RELEASE" || ! -f "$CHANGELOG_FILE" ]]; then
      echo "--from-changelog requires --version and an existing changelog" >&2
      return 1
    fi
    entries=$(awk -v heading="## [${NEW_RELEASE}]" '
      /^## \[/ {
        if (found) exit
        if ($0 == heading || index($0, heading " ") == 1) { found=1; next }
      }
      found { print }
      END { if (!found) exit 1 }
    ' "$CHANGELOG_FILE") || {
      echo "Missing changelog section for ${NEW_RELEASE}" >&2
      return 1
    }
    ai_description=""
  else
    entries=$(collect_entries) || return 1
    ai_description=$(generate_ai_description "$entries")
  fi

  local output
  output=$(
    {
      if [[ -n "$NEW_RELEASE" ]]; then
        echo "# Release ${NEW_RELEASE}"
      fi
      echo ""
      if [[ -n "$ai_description" ]]; then
        echo "$ai_description"
        echo ""
      fi
      echo "## Changes"
      echo ""
      if [[ -n "$entries" ]]; then
        echo "$entries"
      else
        echo "- chore: maintenance release"
      fi
    }
  )

  if [[ "$OUTPUT_TO_FILE" == "true" ]]; then
    echo "$output" > "$RELEASE_NOTES_FILE"
    log "Release notes written to ${RELEASE_NOTES_FILE}"
  else
    echo "$output"
  fi
}

# ── ENTRIES mode: print collected raw bullets only ───────────────────────────
run_entries() {
  if [[ -z "$REPO_NAME" ]]; then
    echo "--repo is required" >&2; exit 1
  fi
  collect_entries
}

# ── Dispatch ──────────────────────────────────────────────────────────────────
case "$MODE" in
  GENERATE) run_generate ;;
  RELEASE)  run_release  ;;
  ENTRIES)  run_entries  ;;
  *) echo "Unknown mode: $MODE (use GENERATE, RELEASE or ENTRIES)" >&2; exit 1 ;;
esac
