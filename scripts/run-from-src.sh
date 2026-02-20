#!/bin/bash

DESTINATION="/usr/local/bin"

while [[ $# -gt 0 ]]; do
  case $1 in
  -p|--path)
    DESTINATION=$2
    shift # past argument
    shift # past argument
    ;;
  *)
    shift # past argument
    ;;
  esac
done

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
SERVICE_PLIST="/Library/LaunchDaemons/com.parallels.prl-devops-service.plist"
SYSTEMD_SERVICE="/etc/systemd/system/prl-devops-service.service"
LINUX_SERVICE_NAME="prl-devops-service"
BINARY_NAME="prldevops"

# Move to the root of the project to ensure make build runs correctly
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT" || { echo "Failed to change to project root directory"; exit 1; }

# Stop the service down based on OS
if [ "$OS" = "darwin" ] && [ -f "$SERVICE_PLIST" ]; then
  echo "Stopping prl-devops-service (macOS)..."
  sudo launchctl unload "$SERVICE_PLIST"
elif [ "$OS" = "linux" ] && [ -f "$SYSTEMD_SERVICE" ]; then
  echo "Stopping prl-devops-service (Linux)..."
  sudo systemctl stop "$LINUX_SERVICE_NAME"
fi

echo "Building prldevops from source..."
make build

if [ $? -ne 0 ]; then
  echo "Build failed. Exiting."
  exit 1
fi

if [ ! -f "out/binaries/$BINARY_NAME" ]; then
  echo "Binary out/binaries/$BINARY_NAME not found. Exiting."
  exit 1
fi

if [ ! -d "$DESTINATION" ]; then
  echo "Creating destination directory: $DESTINATION"
  sudo mkdir -p "$DESTINATION"
fi

if [ -f "$DESTINATION/$BINARY_NAME" ]; then
  echo "Removing existing $BINARY_NAME from $DESTINATION"
  sudo rm "$DESTINATION/$BINARY_NAME"
fi

echo "Copying $BINARY_NAME to $DESTINATION..."
sudo cp "out/binaries/$BINARY_NAME" "$DESTINATION/$BINARY_NAME"
sudo chmod +x "$DESTINATION/$BINARY_NAME"

# Start the service back up and apply explicit permissions based on OS
if [ "$OS" = "darwin" ]; then
  sudo xattr -d com.apple.quarantine "$DESTINATION/$BINARY_NAME" 2>/dev/null || true
  
  if [ -f "$SERVICE_PLIST" ]; then
    echo "Starting prl-devops-service (macOS)..."
    sudo launchctl load "$SERVICE_PLIST"
  else
    echo "Service plist not found at $SERVICE_PLIST, skipping service start."
  fi
elif [ "$OS" = "linux" ]; then
  if [ -f "$SYSTEMD_SERVICE" ]; then
    echo "Starting prl-devops-service (Linux)..."
    sudo systemctl daemon-reload
    sudo systemctl start "$LINUX_SERVICE_NAME"
  else
    echo "Systemd service not found at $SYSTEMD_SERVICE, skipping service start."
  fi
fi

echo "$BINARY_NAME has been built from source, installed to $DESTINATION, and restarted."
