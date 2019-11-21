#!/bin/sh

set -exuf -o pipefail


source $(dirname "$0")/core.sh

init_chain statechain password

if [ -z "${BOND_ADDRESS:-}" ]; then
    BOND_ADDRESS=tbnb1czyqwfxptfnk7aey99cu820ftr28hw2fcvrh74
    echo "empty bond address"
  fi

  if [ -z "${POOL_PUB_KEY:-}" ]; then
    echo "empty pool pub key"
    POOL_PUB_KEY=bnbp1addwnpepq2kdyjkm6y9aa3kxl8wfaverka6pvkek2ygrmhx6sj3ec6h0fegwsskxr6j
  fi
  if [ -z "${POOL_ADDRESS:-}" ]; then
    echo "empty pool address"
    POOL_ADDRESS=tbnb1tdfqy34uptx207scymqsy4k5uzfmry5s8lujqt
    SEQNO=0
  fi

if [ -z "${SEQNO:-}" ]; then
    SEQNO=$(curl https://testnet-dex.binance.org/api/v1/account/${POOL_ADDRESS} | jq '.sequence | tonumber')
    echo $SEQNO
fi

VERSION="$(thorcli query thorchain version | jq -r .version)"
VALIDATOR="$(thord tendermint show-validator)"
NODE_ADDRESS="$(thorcli keys show statechain -a)"
NODE_PUB_KEY="$(thorcli keys show statechain -p)"

add_node_account $NODE_ADDRESS $VALIDATOR $NODE_PUB_KEY $VERSION $BOND_ADDRESS $POOL_PUB_KEY
add_pool_address $POOL_ADDRESS $POOL_PUB_KEY $SEQNO

cat ~/.thord/config/genesis.json
thord validate-genesis
