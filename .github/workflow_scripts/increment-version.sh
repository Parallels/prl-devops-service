#!/bin/bash

while getopts ":t:f:" opt; do
	case $opt in
	t)
		TYPE=$OPTARG
		;;
	f)
		FILE=$OPTARG
		;;
	\?)
		echo "Invalid option: -$OPTARG" >&2
		exit 1
		;;
	esac
done

if [ -z "$FILE" ]; then
	echo "You need to specify the version file with the -f flag"
fi

VERSION=$(cat "$FILE")

MAJOR=$(echo "$VERSION" | cut -d. -f1)
MINOR=$(echo "$VERSION" | cut -d. -f2)
PATCH=$(echo "$VERSION" | cut -d. -f3)

if [ "$TYPE" == "major" ]; then
	MAJOR=$((MAJOR + 1))
	MINOR=0
	PATCH=0
elif [ "$TYPE" == "minor" ]; then
	MINOR=$((MINOR + 1))
	PATCH=0
elif [ "$TYPE" == "patch" ]; then
	PATCH=$((PATCH + 1))
else
	echo "Invalid version type. Use 'major', 'minor' or 'patch'"
	exit 1
fi

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "$NEW_VERSION"
