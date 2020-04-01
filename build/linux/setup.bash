#!/bin/bash

set -ex

BRANCH=${BRANCH:=build}

apt install make

iptables -A INPUT -p tcp  --match multiport --dports 1317,26656,26657,8080,6040,5040,4040 -j ACCEPT
/sbin/iptables-save

[ ! -d "~/thornode" ] && git clone https://gitlab.com/thorchain/thornode.git
cd ~/thornode
git fetch origin
git checkout $BRANCH
git reset --hard origin/$BRANCH

PEER=$PEER BINANCE_HOST=$BINANCE_HOST make -C build/docker reset-testnet-validator
