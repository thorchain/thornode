#!/bin/sh
set -ex

source $(dirname "$0")/core.sh

PEER="${PEER:=none}" # the hostname of a seed node
SIGNER_NAME="${SIGNER_NAME:=thorchain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"

if [ ! -f ~/.thord/config/genesis.json ]; then
    if [[ "$PEER" == "none" ]]; then
        echo "Missing PEER"
        exit 1
    fi

    gen_bnb_address

    thorcli keys show $SIGNER_NAME || echo $SIGNER_PASSWD | thorcli --trace keys add $SIGNER_NAME 2>&1

    # write private key to tss volume
    if [ ! -z ${TSSPRIVKEY+x} ]; then
        echo $SIGNER_PASSWD | thorcli keys tss $SIGNER_NAME 2> $TSSPRIVKEY
    fi

    NODE_ADDRESS=$(thorcli keys show $SIGNER_NAME -a)
    init_chain $NODE_ADDRESS

    fetch_genesis $PEER

    NODE_ID=$(fetch_node_id $PEER)
    peer_list $NODE_ID $PEER

    # thorcli tx thorchain set-trust-account $(thorcli keys show $SIGNER_NAME --pubkey) $(thorcli keys show $SIGNER_NAME --pubkey) $(thord tendermint show-validator) --from $SIGNER_NAME
    echo "YOUR NODE ADDRESS: $NODE_ADDRESS . Send your bond with this as your address."
fi

exec "$@"
