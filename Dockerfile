# Adapted from https://www.docker.com/blog/tag/go-env-series/

FROM node:19.6.0-alpine3.17 AS js-lint
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm clean-install
COPY assets/*.js .eslintrc.json ./
RUN npx eslint ./

FROM golang:1.19.5-bullseye AS go-base
# Disable CGO to produce statically linked executables.
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.* *.go ./

FROM go-base AS go-lint
COPY --from=golangci/golangci-lint:v1.51.0-alpine \
/usr/bin/golangci-lint /usr/bin/golangci-lint
# The --mount arguments to RUN cause the Go module cache, Go build cache, and
# golangci-lint cache to be persisted and reused across Docker builds.
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
--mount=type=cache,target=/root/.cache/golangci-lint \
golangci-lint run -E gofmt,revive --exclude-use-default=false

FROM go-base AS go-test
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
go test

FROM go-base AS go-build
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
go build -o /out/server .

FROM scratch AS bin
COPY --from=go-build /out/server /
COPY assets/ /assets/
ENTRYPOINT ["/server"]
