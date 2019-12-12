#!/bin/sh
set -ex

. $(dirname "$0")/core.sh

CHAIN_ID="${CHAIN_ID:=thorchain}"
BINANCE_HOST="${BINANCE_HOST:=https://data-seed-pre-0-s3.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_API="${CHAIN_API:=127.0.0.1:1317}"
CHAIN_RPC="${CHAIN_RPC:=127.0.0.1:26657}"
SIGNER_NAME="${SIGNER_NAME:=thorchain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=1}"
USE_TSS="${USE_TSS:=false}"
TSS_SCHEME="${TSS_SCHEME:=http}"
TSS_HOST="${TSS_HOST:=127.0.0.1}"
TSS_PORT="${TSS_PORT:=4040}"

$(dirname "$0")/wait-for-thorchain-api.sh $CHAIN_API

gen_bnb_address
BINANCE_PRIVATE_KEY=$(cat ~/.signer/private_key.txt)

SIGNER_PATH=$DB_PATH/signer/
mkdir -p $SIGNER_PATH
mkdir -p /etc/observe/signd

# Generate Signer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"binance_host\": \"$BINANCE_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$CHAIN_RPC\",
    \"start_block_height\": $START_BLOCK_HEIGHT,
    \"enforce_block_height\": false,
    \"block_scan_processors\": 1,
    \"block_height_discover_back_off\": \"1s\",
    \"block_retry_interval\": \"10s\",
    \"scheme\": \"http\"
  },
  \"state_chain\": {
    \"chain_id\": \"$CHAIN_ID\",
    \"chain_host\": \"$CHAIN_API\",
    \"signer_name\": \"$SIGNER_NAME\",
    \"signer_passwd\": \"$SIGNER_PASSWD\"
  },
  \"metric\": {
    \"enabled\": true
  },
  \"signer_db_path\": \"$SIGNER_PATH\",
  \"binance\": {
    \"private_key\": \"$BINANCE_PRIVATE_KEY\",
    \"rpc_host\": \"$BINANCE_HOST\"
  },
  \"use_tss\": $USE_TSS,
  \"tss\": {
    \"scheme\": \"$TSS_SCHEME\",
    \"host\": \"$TSS_HOST\",
    \"port\": $TSS_PORT
  }
}" > /etc/observe/signd/config.json

exec "$@"
