#!/bin/bash

searchguard="/elasticsearch/plugins/search-guard-6"
certs="/elasticsearch/config/certs"

sync

SERVER='http://localhost:9200'

if [ "$SSL_ENABLE" == true ]; then
  SERVER='https://localhost:9200'
fi

until curl -s "$SERVER" --insecure; do
  sleep 0.1
done

"$searchguard"/tools/sgadmin.sh \
  -ks "$certs"/sgadmin.jks \
  -kspass "$KEY_PASS" \
  -ts "$certs"/root.jks \
  -tspass "$KEY_PASS" \
  -cd "$searchguard"/sgconfig -icl -nhnv
