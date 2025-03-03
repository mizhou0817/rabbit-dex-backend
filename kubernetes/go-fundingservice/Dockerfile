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

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH  go build -tags=go_tarantool_msgpack_v5 -o /usr/bin/go-fundingservice cmd/fundingservice/main.go

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

COPY kubernetes/_configs/funding.yaml /root/.rabbit/funding.yaml
COPY kubernetes/_configs/broker.yaml /root/.rabbit/broker.yaml

COPY --from=builder /usr/bin/go-fundingservice /usr/bin/go-fundingservice
