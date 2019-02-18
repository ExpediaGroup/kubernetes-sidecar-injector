SHELL := /bin/bash
TAG := $(shell date "+alpha-%Y%m%d%H%M%S")
CONTAINER_NAME=mageshcmouli/haystack-kube-sidecar-injector
CONTAINER_VERSION=$(CONTAINER_NAME):$(TAG)
CONTAINER_LATEST=$(CONTAINER_NAME):latest
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

lint:
	go list ./... | xargs golint -min_confidence 1.0 

vet:
	go vet ./...

imports:
	goimports -w ${SRC}

clean:
	go clean

ensure: clean vet lint
	dep ensure

build: ensure
	go build

docker: ensure
	CGO_ENABLED=0 GOOS=linux go build
	docker build --no-cache -t ${CONTAINER_VERSION} -t ${CONTAINER_LATEST} .
	rm haystack-kube-sidecar-injector

release: docker
	docker push ${CONTAINER_VERSION}
	docker push ${CONTAINER_LATEST}

