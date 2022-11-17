#!/bin/bash

COUNTER=0

USER='test'
PROXYSQL_NAME='proxy-server'
NAMESPACE='demo'
PASS='pass'

VAR="x"

while [  $COUNTER -lt 100 ]; do
    let COUNTER=COUNTER+1
    VAR=a$VAR
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e 'select 1;' > /dev/null 2>&1
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "INSERT INTO test.testtb(name) VALUES ('$VAR');" > /dev/null 2>&1
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "select * from test.testtb;" > /dev/null 2>&1
    sleep 0.0001
done