# Copyright 2023 The Glove Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

# syntax=docker/dockerfile:1
ARG GO_VERSION=1.21

FROM golang:${GO_VERSION} as build-stage
ARG VCS_COMMIT
ARG VCS_BRANCH
ARG VCS_TAG
ARG BUILD_TIME
ARG BUILD_ENVIRONMENT
WORKDIR /build
COPY cmd /build/cmd
COPY internal /build/internal
COPY pkg /build/pkg
COPY go.mod /build
RUN go mod tidy
RUN go generate && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
 -ldflags="-X glove/internal/version.commitHash=${VCS_COMMIT} \
 -X glove/internal/version.branch=${VCS_BRANCH} \
 -X glove/internal/version.version=${VCS_TAG} \
 -X glove/internal/version.buildTime=${BUILD_TIME} \
 -X glove/internal/version.environment=${BUILD_ENVIRONMENT}" \
 -o glove ./cmd/glove/main.go

FROM gcr.io/distroless/base-debian11 AS release-stage
WORKDIR /
COPY --from=build-stage /build/glove /glove
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/glove"]
