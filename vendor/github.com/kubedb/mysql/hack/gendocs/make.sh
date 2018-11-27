#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/mysql/hack/gendocs
go run main.go
popd
