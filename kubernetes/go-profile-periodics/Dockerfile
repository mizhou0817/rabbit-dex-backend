ARG GO_VERSION=1.21.5
ARG TARGETOS TARGETARCH
FROM --platform=$BUILDPLATFORM  golang:${GO_VERSION} as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache go mod download && go mod verify

COPY . .

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags=go_tarantool_msgpack_v5 -o /usr/bin/go-profile-periodics cmd/profile/periodics/main.go

# production
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

COPY --from=builder /usr/bin/go-profile-periodics /usr/bin/
COPY --from=builder /usr/src/app/kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx
COPY --from=builder /usr/src/app/kubernetes/_configs/profile-periodics.yaml /_configs/
COPY --from=builder /usr/src/app/kubernetes/_configs/broker.yaml /_configs/

RUN ln -ns /_configs/ $HOME/.rabbit
