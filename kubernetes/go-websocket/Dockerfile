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

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH  go build -tags=go_tarantool_msgpack_v5 -o /usr/bin/go-websocket cmd/websocket/main.go

# production
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

COPY --from=builder /usr/bin/go-websocket /usr/bin/
COPY --from=builder /usr/src/app/kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx
COPY --from=builder /usr/src/app/kubernetes/_configs/websocket.yaml /_configs/
COPY --from=builder /usr/src/app/kubernetes/_configs/broker.yaml /_configs/

RUN ln -ns /_configs/ $HOME/.rabbit
