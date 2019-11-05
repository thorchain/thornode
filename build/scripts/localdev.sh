#!/bin/sh

set -exuf -o pipefail


source $(dirname "$0")/core.sh

init_chain statechain password

if [ -z "${POOL_ADDRESS:-}" ]; then
    echo "empty pool address"
    POOL_ADDRESS=bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
    SEQNO=0
fi
if [ -z "${SEQNO:-}" ]; then
    SEQNO=$(curl https://testnet-dex.binance.org/api/v1/account/${POOL_ADDRESS} | jq '.sequence | tonumber')
    echo $SEQNO
fi

VERSION="$(thorcli query swapservice version | jq -r .version)"
VALIDATOR="$(thord tendermint show-validator)"
NODE_ADDRESS="$(thorcli keys show statechain -a)"
OBSERVER_ADDRESS="$(thorcli keys show statechain -a)"

add_node_account $NODE_ADDRESS $VALIDATOR $OBSERVER_ADDRESS $VERSION $POOL_ADDRESS
add_pool_address $POOL_ADDRESS $SEQNO

cat ~/.thord/config/genesis.json
thord validate-genesis
