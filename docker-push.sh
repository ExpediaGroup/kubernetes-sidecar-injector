#!/usr/bin/env bash
CONTAINER_NAME=expediadotcom/kubernetes-sidecar-injector
CONTAINER_LATEST=${CONTAINER_NAME}:latest

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

if [[ "${BRANCH}" == 'master' && "${PULL_REQUEST}" == 'false' ]]; then
  CONTAINER_VERSION=${CONTAINER_NAME}:v1.0-RC-$(date "+%Y%m%d%H%M%S")
  docker tag ${CONTAINER_NAME} ${CONTAINER_LATEST}
  docker push ${CONTAINER_LATEST}
else
  CONTAINER_VERSION=${CONTAINER_NAME}:${BRANCH}-$(date "+%Y%m%d%H%M%S")
fi

docker tag ${CONTAINER_NAME} ${CONTAINER_VERSION}
docker push ${CONTAINER_VERSION}