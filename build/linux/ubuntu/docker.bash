#!/bin/bash

apt install make

git clone https://gitlab.com/thorchain/thornode.git
cd thornode
git checkout validator-run

PEER=67.205.172.201 BINANCE_HOST=http://67.205.172.201:26660 TAG=sandbox TAGS=sandbox make -C build/docker reset-mocknet-validator
