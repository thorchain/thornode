#!/bin/sh
set -ex

NODE_ID="${NODE_ID:=none}"
SEED="${SEED:=none}" # the hostname of a seed node
GENESIS_URL="${GENESIS_URL:=none}"

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

exec "$@"
