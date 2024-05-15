#!/bin/bash

while getopts ":v:p:u" opt; do
  case $opt in
  p)
    DESTINATION="$OPTARG"
    ;;
  v)
    VERSION="$OPTARG"
    ;;
  u)
    UNINSTALL="true"
    ;;
  \?)
    echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

if [ -z "$DESTINATION" ]; then
  DESTINATION="/usr/local/bin"
fi

function uninstall() {
  if [ -f "$DESTINATION/prldevops" ]; then
    echo "Removing prldevops from $DESTINATION"
    sudo rm "$DESTINATION/prldevops"
    echo "prldevops has been uninstalled"
  else
    echo "prldevops is not installed in $DESTINATION"
  fi
}

function install() {
  if [ -z "$VERSION" ]; then
    # Get latest version from github
    VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"')
  fi

  if [[ ! $VERSION == release-v* ]]; then
    VERSION="release-v$VERSION"
  fi
  SHORT_VERSION="$(echo $VERSION | cut -d '-' -f 2)"

  ARCHITECTURE=$(uname -m)
  if [ "$ARCHITECTURE" = "aarch64" ]; then
    ARCHITECTURE="arm64"
  fi
  if [ "$ARCHITECTURE" = "x86_64" ]; then
    ARCHITECTURE="amd64"
  fi

  OS=$(uname -s)
  OS=$(echo "$OS" | tr '[:upper:]' '[:lower:]')
  echo "Installing prldevops $SHORT_VERSION for $OS-$ARCHITECTURE"

  DOWNLOAD_URL="https://github.com/Parallels/prl-devops-service/releases/download/$VERSION/prldevops--$OS-$ARCHITECTURE.tar.gz"

  echo "Downloading prldevops release from GitHub Releases"
  curl -sL $DOWNLOAD_URL -o prldevops.tar.gz

  echo "Extracting prldevops"
  tar -xzf prldevops.tar.gz

  if [ ! -d "$DESTINATION" ]; then
    echo "Creating destination directory: $DESTINATION"
    mkdir -p "$DESTINATION"
  fi

  if [ -f "$DESTINATION/prldevops" ]; then
    echo "Removing existing prldevops"
    sudo rm "$DESTINATION/prldevops"
  fi
  echo "Moving prldevops to $DESTINATION"
  sudo mv prldevops $DESTINATION/prldevops
  sudo chmod +x $DESTINATION/prldevops

  if [ "$OS" = "darwin" ]; then
    if [ -f "/Library/LaunchDaemons/com.parallels.prl-devops-service.plist" ]; then
      echo "Restarting prl-devops-service"
      sudo launchctl unload /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
      sudo launchctl load /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
    fi

    sudo xattr -d com.apple.quarantine $DESTINATION/prldevops
  fi

  echo "Cleaning up"
  rm prldevops.tar.gz
  echo "prldevops $SHORT_VERSION has been installed to $DESTINATION"
}

if [ -n "$UNINSTALL" ]; then
  uninstall
else
  install
fi
