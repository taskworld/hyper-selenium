#!/bin/bash -e

echo ">>> Building binary..."
export CGO_ENABLED=0
export GOOS=linux
mkdir -p build
go build -v -installsuffix cgo -o ./build/hyper-selenium-agent ./agent

echo ">>> Running container..."
docker run -ti --rm -v "$(pwd)/build/hyper-selenium-agent:/hyper-selenium/hyper-selenium-agent" hyper-selenium-agent