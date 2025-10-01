# syntax = docker/dockerfile:1.4

# get modules, if they don't change the cache can be used for faster builds
FROM golang:1.25.1@sha256:8305f5fa8ea63c7b5bc85bd223ccc62941f852318ebfbd22f53bbd0b358c07e1 AS base
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /src
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# build th application
FROM base AS build

# temp mount all files instead of loading into image with COPY
# temp mount module cache
# temp mount go build cache
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" -o /app/main ./cmd/wakatime-profile-stats/*.go

# Import the binary from build stage

FROM alpine as prd
COPY --link --from=build /app/main /

RUN mkdir -p /tmp && chmod 777 /tmp

ENTRYPOINT ["/main"]
