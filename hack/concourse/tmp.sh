#!/bin/bash

set -e

EXIT_CODE=0

if ! (echo 1 && true); then
    echo 11
    EXIT_CODE=1
fi

if ! (echo 2 && false); then
    echo 22
    EXIT_CODE=1
fi

if ! (echo 3 && true); then
    echo 33
    EXIT_CODE=1
fi

echo $EXIT_CODE
exit $EXIT_CODE
