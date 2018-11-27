#!/bin/bash
set -x
CONFIG_DIR="/elasticsearch/config"
CUSTOM_CONFIG_DIR="/elasticsearch/custom-config"

CONFIG_FILE=$CONFIG_DIR/elasticsearch.yml

# yq changes the file permissions after merging custom configuration.
# we need to restore the original permissions after merging done.
ORIGINAL_PERMISSION=$(stat -c '%a' $CONFIG_FILE)

# if common-config file exist then apply it
if [ -f $CUSTOM_CONFIG_DIR/common-config.yml ]; then
  yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/common-config.yml
elif [ -f $CUSTOM_CONFIG_DIR/common-config.yaml ]; then
  yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/common-config.yaml
fi

# if it is data node and data-config file exist then apply it
if [[ "$NODE_DATA" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/data-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/data-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/data-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/data-config.yaml
  fi
fi

# if it is client node and client-config file exist then apply it
if [[ "$NODE_INGEST" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/client-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/client-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/client-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/client-config.yaml
  fi
fi

# if it is master node and mater-config file exist then apply it
if [[ "$NODE_MASTER" == true ]]; then
  if [ -f $CUSTOM_CONFIG_DIR/master-config.yml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/master-config.yml
  elif [ -f $CUSTOM_CONFIG_DIR/master-config.yaml ]; then
    yq merge -i --overwrite $CONFIG_FILE $CUSTOM_CONFIG_DIR/master-config.yaml
  fi
fi

# restore original permission of elasticsearh.yml file
chmod $ORIGINAL_PERMISSION $CONFIG_FILE
