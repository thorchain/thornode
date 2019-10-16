#!/bin/sh

set -exuf -o pipefail

while true; do

  ssd init local --chain-id statechain
  
  echo "password" | sscli keys add jack
  echo "password" | sscli keys add alice

  ssd add-genesis-account $(sscli keys show jack -a) 1000bep,100000000bepv
  ssd add-genesis-account $(sscli keys show alice -a) 1000bep,100000000bepv

  sscli config chain-id statechain
  sscli config output json
  sscli config indent true
  sscli config trust-node true

  # add jack as a trusted account
  {
      jq --arg VERSION "$(sscli query swapservice version | jq -r .version)" --arg VALIDATOR "$(ssd tendermint show-validator)" --arg NODE_ADDRESS "$(sscli keys show jack -a)" --arg OBSERVER_ADDRESS "$(sscli keys show jack -a)" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS, "version": $VERSION, "status":"active","accounts":{"bnb_signer_acc":"bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlYYY", "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}}'
  } <~/.ssd/config/genesis.json >/tmp/genesis.json
  mv /tmp/genesis.json ~/.ssd/config/genesis.json

  ssd validate-genesis

  break

done
