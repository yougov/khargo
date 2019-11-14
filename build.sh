#!/bin/sh

VERSION=$(git describe --always --long --dirty)

GOOS="linux" GOARCH="amd64" go build -ldflags "-s -w -X main.version=${VERSION}"
docker build -t registry.bizmate.it/view/view-gridfs:linux .


