#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/postgres/hack/gendocs
go run main.go
popd
