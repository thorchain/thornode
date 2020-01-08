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

    if [[ "$NET" == "mocknet" ]]; then
        # create a binance wallet and bond/register
        gen_bnb_address
        ADDRESS=$(cat ~/.bond/address.txt)

        # send bond transaction to mock binance
        $(dirname "$0")/mock-bond.sh $PEER $ADDRESS $NODE_ADDRESS

        sleep 10 # wait for thorchain to register the new node account

        # set node keys
        echo $SIGNER_PASSWD | thorcli tx thorchain set-node-keys $(thorcli keys show thorchain --pubkey) $(thorcli keys show thorchain --pubkey) $(thord tendermint show-validator) --node tcp://$PEER:26657 --from $SIGNER_NAME --yes
    else
        echo "YOUR NODE ADDRESS: $NODE_ADDRESS . Send your bond with this as your address."
    fi

fi

exec "$@"
