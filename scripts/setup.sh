#!/bin/sh

set -exuf -o pipefail

while true; do

  make install
  ssd init local --chain-id statechain

  echo "password" | sscli keys add jack
  echo "password" | sscli keys add alice

  ssd add-genesis-account $(sscli keys show jack -a) 1000bep,100000000stake
  ssd add-genesis-account $(sscli keys show alice -a) 1000bep,100000000stake

  sscli config chain-id statechain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  # add jack as a trusted account
  {
    jq --arg OBSERVER_ADDRESS "$(sscli keys show jack -a)" '.app_state.swapservice.admin_configs += [{"key":"PoolAddress", "value": "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlYYY", "address":$OBSERVER_ADDRESS}]' |
      jq --arg VALIDATOR "$(ssd tendermint show-validator)" --arg NODE_ADDRESS "$(sscli keys show jack -a)" --arg OBSERVER_ADDRESS "$(sscli keys show jack -a)" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS ,"status":"active","accounts":{"bnb_signer_acc":"bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlYYY", "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}}'
  } <~/.ssd/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json ~/.ssd/config/genesis.json
  {
    jq --arg OBSERVER_ADDRESS "$(sscli keys show jack -a)" '.app_state.swapservice.admin_configs += [{"key":"PoolExpiry", "value": "2020-01-01T00:00:00Z", "address":$OBSERVER_ADDRESS}]'
  } <~/.ssd/config/genesis.json >/tmp/genesis.json

  mv /tmp/genesis.json ~/.ssd/config/genesis.json
  ssd validate-genesis
  cat ~/.ssd/config/genesis.json
  break

done
