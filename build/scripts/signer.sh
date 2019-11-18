#!/bin/sh
set -ex

. $(dirname "$0")/core.sh

CHAIN_ID="${CHAIN_ID:=thorchain}"
DEX_HOST="${DEX_HOST:=testnet-dex.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_HOST="${RPC_HOST:=127.0.0.1:26657}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=1}"
USE_TSS="${USE_TSS:=false}"
TSS_SCHEME="${TSS_SCHEME:=http}"
TSS_HOST="${TSS_HOST:=127.0.0.1}"
TSS_PORT="${TSS_PORT:=8321}"
NODE_ID="${NODE_ID:=none}"

$(dirname "$0")/wait-for-statechain-api.sh $CHAIN_HOST

gen_bnb_address
BINANCE_PRIVATE_KEY=$(cat ~/.signer/private_key.txt)

SIGNER_PATH=$DB_PATH/signer/
mkdir -p $SIGNER_PATH
mkdir -p /etc/observe/signd

# Generate Signer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"dex_host\": \"$DEX_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$RPC_HOST\",
    \"start_block_height\": $START_BLOCK_HEIGHT,
    \"enforce_block_height\": false,
    \"block_scan_processors\": 1,
    \"block_height_discover_back_off\": \"1s\",
    \"block_retry_interval\": \"10s\",
    \"scheme\": \"http\"
  },
  \"state_chain\": {
    \"chain_id\": \"$CHAIN_ID\",
    \"chain_host\": \"$CHAIN_HOST\",
    \"signer_name\": \"$SIGNER_NAME\",
    \"signer_passwd\": \"$SIGNER_PASSWD\"
  },
  \"metric\": {
    \"enabled\": true
  },
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"dex_host\": \"$DEX_HOST\"
  },
  \"use_tss\": $USE_TSS,
  \"key_gen\": {
    \"scheme\": \"$TSS_SCHEME\",
    \"host\": \"$TSS_HOST\",
    \"port\": $TSS_PORT,
    \"node_id\": $NODE_ID,
  },
  \"key_sign\": {
    \"scheme\": \"$TSS_SCHEME\",
    \"host\": \"$TSS_HOST\",
    \"port\": $TSS_PORT,
    \"node_id\": $NODE_ID,
  }
}" > /etc/observe/signd/config.json

exec "$@"
