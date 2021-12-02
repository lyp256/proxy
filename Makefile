VER=develop
COMMIT= $(shell git rev-parse --short HEAD)
GOMODULE= $(shell go list -m)
VERPKG=$(GOMODULE)/version
IMAGE=docker.io/lyp256/proxy:latest

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: all-check
all-check:

.PHONY: git-check
git-check:
	git diff --exit-code

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags "-X $(VERPKG).Version=$(VER) -X $(VERPKG).CommitID=$(COMMIT)" -o build/ ./cmd/...

.PHONY: build
docker:
	docker build --tag $(IMAGE) .

.PHONY: docker-push
docker-push:
	docker push $(IMAGE)

.PHONY: docker-release
docker-release: docker-build docker-push
