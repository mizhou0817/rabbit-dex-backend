ARG GO_VERSION=1.19.4
ARG TARGETOS TARGETARCH
FROM --platform=$BUILDPLATFORM  golang:${GO_VERSION} as builder

WORKDIR /go/src/app
RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY go.mod go.mod
COPY go.sum go.sum

RUN --mount=type=cache,target=/gomod-cache go mod download

COPY . .

RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache GOOS=$TARGETOS GOARCH=$TARGETARCH go build -tags=go_tarantool_msgpack_v5 -o /go/bin/go-dashboards cmd/dashboards/main.go

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

WORKDIR /bin

COPY kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx

RUN mkdir /root/.rabbit
COPY kubernetes/_configs/dashboards.yaml /root/.rabbit/dashboards.yaml

COPY --from=builder /go/bin /bin

CMD ["/bin/go-dashboards"]
