#!/usr/bin/env bash
CONTAINER_NAME=expediadotcom/haystack-kube-sidecar-injector
CONTAINER_VERSION=${CONTAINER_NAME}:$(date "+v1.0-RC-%Y%m%d%H%M%S")
CONTAINER_LATEST=${CONTAINER_NAME}:latest

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

docker tag ${CONTAINER_NAME} ${CONTAINER_VERSION}
docker push ${CONTAINER_VERSION}

docker tag ${CONTAINER_NAME} ${CONTAINER_LATEST}
docker push ${CONTAINER_LATEST}