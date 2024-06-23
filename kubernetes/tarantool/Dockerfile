FROM --platform=linux/amd64 ubuntu:20.04 as rustdev

RUN apt update && apt install -y curl make gcc protobuf-compiler
RUN curl https://sh.rustup.rs -sSf > /tmp/install && sh /tmp/install -y

WORKDIR /usr/src/app

COPY model/tnt/rbx /usr/src/app/rbx
COPY model/tnt/Makefile /usr/src/app

RUN --mount=type=cache,target=/usr/src/app/rbx/target bash -c "source $HOME/.cargo/env && make lib"


FROM --platform=linux/amd64 tarantool/tarantool:2.10.8-ubuntu20.04 as tntdev

RUN apt-get update && apt-get install -y cartridge-cli unzip curl git gcc cmake g++

COPY model/tnt /usr/src/app

WORKDIR /usr/src/app

RUN rm -rf tmp/* && rm -rf .rocks

RUN cartridge build


FROM --platform=linux/amd64 tarantool/tarantool:2.10.8-ubuntu20.04

WORKDIR /usr/src/app
RUN apt-get update && apt-get install -y cartridge-cli

COPY --from=tntdev /usr/src/app/app /usr/src/app/app
COPY --from=tntdev /usr/src/app/migrations /usr/src/app/migrations
COPY --from=tntdev /usr/src/app/tnt-scm-1.rockspec /usr/src/app
COPY --from=tntdev /usr/src/app/.rocks /usr/src/app/.rocks
COPY --from=tntdev /usr/src/app/*.yml /usr/src/app
COPY --from=tntdev /usr/src/app/*.lua /usr/src/app

COPY --from=rustdev /usr/src/app/librbx.* /usr/src/app
COPY --from=rustdev /usr/src/app/.rocks /usr/src/app/.rocks

COPY kubernetes/tarantool/start.sh /usr/src/app
COPY kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx
