#!/bin/bash
# this script is used to generate release notes for a given release
# first argument is the repository name
# the secon is the pull request id for the last release
# and the third is the new release number

# the script then grabs every pull request merged since that pull request
# and outputs a string of release notes

echo "Generating release notes for $2"

# get the repository name
REPO_NAME=$1

# get the last release pull request id
LAST_RELEASE_PR=$2

# get the new release number
NEW_RELEASE=$3

#get when the last release was merged
LAST_RELEASE_MERGED_AT=$(gh pr view "$LAST_RELEASE_PR" --repo "$REPO_NAME" --json mergedAt | jq -r '.mergedAt')

CHANGELIST=$(gh pr list --repo "$REPO_NAME" --base main --state merged --json title --search "merged:>$LAST_RELEASE_MERGED_AT -label:no-release")

# store the release notes in a variable so we can use it later

echo "Release $NEW_RELEASE" >>releasenotes.md

echo "$CHANGELIST" | jq -r '.[].title' | while read -r line; do
	echo " - $line" >>releasenotes.md
done

echo " "
