cartridge build

(sleep 30 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard)&

cartridge start --cfg ./instances.yml
