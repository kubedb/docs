#!/bin/bash

set -x
set -o errexit
set -o pipefail

sync

# if custom config file exist then process them
CUSTOM_CONFIG_DIR="/elasticsearch/custom-config"

if [ -d "$CUSTOM_CONFIG_DIR" ]; then

  configs=($(find $CUSTOM_CONFIG_DIR -maxdepth 1 -name "*.yaml"))
  configs+=($(find $CUSTOM_CONFIG_DIR -maxdepth 1 -name "*.yml"))
  if [ ${#configs[@]} -gt 0 ]; then
    config-merger.sh
  fi
fi

echo "Starting runit..."
exec /sbin/runsvdir -P /etc/service
