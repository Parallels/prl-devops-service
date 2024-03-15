#!/bin/bash

# Initialize variables
version=""
date=""
content=""

# Read the CHANGELOG.md line by line
while IFS= read -r line
do
  # Check if the line contains a version number
  if [[ $line =~ "## ["[0-9]+\.[0-9]+\.[0-9]+"]" ]]; then
    # If a version was previously found, write the content to the corresponding file
    if [ ! -z "$version" ]; then
      lines+="# Whats New\n\n"
        lines+="$content"
      echo -e "$content" > "./docs/_posts/${date}-v${version}.markdown"
    fi

    # Extract the new version number, date and reset the content
    version=$(echo $line | awk -F '[\\[\\]]' '{print $2}')
    date=$(echo $line | awk '{print $NF}')
    content=""
  else
    # Append the line to the content
    content+="$line\n"
  fi
done < "CHANGELOG.md"

# Write the content for the last version
if [ ! -z "$version" ]; then
  lines+="# Whats New\n\n"
  lines+="$content"
  echo -e "$lines" > "./docs/_posts/${date}-v${version}.markdown"
fi