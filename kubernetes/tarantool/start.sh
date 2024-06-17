#!/bin/bash

#set -ux

(sleep 60 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard)&

rm -rf tmp/run/*

COMMAND="cartridge start --cfg ./instances.yml"

if [[ "$RECOVERY" == "true" ]]; then
    COMMAND="TARANTOOL_FORCE_RECOVERY=true $COMMAND"
fi

eval "$COMMAND"


