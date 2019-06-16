#!/bin/bash

set -e

tx () {
    echo ">$ sscli tx swapservice $@"
    echo "password" | sscli tx swapservice $@ --from jack --yes --broadcast-mode block
}

query () {
    echo ">$ sscli query swapservice $@"
    sscli query swapservice $@
}

# Setup wallets
tx set-account alice ATOM 100000
tx set-account alice BTC 38
tx set-account jack ATOM 100000
tx set-account jack BTC 104
tx set-account jack COIN2 1008

query accstruct alice
query accstruct jack

# Stake Coins
tx set-stake alice BTC 58 12
tx set-stake alice COIN2 578 308
tx set-stake jack BTC 55 88
tx set-stake jack COIN2 600 97

query stakestruct BTC
query stakestruct COIN2

# Swap Coins
query accstruct alice
tx set-swap ATOM BTC 10 alice alice
tx set-swap BTC COIN2 10 alice alice
query accstruct alice
