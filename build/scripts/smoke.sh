#!/bin/sh

#
# Smoke Tests.
#
# Run our smoke tests against a Thorchain instance.
#

#
# Clean/prep the environment.
#
setup() {
  make -C $(dirname "$0")/../docker clean
}

#
# Clear those ENV vars that THORNode don't want to use
#
clear_vars() {
  unset CHAIN_HOST
  unset RPC_HOST
  unset SIGNER_NAME
  unset SIGNER_PASSWD
}

#
# Configure and run all services.
#
run_services() {
  export NODES=1
  export SEED="$(hostname)"

  $(dirname "$0")/genesis.sh
  run_thord

  $(dirname "$0")/rest.sh
  run_rest

  sleep 5

  $(dirname "$0")/observer.sh
  run_observed

  $(dirname "$0")/signer.sh
  run_signd
}

#
# Statechain
#
run_thord() {
  thord start --rpc.laddr tcp://0.0.0.0:26657 &
}

#
# Observer
#
run_observed() {
  observed -c /etc/observe/observed/config.json &
}

#
# Signer
#
run_signd() {
  signd -c /etc/observe/signd/config.json &
}

#
# Statechain REST API
#
run_rest() {
  thorcli rest-server --chain-id thorchain --laddr tcp://0.0.0.0:1317 --node tcp://localhost:26657 &
}

#
# Smoke Tests
#
run_tests() {
  make -C $(dirname "$0")/../docker NET="$1" FAUCET_KEY="$2" PRIV_KEY="$3" validate-smoke-test
}

NET=${NET:-testnet}
FAUCET_KEY=${FAUCET_KEY}

setup
clear_vars
run_services

PRIV_KEY=$(cat ~/.signer/private_key.txt)
run_tests "$NET" "$FAUCET_KEY" "$PRIV_KEY"
