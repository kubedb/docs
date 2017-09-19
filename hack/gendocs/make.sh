#!/usr/bin/env bash

pushd $GOPATH/src/github.com/k8sdb/operator/hack/gendocs
go run main.go
popd
