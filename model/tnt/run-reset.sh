#!/bin/bash -e

export LD_LIBRARY_PATH=$PWD:$LD_LIBRARY_PATH

rm -rf tmp/*
rm -rf .rocks

make lib
cartridge build

(sleep 60 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard)&

cartridge start --cfg ./instances.yml
