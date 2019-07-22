#!/bin/bash

set -e

RED='\033[0;31m'
NC='\033[0m' # No Color
LTGREEN='\033]1;32m'

tx () {
    tput setaf 2; echo ">$ sscli tx swapservice $@"; tput sgr0
    command=$1; shift;
    echo "password" | sscli tx swapservice $command --from jack --yes --broadcast-mode block -- $@
}

query () {
    tput setaf 2; echo ">$ sscli query swapservice $@"; tput sgr0
    sscli query swapservice $@
}

# Setup wallets
tx set-account alice ATOM 100000
tx set-account alice BTC 38
tx set-account alice ETH 447
tx set-account jack ATOM 100000
tx set-account jack BTC 104
tx set-account jack ETH 1008

query accstruct alice
query accstruct jack

# Stake Coins
tx set-stake alice BTC 58 12
tx set-stake alice ETH 578 308
tx set-stake jack BTC 55 88
tx set-stake jack ETH 600 97

query stakestruct BTC
query stakestruct ETH

# Swap Coins
query accstruct alice
tx set-swap ATOM BTC 10 alice alice
tx set-swap BTC ETH 10 alice alice
query accstruct alice
