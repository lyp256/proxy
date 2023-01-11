VER=develop
COMMIT=$(shell git rev-parse --short HEAD)
GOMODULE=$(shell go list -m)
VERPKG=$(GOMODULE)/version

PKGS=$(shell go list ./... |grep -v vendor |xargs echo)

ifdef REF_NAME
TAG:=$(REF_NAME)
else
TAG:=dev
endif

ifneq ($(REF_TYPE),tag)
TAG:=$(TAG)-$(COMMIT)
endif

IMAGE=docker.io/lyp256/proxy:$(TAG)

print:
	echo $(IMAGE)


.PHONY: fmt
fmt:
	go fmt $(PKGS)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: all-check
all-check:tidy fmt vet git-check

.PHONY: git-check
git-check:
	git diff --exit-code

.PHONY: test
test:
	go test $(PKGS)

.PHONY: build
build:
	CGO_ENABLED=0 go build -tags include_oss,include_gcs -ldflags "-X $(VERPKG).Version=$(VER) -X $(VERPKG).CommitID=$(COMMIT)" -o build/ ./cmd/...

.PHONY: docker-build
docker-build:
	docker build --tag $(IMAGE) .

.PHONY: docker-push
docker-push:
	docker images
	docker push $(IMAGE)

.PHONY: docker-release
docker-release: docker-build docker-push

oci-release:
	podman build --tag dst-image .
	podman push dst-image docker://$(IMAGE)
