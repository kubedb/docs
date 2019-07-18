#!/usr/bin/env bash

pushd $GOPATH/src/kubedb.dev/operator/hack/gendocs
go run main.go
popd
