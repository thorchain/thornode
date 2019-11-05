#!/bin/sh
set -ex

source $(dirname "$0")/core.sh

PEER="${PEER:=none}" # the hostname of a seed node
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"

if [ ! -f ~/.thord/config/genesis.json ]; then
    if [[ "$PEER" == "none" ]]; then
        echo "Missing PEER"
        exit 1
    fi

    gen_bnb_address

    init_chain $SIGNER_NAME $SIGNER_PASSWD

    fetch_genesis $PEER

    NODE_ID=$(fetch_node_id $PEER)
    peer_list $NODE_ID $PEER

    NODE_ADDRESS=$(thorcli keys show $SIGNER_NAME -a)
    echo "YOUR NODE ADDRESS: $NODE_ADDRESS. Send your bond with this as your address."
fi

exec "$@"
