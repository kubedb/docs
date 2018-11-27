#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/etcd/hack/gendocs
go run main.go
popd
