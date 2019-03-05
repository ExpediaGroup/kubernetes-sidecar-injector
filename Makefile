SHELL := /bin/bash
CONTAINER_NAME=expediadotcom/kubernetes-sidecar-injector

SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

lint:
	go list ./... | xargs golint -min_confidence 1.0 

vet:
	go vet ./...

imports:
	goimports -w ${SRC}

clean:
	go clean

ensure:
	dep ensure

build: ensure clean vet lint
	go build -o kubernetes-sidecar-injector

release: ensure clean vet lint
	CGO_ENABLED=0 GOOS=linux go build -o kubernetes-sidecar-injector
	docker build --no-cache -t ${CONTAINER_NAME} .
	rm kubernetes-sidecar-injector

