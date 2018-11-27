#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/osm/hack/gendocs
go run main.go
popd
