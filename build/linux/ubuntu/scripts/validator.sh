#!/bin/bash
set -ex

PEER="${PEER:=none}" # the hostname of a seed node
GENESIS_URL="${GENESIS_URL:=none}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"

if [ ! -f ~/.thord/config/genesis.json ]; then
    if [[ "$PEER" == "none" ]]; then
        echo "Missing PEER"
        exit 1
    fi

    if [ ! -f /tmp/shared/genesis.json ]; then
      if [[ "$GENESIS_URL" == "none" ]]; then
        echo "Missing GENESIS_URL"
        exit 1
      fi
    fi

    if [ -f ~/.signer/private_key.txt ]; then
        ADDRESS=$(cat ~/.signer/address.txt)
    else
        echo "GENERATING BNB ADDRESSES"
        # because the generate command can get API rate limited, we may need to retry
        n=0
        until [ $n -ge 60 ]; do
            generate > /tmp/bnb && break
            n=$[$n+1]
            sleep 1
        done
        mkdir -p ~/.signer
        ADDRESS=$(cat /tmp/bnb | grep MASTER= | awk -F= '{print $NF}')
        echo $ADDRESS > ~/.signer/address.txt
        BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
        echo $BINANCE_PRIVATE_KEY > ~/.signer/private_key.txt
        PUBKEY=$(cat /tmp/bnb | grep MASTER_PUBKEY= | awk -F= '{print $NF}')
        echo $PUBKEY > ~/.signer/pubkey.txt
    fi

    # create statechain user
    echo $SIGNER_PASSWD | thorcli keys add $SIGNER_NAME

    NODE_ADDRESS=$(thorcli keys show statechain -a)
    echo "YOUR NODE ADDRESS: $NODE_ADDRESS. Send your bond with this as your address."

    # Setup SSD
    thord init local --chain-id statechain
    thorcli config chain-id statechain
    thorcli config output json
    thorcli config indent true
    thorcli config trust-node true

    if [ -f /tmp/shared/genesis.json ]; then
      cp /tmp/shared/genesis.json ~/.thord/config/genesis.json
    else
      curl $GENESIS_URL --output ~/.thord/config/genesis.json
    fi

    thord validate-genesis

    NODE_ID=$(curl $PEER/status | jq -r .result.node_info.id)
    PEER="$NODE_ID@$PEER"
    ADDR='addr_book_strict = true'
    ADDR_STRICT_FALSE='addr_book_strict = false'
    PEERSISTENT_PEER_TARGET='persistent_peers = ""'

    sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.thord/config/config.toml
    sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.thord/config/config.toml

fi

thord start --rpc.laddr tcp://0.0.0.0:26657 &> ~/daemon.log &
