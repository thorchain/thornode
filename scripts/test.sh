#!/bin/bash

set -e

tx () {
    echo "sscli tx swapservice $@"
    echo "password" | sscli tx swapservice $@ --from jack --yes --broadcast-mode block
}

query () {
    echo "sscli query swapservice $@"
    sscli query swapservice $@
}

# Setup wallets
tx set-account alice ATOM 100000
tx set-account alice COIN1 38
tx set-account alice COIN2 899
tx set-account jack ATOM 100000
tx set-account jack COIN1 104
tx set-account jack COIN2 1008

query accstruct alice
query accstruct jack

# Stake Coins
tx set-stake alice COIN1 58 12
tx set-stake alice COIN2 578 308
tx set-stake jack COIN1 55 88
tx set-stake jack COIN2 600 97

query stakestruct COIN1
query stakestruct COIN2

# Swap Coins
tx set-swap ATOM COIN1 10 alice alice
query accstruct alice
