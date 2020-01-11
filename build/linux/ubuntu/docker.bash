#!/bin/bash

BRANCH=${VARIABLE:=master}

apt install make

git clone https://gitlab.com/thorchain/thornode.git
cd thornode
git checkout $BRANCH

PEER=$PEER BINANCE_HOST=$BINANCE_HOST make -C build/docker reset-mocknet-validator
