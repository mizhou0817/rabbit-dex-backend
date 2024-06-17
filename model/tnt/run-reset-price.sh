#!/bin/zsh -e

rm -rf tmp/*
#rm -rf .rocks

#cartridge build

(
    sleep 20 && cartridge replicasets setup --cfg ./instances.yml --file ./replicasets.yml --bootstrap-vshard
    echo "local d=require('decimal'); market.update_index_price('BTC-USD', d.new(100));" | cartridge enter btc-market
    echo "local d=require('decimal'); market.update_index_price('ETH-USD', d.new(100));" | cartridge enter eth-market
    echo "local d=require('decimal'); market.update_index_price('SOL-USD', d.new(100));" | cartridge enter sol-market
    echo "local d=require('decimal'); market.update_index_price('ARB-USD', d.new(100));" | cartridge enter arb-market
    echo "local d=require('decimal'); market.update_index_price('DOGE-USD', d.new(100));" | cartridge enter doge-market
    echo "local d=require('decimal'); market.update_index_price('LDO-USD', d.new(100));" | cartridge enter ldo-market
    echo "local d=require('decimal'); market.update_index_price('SUI-USD', d.new(100));" | cartridge enter sui-market
    echo "local d=require('decimal'); market.update_index_price('PEPE1000-USD', d.new(100));" | cartridge enter pepe1000-market
    echo "local d=require('decimal'); market.update_index_price('BCH-USD', d.new(100));" | cartridge enter bch-market
    echo "local d=require('decimal'); market.update_index_price('XRP-USD', d.new(100));" | cartridge enter xrp-market
    echo "local d=require('decimal'); market.update_index_price('WLD-USD', d.new(100));" | cartridge enter wld-market
    echo "local d=require('decimal'); market.update_index_price('TON-USD', d.new(100));" | cartridge enter ton-market
    echo "local d=require('decimal'); market.update_index_price('STX-USD', d.new(100));" | cartridge enter stx-market
    echo "local d=require('decimal'); market.update_index_price('MATIC-USD', d.new(100));" | cartridge enter matic-market
    echo "local d=require('decimal'); market.update_index_price('TRB-USD', d.new(100));" | cartridge enter trb-market
    echo "local d=require('decimal'); market.update_index_price('APT-USD', d.new(100));" | cartridge enter apt-market
    echo "local d=require('decimal'); market.update_index_price('INJ-USD', d.new(100));" | cartridge enter inj-market
    echo "local d=require('decimal'); market.update_index_price('AAVE-USD', d.new(100));" | cartridge enter aave-market
    echo "local d=require('decimal'); market.update_index_price('LINK-USD', d.new(100));" | cartridge enter link-market
    echo "local d=require('decimal'); market.update_index_price('BNB-USD', d.new(100));" | cartridge enter bnb-market
    echo "local d=require('decimal'); market.update_index_price('RNDR-USD', d.new(100));" | cartridge enter rndr-market
    echo "local d=require('decimal'); market.update_index_price('MKR-USD', d.new(100));" | cartridge enter mkr-market
    echo "local d=require('decimal'); market.update_index_price('STG-USD', d.new(100));" | cartridge enter stg-market
    echo "local d=require('decimal'); market.update_index_price('ORDI-USD', d.new(100));" | cartridge enter ordi-market
    echo "local d=require('decimal'); market.update_index_price('RLB-USD', d.new(100));" | cartridge enter rlb-market
    echo "local d=require('decimal'); market.update_index_price('SATS1000000-USD', d.new(100));" | cartridge enter sats1000000-market
    echo "local d=require('decimal'); market.update_index_price('TIA-USD', d.new(100));" | cartridge enter tia-market
    echo "local d=require('decimal'); market.update_index_price('BLUR-USD', d.new(100));" | cartridge enter blur-market
    echo "local d=require('decimal'); market.update_index_price('JTO-USD', d.new(100));" | cartridge enter jto-market
    echo "local d=require('decimal'); market.update_index_price('MEME-USD', d.new(100));" | cartridge enter meme-market
    echo "local d=require('decimal'); market.update_index_price('SEI-USD', d.new(100));" | cartridge enter sei-market
    echo "local d=require('decimal'); market.update_index_price('YES-USD', d.new(0.5));" | cartridge enter yes-market
    echo "local d=require('decimal'); market.update_index_price('WIF-USD', d.new(1));" | cartridge enter wif-market
    echo "local d=require('decimal'); market.update_index_price('STRK-USD', d.new(1));" | cartridge enter strk-market
    echo "local d=require('decimal'); market.update_index_price('SHIB1000-USD', d.new(1));" | cartridge enter shib1000-market
    echo "local d=require('decimal'); market.update_index_price('BOME-USD', d.new(1));" | cartridge enter bome-market
    echo "local d=require('decimal'); market.update_index_price('SLERF-USD', d.new(1));" | cartridge enter slerf-market
    echo "local d=require('decimal'); market.update_index_price('W-USD', d.new(1));" | cartridge enter w-market
    echo "local d=require('decimal'); market.update_index_price('ENA-USD', d.new(1));" | cartridge enter ena-market
    echo "local d=require('decimal'); market.update_index_price('PAC-USD', d.new(1));" | cartridge enter pac-market
    echo "local d=require('decimal'); market.update_index_price('MAGA-USD', d.new(1));" | cartridge enter maga-market
    echo "local d=require('decimal'); market.update_index_price('TRUMP-USD', d.new(1));" | cartridge enter trump-market
    echo "local d=require('decimal'); market.update_index_price('MOG1000-USD', d.new(1));" | cartridge enter mog1000-market
    echo "local d=require('decimal'); market.update_index_price('NOT-USD', d.new(1));" | cartridge enter not-market
    echo "local d=require('decimal'); market.update_index_price('MOTHER-USD', d.new(1));" | cartridge enter mother-market
    echo "local d=require('decimal'); market.update_index_price('BONK1000-USD', d.new(1));" | cartridge enter bonk1000-market
    echo "local d=require('decimal'); market.update_index_price('TAIKO-USD', d.new(1));" | cartridge enter taiko-market
    echo "local d=require('decimal'); market.update_index_price('FLOKI1000-USD', d.new(1));" | cartridge enter floki1000-market
)&

cartridge start --cfg ./instances.yml
