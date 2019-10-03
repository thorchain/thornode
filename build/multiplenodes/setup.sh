#!/bin/sh
set -exuf -o pipefail

if [ -z "$SIGNER_PASSWD" ]; then
    echo "SIGNER_PASSWD is empty"
    return
  fi
  if [ -z "$POOL_ADDRESS" ]; then
    echo "POOL_ADDRESS is empty"
    return
  fi
  if [ -z "$TRUSTED_BNB_ADDRESS" ]; then
    echo "$TRUSTED_BNB_ADDRESS is empty"
    return
  fi
# create the first node
docker run --rm -it -v $(pwd)/build:/statechain \
  -e POOL_ADDRESS="$POOL_ADDRESS" \
  -e SIGNER_PASSWD="$SIGNER_PASSWD" \
  -e TRUSTED_BNB_ADDRESS="$TRUSTED_BNB_ADDRESS" \
  -e NODE_ID=zero \
  -e SS_HOME="/statechain/zero/ssd" \
  -e SSC_HOME="/statechain/zero/sscli" \
  thorchain/statechainnode /usr/bin/init.sh

NODE_ID=$(docker run --rm -it -v $(pwd)/build:/statechain \
  -e POOL_ADDRESS="$POOL_ADDRESS" \
  -e SIGNER_PASSWD="$SIGNER_PASSWD" \
  -e TRUSTED_BNB_ADDRESS="$TRUSTED_BNB_ADDRESS" \
  -e NODE_ID=zero \
  -e SS_HOME="/statechain/zero/ssd" \
  -e SSC_HOME="/statechain/zero/sscli" \
  thorchain/statechainnode /usr/bin/ssd tendermint show-node-id | tr -d '\r')

# 172.17.0.2 is the address I got
PEER="$NODE_ID@192.168.10.2:26656"
echo $PEER

docker run --rm -it -v $(pwd)/build:/statechain \
  -e POOL_ADDRESS="$POOL_ADDRESS" \
  -e SIGNER_PASSWD="$SIGNER_PASSWD" \
  -e TRUSTED_BNB_ADDRESS="$TRUSTED_BNB_ADDRESS" \
  -e PEER="$PEER" \
  -e NODE_ID="first" \
  -e SS_HOME="/statechain/first/ssd" \
  -e SSC_HOME="/statechain/first/sscli" \
   thorchain/statechainnode /usr/bin/init.sh
ADDR='addr_book_strict = true'
ADDR_STRICT_FALSE='addr_book_strict = false'
PEERSISTENT_PEER_TARGET='persistent_peers = ""'

sed -i -e "s/$ADDR/$ADDR_STRICT_FALSE/g" $(pwd)/build/first/ssd/config/config.toml
sed -i -e "s/$PEERSISTENT_PEER_TARGET/persistent_peers = \"$PEER\"/g" "$(pwd)"/build/first/ssd/config/config.toml
FIRST_ACCOUNT=$(jq '.app_state.accounts[0]' <$(pwd)/build/first/ssd/config/genesis.json)
{
  jq --argjson FIRST_ACCOUNT "$FIRST_ACCOUNT" '.app_state.accounts +=[$FIRST_ACCOUNT]'
} < $(pwd)/build/zero/ssd/config/genesis.json > /tmp/genesis.json
mv /tmp/genesis.json $(pwd)/build/zero/ssd/config/genesis.json
FIRST_NODE_ACCOUNT=$(jq '.app_state.swapservice.node_accounts[0]' <$(pwd)/build/first/ssd/config/genesis.json)
{
  jq --argjson FIRST_NODE_ACCOUNT "$FIRST_NODE_ACCOUNT" '.app_state.swapservice.node_accounts +=[$FIRST_NODE_ACCOUNT]'
} < $(pwd)/build/zero/ssd/config/genesis.json > /tmp/genesis.json
mv /tmp/genesis.json $(pwd)/build/zero/ssd/config/genesis.json

cp $(pwd)/build/zero/ssd/config/genesis.json $(pwd)/build/first/ssd/config/genesis.json



