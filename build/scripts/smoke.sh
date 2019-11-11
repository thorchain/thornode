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
  rm -rf ~/.signer
  rm -rf ~/.thor*
  rm -rf /tmp/shared

  mkdir ~/.signer
  mkdir -p /tmp/shared

  make -C ../docker clean
  make -C ../../ install tools
}

#
# Configure and run all services.
#
run_services() {
  export NODES=1
  export SEED="$(hostname)"

  ./genesis.sh
  run_thord

  ./rest.sh
  run_rest

  sleep 5

  ./observer.sh
  run_observed

  ./signer.sh
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
  make -C ../docker NET="$1" FAUCET_KEY="$2" PRIV_KEY="$3" validate-smoke-test
}

cd "$(dirname "$0")"

NET=${NET:-testnet}
FAUCET_KEY=${FAUCET_KEY}

setup
run_services

PRIV_KEY=$(cat ~/.signer/private_key.txt)
run_tests "$NET" "$FAUCET_KEY" "$PRIV_KEY"
