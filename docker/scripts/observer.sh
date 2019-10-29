#!/bin/sh
set -ex

CHAIN_ID="${CHAIN_ID:=thorchain}"
DEX_HOST="${DEX_HOST:=testnet-dex.binance.org}"
DB_PATH="${DB_PATH:=/var/data}"
CHAIN_HOST="${CHAIN_HOST:=127.0.0.1:1317}"
RPC_HOST="${RPC_HOST:=data-seed-pre-0-s3.binance.org}"
SIGNER_NAME="${SIGNER_NAME:=statechain}"
SIGNER_PASSWD="${SIGNER_PASSWD:=password}"
START_BLOCK_HEIGHT="${START_BLOCK_HEIGHT:=0}"
NODES="${NODES:=1}"
MASTER="${MASTER:=node1}" # the hostname of the master node
ROTATE_BLOCK_HEIGHT="${ROTATE_BLOCK_HEIGHT:=0}" # how often the pools in statechain should rotate

OBSERVER_PATH=$DB_PATH/observer/
mkdir -p $OBSERVER_PATH
mkdir -p /etc/observe/observed

# Generate observer config file
echo "{
  \"chain_id\": \"$CHAIN_ID\",
  \"dex_host\": \"$DEX_HOST\",
  \"observer_db_path\": \"$OBSERVER_PATH\",
  \"block_scanner\": {
    \"rpc_host\": \"$RPC_HOST\",
    \"start_block_height\": \"$START_BLOCK_HEIGHT\",
    \"enforce_block_height\": false,
    \"block_scan_processors\": 1,
    \"block_height_discover_back_off\": \"1s\",
    \"block_retry_interval\": \"10s\"
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
  }
}" > /etc/observe/observed/config.json

ls /go/bin
exec "$@"
