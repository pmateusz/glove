# Copyright 2023 The Glove Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

# syntax=docker/dockerfile:1
ARG GO_VERSION=1.21

FROM golang:${GO_VERSION} as build-stage
ARG TARGETOS
ARG TARGETPLATFORM
ARG TARGETARCH
ARG BUILDPLATFORM
ARG VERSION_PACKAGE="github.com/pmateusz/glove/internal/version"
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
RUN go generate && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
 -ldflags="-X ${VERSION_PACKAGE}.commitHash=${VCS_COMMIT} \
 -X ${VERSION_PACKAGE}.branch=${VCS_BRANCH} \
 -X ${VERSION_PACKAGE}.version=${VCS_TAG} \
 -X ${VERSION_PACKAGE}.buildTime=${BUILD_TIME} \
 -X ${VERSION_PACKAGE}.environment=${BUILD_ENVIRONMENT}" \
 -o glove ./cmd/glove/main.go

FROM gcr.io/distroless/base AS release-stage
WORKDIR /
COPY --from=build-stage /build/glove /glove
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/glove"]
