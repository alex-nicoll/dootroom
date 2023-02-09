# Adapted from https://www.docker.com/blog/tag/go-env-series/

FROM node:19.6.0-alpine3.17 AS eslint
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm clean-install
COPY *.js .eslintrc.json ./
RUN npx eslint ./

FROM golangci/golangci-lint:v1.51.0-alpine AS linter

FROM golang:1.19.5-bullseye AS build
COPY --from=linter /usr/bin/golangci-lint /usr/bin/golangci-lint
WORKDIR /src
COPY go.* *.go ./
# The --mount arguments to RUN cause the Go module cache, Go build cache, and
# golangci-lint cache to be persisted and reused across Docker builds.
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
--mount=type=cache,target=/root/.cache/golangci-lint \
golangci-lint run -E gofmt,revive --exclude-use-default=false
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
go test
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
go build -o /out/server .

FROM scratch AS bin
COPY --from=build /out/server /
