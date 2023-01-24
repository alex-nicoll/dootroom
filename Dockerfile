FROM alexnicoll/go-build-env AS build

WORKDIR /src

COPY go.mod go.sum *.go build ./

# The --mount arguments to RUN cause the Go module cache, Go build cache, and
# staticcheck cache to be persisted and reused across Docker builds.
RUN \
--mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
--mount=type=cache,target=/root/.cache/staticcheck \
./build /out/server

FROM scratch AS bin
COPY --from=build /out/server /
