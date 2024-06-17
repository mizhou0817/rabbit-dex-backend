ARG GO_VERSION=1.19.4
ARG TARGETOS TARGETARCH
FROM --platform=$BUILDPLATFORM  golang:1.19.4 as builder

WORKDIR /usr/src/app
RUN mkdir /root/.rabbit

RUN git clone -b tarantool_use_json --depth 1 https://github.com/centrifugal/centrifugo.git /usr/src/app

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx


RUN --mount=type=cache,target=/gomod-cache go mod download && go mod verify
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /usr/bin/centrifugo

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

COPY kubernetes/_configs/centrifugo.json /root/.rabbit/config.json

COPY --from=builder /usr/bin/centrifugo /usr/bin/centrifugo
