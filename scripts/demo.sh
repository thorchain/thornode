#!/bin/bash

set -e

RED='\033[0;31m'
NC='\033[0m' # No Color
LTGREEN='\033]1;32m'

tx () {
    tput setaf 2; echo ">$ sscli tx swapservice $@"; tput sgr0
    echo "password" | sscli tx swapservice $@ --from jack --yes --broadcast-mode block
}

query () {
    tput setaf 2; echo ">$ sscli query swapservice $@"; tput sgr0
    sscli query swapservice $@
}

# Setup wallets
tx set-account alice ATOM 100000
tx set-account alice CANYA 38
tx set-account jack ATOM 100000
tx set-account jack CANYA 104
tx set-account jack MARVEL 1008

query accstruct alice
query accstruct jack

# Stake Coins
tx set-stake alice CANYA 58 12
tx set-stake alice MARVEL 578 308
tx set-stake jack CANYA 55 88
tx set-stake jack MARVEL 600 97

query stakestruct CANYA
query stakestruct MARVEL

# Swap Coins
query accstruct alice
tx set-swap ATOM CANYA 10 alice alice
tx set-swap CANYA MARVEL 10 alice alice
query accstruct alice
