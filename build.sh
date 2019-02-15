#!/usr/bin/env bash
# clean pkg
go clean

# get dependencies
dep ensure

# build with linux OS
CGO_ENABLED=0 GOOS=linux go build

# build docker image
docker build --no-cache -t mchandramouli/haystack-kube-sidecar-injector:1.0 .

# remove binary
rm haystack-kube-sidecar-injector
