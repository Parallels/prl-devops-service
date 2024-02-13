#!/bin/bash

while getopts ":f:" opt; do
	case $opt in
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
echo "$VERSION"
