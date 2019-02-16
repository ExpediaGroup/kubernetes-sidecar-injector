lint:
	go list ./... | xargs golint -min_confidence 1.0

vet:
	go vet ./...

all: vet lint
