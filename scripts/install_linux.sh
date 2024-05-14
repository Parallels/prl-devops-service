#!/bin/bash

while getopts ":v:p" opt; do
  case $opt in
  p)
    DESTINATION="$OPTARG"
    ;;
  v)
    VERSION="$OPTARG"
    ;;
  \?)
    echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

# Check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq could not be found"
    echo "Please install jq before running this script"
    exit
fi

if [ -z "$DESTINATION" ]; then
  DESTINATION="/usr/local/bin"
fi

if [ -z "$VERSION" ]; then
  # Get latest version from github
  VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases/latest | jq -r .tag_name)
fi

DOWNLOAD_URL="https://github.com/Parallels/prl-devops-service/releases/download/$VERSION/prldevops--linux-amd64.tar.gz"

echo "Downloading prldevops $VERSION from $DOWNLOAD_URL"
curl -L $DOWNLOAD_URL -o prldevops.tar.gz

echo "Extracting prldevops"
tar -xzf prldevops.tar.gz

echo "Moving prldevops to $DESTINATION"
mv prldevops $DESTINATION
chmod +x $DESTINATION/prldevops

echo "Cleaning up"
rm prldevops.tar.gz