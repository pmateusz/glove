# Copyright 2023 The Glove Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

ifndef VCS_COMMIT
	VCS_COMMIT := $(shell git rev-parse HEAD)
endif

ifndef VCS_BRANCH
	VCS_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif

ifndef VCS_TAG
	VCS_TAG := $(shell git describe --tags --abbrev 2>/dev/null || echo -n "v0.0.0")
endif

ifndef BUILD_TIME
	BUILD_TIME := $(shell date --utc --iso-8601=seconds)
endif

ifndef BUILD_ENVIRONMENT
	BUILD_ENVIRONMENT := local
endif
VERSION_PACKAGE = "github.com/pmateusz/glove/internal/version"
LDFLAGS = "-X ${VERSION_PACKAGE}.commitHash=${VCS_COMMIT} \
-X ${VERSION_PACKAGE}.branch=${VCS_BRANCH} \
-X ${VERSION_PACKAGE}.version=${VCS_TAG} \
-X ${VERSION_PACKAGE}.buildTime=${BUILD_TIME} \
-X ${VERSION_PACKAGE}.environment=${BUILD_ENVIRONMENT}"

.PHONY: build clean cli deps-install deps-update format image lint test test-cover update

cli: deps-install build

format:
	gofmt -l -s -w .

lint:
	go vet ./...
	staticcheck ./...

test:
	go test ./...

test-cover:
	go test ./... -coverpkg=glove/internal/...,glove/pkg/... -coverprofile=coverage.out

deps-install:
	go get ./...

deps-update:
	go get -u ./...

build:
	echo $(LDFLAGS)
	go generate && go build -ldflags=$(LDFLAGS) -o ./bin/glove ./cmd/glove/main.go

clean:
	rm ./bin/glove

image:
	docker buildx build --platform linux/amd64 \
	--build-arg="VCS_COMMIT=${VCS_COMMIT}" \
	--build-arg="VCS_BRANCH=${VCS_BRANCH}" \
	--build-arg="VCS_TAG=${VCS_TAG}" \
	--build-arg="BUILD_TIME=${BUILD_TIME}" \
	--build-arg="BUILD_ENVIRONMENT=${BUILD_ENVIRONMENT}" \
	-t glove -f ./build/Dockerfile . \
	--load
