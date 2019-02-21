#!/usr/bin/env bash
set -eo pipefail
set -o errexit

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

make release