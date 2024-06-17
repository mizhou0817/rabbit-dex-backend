rm -rf tmp/*

cartridge build

(sleep 5 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard)&

cartridge start --cfg ./instances.yml
