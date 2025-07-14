#!/bin/sh

mkdir -p ./bin/dist

service_name="user_service"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "./bin/dist/$service_name" ./cmd
wait

VERSION=$(cat VERSION)
echo "Building Docker image for $service_name:$VERSION" >/dev/stderr
docker build -t "$service_name:$VERSION" --build-arg BIN_NAME=$service_name -f Dockerfile .

registry_image="$CI_REGISTRY_IMAGE:$VERSION"
echo "$CI_JOB_TOKEN" | docker login -u gitlab-ci-token --password-stdin "$CI_REGISTRY"

echo "Tagging image as $registry_image" >/dev/stderr
docker tag "$service_name:$VERSION" "$registry_image"
docker push "$registry_image"

echo "Tagging image as $registry_image" >/dev/stderr
docker tag "$service_name:$VERSION" "$CI_REGISTRY_IMAGE:latest"
docker push "$CI_REGISTRY_IMAGE:latest"

echo "Removing local image and tagged image" >/dev/stderr
docker rmi "$service_name:$VERSION"
docker rmi "$registry_image"
docker rmi "$CI_REGISTRY_IMAGE:latest"
