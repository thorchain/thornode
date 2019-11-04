#!/bin/sh

set -exuf -o pipefail

while true; do

  make install
  thord init local --chain-id statechain

  echo "password" | thorcli keys add jack
  echo "password" | thorcli keys add alice

  thord add-genesis-account $(thorcli keys show jack -a) 1000thor
  thord add-genesis-account $(thorcli keys show alice -a) 1000thor

  thorcli config chain-id statechain
  thorcli config output json
  thorcli config indent true
  thorcli config trust-node true

  if [ -z "${POOL_ADDRESS:-}" ]; then
    echo "empty pool address"
    POOL_ADDRESS=bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6
  fi
  # add jack as a trusted account
  {
    jq --arg VERSION "$(thorcli query swapservice version | jq -r .version)" --arg POOL_ADDRESS "$POOL_ADDRESS" --arg VALIDATOR "$(thord tendermint show-validator)" --arg NODE_ADDRESS "$(thorcli keys show jack -a)" --arg OBSERVER_ADDRESS "$(thorcli keys show jack -a)" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active","bond_address":$POOL_ADDRESS,"accounts":{"bnb_signer_acc": $POOL_ADDRESS, "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}} | .app_state.swapservice.pool_addresses.rotate_at="28800" | .app_state.swapservice.pool_addresses.rotate_window_open_at="27800" | .app_state.swapservice.pool_addresses.current[0] = {"chain":"BNB","seq_no":"4","pub_key":$POOL_ADDRESS}'
  } <~/.thord/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json ~/.thord/config/genesis.json
  cat ~/.thord/config/genesis.json
  thord validate-genesis
  break

done
