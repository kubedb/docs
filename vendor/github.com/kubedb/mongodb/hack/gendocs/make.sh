#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/mongodb/hack/gendocs
go run main.go
popd
