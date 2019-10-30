#!/bin/sh
set -ex

NODE_ID="${NODE_ID:=none}"
SEED="${SEED:=none}" # the hostname of a seed node
GENESIS_URL="${GENESIS_URL:=none}"

if [ ! -f ~/.ssd/config/genesis.json ]; then
    if [[ "$NODE_ID" == "none" ]]; then
        echo "Mising NODE_ID"
        exit 1
    fi

    if [[ "$SEED" == "none" ]]; then
        echo "Mising SEED"
        exit 1
    fi

    if [[ "$GENESIS_URL" == "none" ]]; then
        echo "Mising GENESIS_URL"
        exit 1
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
        ADDRESS=$(cat /tmp/bnb | grep MASTER= | awk -F= '{print $NF}')
        echo $ADDRESS > ~/.signer/address.txt
        BINANCE_PRIVATE_KEY=$(cat /tmp/bnb | grep MASTER_KEY= | awk -F= '{print $NF}')
        echo $BINANCE_PRIVATE_KEY > ~/.signer/private_key.txt
    fi

    # create statechain user
    echo $SIGNER_PASSWD | sscli keys add statechain

    NODE_ADDRESS=$(sscli keys show statechain -a)
    echo "YOUR NODE ADDRESS: $NODE_ADDRESS. Send your bond with this as your address."

    # Setup SSD
    ssd init local --chain-id statechain
    sscli config chain-id statechain
    sscli config output json
    sscli config indent true
    sscli config trust-node true

    curl $GENESIS_URL --output ~/.ssd/config/genesis.json

    ssd validate-genesis

    PEER="$NODE_ID@$SEED:26656"
    ADDR='addr_book_strict = true'
    ADDR_STRICT_FALSE='addr_book_strict = false'
    PEERSISTENT_PEER_TARGET='persistent_peers = ""'

    sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" ~/.ssd/config/config.toml
    sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" ~/.ssd/config/config.toml

fi

exec "$@"
