#!/bin/bash
MODE="INSTALL"
INSTALL_SERVICE="true"
STD_USER="false"
PRE_RELEASE="false"
while [[ $# -gt 0 ]]; do
  case $1 in
  -i)
    MODE="INSTALL"
    shift # past argument
    ;;
  --install)
    MODE="INSTALL"
    shift # past argument
    ;;
  -u)
    MODE="UNINSTALL"
    shift # past argument
    ;;
  --uninstall)
    MODE="UNINSTALL"
    shift # past argument
    ;;
  -p)
    DESTINATION=$2
    shift # past argument
    shift # past argument
    ;;
  --path)
    DESTINATION=$2
    shift # past argument
    shift # past argument
    ;;
  --no-service)
    INSTALL_SERVICE="false"
    shift # past argument
    ;;
  -v)
    VERSION=$2
    shift # past argument
    shift # past argument
    ;;
  --version)
    VERSION=$2
    shift # past argument
    shift # past argument
    ;;
  --std-user)
    STD_USER="true"
    shift # past argument
    ;;
  --pre-release)
    PRE_RELEASE="true"
    shift # past argument
    ;;
  *)
    echo "Invalid option $1" >&2
    exit 1
    ;;
  esac
done

if [ -z "$DESTINATION" ]; then
  DESTINATION="/usr/local/bin"
fi

function install() {
  if [ -z "$VERSION" ]; then
    # Get latest version from github
    if [ "$PRE_RELEASE" = "true" ]; then
      # $(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases | jq '[.[] | select(.prerelease == true)] | sort_by(.created_at) | .[0]' | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"')
      VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"' | head -n 1)
    else
      VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"')
    fi
  else
    if [[ ! $VERSION == v* ]]; then
      VERSION="v$VERSION"
    fi

    echo "Checking if version $VERSION exists in GitHub releases"
    TARGET_VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases | grep -o "\"tag_name\": \"$VERSION\"")
    if [ -z "$TARGET_VERSION" ]; then
      echo "Error: Version $VERSION not found in GitHub releases"
      exit 1
    fi
  fi

  if [[ $VERSION == release-v* ]]; then
    VERSION="v${VERSION#release-v}"
  fi

  if [[ ! $VERSION == v* ]]; then
    VERSION="v$VERSION"
  fi
  SHORT_VERSION=$VERSION

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
  if ! curl -sL "$DOWNLOAD_URL" -o prldevops.tar.gz; then
    echo "Failed to download prldevops release from GitHub Releases"
    exit 1
  fi

  echo "Extracting prldevops"
  if ! tar -xzf prldevops.tar.gz; then
    echo "Failed to extract prldevops"
    exit 1
  fi

  if [ ! -d "$DESTINATION" ]; then
    echo "Creating destination directory: $DESTINATION"
    mkdir -p "$DESTINATION"
  fi

  if [ -f "$DESTINATION/prldevops" ]; then
    echo "Removing existing prldevops"
    sudo rm "$DESTINATION/prldevops"
  fi
  echo "Moving prldevops to $DESTINATION"
  sudo mv prldevops "$DESTINATION"/prldevops
  sudo chmod +x "$DESTINATION"/prldevops

  if [ "$INSTALL_SERVICE" = "true" ]; then
    if [ "$OS" = "darwin" ]; then
      echo "Installing prldevops service"
      sudo "$DESTINATION"/prldevops install service
      if [ -f "/Library/LaunchDaemons/com.parallels.prl-devops-service.plist" ]; then
        echo "Restarting prl-devops-service"
        sudo launchctl unload /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
        sudo launchctl load /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
      fi

      sudo xattr -d com.apple.quarantine "$DESTINATION"/prldevops
    fi
  fi

  echo "Cleaning up"
  rm prldevops.tar.gz
  echo "prldevops $SHORT_VERSION has been installed to $DESTINATION"
}

