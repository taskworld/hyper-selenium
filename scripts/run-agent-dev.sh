#!/bin/bash -e

echo ">>> Building binary..."
export CGO_ENABLED=0
export GOOS=linux
go build -v -installsuffix cgo -o hyper-selenium-agent ./agent

echo ">>> Running container..."
docker run -ti --rm -v "$(pwd)/hyper-selenium-agent:/hyper-selenium/hyper-selenium-agent" hyper-selenium-agent