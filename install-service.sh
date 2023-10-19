#!/bin/bash

install() {
 echo "Installing Service on $1"
 echo "<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>UserName</key>
  <string>root</string>
  <key>Label</key>
  <string>com.parallels.api-service</string>
  <key>Program</key>
  <string>$1/pd-api-service</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardErrorPath</key>
  <string>/tmp/api-service.job.err</string>
  <key>StandardOutPath</key>
  <string>/tmp/api-service.job.out</string> 
</dict>
</plist>" > /Library/LaunchDaemons/com.parallels.api-service.plist
  
  chown root:wheel /Library/LaunchDaemons/com.parallels.api-service.plist
  chmod 644 /Library/LaunchDaemons/com.parallels.api-service.plist

  launchctl unload /Library/LaunchDaemons/com.parallels.api-service.plist
  launchctl load /Library/LaunchDaemons/com.parallels.api-service.plist
  launchctl start /Library/LaunchDaemons/com.parallels.api-service
  echo "Done"
}

install $1