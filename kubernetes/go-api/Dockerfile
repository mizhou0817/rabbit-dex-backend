ARG GO_VERSION=1.21.5
ARG TARGETOS TARGETARCH
FROM --platform=$BUILDPLATFORM  golang:${GO_VERSION} as builder

RUN mkdir /root/.rabbit

WORKDIR /usr/src/app
RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache go mod download && go mod verify

COPY . .
COPY kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags=go_tarantool_msgpack_v5 -o /usr/bin/go-api cmd/api/main.go

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

COPY kubernetes/_configs/rest.yaml /root/.rabbit/rest.yaml
COPY kubernetes/_configs/broker.yaml /root/.rabbit/broker.yaml
COPY kubernetes/_configs/airdrop.yaml /root/.rabbit/airdrop.yaml

COPY --from=builder /usr/bin/go-api /usr/bin/go-api
