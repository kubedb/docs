#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/operator/hack/gendocs
go run main.go
popd
