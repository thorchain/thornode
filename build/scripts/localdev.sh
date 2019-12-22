#!/bin/sh

set -exuf -o pipefail


source $(dirname "$0")/core.sh

if [ -z "${BOND_ADDRESS:-}" ]; then
    BOND_ADDRESS=tbnb1czyqwfxptfnk7aey99cu820ftr28hw2fcvrh74
    echo "empty bond address"
fi

if [ -z "${POOL_PUB_KEY:-}" ]; then
    echo "empty pool pub key"
    POOL_PUB_KEY=bnbp1addwnpepq2kdyjkm6y9aa3kxl8wfaverka6pvkek2ygrmhx6sj3ec6h0fegwsskxr6j
fi

echo "password" | thorcli keys add thorchain

VALIDATOR="$(thord tendermint show-validator)"
NODE_ADDRESS="$(thorcli keys show thorchain -a)"
NODE_PUB_KEY="$(thorcli keys show thorchain -p)"

init_chain $NODE_ADDRESS

VERSION="$(thorcli query thorchain version | jq -r .version)"

add_node_account $NODE_ADDRESS $VALIDATOR $NODE_PUB_KEY $VERSION $BOND_ADDRESS $POOL_PUB_KEY

cat ~/.thord/config/genesis.json
thord validate-genesis
