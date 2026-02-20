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


function get_latest_release() {
  if [ "$PRE_RELEASE" = "true" ]; then
    URL="https://api.github.com/repos/Parallels/prl-devops-service/releases"
  else
    URL="https://api.github.com/repos/Parallels/prl-devops-service/releases/latest"
  fi

  HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" "$URL")
  HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS\:.*//g')
  HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

  if [ "$HTTP_STATUS" = "403" ] || [ "$HTTP_STATUS" = "429" ]; then
    echo "Error: GitHub API rate limit exceeded. Please try again later."
    exit 1
  fi

  if [ "$HTTP_STATUS" != "200" ]; then
    echo "Error: Failed to fetch release information. HTTP Status: $HTTP_STATUS"
    exit 1
  fi

  if [ "$PRE_RELEASE" = "true" ]; then
    echo "$HTTP_BODY" | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"' | head -n 1
  else
    echo "$HTTP_BODY" | grep -o '"tag_name": "[^"]*"' | cut -d ' ' -f 2 | tr -d '"'
  fi
}

function install() {
  if [ -z "$VERSION" ]; then
    echo "Getting latest version from GitHub"
    # Get latest version from github
    VERSION=$(get_latest_release)
  fi

  if [[ ! $VERSION == *-beta ]]; then
    if [[ ! $VERSION == v* ]]; then
      VERSION="v$VERSION"
    fi
    is_found=$(curl -s -o /dev/null -w "%{http_code}" https://github.com/Parallels/prl-devops-service/releases/$VERSION)
    if [ "$is_found" != "200" ]; then
      echo "Version not found with new format, attempting old format"
      VERSION="release-$VERSION"
      is_found=$(curl -s -o /dev/null -w "%{http_code}" https://github.com/Parallels/prl-devops-service/releases/$VERSION)
      if [ "$is_found" != "200" ]; then
        echo "Version $VERSION not found in GitHub releases"
        exit 1
      fi
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
  # Download the file and capture HTTP status code
  HTTP_STATUS=$(curl -sL -w "%{http_code}" "$DOWNLOAD_URL" -o prldevops.tar.gz)

  if [ "$HTTP_STATUS" = "403" ] || [ "$HTTP_STATUS" = "429" ]; then
    echo "Error: GitHub API rate limit exceeded during download. Please try again later."
    exit 1
  fi

  # Check if the status code is 200 (OK)
  if [ "$HTTP_STATUS" != "200" ]; then
    echo "Failed to download prldevops release from GitHub Releases (HTTP status: $HTTP_STATUS)"
    exit 1
  fi

  # Check if the file exists and has content
  if [ ! -s "prldevops.tar.gz" ]; then
    echo "Downloaded file is empty or does not exist"
    exit 1
  fi

  echo "Extracting prldevops"
  if ! tar -xzf prldevops.tar.gz; then
    echo "Failed to extract prldevops"
    exit 1
  fi

  if [ $? -ne 0 ]; then
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
      SERVICE_PLIST="/Library/LaunchDaemons/com.parallels.prl-devops-service.plist"
      CONFIG_EXISTS="false"
      
      if [ -f "$DESTINATION/config.yaml" ] || [ -f "$DESTINATION/config.yml" ]; then
        CONFIG_EXISTS="true"
      fi

      if [ -f "$SERVICE_PLIST" ] && [ "$CONFIG_EXISTS" = "true" ]; then
         echo "Service already installed and configured. Updating binary and restarting service."
         sudo launchctl unload "$SERVICE_PLIST"
         sudo launchctl load "$SERVICE_PLIST"
      else
          echo "Installing prldevops service"
          sudo "$DESTINATION"/prldevops install service
          if [ -f "$SERVICE_PLIST" ]; then
            echo "Restarting prl-devops-service"
            sudo launchctl unload "$SERVICE_PLIST"
            sudo launchctl load "$SERVICE_PLIST"
          fi
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
    echo "Getting latest version from GitHub"
    # Get latest version from github
    VERSION=$(get_latest_release)
  fi

  if [[ ! $VERSION == *-beta ]]; then
    if [[ ! $VERSION == v* ]]; then
      VERSION="v$VERSION"
    fi
    is_found=$(curl -s -o /dev/null -w "%{http_code}" https://github.com/Parallels/prl-devops-service/releases/$VERSION)
    if [ "$is_found" != "200" ]; then
      echo "Version not found with new format, attempting old format"
      VERSION="release-$VERSION"
      is_found=$(curl -s -o /dev/null -w "%{http_code}" https://github.com/Parallels/prl-devops-service/releases/$VERSION)
      if [ "$is_found" != "200" ]; then
        echo "Version $VERSION not found in GitHub releases"
        exit 1
      fi
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
  # Download the file and capture HTTP status code
  HTTP_STATUS=$(curl -sL -w "%{http_code}" "$DOWNLOAD_URL" -o prldevops.tar.gz)

  if [ "$HTTP_STATUS" = "403" ] || [ "$HTTP_STATUS" = "429" ]; then
    echo "Error: GitHub API rate limit exceeded during download. Please try again later."
    exit 1
  fi

  # Check if the status code is 200 (OK)
  if [ "$HTTP_STATUS" != "200" ]; then
    echo "Failed to download prldevops release from GitHub Releases (HTTP status: $HTTP_STATUS)"
    exit 1
  fi

  # Check if the file exists and has content
  if [ ! -s "prldevops.tar.gz" ]; then
    echo "Downloaded file is empty or does not exist"
    exit 1
  fi

  echo "Extracting prldevops"
  tar -xzf prldevops.tar.gz

  if [ $? -ne 0 ]; then
    echo "Failed to extract prldevops"
    exit 1
  fi

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
      SERVICE_PLIST="/Library/LaunchDaemons/com.parallels.prl-devops-service.plist"
      CONFIG_EXISTS="false"
      
      if [ -f "$DESTINATION/config.yaml" ] || [ -f "$DESTINATION/config.yml" ]; then
        CONFIG_EXISTS="true"
      fi

      if [ -f "$SERVICE_PLIST" ] && [ "$CONFIG_EXISTS" = "true" ]; then
         echo "Service already installed and configured. Updating binary and restarting service."
         launchctl unload "$SERVICE_PLIST"
         launchctl load "$SERVICE_PLIST"
      else
          echo "Installing prldevops service"
          "$DESTINATION"/prldevops install service
          if [ -f "$SERVICE_PLIST" ]; then
            echo "Restarting prl-devops-service"
            launchctl unload "$SERVICE_PLIST"
            launchctl load "$SERVICE_PLIST"
          fi
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
