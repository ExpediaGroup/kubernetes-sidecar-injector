#!/usr/bin/env bash
set -euxo pipefail
set -o errexit
set -o nounset

CONTAINER_NAME=mageshcmouli/haystack-kube-sidecar-injector:alpha

# clean pkg
go clean

# get dependencies
dep ensure

# build with linux OS
CGO_ENABLED=0 GOOS=linux go build

# build docker image
docker build --no-cache -t ${CONTAINER_NAME} .

# remove binary
rm haystack-kube-sidecar-injector

if [[ "$1" == "push" ]]; then
  docker push ${CONTAINER_NAME}
fi