function install_standard() {
  if [ -z "$VERSION" ]; then
    # Get latest version from github
    if [ "$PRE_RELEASE" = "true" ]; then
      VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"' | head -n 1)
    else
      VERSION=$(curl -s https://api.github.com/repos/Parallels/prl-devops-service/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"')
    fi
  fi

  if [[ ! $VERSION == *-beta ]]; then
    if [[ ! $VERSION == release-v* ]]; then
      VERSION="release-v$VERSION"
    fi
    SHORT_VERSION="$(echo $VERSION | cut -d '-' -f 2)"
  else
    if [[ ! $VERSION == v* ]]; then
      VERSION="v$VERSION"
    fi
    SHORT_VERSION=$VERSION
  fi

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
  curl -sL "$DOWNLOAD_URL" -o prldevops.tar.gz

  echo "Extracting prldevops"
  tar -xzf prldevops.tar.gz

  if [ ! -d "$DESTINATION" ]; then
    echo "Creating destination directory: $DESTINATION"
    mkdir -p "$DESTINATION"
  fi

  if [ -f "$DESTINATION/prldevops" ]; then
    echo "Removing existing prldevops"
    rm "$DESTINATION/prldevops"
  fi
  echo "Moving prldevops to $DESTINATION"
  mv prldevops "$DESTINATION"/prldevops
  chmod +x "$DESTINATION"/prldevops

  if [ "$INSTALL_SERVICE" = "true" ]; then
    if [ "$OS" = "darwin" ]; then
      echo "Installing prldevops service"

      "$DESTINATION"/prldevops install service
      if [ -f "/Library/LaunchDaemons/com.parallels.prl-devops-service.plist" ]; then
        echo "Restarting prl-devops-service"
        launchctl unload /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
        launchctl load /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
      fi

      xattr -d com.apple.quarantine "$DESTINATION"/prldevops
    fi
  fi

  echo "Cleaning up"
  rm prldevops.tar.gz
  echo "prldevops $SHORT_VERSION has been installed to $DESTINATION"
}

function uninstall() {
  OS=$(uname -s)
  OS=$(echo "$OS" | tr '[:upper:]' '[:lower:]')

  if [ -f "$DESTINATION/prldevops" ]; then
    if [ "$OS" = "darwin" ]; then
      if [ -f "/Library/LaunchDaemons/com.parallels.prl-devops-service.plist" ]; then
        echo "Uninstalling prldevops service"
        echo "Stopping prl-devops-service"
        sudo launchctl unload /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
        sudo rm /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
      fi
    fi

    echo "Removing prldevops from $DESTINATION"
    sudo rm "$DESTINATION/prldevops"
    if [ -f "$DESTINATION/config.yml" ]; then
      echo "Removing configuration file from $DESTINATION"
      sudo rm "$DESTINATION/config.yml"
    fi
    if [ -f "$DESTINATION/config.yaml" ]; then
      echo "Removing configuration file from $DESTINATION"
      sudo rm "$DESTINATION/config.yaml"
    fi
    if [ -f "$DESTINATION/config.json" ]; then
      echo "Removing configuration file from $DESTINATION"
      sudo rm "$DESTINATION/config.json"
    fi
    echo "Removing current database"
    if [ -f "/etc/prl-devops-service" ]; then
      sudo rm -rf "/etc/prl-devops-service"
    fi

    echo "prldevops has been uninstalled"
  else
    echo "prldevops is not installed in $DESTINATION"
  fi
}

function uninstall_standard() {
  OS=$(uname -s)
  OS=$(echo "$OS" | tr '[:upper:]' '[:lower:]')

  if [ -f "$DESTINATION/prldevops" ]; then
    if [ "$OS" = "darwin" ]; then
      if [ -f "/Library/LaunchDaemons/com.parallels.prl-devops-service.plist" ]; then
        echo "Uninstalling prldevops service"
        echo "Stopping prl-devops-service"
        launchctl unload /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
        rm /Library/LaunchDaemons/com.parallels.prl-devops-service.plist
      fi
    fi

    echo "Removing prldevops from $DESTINATION"
    rm "$DESTINATION/prldevops"
    sudo rm "$DESTINATION/prldevops"
    if [ -f "$DESTINATION/config.yml" ]; then
      echo "Removing configuration file from $DESTINATION"
      rm "$DESTINATION/config.yml"
    fi
    if [ -f "$DESTINATION/config.yaml" ]; then
      echo "Removing configuration file from $DESTINATION"
      rm "$DESTINATION/config.yaml"
    fi
    if [ -f "$DESTINATION/config.json" ]; then
      echo "Removing configuration file from $DESTINATION"
      rm "$DESTINATION/config.json"
    fi
    echo "prldevops has been uninstalled"
  else
    echo "prldevops is not installed in $DESTINATION"
  fi
}

if [ "$MODE" = "UNINSTALL" ]; then
  if [ "$STD_USER" = "true" ]; then
    uninstall_standard
  else
    uninstall
  fi
else
  if [ "$STD_USER" = "true" ]; then
    install_standard
  else
    install
  fi
fi
