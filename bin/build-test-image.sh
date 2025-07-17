#!/bin/sh

mkdir -p ./bin/dist

service_name="user_service"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "./bin/dist/$service_name" ./cmd
wait

echo "Building Docker image for $CI_REGISTRY_IMAGE:test" >/dev/stderr
docker build -t "$CI_REGISTRY_IMAGE:test" --build-arg BIN_NAME=$service_name -f Dockerfile .

rm -rf ./bin/dist