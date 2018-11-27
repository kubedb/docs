#!/usr/bin/env bash

pushd $GOPATH/src/github.com/kubedb/elasticsearch/hack/gendocs
go run main.go
popd
