#!/bin/bash -e

rm -rf tmp/*
rm -rf .rocks

cartridge build

(sleep 60 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard)&

cartridge start --cfg ./instances.yml
