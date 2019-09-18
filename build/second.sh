#!/bin/sh
set -exuf -o pipefail

# initialize statechain
second() {

  if [ -z "$SIGNER_PASSWD" ]; then
    echo "SIGNER_PASSWD is empty"
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
  ssd validate-genesis
}

second
