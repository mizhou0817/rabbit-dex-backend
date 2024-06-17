FROM tarantool/tarantool:2.10.8-ubuntu20.04

RUN  apt-get update && apt-get install -y cartridge-cli unzip curl git gcc cmake g++ 

COPY model/tnt /usr/src/app
COPY kubernetes/tarantool/start.sh /usr/src/app
COPY kubernetes/.buildinfo-rabbitx /.buildinfo-rabbitx

WORKDIR /usr/src/app

RUN rm -rf tmp/* && rm -rf .rocks

RUN cartridge build