#!/bin/bash -e

if [ -z "$(docker images -q hyper-selenium-env)" ]
then
    echo ">>> Building Selenium docker image..."
    docker build --target env -t hyper-selenium-env .
else
    echo ">>> hyper-selenium-env image already exists, so we are going to use it."
    echo "    If you want to rebuild the image, run:"
    echo "        docker rmi hyper-selenium-env"
    echo "    and re-run the script."
fi

echo ">>> Building binary..."
export CGO_ENABLED=0
export GOOS=linux
mkdir -p build
go build -v -installsuffix cgo -o ./build/hyper-selenium-agent ./cmd/hyper-selenium-agent

echo ">>> Running container..."
docker run -ti --rm -v "$(pwd)/build/hyper-selenium-agent:/hyper-selenium/hyper-selenium-agent" hyper-selenium-env ./hyper-selenium-agent "$@"