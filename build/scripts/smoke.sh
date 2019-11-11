#!/bin/sh

#
# Smoke Tests.
#
# Run our smoke tests against a Thorchain instance.
#

cd "$(dirname "$0")"/../docker
NET=$1
FAUCET_KEY=$2

#
# Clean/prep the environment.
#
setup() {
  rm -rf ~/.signer
  rm -rf ~/.thor*
  rm -rf /tmp/shared

  make clean
  make -C ../../ install tools

  mkdir ~/.signer
  mkdir -p /tmp/shared
}

#
# Configure and run all services.
#
run_services() {
  export NODES=1
  export SEED="$(hostname)"

  ../scripts/genesis.sh
  run_thord

  ../scripts/rest.sh
  run_rest

  sleep 5

  ../scripts/observer.sh
  run_observed

  ../scripts/signer.sh
  run_signd
}

#
# Statechain
#
run_thord() {
  thord start --rpc.laddr tcp://0.0.0.0:26657 &>/dev/null &
}

#
# Observer
#
run_observed() {
  observed -c /etc/observe/observed/config.json &>/dev/null &
}

#
# Signer
#
run_signd() {
  signd -c /etc/observe/signd/config.json &>/dev/null &
}

#
# Statechain REST API
#
run_rest() {
  thorcli rest-server --chain-id thorchain --laddr tcp://0.0.0.0:1317 --node tcp://localhost:26657 &>/dev/null &
}

#
# Smoke Tests
#
run_tests() {
  make NET="$1" FAUCET_KEY="$2" PRIV_KEY="$3" validate-smoke-test
}

setup
run_services

PRIV_KEY=$(cat ~/.signer/private_key.txt)
run_tests "$NET" "$FAUCET_KEY" "$PRIV_KEY"
