#!/bin/sh
set -exuf -o pipefail

# initialize statechain
initialize() {

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
  if [ -z "$SS_HOME" ]; then
    echo "SS_HOME is empty"
  else
    echo "SS_HOME is $SS_HOME"
  fi
  if [ -z "$SSC_HOME" ]; then
    echo "SSC_HOME is empty"
  else
    echo "SSC_HOME is $SSC_HOME"
  fi

  echo "$SIGNER_PASSWD" | sscli keys add jack
  ssd init "$NODE_ID" --chain-id statechain
  ssd add-genesis-account "$(sscli keys show jack -a)" 1000rune,100000000stake
  sscli config chain-id statechain
  sscli config output json
  sscli config indent true
  sscli config trust-node true
  #echo "$SIGNER_PASSWD" | ssd gentx --name jack --home-client "$SSC_HOME"
  #ssd collect-gentxs
  {
      jq --arg TRUSTED_BNB_ADDRESS "$TRUSTED_BNB_ADDRESS" '.app_state.swapservice.admin_configs[1] = {"key":"PoolExpiry", "value": "2020-01-01T00:00:00Z" , "address": $TRUSTED_BNB_ADDRESS}' |
      jq --arg POOL_ADDRESS "$POOL_ADDRESS" --arg TRUSTED_BNB_ADDRESS "$TRUSTED_BNB_ADDRESS" '.app_state.swapservice.admin_configs[0] = {"key":"PoolAddress", "value": $POOL_ADDRESS , "address": $TRUSTED_BNB_ADDRESS}' |
      jq --arg TRUSTED_BNB_ADDRESS "$TRUSTED_BNB_ADDRESS" --arg RUNE_ADDRESS "$(sscli keys show jack -a)" --arg VALIDATOR "$(./ssd tendermint show-validator)" '.app_state.swapservice.trust_accounts[0] = {"signer_address": $TRUSTED_BNB_ADDRESS, "admin_address": $TRUSTED_BNB_ADDRESS, "observer_address": $RUNE_ADDRESS, "validator_pub_key": $VALIDATOR}'

  } < "$SS_HOME/config/genesis.json" >/tmp/genesis.json

  mv /tmp/genesis.json "$SS_HOME/config/genesis.json"
  cat "$SS_HOME/config/genesis.json"
  ssd validate-genesis
}

initialize
