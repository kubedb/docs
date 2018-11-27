#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/redis/hack/gendocs
go run main.go
popd
