#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/memcached/hack/gendocs
go run main.go
popd
